package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		ctx.Next()

		log.Printf("%s %s %v", ctx.Request.URL.Path, ctx.Request.Method, time.Since(start))
	}
}
