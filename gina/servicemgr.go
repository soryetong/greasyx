package gina

import (
	"fmt"
	"github.com/soryetong/greasyx/helper"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"os"
	"github.com/soryetong/greasyx/console"
)

func init() {
	console.Append(serviceMgrCmd)
}

var serviceMgrCmd = &cobra.Command{
	Use:   "Start", // 命令名称, 不要修改
	Short: "Web项目的服务启动",
	Long:  `通过注册你指定的路由启动一个HTTP服务`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(serviceList) <= 0 {
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-ERROR] "+
				"请务必通过实现接口 `gina.IService` 注册你要启动的服务 \n")
			os.Exit(124)
		}

		var eg errgroup.Group
		for _, service := range serviceList {
			eg.Go(func() error {
				if err := service.OnStart(); err != nil {
					err = fmt.Errorf("服务 %s: %v", helper.GetCallerName(service), err)
					_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-ERROR] %v", err)

					return err
				}

				return nil
			})
		}

		// 等待所有任务完成
		_ = eg.Wait()
		_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-DEBUG] 服务已关闭 \n")
		os.Exit(124)
	},
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
