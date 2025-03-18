package httpgenerator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/helper"
)

const serverContentTemplate = `
package {{.PackageName}}

import (
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/modules/httpmodule"
	"{{ .RouterPackagePath}}"
)

func init() {
	gina.Register(&{{ .ServerName}}{})
}

type {{ .ServerName}} struct {
	*gina.IServer

	httpModule httpmodule.IHttp
}

func (self *{{ .ServerName}}) OnStart() (err error) {
	// 添加回调函数
	self.httpModule.OnStop(self.exitCallback())

	self.httpModule.Init(self, "127.0.0.1:9888", 5, router.InitRouter())
	err = self.httpModule.Start()

	return
}

// TODO 添加回调函数, 无逻辑可直接删除这个方法
func (self *{{ .ServerName}}) exitCallback() *httpmodule.CallbackMap {
	callback := httpmodule.NewStopCallbackMap()
	callback.Append("exit", func() {
		gina.Log.Info("这是程序退出后的回调函数, 执行你想要执行的逻辑, 无逻辑可以直接删除这段代码")
	})
	
	return callback
}
`

func (self *HttpGenerator) GenServer() (err error) {
	split := strings.Split(strings.TrimLeft(self.Output, "./"), "/")
	outputName := "http"
	if len(split) >= 2 {
		outputName = split[0]
	}

	serverName := fmt.Sprintf("%sServer", helper.CapitalizeFirst(outputName))
	path := filepath.Join(self.Output, "server")
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	filename := filepath.Join(path, fmt.Sprintf("%s.go", serverName))
	if _, err = os.Stat(filename); err == nil {
		console.Echo.Info(fmt.Sprintf("服务文件: %s 已存在,不进行重写", filename))
		return nil
	}

	contentTmpl, err := template.New("server").Parse(serverContentTemplate)
	if err != nil {
		return err
	}

	var builder strings.Builder
	data := map[string]interface{}{
		"PackageName":       "server",
		"ServerName":        serverName,
		"RouterPackagePath": self.RouterPath,
	}
	if err = contentTmpl.Execute(&builder, data); err != nil {
		return err
	}
	builder.WriteString("")

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	console.Echo.Info("正在生成服务文件: ", filename)
	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.formatFileWithGofmt(filename)

	return nil
}
