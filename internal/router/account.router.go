package router

import (
	"golang-exercise/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupAccountRoutes(router *gin.RouterGroup, accountHandler *handler.AccountHandler) {
	accounts := router.Group("/accounts")
	{
		accounts.POST("/", accountHandler.CreateAccount)
		accounts.GET("/:account_number", accountHandler.GetAccount)

		// Direct funds processing endpoint
		accounts.POST("/funds", accountHandler.ProcessFunds)

		accounts.GET("/:account_number/balance", accountHandler.GetAccountBalance)
	}
}
