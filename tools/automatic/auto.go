package automatic

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/ginahelper"
	"github.com/soryetong/greasyx/tools/automatic/config"
	"github.com/soryetong/greasyx/tools/automatic/httpgenerator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	console.Echo = ginahelper.InitSugaredLogger()
	moduleName, err := ginahelper.GetModuleName()
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

	if argMap["src"] == "" || argMap["output"] == "" {
		argMap["src"] = promptForInput("请输入API文件路径 - 必填")
		if argMap["src"] == "" {
			console.Echo.Fatalf("❌ 错误: 输入api文件所在的路径 \n")
		}
		argMap["output"] = promptForInput("请输入生成的代码存放路径 - 必填")
		if argMap["output"] == "" {
			console.Echo.Fatalf("❌ 错误: 输入生成的代码存放路径 \n")
		}
	}

	hasConfig := false
	if argMap["config"] != "" {
		viper.SetConfigFile(argMap["config"])
		if err := viper.ReadInConfig(); err != nil {
			console.Echo.Fatalf("❌ 错误: 读取配置文件错误: %s", err)
		}
		hasConfig = true
	}

	routerEnterGo := filepath.Join(argMap["output"], "router", "enter.go")
	_, err = os.Stat(routerEnterGo)
	routerPrefix := "/api/v1"
	needRequestLog := "NO"
	port := ":9888"
	if os.IsNotExist(err) {
		var routerPrefixZ string
		var portZ string
		if !hasConfig {
			routerPrefixZ = promptForInput("自定义的路由前缀 - 选填(默认 \"/api/v1\" )")
			portZ = promptForInput("自定义的服务器端口 - 选填(默认 :9888 )")
		} else {
			routerPrefixZ = viper.GetString("App.RouterPrefix")
			portZ = viper.GetString("App.Addr")
			if portZ == "" {
				viper.SetDefault("App.Addr", ":9888")
			}
		}

		if routerPrefixZ != "" {
			routerPrefix = routerPrefixZ
		}
		if portZ != "" {
			port = portZ
		}

		items := []string{"YES", "NO"}
		needRequestLog = promptForSelect("是否需要引入网络消息日志", items)
	}

	typePackageName := "types"
	xCtx := new(httpgenerator.XContext)
	switch argMap["mode"] {
	default:
		xCtx.ModuleName = moduleName
		xCtx.Output = argMap["output"]
		xCtx.Src = argMap["src"]
		xCtx.RouterPrefix = "/" + strings.Trim(routerPrefix, "/")
		xCtx.Port = port
		xCtx.NeedRequestLog = needRequestLog == "YES"
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

func promptForInput(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}
	result, err := prompt.Run()
	if err != nil {
		return ""
	}

	return result
}

func promptForSelect(label string, items []string) string {
	prompt := promptui.Select{
		Label:    label,
		Items:    items,
		HideHelp: true,
	}
	_, result, err := prompt.Run()
	if err != nil {
		return ""
	}

	return result
}
