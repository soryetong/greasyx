package ginamiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/libs/ginaerror"
	"github.com/soryetong/greasyx/libs/ginasrv"
)

func Limiter(limiterStore *ginasrv.LimiterStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !limiterStore.Allow(ctx) {
			gina.Fail(ctx, ginaerror.RequestLimit)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
