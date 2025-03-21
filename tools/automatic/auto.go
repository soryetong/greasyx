package automatic

import (
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/helper"
	"github.com/soryetong/greasyx/tools/automatic/config"
	"github.com/soryetong/greasyx/tools/automatic/httpgenerator"
	"github.com/spf13/cobra"
	"strings"
)

func init() {
	console.Append(autoCCmd)
}

var argMap = make(map[string]string)

var autoCCmd = &cobra.Command{
	Use:   "autoc", // 命令名称, 不要修改
	Short: "Greasyx-自动生成代码工具",
	Long:  `Greasyx-自动生成代码工具`,
	Run:   runFunc,
}

func runFunc(cmd *cobra.Command, args []string) {
	console.Echo = helper.InitSugaredLogger()
	moduleName, err := helper.GetModuleName()
	if err != nil {
		console.Echo.Fatalf("❌ 错误: 无法获取当前项目Module名, 错误为: [%v] \n", err)
	}
	for _, arg := range args {
		argArr := strings.Split(arg, "=")
		if len(argArr) != 2 {
			continue
		}

		argMap[argArr[0]] = argArr[1]
	}

	if argMap["src"] == "" {
		console.Echo.Fatalf("❌ 错误: 请通过: [%s], 输入api文件所在的路径 \n", "autoc src={path}")
	}

	typePackageName := "types"
	xCtx := new(httpgenerator.XContext)
	switch argMap["mode"] {
	default:
		xCtx.ModuleName = moduleName
		xCtx.Output = argMap["output"]
		xCtx.Src = argMap["src"]
		xCtx.TypesPackageName = typePackageName
		xCtx.FileType = config.FileType(strings.ToUpper(argMap["type"]))
		xCtx.LogicPackagePath = make(map[string]string)
		xCtx.LogicFuncName = make(map[string]string)
		xCtx.LogicPackageName = make(map[string]string)
		xCtx.LogicName = make(map[string]string)
		xCtx.HandlerPackPath = make(map[string]string)
		xCtx.HandlerPackName = make(map[string]string)
		httpGen := httpgenerator.NewGenerator(xCtx)
		err = httpGen.Generate()
	}

	if err != nil {
		console.Echo.Fatalf("❌ 错误: 自动生成代码失败, 错误为: [%v] \n", err)
	}
}
