package ginamiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/ginahelper"
	"github.com/soryetong/greasyx/libs/ginaauth"
	"github.com/soryetong/greasyx/libs/ginaerror"
	"github.com/spf13/viper"
)

func Casbin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.ToLower(ctx.Request.Method) == "OPTIONS" {
			ctx.Next()
			return
		}

		roleId := ginaauth.GetTokenData[int64](ctx, "role_id")
		if roleId == 0 {
			console.Echo.Info("ℹ️ 提示: 无法使用 `Casbin` 权限校验, 请确保 `Token` 中包含了字段 `role_id`")
			ctx.Next()
			return
		}

		path := ginahelper.ConvertToRestfulURL(strings.TrimPrefix(ctx.Request.URL.Path, viper.GetString("App.RouterPrefix")))
		success, _ := gina.Casbin.Enforce(ginahelper.Int64ToString(roleId), path, ctx.Request.Method)
		if !success {
			gina.Fail(ctx, ginaerror.NoAuth)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
