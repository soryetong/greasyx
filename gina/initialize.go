package gina

import (
	"strings"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/modules/cachemodule"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

func init() {
	console.RootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	console.RootCmd.CompletionOptions.DisableDefaultCmd = true
	console.Append(greasyxCmd)
}

var greasyxCmd = &cobra.Command{
	Use:   "Gina", // 命令名称, 不要修改
	Short: "Greasyx框架初始化",
	Long:  `Greasyx框架初始化`,
	Run: func(cmd *cobra.Command, args []string) {
		// 初始化配置文件
		initConfig()
		// 初始化日志
		initILog()
		// 初始化缓存模块
		Cache = cachemodule.New(10000, 64, 0)
	},
}

func initConfig() {
	// 加载配置
	if configFile == "" {
		configFile = "./config.json"
	}
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		console.Echo.Fatalf("❌ 错误: 读取配置文件错误: %s", err)
	}

	// 设置默认值
	viper.SetDefault("App.Env", "test")
	routerPrefix := viper.GetString("App.RouterPrefix")
	if routerPrefix == "" {
		viper.SetDefault("App.RouterPrefix", "/api/v1")
	} else {
		viper.Set("App.RouterPrefix", "/"+strings.Trim(routerPrefix, "/"))
	}
}
