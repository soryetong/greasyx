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

const routerContentTemplate = `
package router

import (
	"github.com/gin-gonic/gin"
	"{{ .HandlerPackPath}}"
	{{if .HasGinaMiddleware}} "github.com/soryetong/greasyx/libs/middleware" {{end}}
	{{if .HasProjMiddleware}} middleware2 "{{ .ProjMiddlewarePath}}" {{end}}
)

func Init{{.NowGroupName}}Router(routerGroup *gin.RouterGroup) {
{{range .Routes}}{{.GroupName}}Group := routerGroup.Group("/{{.RouteGroup}}"){{range .Middleware}}
{{.GroupName}}Group.Use({{if .InGina}} middleware {{ else }} middleware2 {{end}}.{{ .Name}}()){{end}}
{{"{"}}{{range .Routes}}
	{{.GroupName}}Group.{{.Method}}("/{{.Path}}", {{.HandlerPackName}}.{{.HandlerName}}){{end}}
{{"}"}}{{end}}
}
`

type RouteTemplateData struct {
	NowGroupName       string
	HandlerPackPath    string
	HasGinaMiddleware  bool
	HasProjMiddleware  bool
	ProjMiddlewarePath string
	Routes             []RouteGroupTemplateData
}

type RouteGroupTemplateData struct {
	GroupName  string
	RouteGroup string
	Middleware []RouteMiddleware
	Routes     []RouteSpecTemplateData
}

type RouteSpecTemplateData struct {
	GroupName       string
	Method          string
	Path            string
	HandlerPackName string
	HandlerName     string
}

type RouteMiddleware struct {
	GroupName string
	InGina    bool
	Name      string
}

var ownedMiddleware = map[string]struct{}{
	"Jwt":   {},
	"Cross": {},
}

func (self *HttpGenerator) GenRouter(groupName string) (err error) {
	nowRouterPath := filepath.Join(self.Output, "router")
	middlewarePath := filepath.Join(self.Output, "middleware")
	self.RouterPath = filepath.Join(self.ModuleName, nowRouterPath)
	if err = os.MkdirAll(nowRouterPath, os.ModePerm); err != nil {
		return err
	}

	handlerPackPath := ""
	for _, service := range self.Services {
		handlerPackPath = self.HandlerPackPath[strings.ToLower(service.Name)]
	}

	templateData := RouteTemplateData{
		// NowGroupName:    helper.CapitalizeFirst(groupName),
		HandlerPackPath:    handlerPackPath,
		ProjMiddlewarePath: self.ModuleName + "/" + middlewarePath,
		Routes:             []RouteGroupTemplateData{},
	}

	newFileName := ""
	projMiddleware := make(map[string]struct{})
	for _, service := range self.Services {
		templateData.NowGroupName = helper.CapitalizeFirst(service.Name)
		newGroupName := strings.ToLower(service.Name[:1]) + service.Name[1:]
		split := strings.Split(helper.SeparateCamel(service.Name, "/"), "/")
		newFileName = strings.ToLower(strings.Join(split, "_"))
		group := RouteGroupTemplateData{
			GroupName:  newGroupName,
			RouteGroup: strings.ToLower(groupName),
			Routes:     []RouteSpecTemplateData{},
		}
		middlewareArr := strings.Split(service.Middleware, ",")
		for _, route := range middlewareArr {
			if route == "" {
				continue
			}
			_, ok := ownedMiddleware[route]
			group.Middleware = append(group.Middleware, RouteMiddleware{
				GroupName: newGroupName,
				Name:      route,
				InGina:    ok,
			})
			if ok {
				templateData.HasGinaMiddleware = true
			} else {
				templateData.HasProjMiddleware = true
				projMiddleware[route] = struct{}{}
			}
		}
		for _, route := range service.Routes {
			routeData := RouteSpecTemplateData{
				GroupName:       newGroupName,
				Method:          strings.ToUpper(route.Method),
				Path:            strings.ToLower(route.Path),
				HandlerName:     helper.CapitalizeFirst(service.Name) + helper.CapitalizeFirst(route.Name),
				HandlerPackName: self.HandlerPackName[strings.ToLower(route.Name+service.Name)],
			}
			group.Routes = append(group.Routes, routeData)
		}
		templateData.Routes = append(templateData.Routes, group)
	}

	var builder strings.Builder
	tmpl, err := template.New("router").Parse(routerContentTemplate)
	if err != nil {
		return err
	}

	if err = tmpl.Execute(&builder, templateData); err != nil {
		return err
	}

	filename := filepath.Join(nowRouterPath, fmt.Sprintf("%s.go", newFileName))
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	console.Echo.Info("正在生成路由文件: ", filename)
	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.formatFileWithGofmt(filename)

	// 检测并创建中间件
	for middlewareName := range projMiddleware {
		if err = os.MkdirAll(middlewarePath, os.ModePerm); err != nil {
			return err
		}

		if exists, _ := helper.FunctionExists(middlewarePath, middlewareName); !exists {
			if err = self.GenMiddleware(middlewarePath, middlewareName); err != nil {
				return err
			}
		}
	}

	// 更新入口文件
	err = self.updateEnterGo(nowRouterPath, fmt.Sprintf("Init%sRouter", templateData.NowGroupName))

	return
}

const enterGoTemplate = `package router

import (
	"github.com/gin-gonic/gin"
	"github.com/soryetong/greasyx/libs/middleware"
	"net/http"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	fs := "/uploads"
	r.StaticFS(fs, http.Dir("./"+fs))

	r.Use(middleware.Cross())
	groups := r.Group("/api")
	{{.InitFunctions}}
	return r
}
`

func (self *HttpGenerator) updateEnterGo(nowRouterPath, newRouter string) (err error) {
	filename := filepath.Join(nowRouterPath, "enter.go")

	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		content := strings.ReplaceAll(enterGoTemplate, "{{.InitFunctions}}", fmt.Sprintf("\t%s(groups)", newRouter))
		err = os.WriteFile(filename, []byte(content), 0644)

		return
	}

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	if strings.Contains(string(fileContent), fmt.Sprintf("%s(groups)", newRouter)) {
		return nil
	}

	newContent := strings.Replace(string(fileContent), "return r", fmt.Sprintf("\t%s(groups)\n\treturn r", newRouter), 1)
	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		return
	}

	self.formatFileWithGofmt(filename)

	return err
}
