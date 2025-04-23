package httpgenerator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/soryetong/greasyx/console"
	"github.com/soryetong/greasyx/helper"
	"github.com/soryetong/greasyx/tools/automatic/config"
)

const logicHeaderTemplate = `
package {{.PackageName}}

import (
	"context"
	{{if .HasTypes}} "{{.TypesPackagePath}}" {{end}}
)

type {{.LogicName}} struct {
}

func New{{.LogicName}}() *{{.LogicName}} {
	return &{{.LogicName}}{}
}
`

const logicContentTemplate = `

// @Summary {{ .Summary }}
func (self *{{.LogicName}}) {{.FuncName}}(ctx context.Context,{{if .PathParam}} {{.PathParam}} int64,{{end}}{{if .RequestType}} params *{{.TypesPackageName}}.{{.RequestType}}{{end}}) ({{if .ResponseType}} resp {{if not (hasPrefix .ResponseType "[]")}}*{{.TypesPackageName}}.{{end}}{{.ResponseType}},{{end}} err error) {
    // TODO implement

    return
}
`

func (self *HttpGenerator) GenLogic() (err error) {
	logicPath := filepath.Join(self.Output, "logic")
	for _, datum := range self.Services {
		if datum.Name == "" {
			continue
		}

		logicTempPath := helper.SeparateCamel(datum.Name, "/")
		if self.FileType == config.Logic_Handler_File_Type {
			logicPath = filepath.Join(logicPath, logicTempPath)
			self.LogicPackagePath[strings.ToLower(datum.Name)] = filepath.Join(self.ModuleName, logicPath)
			if err = os.MkdirAll(logicPath, os.ModePerm); err != nil {
				return err
			}
			return self.tileLogicWrite(datum, logicPath)
		}

		// split := strings.Split(logicTempPath, "/")
		// logicPath = filepath.Join(logicPath, strings.Join(split[:len(split)-1], "/"))
		self.LogicPackagePath[strings.ToLower(datum.Name)] = filepath.Join(self.ModuleName, logicPath)
		if err = os.MkdirAll(logicPath, os.ModePerm); err != nil {
			return err
		}

		return self.combineLogicWrite(datum, logicPath)
	}

	return nil
}

func (self *HttpGenerator) tileLogicWrite(service *ServiceSpec, nowLogicPath string) (err error) {
	tmpl, err := template.New("logic").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(logicHeaderTemplate + logicContentTemplate)
	if err != nil {
		return err
	}

	tempPackageName := helper.SeparateCamel(service.Name, "/")
	split := strings.Split(tempPackageName, "/")
	for _, route := range service.Routes {
		filename := filepath.Join(nowLogicPath, fmt.Sprintf("%s.go", strings.ToLower(route.Name)))
		if _, err = os.Stat(filename); err == nil {
			console.Echo.Info(filename, " 已存在,不进行重写")
			continue
		}

		var builder strings.Builder
		newName := helper.CapitalizeFirst(route.Name)
		packageName := strings.ToLower(split[len(split)-1])
		mapKey := strings.ToLower(newName)
		self.LogicName[mapKey] = fmt.Sprintf("New%sLogic()", newName)
		self.LogicPackageName[mapKey] = packageName
		self.LogicFuncName[mapKey] = newName
		logicData := map[string]interface{}{
			"Summary":          route.Summary,
			"PackageName":      packageName,
			"LogicName":        fmt.Sprintf("%sLogic", newName),
			"FuncName":         newName,
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"TypesPackagePath": self.TypesPackagePath,
			"TypesPackageName": self.TypesPackageName,
			"HasTypes":         route.ResponseType != "" || route.RequestType != "",
			"PathParam":        route.RustFulKey,
		}
		if err = tmpl.Execute(&builder, logicData); err != nil {
			return err
		}
		builder.WriteString("\n")

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

func (self *HttpGenerator) combineLogicWrite(service *ServiceSpec, nowLogicPath string) (err error) {
	split := strings.Split(helper.SeparateCamel(service.Name, "/"), "/")
	filename := filepath.Join(nowLogicPath, fmt.Sprintf("%s.go", strings.ToLower(strings.Join(split, "_"))))

	var hasTypes bool
	logicName := fmt.Sprintf("%sLogic", helper.CapitalizeFirst(service.Name))
	self.LogicName[strings.ToLower(service.Name)] = fmt.Sprintf("New%s()", logicName)
	for _, route := range service.Routes {
		if hasTypes == false {
			hasTypes = route.ResponseType != "" || route.RequestType != ""
		}
	}

	packageName := "logic"
	self.LogicPackageName[strings.ToLower(service.Name)] = packageName

	var fileContent []byte
	if _, err = os.Stat(filename); err == nil {
		fileContent, err = os.ReadFile(filename)
		if err != nil {
			return err
		}
	}

	if fileContent == nil {
		headerData := map[string]interface{}{
			"LogicName":        logicName,
			"PackageName":      packageName,
			"HasTypes":         hasTypes,
			"TypesPackagePath": self.TypesPackagePath,
		}
		headerTmpl, tmplErr := template.New("logic").Funcs(template.FuncMap{
			"hasPrefix": strings.HasPrefix,
		}).Parse(logicHeaderTemplate)
		if tmplErr != nil {
			return tmplErr
		}

		var builder strings.Builder
		if err = headerTmpl.Execute(&builder, headerData); err != nil {
			return err
		}
		// 写入文件
		if err = os.WriteFile(filename, []byte(builder.String()), 0644); err != nil {
			return err
		}
		// 重新读取文件内容供后续检测使用
		fileContent, err = os.ReadFile(filename)
		if err != nil {
			return err
		}
	}

	contentTmpl, err := template.New("logic").Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(logicContentTemplate)
	if err != nil {
		return err
	}
	for _, route := range service.Routes {
		newName := helper.CapitalizeFirst(route.Name)
		self.LogicFuncName[strings.ToLower(route.Name)] = newName
		logicData := map[string]string{
			"Summary":          route.Summary,
			"LogicName":        logicName,
			"FuncName":         newName,
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"TypesPackageName": self.TypesPackageName,
			"PathParam":        route.RustFulKey,
		}
		funcSignature := fmt.Sprintf("func (self *%s) %s(", logicName, newName)
		if strings.Contains(string(fileContent), funcSignature) {
			console.Echo.Info(fmt.Sprintf("logic: %s 中 %s 方法已存在，不进行重写", filename, newName))
			continue
		}

		var methodBuilder strings.Builder
		if err = contentTmpl.Execute(&methodBuilder, logicData); err != nil {
			return err
		}

		f, fErr := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
		if fErr != nil {
			return fErr
		}

		if _, err = f.WriteString("\n" + methodBuilder.String()); err != nil {
			f.Close()
			return err
		}
		f.Close()

		fileContent, err = os.ReadFile(filename)
		if err != nil {
			return err
		}
	}

	self.formatFileWithGofmt(filename)

	return nil
}
