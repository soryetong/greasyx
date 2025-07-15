package ginamiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/libs/ginaauth"
	"github.com/soryetong/greasyx/libs/ginaerror"
)

func Jwt() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			gina.Fail(ctx, ginaerror.NeedLogin)
			ctx.Abort()
			return
		}

		claims, err := ginaauth.ParseJwtToken(tokenString[7:])
		if err != nil {
			gina.Fail(ctx, ginaerror.NeedLogin)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}
