package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// http port
const PORT = ":3002"

// grpc port
const GRPCPORT = ":4002"

const GRPC_IMAGE_SERVICE_PORT = ":4001"
const GRPC_IMAGE_SERVICE_NUM_INSTANCE = 2

const IMAGE_SCHEME = "example"
const IMAGE_SERVICE_NAME = "image-service"

// rabbitmq port
const RABBITMQPORT = ":5672"

func main() {

	var wg sync.WaitGroup

	cfg := InitConfig()

	postgresStorage := NewPostgresStorage()
	postgresStorage.Init()

	// set db conn limit
	postgresStorage.db.SetMaxOpenConns(25)
	postgresStorage.db.SetMaxIdleConns(25)
	postgresStorage.db.SetConnMaxLifetime(5 * time.Minute)

	//grpcServer :4002
	grpcServer := NewGrpcServer(GRPCPORT, postgresStorage)

	//http server
	httpServer := NewServer(PORT, postgresStorage, cfg)

	wg.Add(1)
	go func() {
		grpcServer.RunGrpc()
		defer wg.Done()
	}()

	//http server :3002
	wg.Add(1)
	go func() {
		httpServer.Run()
		defer wg.Done()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs
	log.Println("SIGTERM detected, will attempt to graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// shutdown http.server
	if err := httpServer.Server.Shutdown(shutdownCtx); err != nil {
		log.Println("Error when trying to shutdown http server:", err)
	} else {
		log.Println("http server closed")
	}

	// shutdown grpc server
	func() {
		grpcServer.Server.GracefulStop()
		log.Println("GRPC server closed")
	}()

	wg.Wait()
}
