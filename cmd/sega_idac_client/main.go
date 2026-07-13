package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"sega_idac_client/internal/adapter/client/all_net"
	"sega_idac_client/internal/adapter/client/sega_idac"
	grpcserver "sega_idac_client/internal/adapter/grpc"
	"sega_idac_client/internal/config"
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
	_ = segaClient
	_ = allNetClient

	// 3. Init gRPC server
	server := grpcserver.NewServer(cfg)
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("gRPC server error:", err)
		}
	}()

	// 4. Graceful shutdown
	log.Println("✅ sega_idac_client is running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Gracefully shutting down...")
	server.Stop()
}
