package grpcserver

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/alexander-xyz/metrics/internal/proto"
	"github.com/alexander-xyz/metrics/internal/repository"
)

// startServer поднимает gRPC-сервер на буферном соединении и возвращает
// клиента к нему.
func startServer(t *testing.T, store Storage, subnet *net.IPNet) pb.MetricsClient {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := New(store, subnet)

	go func() {
		_ = server.Serve(listener)
	}()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	t.Cleanup(func() {
		conn.Close()
		server.Stop()
		listener.Close()
	})

	return pb.NewMetricsClient(conn)
}

func TestUpdateMetricsStoresBatch(t *testing.T) {
	store := repository.NewMemStorage()
	client := startServer(t, store, nil)

	_, err := client.UpdateMetrics(context.Background(), &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 12.5},
			{Id: "PollCount", Type: pb.Metric_COUNTER, Delta: 3},
		},
	})
	require.NoError(t, err)

	ctx := context.Background()

	gauge, err := store.GetGauge(ctx, "Alloc")
	require.NoError(t, err)
	assert.Equal(t, repository.Gauge(12.5), gauge)

	counter, err := store.GetCounter(ctx, "PollCount")
	require.NoError(t, err)
	assert.Equal(t, repository.Counter(3), counter)
}

func TestUpdateMetricsRejectsEmptyBatch(t *testing.T) {
	client := startServer(t, repository.NewMemStorage(), nil)

	_, err := client.UpdateMetrics(context.Background(), &pb.UpdateMetricsRequest{})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestUpdateMetricsRejectsEmptyName(t *testing.T) {
	client := startServer(t, repository.NewMemStorage(), nil)

	_, err := client.UpdateMetrics(context.Background(), &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{{Id: "", Type: pb.Metric_GAUGE, Value: 1}},
	})

	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestTrustedSubnetAllowsClientFromSubnet(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	client := startServer(t, repository.NewMemStorage(), subnet)

	ctx := metadata.AppendToOutgoingContext(context.Background(), RealIPKey, "10.1.2.3")

	_, err = client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 1}},
	})

	require.NoError(t, err)
}

func TestTrustedSubnetRejectsForeignClient(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	client := startServer(t, repository.NewMemStorage(), subnet)

	testCases := []struct {
		name string
		ip   string
	}{
		{name: "чужая подсеть", ip: "192.168.1.1"},
		{name: "не адрес", ip: "not-an-ip"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.AppendToOutgoingContext(context.Background(), RealIPKey, tc.ip)

			_, callErr := client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
				Metrics: []*pb.Metric{{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 1}},
			})

			assert.Equal(t, codes.PermissionDenied, status.Code(callErr))
		})
	}
}

func TestTrustedSubnetRejectsRequestWithoutMetadata(t *testing.T) {
	_, subnet, err := net.ParseCIDR("10.0.0.0/8")
	require.NoError(t, err)

	client := startServer(t, repository.NewMemStorage(), subnet)

	_, err = client.UpdateMetrics(context.Background(), &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{{Id: "Alloc", Type: pb.Metric_GAUGE, Value: 1}},
	})

	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestServeRejectsBusyAddress(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	defer listener.Close()

	server := New(repository.NewMemStorage(), nil)
	defer server.Stop()

	assert.Error(t, Serve(server, listener.Addr().String()))
}
