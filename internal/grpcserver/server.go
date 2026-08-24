// Package grpcserver реализует приём метрик по протоколу gRPC.
package grpcserver

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	models "github.com/alexander-xyz/metrics/internal/model"
	pb "github.com/alexander-xyz/metrics/internal/proto"
	"github.com/alexander-xyz/metrics/internal/repository"
)

// RealIPKey — ключ метаданных, в котором агент передаёт свой IP-адрес.
const RealIPKey = "x-real-ip"

// Storage описывает хранилище, куда сервис складывает принятые метрики.
type Storage interface {
	repository.BatchUpdater
}

// MetricsServer принимает метрики по gRPC и сохраняет их в хранилище.
type MetricsServer struct {
	pb.UnimplementedMetricsServer

	store Storage
}

// NewMetricsServer создаёт gRPC-сервис метрик поверх хранилища.
func NewMetricsServer(store Storage) *MetricsServer {
	return &MetricsServer{store: store}
}

// UpdateMetrics сохраняет пакет метрик, присланный агентом.
func (s *MetricsServer) UpdateMetrics(
	ctx context.Context,
	req *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	metrics := make([]models.Metrics, 0, len(req.GetMetrics()))

	for _, metric := range req.GetMetrics() {
		if metric.GetId() == "" {
			return nil, status.Error(codes.InvalidArgument, "empty metric name")
		}

		metrics = append(metrics, convert(metric))
	}

	if len(metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty batch")
	}

	if err := s.store.UpdateBatch(ctx, metrics); err != nil {
		return nil, status.Error(codes.Internal, "cannot store metrics")
	}

	return &pb.UpdateMetricsResponse{}, nil
}

// convert превращает метрику из протокола в модель проекта.
func convert(metric *pb.Metric) models.Metrics {
	if metric.GetType() == pb.Metric_COUNTER {
		delta := metric.GetDelta()

		return models.Metrics{ID: metric.GetId(), MType: models.Counter, Delta: &delta}
	}

	value := metric.GetValue()

	return models.Metrics{ID: metric.GetId(), MType: models.Gauge, Value: &value}
}

// TrustedSubnetInterceptor отклоняет вызовы, если IP-адрес клиента
// из метаданных не входит в доверенную подсеть. При пустой подсети
// проверка не выполняется.
func TrustedSubnetInterceptor(subnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if subnet == nil {
			return handler(ctx, req)
		}

		if !subnet.Contains(clientIP(ctx)) {
			return nil, status.Error(codes.PermissionDenied, "untrusted client address")
		}

		return handler(ctx, req)
	}
}

// clientIP достаёт адрес клиента из метаданных запроса.
func clientIP(ctx context.Context) net.IP {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	values := md.Get(RealIPKey)
	if len(values) == 0 {
		return nil
	}

	return net.ParseIP(values[0])
}

// Serve запускает gRPC-сервер на указанном адресе и блокируется,
// пока сервер не остановят.
func Serve(server *grpc.Server, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen %s: %w", address, err)
	}

	if err := server.Serve(listener); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}

	return nil
}

// New собирает gRPC-сервер с зарегистрированным сервисом метрик
// и проверкой доверенной подсети.
func New(store Storage, subnet *net.IPNet) *grpc.Server {
	server := grpc.NewServer(grpc.UnaryInterceptor(TrustedSubnetInterceptor(subnet)))
	pb.RegisterMetricsServer(server, NewMetricsServer(store))

	return server
}
