package httpgenerator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
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
func (self *HttpGenerator) writeToTypeFile2(filename, structContent string) error {
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

// writeToTypeFile writes or updates a Go struct in the specified file.
func (self *HttpGenerator) writeToTypeFile(filename, structContent string) error {
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	structNameRegex := regexp.MustCompile(`(?m)^type\s+(\w+)\s+struct\b`)
	matches := structNameRegex.FindStringSubmatch(structContent)
	if len(matches) < 2 {
		return fmt.Errorf("struct name not found in struct definition")
	}
	structName := matches[1]
	existingContent, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	existingCode := string(existingContent)
	structRegex := regexp.MustCompile(fmt.Sprintf(`(?ms)^type\s+%s\s+struct\s*\{.*?\}\s*(?:\n|$)`, structName))
	cleanedContent := structRegex.ReplaceAllString(existingCode, "")

	var buffer bytes.Buffer
	if !strings.Contains(existingCode, "package ") {
		buffer.WriteString(fmt.Sprintf("package %s\n\n", self.TypesPackageName))
	}

	if strings.TrimSpace(cleanedContent) != "" {
		buffer.WriteString(strings.TrimSpace(cleanedContent))
		buffer.WriteString("\n\n")
	}

	buffer.WriteString(strings.TrimSpace(structContent))
	buffer.WriteString("\n")

	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("failed to format Go code: %w\nGenerated Code:\n%s", err, buffer.String())
	}

	return os.WriteFile(filename, formattedCode, 0644)
}

func (self *HttpGenerator) GenTypes() (err error) {
	typePath := filepath.Join(self.Output, self.TypesPackageName)
	self.TypesPackagePath = filepath.Join(self.ModuleName, typePath)
	var typeFilePrefix string
	if self.Domain != "" {
		typeFilePrefix = fmt.Sprintf("%s_", self.Domain)
	}
	for _, s := range self.Types {
		content := self.formatStruct(s)
		var filename string
		if strings.HasPrefix(s.Name, "Common") {
			filename = filepath.Join(typePath, "common.go")
		} else if strings.HasSuffix(s.Name, "Req") {
			filename = filepath.Join(typePath, fmt.Sprintf("%s%s", typeFilePrefix, "request.go"))
		} else if strings.HasSuffix(s.Name, "Resp") {
			filename = filepath.Join(typePath, fmt.Sprintf("%s%s", typeFilePrefix, "response.go"))
		} else {
			filename = filepath.Join(typePath, "common.go")
		}

		if err = self.writeToTypeFile(filename, content); err != nil {
			return err
		}
	}

	return nil
}
