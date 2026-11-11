package health_test

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	healthv1 "github.com/Duke40199/sega-idac-client/gen/healthv1"
	"github.com/Duke40199/sega-idac-client/internal/adapter/grpc/health"
)

// TestHealthCheck spins up the Health service over an in-memory gRPC connection
// and asserts that Check reports SERVING.
func TestHealthCheck(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)

	srv := grpc.NewServer()
	healthv1.RegisterHealthServiceServer(srv, health.NewHandler())
	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := healthv1.NewHealthServiceClient(conn)
	resp, err := client.Check(context.Background(), &healthv1.CheckRequest{})
	if err != nil {
		t.Fatalf("Check RPC failed: %v", err)
	}

	if resp.GetStatus() != healthv1.CheckResponse_SERVING_STATUS_SERVING {
		t.Fatalf("expected SERVING, got %v", resp.GetStatus())
	}
}
