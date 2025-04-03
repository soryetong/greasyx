package middleware

import (
	"github.com/gin-gonic/gin"
)

func RequestLog() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.Next()
	}
}
