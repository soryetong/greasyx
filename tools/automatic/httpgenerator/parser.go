package httpgenerator

import (
	"fmt"
	"regexp"
	"strings"
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
	serviceRegex := regexp.MustCompile(`service\s+(\w+)(?:\s+Use\s+([\w,]+))?\s*{([^}]*)}`)
	serviceMatches := serviceRegex.FindAllStringSubmatch(content, -1)
	for _, serviceMatch := range serviceMatches {
		var service ServiceSpec
		service.Name = serviceMatch[1]
		service.Middleware = serviceMatch[2]
		routesBlock := serviceMatch[3]

		routeRegex := regexp.MustCompile(`(\w+)\s+([\w/:]+)(?::(\w+))?\s*(?:\(([\[\]\w]+)\))?\s*returns\s*(?:\(([\[\]\w]+)\))?`)
		routeMatches := routeRegex.FindAllStringSubmatch(routesBlock, -1)
		for _, routeMatch := range routeMatches {
			nameArr := strings.Split(routeMatch[2], "/")
			nameVal := routeMatch[2]
			rustFulVal := routeMatch[3]
			if len(nameArr) > 1 {
				nameVal = nameArr[0]
				rustFulVal = strings.Trim(nameArr[1], ":")
			}
			service.Routes = append(service.Routes, &RouteSpec{
				Method:       routeMatch[1],
				Path:         routeMatch[2],
				Name:         nameVal,
				RustFulKey:   rustFulVal,
				RequestType:  routeMatch[4],
				ResponseType: routeMatch[5],
			})
		}

		services = append(services, &service)
	}

	self.Services = services

	return nil
}
