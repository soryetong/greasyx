package gina

import (
	"fmt"
	"os"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/helper"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func init() {
	console.Append(serviceMgrCmd)
}

var serviceMgrCmd = &cobra.Command{
	Use:   "Start", // 命令名称, 不要修改
	Short: "Web项目的服务启动",
	Long:  `通过注册你指定的路由启动一个HTTP服务`,
	Run: func(cmd *cobra.Command, args []string) {
		defer closeServiceMgr()
		if len(serviceList) <= 0 {
			console.Echo.Fatalln("❌ 错误: 请务必通过实现接口 `gina.IService` 注册你要启动的服务")
		}

		var eg errgroup.Group
		for _, service := range serviceList {
			eg.Go(func() error {
				if err := service.OnStart(); err != nil {
					err = fmt.Errorf("服务 %s: %v", helper.GetCallerName(service), err)
					console.Echo.Errorf("❌  错误: %s", err)

					return err
				}

				return nil
			})
		}

		// 等待所有任务完成
		_ = eg.Wait()
		os.Exit(124)
	},
}

func closeServiceMgr() {
	_ = console.Echo.Sync()
	_ = Log.Sync()
	if rotationScheduler != nil {
		rotationScheduler.Stop()
	}
}

var serviceList []IService

func Register(service ...IService) {
	serviceList = append(serviceList, service...)
}

type IService interface {
	OnStart() error
}

type IServer struct {
	IService
}

func (self *IServer) OnStart() error {
	return nil
}
