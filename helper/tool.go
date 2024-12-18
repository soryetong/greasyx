package helper

import (
	"bytes"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
)

func GetCallerName(caller interface{}) string {
	typ := reflect.TypeOf(caller)

	switch typ.Kind() {
	case reflect.Ptr: // 指针类型
		if typ.Elem().Kind() == reflect.Struct {
			return typ.Elem().Name()
		}
		return fmt.Sprintf("%v", reflect.ValueOf(caller).Elem())
	case reflect.Struct: // 结构体类型
		return typ.Name()
	default: // 其他类型
		return fmt.Sprintf("%v", caller)
	}
}

// GetModuleName retrieves the current project's module name using `go list`.
func GetModuleName() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get module name: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// CamelToSlash converts a camelCase string to a slash-separated string.
func CamelToSlash(name string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteString("/")
		}

		result.WriteRune(r | ' ')
	}
	return result.String()
}
