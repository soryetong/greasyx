package gina

import (
	"fmt"
	"github.com/soryetong/greasyx/console"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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
		initConfig()
	},
}

func initConfig() {
	if configFile == "" {
		configFile = "./config.json"
	}
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-error] 读取配置文件错误: %s\n", err)
		os.Exit(104)
	}
}
