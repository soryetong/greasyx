package xresp

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"github.com/soryetong/greasyx/libs/xerror"
	"fmt"
	"github.com/soryetong/greasyx/utils"
)

type Response struct {
	Code    int64       `json:"code"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
	NowTime int64       `json:"nowTime"`
	UseTime string      `json:"useTime"`
}

func Result(ctx *gin.Context, code int64, data interface{}, msg string) {
	ctx.JSON(http.StatusOK, Response{
		Code:    code,
		Msg:     msg,
		Data:    data,
		NowTime: time.Now().Unix(),
		UseTime: useTime(ctx),
	})
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
	runTimeStr := "0.000000"
	if startTime != nil {
		runTimeStr = fmt.Sprintf("%.6f", float64(stopTime-utils.TypeConvert().InterfaceToInt64(startTime))/1000000)
	}

	return runTimeStr
}
