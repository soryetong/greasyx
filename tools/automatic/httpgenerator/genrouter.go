package httpgenerator

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"html/template"
	"os"
	"path/filepath"
	"strings"
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
	Routes     []RouteSpecTemplateData
}

type RouteSpecTemplateData struct {
	GroupName       string
	Method          string
	Path            string
	HandlerPackName string
	HandlerName     string
}

func (self *HttpGenerator) GenRouter(groupName string) (err error) {
	c := cases.Title(language.English)
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
		NowGroupName:    c.String(groupName),
		HandlerPackPath: handlerPackPath,
		Routes:          []RouteGroupTemplateData{},
	}

	for _, service := range self.Services {
		newGroupName := strings.ToLower(service.Name[:1]) + service.Name[1:]
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
				HandlerName:     c.String(route.Name),
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

	filename := filepath.Join(nowRouterPath, fmt.Sprintf("%s.go", strings.ToLower(groupName)))
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("[GREASYX-TOOLS-INFO] 正在生成路由文件: %s", filename))
	if _, err = file.WriteString(builder.String()); err != nil {
		return err
	}
	self.formatFileWithGofmt(filename)

	// 更新入口文件
	err = self.updateEnterGo(nowRouterPath, fmt.Sprintf("Init%sRouter", c.String(groupName)))

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
