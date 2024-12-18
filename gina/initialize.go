package gina

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"github.com/soryetong/greasyx/console"
)

var configFile string

func init() {
	console.RootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file")
	console.Append(greasyxCmd)
}

var greasyxCmd = &cobra.Command{
	Use:   "Gina", // 命令名称, 不要修改
	Short: "Greasyx框架初始化",
	Long:  `Greasyx框架初始化`,
	Run: func(cmd *cobra.Command, args []string) {
		initConfig() // 初始化配置文件
		initILog()   // 初始化日志
	},
}

func initConfig() {
	if configFile == "" {
		configFile = "./config.json"
	}
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-ERROR] 读取配置文件错误: %s\n", err)
		os.Exit(104)
	}
}
