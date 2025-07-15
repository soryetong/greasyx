package httpgenerator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/ginahelper"
)

const middlewareContentTemplate = `
package ginamiddleware

import (
	"github.com/gin-gonic/gin"
)

func {{ .Name }}() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		
		ctx.Next()
	}
}
`

func (self *HttpGenerator) GenMiddleware(middlewarePath, name string) (err error) {
	middlewareTmpl, err := template.New("ginamiddleware").Parse(middlewareContentTemplate)
	if err != nil {
		return err
	}

	var builder strings.Builder
	data := map[string]interface{}{
		"Name": ginahelper.UcFirst(name),
	}
	if err = middlewareTmpl.Execute(&builder, data); err != nil {
		return err
	}
	builder.WriteString("\n")

	newName := ginahelper.SeparateCamel(name, "_")
	filename := filepath.Join(middlewarePath, fmt.Sprintf("%s.go", strings.ToLower(newName)))
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	console.Echo.Info("正在生成中间件文件: ", filename)
	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.formatFileWithGofmt(filename)

	return
}
