package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/alexander-xyz/metrics/internal/grpcserver"
	models "github.com/alexander-xyz/metrics/internal/model"
	pb "github.com/alexander-xyz/metrics/internal/proto"
	"github.com/alexander-xyz/metrics/internal/repository"
)

func TestToProto(t *testing.T) {
	value := 12.5
	delta := int64(3)

	converted := toProto([]models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
		{ID: "PollCount", MType: models.Counter, Delta: &delta},
		{ID: "NoValue", MType: models.Gauge},
	})

	require.Len(t, converted, 3)

	assert.Equal(t, "Alloc", converted[0].GetId())
	assert.Equal(t, pb.Metric_GAUGE, converted[0].GetType())
	assert.InDelta(t, 12.5, converted[0].GetValue(), 0.001)

	assert.Equal(t, pb.Metric_COUNTER, converted[1].GetType())
	assert.Equal(t, int64(3), converted[1].GetDelta())

	assert.Zero(t, converted[2].GetValue(), "метрика без значения переводится с нулём")
}

func TestGRPCSenderSendsMetricsAndIP(t *testing.T) {
	store := repository.NewMemStorage()

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	var gotIP string

	server := grpc.NewServer(grpc.UnaryInterceptor(
		func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
			if md, ok := metadata.FromIncomingContext(ctx); ok {
				if values := md.Get(grpcserver.RealIPKey); len(values) > 0 {
					gotIP = values[0]
				}
			}

			return handler(ctx, req)
		}))
	pb.RegisterMetricsServer(server, grpcserver.NewMetricsServer(store))

	go func() { _ = server.Serve(listener) }()

	defer server.Stop()

	sender, err := newGRPCSender(listener.Addr().String())
	require.NoError(t, err)

	defer sender.Close()

	value := 1.5
	require.NoError(t, sender.Send(context.Background(), []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: &value},
	}))

	stored, err := store.GetGauge(context.Background(), "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(1.5), stored)
	assert.Equal(t, localIP(), gotIP, "агент передал свой адрес в метаданных")
}

func TestNewGRPCSenderFailsOnUnreachableServer(t *testing.T) {
	sender, err := newGRPCSender("localhost:1")
	require.NoError(t, err, "соединение устанавливается лениво")

	defer sender.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	value := 1.5
	err = sender.Send(ctx, []models.Metrics{{ID: "Alloc", MType: models.Gauge, Value: &value}})

	assert.Error(t, err, "недоступный сервер даёт ошибку при отправке")
}

func TestSendUsesGRPCWhenConfigured(t *testing.T) {
	store := repository.NewMemStorage()

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterMetricsServer(server, grpcserver.NewMetricsServer(store))

	go func() { _ = server.Serve(listener) }()

	defer server.Stop()

	conn, err := grpc.NewClient(listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	defer conn.Close()

	sender := &grpcSender{conn: conn, client: pb.NewMetricsClient(conn)}

	value := 2.5
	batch := []models.Metrics{{ID: "Frees", MType: models.Gauge, Value: &value}}

	require.NoError(t, send(context.Background(), batch, &Config{}, nil, sender))

	stored, err := store.GetGauge(context.Background(), "Frees")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(2.5), stored)
}

func TestSendWorksWithCancelledContext(t *testing.T) {
	store := repository.NewMemStorage()

	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	server := grpc.NewServer()
	pb.RegisterMetricsServer(server, grpcserver.NewMetricsServer(store))

	go func() { _ = server.Serve(listener) }()

	defer server.Stop()

	sender, err := newGRPCSender(listener.Addr().String())
	require.NoError(t, err)

	defer sender.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	value := 3.5
	batch := []models.Metrics{{ID: "Sys", MType: models.Gauge, Value: &value}}

	require.NoError(t, send(ctx, batch, &Config{}, nil, sender),
		"метрики досылаются даже после сигнала остановки")

	stored, err := store.GetGauge(context.Background(), "Sys")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(3.5), stored)
}
