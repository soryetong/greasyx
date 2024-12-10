package gina

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/libs/xerror"
	"github.com/soryetong/greasyx/utils"
	"net/http"
	"time"
)

type PageResult struct {
	List        interface{} `json:"list"`
	Total       int64       `json:"total"`
	CurrentPage int64       `json:"current_page"`
	PageSize    int64       `json:"page_size"`
}

type Response struct {
	Code    int64       `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	NowTime int64       `json:"nowTime"`
	UseTime string      `json:"useTime"`
}

func Result(ctx *gin.Context, code int64, data interface{}, msg string) {
	resp := Response{
		Code:    code,
		Msg:     msg,
		Data:    data,
		NowTime: time.Now().Unix(),
	}
	if useTime(ctx) != "" {
		resp.UseTime = useTime(ctx)
	}
	ctx.JSON(http.StatusOK, resp)
}

func Success(ctx *gin.Context, data interface{}) {
	Result(ctx, xerror.OK, data, "success")
}

func SuccessWithMessage(ctx *gin.Context, msg string) {
	Result(ctx, xerror.OK, nil, msg)
}

func FailWithMessage(ctx *gin.Context, msg string) {
	Result(ctx, xerror.Error, nil, msg)
}

func Fail(ctx *gin.Context, code int64, msg ...string) {
	var message string
	if len(msg) > 0 {
		message = msg[0]
	} else {
		message = xerror.GetErrorMessage(code)
	}

	Result(ctx, code, nil, message)
}

func useTime(c *gin.Context) string {
	startTime, _ := c.Get("requestStartTime")
	stopTime := time.Now().UnixMicro()
	if startTime == nil {
		return ""
	}

	return fmt.Sprintf("%.6f", float64(stopTime-utils.InterfaceToInt64(startTime))/1000000)
}
