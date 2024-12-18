package httpgenerator

import (
	"fmt"
	"github.com/soryetong/greasyx/helper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

const serverContentTemplate = `
package {{.PackageName}}

import (
	"github.com/soryetong/greasyx/gina"
	"github.com/soryetong/greasyx/modules/httpmodule"
	"{{ .RouterPackagePath}}"
)

func init() {
	gina.Register(&{{ .ServiceName}}{})
}

type {{ .ServiceName}} struct {
	*gina.IServer

	httpModule httpmodule.IHttp
}

func (self *{{ .ServiceName}}) OnStart() (err error) {
	// 添加回调函数
	self.httpModule.OnStop(self.exitCallback())

	self.httpModule.Init(self, "127.0.0.1:9888", 5, router.InitRouter())
	err = self.httpModule.Start()

	return
}

// TODO 添加回调函数, 无逻辑可直接删除这个方法
func (self *{{ .ServiceName}}) exitCallback() *httpmodule.CallbackMap {
	callback := httpmodule.NewStopCallbackMap()
	callback.Append("exit", func() {
		gina.Log.Info("这是程序退出后的回调函数, 执行你想要执行的逻辑, 无逻辑可以直接删除这段代码")
	})
	
	return callback
}
`

func (self *HttpGenerator) GenServer() (err error) {
	split := strings.Split(self.Output, "/")
	outputName := split[len(split)-1]
	packageName := outputName
	if outputName == "" {
		outputName = "http"
		packageName, _ = helper.GetModuleName()
		if strings.Contains(packageName, "-") {
			packageName = strings.ReplaceAll(packageName, "-", "_")
		}
	}

	c := cases.Title(language.English)
	contentTmpl, err := template.New("service").Parse(serverContentTemplate)
	if err != nil {
		return err
	}

	var builder strings.Builder
	serviceName := fmt.Sprintf("%sService", c.String(outputName))
	data := map[string]interface{}{
		"PackageName":       strings.ToLower(packageName),
		"ServiceName":       serviceName,
		"RouterPackagePath": self.RouterPath,
	}
	if err = contentTmpl.Execute(&builder, data); err != nil {
		return err
	}
	builder.WriteString("")

	path := filepath.Join(self.Output, "server")
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	filename := filepath.Join(path, fmt.Sprintf("%s.go", serviceName))
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("[GREASYX-TOOLS-INFO] 正在生成服务文件: %s", filename))
	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.normalFormatFileWithGofmt(filename)

	return nil
}
