package zhttp

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func New(routerProvider ZRouterProvider) *cobra.Command {
	return &cobra.Command{
		Use:   "Start", // 命令名称, 不要修改
		Short: "Web项目的服务启动",
		Long:  `通过注册你指定的路由启动一个HTTP服务`,
		Run: func(cmd *cobra.Command, args []string) {
			viper.SetDefault("App.Addr", "127.0.0.1:9901")
			viper.SetDefault("App.Timeout", 5)
			server := newHttp(
				viper.GetString("App.Addr"),
				time.Duration(viper.GetInt64("App.Timeout")),
				routerProvider.Register(),
			)
			server.setExitCallback(routerProvider.ExitCallback())
			server.start()
		},
	}
}

type IHttp struct {
	*gin.Engine
	srv *http.Server

	listenAddr    string
	tls           bool
	handleTimeout time.Duration

	exit         chan error
	exitCallback *ExitCallbackMap
}

func newHttp(addr string, timeout time.Duration, engine *gin.Engine) *IHttp {
	iHttp := new(IHttp)
	iHttp.listenAddr = addr
	iHttp.handleTimeout = timeout
	if engine == nil {
		engine = gin.Default()
	}
	iHttp.Engine = engine
	iHttp.srv = &http.Server{
		Addr:    iHttp.listenAddr,
		Handler: iHttp.Engine,
	}
	iHttp.exit = make(chan error)
	iHttp.exitCallback = NewExitCallbackMap()

	return iHttp
}

func (self *IHttp) setExitCallback(data *ExitCallbackMap) {
	self.exitCallback = data
}

func (self *IHttp) start() {
	go func() {
		if err := self.srv.ListenAndServe(); err != nil {
			self.exit <- err
		}
	}()

	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n\033[32m [GREASYX-info] "+
		"HTTP服务启动成功，当前地址: %s \033[0m\n\n", self.getServerAddr()))
	self.running()
}

func (self *IHttp) running() {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case _, _ = <-self.exit:
			self.exitCallback.Iterate()
			os.Exit(0)
		case <-quit:
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-info] 已接收到退出信号\n")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
			defer cancel()
			if err := self.srv.Shutdown(ctx); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-error] "+
					"服务停机失败: %s\n", err))
			}
		}
	}
}

func (self *IHttp) getServerAddr() string {
	prefix := "http"
	if self.tls {
		prefix = "https"
	}

	return fmt.Sprintf("%s://%s", prefix, self.listenAddr)
}
