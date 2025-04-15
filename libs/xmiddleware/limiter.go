package xmiddleware

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/libs/xapp"
	"github.com/soryetong/greasyx/libs/xerror"
)

func Limiter(limiterStore *xapp.LimiterStore) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !limiterStore.Allow(ctx) {
			gina.Fail(ctx, xerror.RequestLimit)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
