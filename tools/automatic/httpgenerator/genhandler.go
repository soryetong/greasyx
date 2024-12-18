package httpgenerator

import (
	"fmt"
	"github.com/soryetong/greasyx/helper"
	"github.com/soryetong/greasyx/tools/automatic/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

const handlerHeaderTemplate = `
package {{.PackageName}}

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/gina"
	"{{ .LogicPackagePath}}"
	{{if .HasTypes}} "{{.TypesPackagePath}}" {{end}}
	{{if .HasTypes}} "github.com/soryetong/greasyx/libs/xerror" {{end}}
)
`

const handlerContentTemplate = `

func {{ .HandlerName }}(ctx *gin.Context) {
{{if .RequestType}}	var req {{.TypesPackageName}}.{{.RequestType}}
	if err := ctx.ShouldBind(&req); err != nil {
		gina.FailWithMessage(ctx, xerror.Trans(err))
		return
	}

	{{if .ResponseType}}resp, err := {{ .LogicPackageName}}.{{ .LogicName}}.{{ .LogicFuncName}}(ctx, &req)
	if err != nil {
		gina.FailWithMessage(ctx, err.Error())
		return
	}

	gina.Success(ctx, resp)
	{{else}}if err := {{ .LogicPackageName}}.{{ .LogicName}}.{{ .LogicFuncName}}(ctx, &req);err != nil {
		gina.FailWithMessage(ctx, err.Error())
		return
	}

	gina.Success(ctx, nil){{end}}{{else}}{{if .ResponseType}}resp, err := {{ .LogicPackageName}}.{{ .LogicName}}.{{ .LogicFuncName}}(ctx)
	if err != nil {
		gina.FailWithMessage(ctx, err.Error())
		return
	}

	gina.Success(ctx, resp)
	{{else}}if err := {{ .LogicPackageName}}.{{ .LogicName}}.{{ .LogicFuncName}}(ctx);err != nil {
		gina.FailWithMessage(ctx, err.Error())
		return
	}

	gina.Success(ctx, nil){{end}}{{end}}}
`

func (self *HttpGenerator) GenHandler() (err error) {
	handlerPath := filepath.Join(self.Output, "handler")
	for _, datum := range self.Services {
		if datum.Name == "" {
			continue
		}

		logicTempPath := helper.CamelToSlash(datum.Name)
		if self.FileType == config.Logic_Handler_File_Type {
			handlerPath = filepath.Join(handlerPath, logicTempPath)
			self.HandlerPackPath[strings.ToLower(datum.Name)] = filepath.Join(self.ModuleName, handlerPath)
			if err = os.MkdirAll(handlerPath, os.ModePerm); err != nil {
				return err
			}
			return self.tileHandlerWrite(datum, handlerPath)
		}

		split := strings.Split(logicTempPath, "/")
		handlerPath = filepath.Join(handlerPath, strings.Join(split[:len(split)-1], "/"))
		self.HandlerPackPath[strings.ToLower(datum.Name)] = filepath.Join(self.ModuleName, handlerPath)
		if err = os.MkdirAll(handlerPath, os.ModePerm); err != nil {
			return err
		}

		return self.combineHandlerWrite(datum, handlerPath, split[len(split)-1])
	}

	return nil
}

func (self *HttpGenerator) combineHandlerWrite(service *ServiceSpec, nowHandlerPath, fileName string) (err error) {
	c := cases.Title(language.English)
	headerTmpl, err := template.New("handler").Parse(handlerHeaderTemplate)
	if err != nil {
		return err
	}
	contentTmpl, err := template.New("handler").Parse(handlerContentTemplate)
	if err != nil {
		return err
	}

	var builder strings.Builder
	var hasTypes bool
	for _, route := range service.Routes {
		if hasTypes == false {
			hasTypes = route.RequestType != ""
		}
	}

	split := strings.Split(helper.CamelToSlash(service.Name), "/")
	packageName := "handler"
	if len(split) >= 2 {
		packageName = split[len(split)-2]
	}
	headerData := map[string]interface{}{
		"PackageName":      packageName,
		"HasTypes":         hasTypes,
		"TypesPackagePath": self.TypesPackagePath,
		"LogicPackagePath": self.LogicPackagePath[strings.ToLower(service.Name)],
	}
	if err = headerTmpl.Execute(&builder, headerData); err != nil {
		return err
	}
	builder.WriteString("\n")

	for _, route := range service.Routes {
		self.HandlerPackName[strings.ToLower(route.Name+service.Name)] = packageName
		logicData := map[string]string{
			"HandlerName":      c.String(route.Name),
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"TypesPackageName": self.TypesPackageName,
			"LogicFuncName":    self.LogicFuncName[strings.ToLower(route.Name)],
			"LogicPackageName": self.LogicPackageName[strings.ToLower(service.Name)],
			"LogicName":        self.LogicName[strings.ToLower(service.Name)],
		}
		if err = contentTmpl.Execute(&builder, logicData); err != nil {
			return err
		}
		builder.WriteString("")
	}

	filename := filepath.Join(nowHandlerPath, fmt.Sprintf("%s.go", strings.ToLower(fileName)))
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.formatFileWithGofmt(filename)

	return nil
}

func (self *HttpGenerator) tileHandlerWrite(service *ServiceSpec, nowLogicPath string) (err error) {
	tmpl, err := template.New("handler").Parse(handlerHeaderTemplate + handlerContentTemplate)
	if err != nil {
		return err
	}

	split := strings.Split(helper.CamelToSlash(service.Name), "/")
	c := cases.Title(language.English)
	for _, route := range service.Routes {
		packageName := strings.ToLower(split[len(split)-1])
		self.HandlerPackName[strings.ToLower(route.Name+service.Name)] = packageName

		var builder strings.Builder
		newName := strings.ToLower(c.String(route.Name))
		logicData := map[string]interface{}{
			"PackageName":      packageName,
			"HasTypes":         route.RequestType != "",
			"TypesPackagePath": self.TypesPackagePath,
			"LogicPackagePath": self.LogicPackagePath[strings.ToLower(service.Name)],
			"HandlerName":      c.String(route.Name),
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"TypesPackageName": self.TypesPackageName,
			"LogicFuncName":    self.LogicFuncName[newName],
			"LogicPackageName": self.LogicPackageName[newName],
			"LogicName":        self.LogicName[newName],
		}
		if err = tmpl.Execute(&builder, logicData); err != nil {
			return err
		}
		builder.WriteString("\n")

		filename := filepath.Join(nowLogicPath, fmt.Sprintf("%s.go", strings.ToLower(route.Name)))
		file, err := os.Create(filename)
		defer file.Close()
		if err != nil {
			return err
		}

		if _, err = file.WriteString(builder.String()); err != nil {
			return err
		}
		self.formatFileWithGofmt(filename)
	}

	return nil
}
