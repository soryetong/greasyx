package httpmodule

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/helper"
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
			console.Echo.Errorf("❌  错误: 服务启动异常 %s", err)
			self.exit <- err
		}
	}()

	self.tls = false
	console.Echo.Infof("ℹ️ 提示: 服务 %s 启动成功，地址为: %s\n", self.name, self.getServerAddr())

	return self.running()
}

func (self *IHttp) StartTLS(certFile, keyFile string) error {
	self.OnInit()
	go func() {
		if err := self.srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			console.Echo.Errorf("❌  错误: 服务启动异常 %s", err)
			self.exit <- err
		}
	}()

	self.tls = true
	console.Echo.Infof("ℹ️ 提示: 服务 %s 启动成功，地址为: %s\n", self.name, self.getServerAddr())

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
				// console.Echo.Warnf("⚠️ 警告: 服务停机失败: %s\n", err)

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

	addr := self.listenAddr
	addrArr := strings.Split(self.listenAddr, ":")
	if addrArr[0] == "" {
		addrArr[0] = helper.GetLocalIP()
		addr = strings.Join(addrArr, ":")
	}

	return fmt.Sprintf("%s://%s", prefix, addr)
}
