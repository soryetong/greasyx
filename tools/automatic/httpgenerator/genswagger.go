package httpgenerator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

type SchemaField struct {
	Name     string
	Type     string
	JSONName string
	Desc     string
}

type SchemaStruct struct {
	Name     string
	Fields   []SchemaField
	Required []string
}

type ServiceSpecSwagger struct {
	Name      string
	Summary   string
	GroupName string
	Routes    []RouteSpec
	Structs   map[string]SchemaStruct
}

func parseStructs(dir string) (map[string]SchemaStruct, error) {
	structs := make(map[string]SchemaStruct)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		for _, decl := range node.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				var fields []SchemaField
				var requiredArr []string
				for _, field := range st.Fields.List {
					tag := ""
					desc := ""
					if field.Tag != nil {
						tag = reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("json")
						desc = reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("description")
						bindingStr := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Get("binding")
						if strings.Contains(bindingStr, "required") {
							requiredArr = append(requiredArr, tag)
						}
					}
					for _, name := range field.Names {
						fields = append(fields, SchemaField{
							Name:     name.Name,
							Type:     exprToType(field.Type),
							JSONName: tag,
							Desc:     desc,
						})
					}
				}

				structs[ts.Name.Name] = SchemaStruct{
					Name:     ts.Name.Name,
					Fields:   fields,
					Required: requiredArr,
				}
			}
		}
		return nil
	})
	return structs, err
}

func exprToType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return "[]" + exprToType(t.Elt)
	case *ast.StarExpr:
		return exprToType(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return "object"
	}
}

// --- 模板 ---
const openapiTemplate = `
openapi: 3.0.0
info:
  title: {{ .Summary }}
  version: "1.0"
paths:
{{- range .Routes }}
  /{{ $.GroupName }}/{{ .Path }}:
    {{ .Method }}:
      summary: {{ .Summary }}
      {{- if .RequestType }}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .RequestType }}'
      {{- end }}
      responses:
        '200':
          description: OK
          {{- if .ResponseType }}
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{ .ResponseType }}'
          {{- end }}
{{- end }}

components:
  schemas:
{{- range $name, $schema := .Structs }}
    {{ $name }}:
      type: object
      properties:
      {{- range $schema.Fields }}
        {{ .JSONName }}:
          type: {{ TypeToOpenAPI .Type }}
          description: "{{ .Desc }}"
      {{- end }}
      required:
	  {{- range $schema.Required }}
        - {{ . }}
	  {{- end }}
{{- end }}
`

func typeToOpenAPI(goType string) string {
	switch goType {
	case "int", "int64":
		return "integer"
	case "float64", "float32":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	default:
		if strings.HasPrefix(goType, "[]") {
			return "array"
		}
		return "object"
	}
}

// --- 模板渲染 ---
func generateOpenAPIDoc(service ServiceSpecSwagger, output string) error {
	funcMap := template.FuncMap{
		"TypeToOpenAPI": typeToOpenAPI,
	}
	tmpl := template.Must(template.New("openapi").Funcs(funcMap).Parse(openapiTemplate))
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, service)
}

func (self *HttpGenerator) GenSwagger() (err error) {
	groupName := self.FileName
	structs, err := parseStructs("./" + self.Output + "/types")
	if err != nil {
		return err
	}

	service := ServiceSpecSwagger{
		Name:      self.Services[0].Name,
		Summary:   self.Services[0].Summary,
		GroupName: groupName,
		Structs:   structs,
	}
	for _, spec := range self.Services {
		for _, route := range spec.Routes {
			service.Routes = append(service.Routes, RouteSpec{
				Method:       route.Method,
				Name:         route.Name,
				Path:         route.Path,
				RequestType:  route.RequestType,
				ResponseType: route.ResponseType,
				RustFulKey:   route.RustFulKey,
				Summary:      route.Summary,
			})
		}
	}

	outputArr := strings.Split(self.Output, "/")
	swaggerDir := strings.Join(outputArr[:len(outputArr)-1], "/") + "/swagger"
	if err = os.MkdirAll(swaggerDir, os.ModePerm); err != nil {
		fmt.Println("创建目录失败：", err)
		return err
	}

	err = generateOpenAPIDoc(service, fmt.Sprintf("%s/api_%s.yaml", swaggerDir, groupName))

	return
}
