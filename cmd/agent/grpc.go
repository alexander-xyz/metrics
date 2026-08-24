package main

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/alexander-xyz/metrics/internal/grpcserver"
	models "github.com/alexander-xyz/metrics/internal/model"
	pb "github.com/alexander-xyz/metrics/internal/proto"
)

// grpcSender отправляет метрики на сервер по протоколу gRPC.
type grpcSender struct {
	conn   *grpc.ClientConn
	client pb.MetricsClient
}

// newGRPCSender подключается к gRPC-серверу по указанному адресу.
func newGRPCSender(address string) (*grpcSender, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect to grpc server %s: %w", address, err)
	}

	return &grpcSender{conn: conn, client: pb.NewMetricsClient(conn)}, nil
}

// Close закрывает соединение с сервером.
func (s *grpcSender) Close() error {
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("close grpc connection: %w", err)
	}

	return nil
}

// Send отправляет пакет метрик, добавляя свой IP-адрес в метаданные.
func (s *grpcSender) Send(ctx context.Context, metrics []models.Metrics) error {
	request := &pb.UpdateMetricsRequest{Metrics: toProto(metrics)}

	if ip := localIP(); ip != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, grpcserver.RealIPKey, ip)
	}

	if _, err := s.client.UpdateMetrics(ctx, request); err != nil {
		return fmt.Errorf("send metrics over grpc: %w", err)
	}

	return nil
}

// toProto переводит метрики проекта в сообщения протокола.
func toProto(metrics []models.Metrics) []*pb.Metric {
	converted := make([]*pb.Metric, 0, len(metrics))

	for _, metric := range metrics {
		item := &pb.Metric{Id: metric.ID}

		if metric.MType == models.Counter {
			item.Type = pb.Metric_COUNTER

			if metric.Delta != nil {
				item.Delta = *metric.Delta
			}
		} else {
			item.Type = pb.Metric_GAUGE

			if metric.Value != nil {
				item.Value = *metric.Value
			}
		}

		converted = append(converted, item)
	}

	return converted
}
