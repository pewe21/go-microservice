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
const PORT = ":3004"

// grpc port
const GRPC_USER_SERVICE_PORT = ":4002"
const GRPC_USER_SERVICE_NUM_INSTANCE = 2

const GRPC_IMAGE_SERVICE_PORT = ":4001"
const GRPC_IMAGE_SERVICE_NUM_INSTANCE = 2

const USER_SCHEME = "user"
const USER_SERVICE_NAME = "user-service"

const IMAGE_SCHEME = "example"
const IMAGE_SERVICE_NAME = "image-service"

// rabbitmq port
const RABBITMQ_PORT = ":5672"

func main() {

	var wg sync.WaitGroup

	cfg := InitConfig()

	postgresStorage := NewPostgresStorage()
	postgresStorage.Init()

	// set db conn limit
	postgresStorage.db.SetMaxOpenConns(25)
	postgresStorage.db.SetMaxIdleConns(25)
	postgresStorage.db.SetConnMaxLifetime(5 * time.Minute)

	// http server
	s := NewServer(PORT, postgresStorage, cfg)

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Run()
	}()

	// rabbitmq consumer
	rabbitMq := NewRabbitMQ(cfg, postgresStorage)
	go rabbitMq.Run()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	<-sigs
	log.Println("SIGTERM detected, will attempt to graceful shutdown...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// shutdown http.server
	if err := s.Server.Shutdown(shutdownCtx); err != nil {
		log.Println("Error when trying to shutdown http server:", err)
	} else {
		log.Println("http server closed")
	}

	// shutdown rabbitmq
	rabbitMq.Close()

	//biar main func tidak exit duluan
	wg.Wait()
}
