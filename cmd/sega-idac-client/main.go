package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Duke40199/sega-idac-client/internal/adapter/client/all_net"
	"github.com/Duke40199/sega-idac-client/internal/adapter/client/sega_idac"
	grpcserver "github.com/Duke40199/sega-idac-client/internal/adapter/grpc"
	httpgateway "github.com/Duke40199/sega-idac-client/internal/adapter/http"
	"github.com/Duke40199/sega-idac-client/internal/config"
)

func main() {
	// 1. Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	// 2. Init clients
	segaClient := sega_idac.NewSegaIDACClient(cfg)
	allNetClient := all_net.NewAllNetClient(cfg)

	// 3. Init gRPC server
	server := grpcserver.NewServer(cfg, segaClient, allNetClient)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("gRPC server error:", err)
		}
	}()

	// 4. Init HTTP gateway (REST/JSON -> gRPC). Its upstream connections live for
	// the process lifetime, so cancel on shutdown.
	gwCtx, cancelGW := context.WithCancel(context.Background())
	defer cancelGW()
	gateway, err := httpgateway.NewGateway(gwCtx, cfg)
	if err != nil {
		log.Fatal("HTTP gateway init error:", err)
	}
	go func() {
		if err := gateway.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("HTTP gateway error:", err)
		}
	}()

	// 5. Graceful shutdown
	log.Println("✅ sega-idac-client is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Gracefully shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := gateway.Stop(shutdownCtx); err != nil {
		log.Printf("HTTP gateway shutdown error: %v", err)
	}
	server.Stop()
}
