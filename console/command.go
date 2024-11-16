package console

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

type runFunc func(cmd *cobra.Command, args []string)

var mapCommand = make(map[string]runFunc)
var RootCmd = &cobra.Command{
	Use:   "Root",
	Short: "go gin frame",
	Long:  `Web project scaffolding based on go+gin framework`,
	Run: func(cmd *cobra.Command, args []string) {
		if mapCommand["Gina"] == nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-error] "+
				"请务必在入口函数 `main()` 中通过 `_ github.com/soryetong/greasyx/gina` 加载Greasyx模块\n")
			os.Exit(104)
		}
		if mapCommand["Start"] == nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-error] "+
				"请务必通过 `zhttp.New(&router.Route{})` 加载Start服务 (`&router.Route{}`是你自定义的路由)\n")
			os.Exit(104)
		}
		// 确保 Gina 命令最先执行
		mapCommand["Gina"](cmd, args)

		for name, runFunc := range mapCommand {
			if name == "Start" || name == "Gina" {
				continue
			}
			runFunc(cmd, args)
		}

		// 确保 Start 命令最后执行
		mapCommand["Start"](cmd, args)
	},
}

func Run(cmdList ...*cobra.Command) {
	Append(cmdList...)

	if err := RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-error] cmd run err: %s\n", err)
		os.Exit(104)
	}
}

func Append(cmdList ...*cobra.Command) {
	for _, cmd := range cmdList {
		RootCmd.AddCommand(cmd)
		mapCommand[cmd.Use] = cmd.Run
	}
}
