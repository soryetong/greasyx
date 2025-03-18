package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/helper"
	"github.com/soryetong/greasyx/libs/auth"
	"github.com/soryetong/greasyx/libs/xerror"
)

func Casbin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		roleId := auth.GetTokenData[int64](ctx, "role_id")
		if roleId == 0 {
			console.Echo.Info("ℹ️ 提示: 无法使用 `Casbin` 权限校验, 请确保 `Token` 中包含了 `role_id`")
			ctx.Next()
			return
		}

		path := helper.ConvertToRestfulURL(strings.TrimPrefix(ctx.Request.URL.Path, "/api"))
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
