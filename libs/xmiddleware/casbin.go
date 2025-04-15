package xmiddleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/helper"
	"github.com/soryetong/greasyx/libs/xauth"
	"github.com/soryetong/greasyx/libs/xerror"
	"github.com/spf13/viper"
)

func Casbin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if strings.ToLower(ctx.Request.Method) == "OPTIONS" {
			ctx.Next()
			return
		}

		roleId := xauth.GetTokenData[int64](ctx, "role_id")
		if roleId == 0 {
			console.Echo.Info("ℹ️ 提示: 无法使用 `Casbin` 权限校验, 请确保 `Token` 中包含了字段 `role_id`")
			ctx.Next()
			return
		}

		path := helper.ConvertToRestfulURL(strings.TrimPrefix(ctx.Request.URL.Path, viper.GetString("App.RouterPrefix")))
		success, _ := gina.Casbin.Enforce(helper.Int64ToString(roleId), path, ctx.Request.Method)
		if success {
			ctx.Next()
		} else {
			gina.Fail(ctx, xerror.NoAuth)
			ctx.Abort()
			return
		}
	}
}
