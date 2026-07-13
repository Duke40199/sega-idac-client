package health

import (
	"context"

	healthv1 "sega_idac_client/internal/adapter/grpc/gen/healthv1"
)

// Handler implements the HealthService gRPC endpoint.
type Handler struct {
	healthv1.UnimplementedHealthServiceServer
}

// NewHandler creates a new Health gRPC handler.
func NewHandler() *Handler {
	return &Handler{}
}

// Check reports the serving status of the service. For now the server is
// considered healthy whenever it is able to serve the request.
func (h *Handler) Check(_ context.Context, _ *healthv1.CheckRequest) (*healthv1.CheckResponse, error) {
	return &healthv1.CheckResponse{
		Status: healthv1.CheckResponse_SERVING_STATUS_SERVING,
	}, nil
}
