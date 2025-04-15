package xmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/libs/xauth"
	"github.com/soryetong/greasyx/libs/xerror"
)

func Jwt() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			gina.Fail(ctx, xerror.NeedLogin)
			ctx.Abort()
			return
		}

		claims, err := xauth.ParseJwtToken(tokenString[7:])
		if err != nil {
			gina.Fail(ctx, xerror.NeedLogin)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
