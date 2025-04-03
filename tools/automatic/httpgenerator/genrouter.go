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

const routerContentTemplate = `
package router

import (
	"github.com/gin-gonic/gin"
	"{{ .HandlerPackPath}}"
)

func Init{{.NowGroupName}}Router(routerGroup *gin.RouterGroup) {
{{range .Routes}}{{.GroupName}}Group := routerGroup.Group("/{{.RouteGroup}}")
{{"{"}}{{range .Routes}}
	{{.GroupName}}Group.{{.Method}}("/{{.Path}}", {{.HandlerPackName}}.{{.HandlerName}}){{end}}
{{"}"}}{{end}}
}
`

type RouteTemplateData struct {
	NowGroupName    string
	HandlerPackPath string
	Routes          []RouteGroupTemplateData
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

func (self *HttpGenerator) GenRouter(groupName string) (err error) {
	nowRouterPath := filepath.Join(self.Output, "router")
	self.RouterPath = filepath.Join(self.ModuleName, nowRouterPath)
	if err = os.MkdirAll(nowRouterPath, os.ModePerm); err != nil {
		return err
	}

	handlerPackPath := ""
	for _, service := range self.Services {
		handlerPackPath = self.HandlerPackPath[strings.ToLower(service.Name)]
	}

	templateData := RouteTemplateData{
		HandlerPackPath: handlerPackPath,
		Routes:          []RouteGroupTemplateData{},
	}

	newFileName := ""
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

	r.Use(middleware.Cross()){{if .NeedRequestLog}}.Use(middleware.RequestLog()){{end}}
	publicGroup := r.Group("{{ .RouterPrefix}}")
	{
		// 健康监测
		publicGroup.GET("/health", func(c *gin.Context) {
			c.JSON(200, "ok")
		})

		{{ if eq .GroupName "Public" }}{{.InitPublicFunctions}}(publicGroup){{ end }}
	}

	{{ if eq .GroupName "Auth" }}
	privateAuthGroup := r.Group("{{ .RouterPrefix}}")
	privateAuthGroup.Use(middleware.Casbin()).Use(middleware.Jwt())
	{
		{{.InitPrivateAuthFunctions}}(privateAuthGroup)
	}{{ end }}

	{{ if eq .GroupName "Token" }}
	privateTokenGroup := r.Group("{{ .RouterPrefix}}")
	privateTokenGroup.Use(middleware.Jwt())
	{
		{{.InitPrivateTokenFunctions}}(privateTokenGroup)
	}{{ end }}

	return r
}
`

type EnterGoTemplateData struct {
	RouterPrefix              string
	NeedRequestLog            bool
	GroupName                 string
	InitPublicFunctions       string
	InitPrivateAuthFunctions  string
	InitPrivateTokenFunctions string
}

func (self *HttpGenerator) updateEnterGo(nowRouterPath, newRouter string) (err error) {
	var nowGroup string
	for _, service := range self.Services {
		nowGroup = service.Group
	}

	filename := filepath.Join(nowRouterPath, "enter.go")
	_, err = os.Stat(filename)

	templateData := EnterGoTemplateData{
		GroupName:      nowGroup,
		RouterPrefix:   self.RouterPrefix,
		NeedRequestLog: self.NeedRequestLog,
	}
	if os.IsNotExist(err) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}

		defer file.Close()
		switch nowGroup {
		case config.Group_Public:
			templateData.InitPublicFunctions = newRouter
		case config.Group_Auth:
			templateData.InitPrivateAuthFunctions = newRouter
		case config.Group_Token:
			templateData.InitPrivateTokenFunctions = newRouter
		default:
		}

		var builder strings.Builder
		tmpl, err := template.New("routerEnter").Parse(enterGoTemplate)
		if err != nil {
			return err
		}
		if err = tmpl.Execute(&builder, templateData); err != nil {
			return err
		}

		console.Echo.Info("正在初始化路由入口文件: ", filename)
		if _, err = file.WriteString(builder.String()); err != nil {
			return err
		}
		self.formatFileWithGofmt(filename)

		return err
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	groupName := strings.ToLower(nowGroup)
	if nowGroup != config.Group_Public {
		groupName = "private" + nowGroup
	}

	lines := strings.Split(string(content), "\n")
	var newContent []string
	foundGroup := false
	inserted := false
	functionExists := false
	groupStartIndex := -1
	groupEndIndex := -1
	returnIndex := -1

	for i, line := range lines {
		newContent = append(newContent, line)
		if strings.TrimSpace(line) == "return r" {
			returnIndex = i
		}
		if strings.Contains(line, newRouter+"(") {
			functionExists = true
		}
		if strings.Contains(line, groupName+"Group := r.Group(") {
			foundGroup = true
			groupStartIndex = i
		}
		if foundGroup && strings.TrimSpace(line) == "}" {
			groupEndIndex = i
			foundGroup = false
		}
	}

	if groupStartIndex == -1 {
		if returnIndex != -1 {
			newContent = append(newContent[:returnIndex], fmt.Sprintf("\t%sGroup := r.Group(\"%s\")\n", groupName, self.RouterPrefix))
			if nowGroup == config.Group_Auth {
				newContent = append(newContent, "\t"+groupName+"Group.Use(middleware.Casbin()).Use(middleware.Jwt())")
			} else if nowGroup == config.Group_Token {
				newContent = append(newContent, "\t"+groupName+"Group.Use(middleware.Jwt())")
			}
			newContent = append(newContent, "\t{")
			newContent = append(newContent, "\t\t"+newRouter+"("+groupName+"Group)")
			newContent = append(newContent, "\t}\n")
			newContent = append(newContent, lines[returnIndex:]...)
		}
		inserted = true
	} else if !functionExists && groupEndIndex != -1 {
		newContent = append(newContent[:groupEndIndex], "\t\t"+newRouter+"("+groupName+"Group)")
		newContent = append(newContent, lines[groupEndIndex:]...)
		inserted = true
	}

	if inserted {
		// 写回文件
		if err = os.WriteFile(filename, []byte(strings.Join(newContent, "\n")), 0644); err != nil {
			return err
		}
	}

	console.Echo.Info("✅ 已完成路由入口文件更新: ", filename, "  ", nowGroup)
	self.formatFileWithGofmt(filename)

	return err
}
