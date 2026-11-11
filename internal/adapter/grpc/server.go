package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	allnetv1 "github.com/Duke40199/sega-idac-client/gen/allnetv1"
	healthv1 "github.com/Duke40199/sega-idac-client/gen/healthv1"
	segaidacv1 "github.com/Duke40199/sega-idac-client/gen/segaidacv1"
	allnet "github.com/Duke40199/sega-idac-client/internal/adapter/grpc/all_net"
	"github.com/Duke40199/sega-idac-client/internal/adapter/grpc/health"
	segaidac "github.com/Duke40199/sega-idac-client/internal/adapter/grpc/sega_idac"
	"github.com/Duke40199/sega-idac-client/internal/config"
	"github.com/Duke40199/sega-idac-client/internal/domain/port"
)

// Server wraps the gRPC server together with the listener address so it can be
// started and gracefully stopped by the caller.
type Server struct {
	grpcServer *grpc.Server
	addr       string
}

// NewServer builds a gRPC server and registers all service handlers on it.
func NewServer(cfg config.Config, segaClient port.SegaIDACClient, allNetClient port.AllNetClient) *Server {
	grpcServer := grpc.NewServer()

	// Register endpoints.
	healthv1.RegisterHealthServiceServer(grpcServer, health.NewHandler())
	segaidacv1.RegisterSegaIDACServiceServer(grpcServer, segaidac.NewHandler(segaClient))
	allnetv1.RegisterAllNetServiceServer(grpcServer, allnet.NewHandler(allNetClient))

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
