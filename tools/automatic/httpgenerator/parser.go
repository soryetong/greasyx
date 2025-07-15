package httpgenerator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/soryetong/greasyx/ginahelper"
)

// PStruct parses the given content and returns a slice of TypesStructSpec structs.
func (self *HttpGenerator) PTypesStruct(content string) (err error) {
	var structs []*TypesStructSpec
	structRegex := regexp.MustCompile(`type\s+(\w+)\s*{([^}]*)}`)
	matches := structRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		name := match[1]
		fieldsBlock := match[2]
		fields := []*FieldSpec{}

		lines := strings.Split(fieldsBlock, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line) // Remove any leading/trailing whitespace
			if line == "" {
				continue
			}
			field, err := self.parseFieldDeclaration(line)
			if err != nil {
				continue
			}

			fields = append(fields, field)
		}

		structs = append(structs, &TypesStructSpec{
			Name:   name,
			Fields: fields,
		})
	}

	self.Types = structs

	return nil
}

// parseFieldDeclaration parses a single field declaration and returns its name, type, and tag.
func (self *HttpGenerator) parseFieldDeclaration(declaration string) (*FieldSpec, error) {
	regexPattern := `^\s*(?:(\w+)\s+)?([\w\*\[\]]+)(?:\s+` + "`" + `([^` + "`" + `]*)` + "`" + `)?\s*$`
	re := regexp.MustCompile(regexPattern)

	matches := re.FindStringSubmatch(declaration)
	if matches == nil || len(matches) < 3 {
		return nil, fmt.Errorf("invalid field declaration format: %s", declaration)
	}

	return &FieldSpec{
		Name: matches[1],
		Type: matches[2],
		Tag:  matches[3],
	}, nil
}

func (self *HttpGenerator) PRoutesService(content string) (err error) {
	var services []*ServiceSpec
	// serviceRegex := regexp.MustCompile(`service\s+(\w+)\s*{([^}]*)}`)
	// serviceRegex := regexp.MustCompile(`service\s+(\w+)(?:\s+Group\s+([\w,]+))?\s*{([^}]*)}`)
	serviceRegex := regexp.MustCompile(`(?m)(?:\s*@Summary\s+([^\n\r]+))?\s*service\s+(\w+)(?:\s+Group\s+([\w,]+))?\s*{([^}]*)}`)
	serviceMatches := serviceRegex.FindAllStringSubmatch(content, -1)
	for _, serviceMatch := range serviceMatches {
		var service ServiceSpec
		service.Summary = serviceMatch[1]
		service.Name = serviceMatch[2]
		service.Group = serviceMatch[3]
		routesBlock := serviceMatch[4]
		if service.Summary == "" {
			service.Summary = service.Name
		}

		routeLineRegex := regexp.MustCompile(`(\w+)\s+([\w/:]+)(?::(\w+))?\s*(?:\(([\[\]\w]+)\))?\s*returns\s*(?:\(([\[\]\w]+)\))?`)
		summaryRegex := regexp.MustCompile(`@Summary\s+(.+)`)
		lines := strings.Split(routesBlock, "\n")
		var lastSummary string

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "@Summary") {
				if matches := summaryRegex.FindStringSubmatch(line); len(matches) == 2 {
					lastSummary = matches[1]
				}
				continue
			}

			if routeMatch := routeLineRegex.FindStringSubmatch(line); len(routeMatch) > 0 {
				uriVal, rustFulVal := ginahelper.ConvertRestfulURLToUri(routeMatch[2])
				nameArr := strings.Split(uriVal, "/")
				for i, s := range nameArr {
					nameArr[i] = ginahelper.UcFirst(s)
				}
				nameVal := strings.Join(nameArr, "")
				if lastSummary == "" {
					lastSummary = ginahelper.UcFirst(service.Name) + nameVal
				}
				service.Routes = append(service.Routes, &RouteSpec{
					Method:       routeMatch[1],
					Path:         routeMatch[2],
					Name:         nameVal,
					RustFulKey:   rustFulVal,
					RequestType:  routeMatch[4],
					ResponseType: routeMatch[5],
					Summary:      lastSummary,
				})

				lastSummary = ""
			}
		}

		services = append(services, &service)
	}

	self.Services = services

	return nil
}
