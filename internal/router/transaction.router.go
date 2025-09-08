package router

import (
	"golang-exercise/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupTransactionRoutes(router *gin.RouterGroup, transactionHandler *handler.TransactionHandler) {
	transactions := router.Group("/transactions")
	{
		// Get transaction history for a specific account
		transactions.GET("/account/:account_number/history", transactionHandler.GetTransactionHistory)

		// Get transaction status by transaction ID
		transactions.GET("/:transaction_id/status", transactionHandler.GetTransactionStatus)
	}
}
