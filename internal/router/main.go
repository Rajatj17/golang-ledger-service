package router

import (
	"golang-exercise/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		SetupAccountRoutes(v1, &handler.AccountHandler{})
		SetupTransactionRoutes(v1, &handler.TransactionHandler{})
	}
}
