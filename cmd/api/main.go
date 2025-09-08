package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"golang-exercise/config"
	"golang-exercise/internal/database"
	"golang-exercise/internal/handler"
	"golang-exercise/internal/messaging"
	"golang-exercise/internal/middleware"
	"golang-exercise/internal/repository"
	"golang-exercise/internal/router"
	"golang-exercise/internal/service"
)

func main() {
	r := gin.New()

	// Load the environment config
	config.Load("config.yaml")

	// Add logger middleware
	r.Use(middleware.Logger())

	// Connect to the database
	database.ConnectDB()

	// Connect to broker (RabbitMQ) for publishing
	rabbitmq := messaging.NewRabbitMQ()
	if err := rabbitmq.Connect(); err != nil {
		fmt.Printf("Warning: Failed to connect to RabbitMQ: %v\n", err)
		fmt.Println("API will run without queue functionality")
	} else {
		defer rabbitmq.Close()

		// Declare the transaction queue
		fmt.Println(config.GetConfig().RabbitMQ.Queue)
		if err := rabbitmq.DeclareQueue(config.GetConfig().RabbitMQ.Queue); err != nil {
			fmt.Printf("Warning: Failed to declare queue: %v\n", err)
		}
	}

	// Health route
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Banking Ledger API is healthy!",
			"service": "api-gateway",
		})
	})

	// Initialize repositories and services
	accountRepo := repository.NewAccountRepository()
	txLogRepo := repository.NewTransactionLogRepository()

	accountService := service.NewAccountService(accountRepo)
	txLogService := service.NewTransactionLogService(txLogRepo)
	transactionService := service.NewTransactionService(accountService, txLogService)

	// Initialize publishers and handlers
	transactionPublisher := messaging.NewTransactionPublisher(rabbitmq)
	accountHandler := handler.NewAccountHandler(accountService, transactionService, txLogService, transactionPublisher)
	transactionHandler := handler.NewTransactionHandler(accountService, txLogService)

	// Setup API routes with properly initialized handlers
	v1 := r.Group("/api/v1")
	{
		router.SetupAccountRoutes(v1, accountHandler)
		router.SetupTransactionRoutes(v1, transactionHandler)
	}

	// Start the API server
	port := config.GetConfig().App.Port
	fmt.Printf("Starting API server on port %s\n", port)
	r.Run(fmt.Sprintf(":%s", port))
}
