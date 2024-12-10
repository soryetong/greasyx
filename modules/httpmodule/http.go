package httpmodule

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/helper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type IHttp struct {
	*gin.Engine

	name       string
	listenAddr string
	timeout    time.Duration
	srv        *http.Server
	tls        bool

	stopCallback *CallbackMap
	exit         chan error
}

func (self *IHttp) Init(caller interface{}, addr string, timeout time.Duration, engine *gin.Engine) {
	self.name = helper.GetCallerName(caller)
	self.exit = make(chan error)
	self.listenAddr = addr
	self.timeout = timeout
	self.Engine = engine
}

func (self *IHttp) OnInit() {
	self.srv = &http.Server{
		Addr:    self.listenAddr,
		Handler: self.Engine,
	}
}

func (self *IHttp) OnStop(data *CallbackMap) {
	self.stopCallback = data
}

func (self *IHttp) Start() error {
	self.OnInit()
	go func() {
		if err := self.srv.ListenAndServe(); err != nil {
			self.exit <- err
		}
	}()

	self.tls = false
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n\033[32m [GREASYX-GINFO] "+
		"服务 %s 启动成功，地址为: %s \033[0m\n", self.name, self.getServerAddr()))

	return self.running()
}

func (self *IHttp) StartTLS(certFile, keyFile string) error {
	self.OnInit()
	go func() {
		if err := self.srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			self.exit <- err
		}
	}()

	self.tls = true
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n\033[32m [GREASYX-GINFO] "+
		"服务 %s 启动成功，地址为: %s \033[0m\n", self.name, self.getServerAddr()))

	return self.running()
}

func (self *IHttp) running() error {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case _, _ = <-self.exit:
			self.stopCallback.Foreach()
			return nil
		case <-quit:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*self.timeout)
			defer cancel()
			if err := self.srv.Shutdown(ctx); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("\n[GREASYX-ERROR] "+
					"服务停机失败: %s\n", err))

				return err
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
