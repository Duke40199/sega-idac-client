package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"sega_idac_client/internal/adapter/grpc/health"
	healthv1 "sega_idac_client/internal/adapter/grpc/gen/healthv1"
	"sega_idac_client/internal/config"
)

// Server wraps the gRPC server together with the listener address so it can be
// started and gracefully stopped by the caller.
type Server struct {
	grpcServer *grpc.Server
	addr       string
}

// NewServer builds a gRPC server and registers all service handlers on it.
func NewServer(cfg config.Config) *Server {
	grpcServer := grpc.NewServer()

	// Register endpoints.
	healthv1.RegisterHealthServiceServer(grpcServer, health.NewHandler())

	// Enable server reflection so tools like grpcurl can introspect the service.
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		addr:       cfg.GetGRPCCfg().Addr(),
	}
}

// Start binds the listener and serves gRPC. It blocks until the server stops.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	log.Printf("🚀 gRPC server listening on %s", s.addr)
	return s.grpcServer.Serve(lis)
}

// Stop gracefully shuts down the gRPC server.
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
