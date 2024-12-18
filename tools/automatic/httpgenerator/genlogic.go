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

func (self *{{.LogicName}}) {{.FuncName}}(ctx context.Context,{{if .RequestType}} params *{{.TypesPackageName}}.{{.RequestType}}{{end}}) ({{if .ResponseType}} resp *{{.TypesPackageName}}.{{.ResponseType}},{{end}} err error) {
    // TODO implement

    return {{if .ResponseType}}resp,{{end}} nil
}
`

func (self *HttpGenerator) GenLogic() (err error) {
	logicPath := filepath.Join(self.Output, "logic")
	for _, datum := range self.Services {
		if datum.Name == "" {
			continue
		}

		logicTempPath := helper.CamelToSlash(datum.Name)
		if self.FileType == config.Logic_Handler_File_Type {
			logicPath = filepath.Join(logicPath, logicTempPath)
			self.LogicPackagePath[strings.ToLower(datum.Name)] = filepath.Join(self.ModuleName, logicPath)
			if err = os.MkdirAll(logicPath, os.ModePerm); err != nil {
				return err
			}
			return self.tileLogicWrite(datum, logicPath)
		}

		split := strings.Split(logicTempPath, "/")
		logicPath = filepath.Join(logicPath, strings.Join(split[:len(split)-1], "/"))
		self.LogicPackagePath[strings.ToLower(datum.Name)] = filepath.Join(self.ModuleName, logicPath)
		if err = os.MkdirAll(logicPath, os.ModePerm); err != nil {
			return err
		}

		return self.combineLogicWrite(datum, logicPath, split[len(split)-1])
	}

	return nil
}

func (self *HttpGenerator) combineLogicWrite(service *ServiceSpec, nowLogicPath, fileName string) (err error) {
	c := cases.Title(language.English)
	headerTmpl, err := template.New("logic").Parse(logicHeaderTemplate)
	if err != nil {
		return err
	}
	contentTmpl, err := template.New("logic").Parse(logicContentTemplate)
	if err != nil {
		return err
	}

	var builder strings.Builder
	var hasTypes bool
	logicName := fmt.Sprintf("%sLogic", c.String(fileName))
	self.LogicName[strings.ToLower(service.Name)] = fmt.Sprintf("New%s()", logicName)
	for _, route := range service.Routes {
		if hasTypes == false {
			hasTypes = route.ResponseType != "" || route.RequestType != ""
		}
	}

	split := strings.Split(helper.CamelToSlash(service.Name), "/")
	packageName := "logic"
	if len(split) >= 2 {
		packageName = split[len(split)-2]
	}
	self.LogicPackageName[strings.ToLower(service.Name)] = packageName
	headerData := map[string]interface{}{
		"LogicName":        logicName,
		"PackageName":      packageName,
		"HasTypes":         hasTypes,
		"TypesPackagePath": self.TypesPackagePath,
	}
	if err = headerTmpl.Execute(&builder, headerData); err != nil {
		return err
	}
	builder.WriteString("\n")

	for _, route := range service.Routes {
		self.LogicFuncName[strings.ToLower(route.Name)] = c.String(route.Name)
		logicData := map[string]string{
			"LogicName":        logicName,
			"FuncName":         c.String(route.Name),
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"TypesPackageName": self.TypesPackageName,
		}
		if err = contentTmpl.Execute(&builder, logicData); err != nil {
			return err
		}
		builder.WriteString("\n")
	}

	filename := filepath.Join(nowLogicPath, fmt.Sprintf("%s.go", strings.ToLower(fileName)))
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

func (self *HttpGenerator) tileLogicWrite(service *ServiceSpec, nowLogicPath string) (err error) {
	tmpl, err := template.New("logic").Parse(logicHeaderTemplate + logicContentTemplate)
	if err != nil {
		return err
	}

	tempPackageName := helper.CamelToSlash(service.Name)
	split := strings.Split(tempPackageName, "/")
	c := cases.Title(language.English)
	for _, route := range service.Routes {
		var builder strings.Builder
		newName := c.String(route.Name)
		packageName := strings.ToLower(split[len(split)-1])
		mapKey := strings.ToLower(newName)
		self.LogicName[mapKey] = fmt.Sprintf("New%sLogic()", newName)
		self.LogicPackageName[mapKey] = packageName
		self.LogicFuncName[mapKey] = newName
		logicData := map[string]interface{}{
			"PackageName":      packageName,
			"LogicName":        fmt.Sprintf("%sLogic", newName),
			"FuncName":         newName,
			"RequestType":      route.RequestType,
			"ResponseType":     route.ResponseType,
			"TypesPackagePath": self.TypesPackagePath,
			"TypesPackageName": self.TypesPackageName,
			"HasTypes":         route.ResponseType != "" || route.RequestType != "",
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
