package httpgenerator

import (
	"fmt"
	"path/filepath"
	"os"
	"regexp"
	"strings"
)

// FormatStruct formats a TypesStructSpec into a string representation.
func (self *HttpGenerator) formatStruct(s *TypesStructSpec) string {
	structDef := fmt.Sprintf("type %s struct {\n", s.Name)
	for _, field := range s.Fields {
		tag := ""
		if field.Tag != "" {
			tag = fmt.Sprintf(" `%s`", field.Tag)
		}

		structDef += fmt.Sprintf("\t%s %s%s\n", field.Name, field.Type, tag)

	}
	structDef += "}\n\n"

	return structDef
}

// writeToFile writes or updates a Go struct to the specified file.
func (self *HttpGenerator) writeToTypeFile(filename, structContent string) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	// Extract the struct name using a regular expression.
	structNameRegex := regexp.MustCompile(`^type\s+(\w+)\s+struct`)
	matches := structNameRegex.FindStringSubmatch(structContent)
	if len(matches) < 2 {
		return fmt.Errorf("struct name not found in struct definition")
	}
	structName := matches[1]

	var existingContent []byte
	if content, err := os.ReadFile(filename); err == nil {
		existingContent = content
	}

	// Replace the existing struct definition with the new one.
	structRegex := regexp.MustCompile(fmt.Sprintf(`(?ms)^type %s struct \{.*?\}\n\n`, structName))
	if structRegex.Match(existingContent) {
		existingContent = structRegex.ReplaceAll(existingContent, []byte(structContent))
	} else {
		if len(existingContent) == 0 {
			existingContent = []byte(fmt.Sprintf("package %s\n\n", self.TypesPackageName))
		}
		existingContent = append(existingContent, []byte(structContent)...)
	}

	return os.WriteFile(filename, existingContent, 0644)
}

func (self *HttpGenerator) GenTypes() (err error) {
	typePath := filepath.Join(self.Output, self.TypesPackageName)
	self.TypesPackagePath = filepath.Join(self.ModuleName, typePath)
	for _, s := range self.Types {
		content := self.formatStruct(s)
		var filename string
		if strings.HasSuffix(s.Name, "Req") {
			filename = filepath.Join(typePath, "request.go")
		} else if strings.HasSuffix(s.Name, "Resp") {
			filename = filepath.Join(typePath, "response.go")
		} else {
			filename = filepath.Join(typePath, "common.go")
		}

		if err = self.writeToTypeFile(filename, content); err != nil {
			return err
		}
		self.formatFileWithGofmt(filename)
	}

	return nil
}
