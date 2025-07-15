package httpgenerator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/ginahelper"
	"github.com/spf13/viper"
)

const serverContentTemplate = `
package {{.PackageName}}

import (
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/modules/httpmodule"
	"{{ .RouterPackagePath}}"
	{{if .HasViper}} "github.com/spf13/viper" {{end}}
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

	{{if .HasViper}} self.httpModule.Init(self, viper.GetString("App.Addr"), 5, router.InitRouter()) {{ else }}
	self.httpModule.Init(self, "{{ .ServerAddr}}", 5, router.InitRouter()) {{end}}
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

	serverName := fmt.Sprintf("%sServer", ginahelper.UcFirst(outputName))
	path := filepath.Join(self.Output, "server")
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	filename := filepath.Join(path, fmt.Sprintf("%s.go", serverName))
	if _, err = os.Stat(filename); err == nil {
		console.Echo.Info(fmt.Sprintf("服务文件: %s 已存在，不进行重写", filename))
		return nil
	}

	contentTmpl, err := template.New("server").Parse(serverContentTemplate)
	if err != nil {
		return err
	}

	hasViper := true
	serverAddr := viper.GetString("App.Addr")
	if serverAddr == "" {
		hasViper = false
		serverAddr = "127.0.0.1" + self.Port
	}
	var builder strings.Builder
	data := map[string]interface{}{
		"PackageName":       "server",
		"ServerName":        serverName,
		"ServerAddr":        serverAddr,
		"HasViper":          hasViper,
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
