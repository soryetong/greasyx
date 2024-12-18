package console

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type runFunc func(cmd *cobra.Command, args []string)

var mapCommand = make(map[string]runFunc)

var RootCmd = &cobra.Command{
	Use:   "Root",
	Short: "go gin frame",
	Long:  `Web project scaffolding based on go+gin framework`,
	Run: func(cmd *cobra.Command, args []string) {
		if mapCommand["Gina"] == nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n[GREASYX-ERROR] "+
				"请务必在入口函数 `main()` 中通过 `_ github.com/soryetong/greasyx/gina` 加载Greasyx模块\n")
			os.Exit(104)
		}
		// 确保 Gina 命令最先执行
		mapCommand["Gina"](cmd, args)

		for name, runFunc := range mapCommand {
			name = strings.ToUpper(name)
			if name == "START" || name == "GINA" || name == "AUTOC" {
				continue
			}
			runFunc(cmd, args)
		}

		// 确保 Start 命令最后执行
		mapCommand["Start"](cmd, args)
	},
}

func Append(cmdList ...*cobra.Command) {
	for _, cmd := range cmdList {
		RootCmd.AddCommand(cmd)
		mapCommand[cmd.Use] = cmd.Run
	}
}
