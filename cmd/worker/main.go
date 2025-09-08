package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"golang-exercise/config"
	"golang-exercise/internal/database"
	"golang-exercise/internal/messaging"
	"golang-exercise/internal/repository"
	"golang-exercise/internal/service"
)

func main() {
	// Load the environment config
	config.Load("config.yaml")

	// Connect to databases
	database.ConnectDB()

	// Connect to RabbitMQ
	rabbitmq := messaging.NewRabbitMQ()
	if err := rabbitmq.Connect(); err != nil {
		panic(fmt.Sprintf("Failed to connect to RabbitMQ: %v", err))
	}
	defer rabbitmq.Close()

	// Declare the transaction queue
	if err := rabbitmq.DeclareQueue(config.GetConfig().RabbitMQ.Queue); err != nil {
		panic(fmt.Sprintf("Failed to declare queue: %v", err))
	}

	// Initialize repositories
	accountRepo := repository.NewAccountRepository()
	txLogRepo := repository.NewTransactionLogRepository()

	// Initialize services
	accountService := service.NewAccountService(accountRepo)
	txLogService := service.NewTransactionLogService(txLogRepo)
	transactionService := service.NewTransactionService(accountService, txLogService)

	// Initialize and start the transaction consumer
	consumer := messaging.NewTransactionConsumer(rabbitmq, accountService, transactionService)

	log.Println("Starting transaction worker")
	if err := consumer.StartConsuming(); err != nil {
		panic(fmt.Sprintf("Failed to start consuming: %v", err))
	}

	log.Println("Transaction worker started successfully!")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down transaction worker")
}
