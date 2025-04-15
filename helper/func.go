package helper

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidatePasswd 校验密码是否一致
func ValidatePasswd(pwd, salt, passwd string) bool {
	return Md5Encode(pwd+salt) == passwd
}

// MakePasswd 生成密码
func MakePasswd(pwd, salt string) string {
	return Md5Encode(pwd + salt)
}

// Md5Encode md5处理
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

type Number interface {
	~int | ~int32 | ~int64 | ~float64 | ~float32 | ~string
}

// IsValidNumber 判断是否是有效数字
func IsValidNumber[T Number](value T) bool {
	switch v := any(value).(type) {
	case int:
		return v > 0
	case int32:
		return v > 0
	case int64:
		return v > 0
	case float64:
		return v > 0
	case float32:
		return v > 0
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num > 0
		}
	}
	return false
}

// GetCallerName 获取调用者的名称
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

// SeparateCamel 按照自定符号分隔驼峰
func SeparateCamel(name, separator string) string {
	var result strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteString(separator)
		}

		result.WriteRune(r | ' ')
	}
	return result.String()
}

// CapitalizeFirst 字符串首字母大写
func CapitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func GetRequestPath(path, prefix string) (uri string, id int64) {
	uri = strings.TrimPrefix(path, prefix)
	re := regexp.MustCompile(`^(.*)/(\d+)$`)
	matches := re.FindStringSubmatch(uri)
	if len(matches) == 3 {
		uri = matches[1]
		id = StringToInt64(matches[2])
	}

	return
}

func ConvertToRestfulURL(url string) string {
	re := regexp.MustCompile(`(^.+?/[^/]+)/\d+$`)
	return re.ReplaceAllString(url, `$1/:id`)
}

// GetMapValue 获取map的值
func GetMapValue[T any](m map[string]interface{}, key string) T {
	var zero T

	value, exists := m[key]
	if exists {
		v := reflect.ValueOf(value)
		if v.Type().ConvertibleTo(reflect.TypeOf(zero)) {
			return v.Convert(reflect.TypeOf(zero)).Interface().(T)
		}
	}

	return zero
}

type MapSupportedTypes interface {
	string | int64 | float64 | bool
}

// GetMapSpecificValue 获取map的特定类型的值, 相较于GetMapValue, 不用每次反射获取值
func GetMapSpecificValue[T MapSupportedTypes](m map[string]interface{}, key string) T {
	var zero T

	value, exists := m[key]
	if exists {
		if v, ok := value.(T); ok {
			return v
		}
	}

	var result any
	switch v := value.(type) {
	case float64:
		if _, ok := any(zero).(int64); ok {
			result = int64(v)
		} else if _, ok := any(zero).(bool); ok {
			result = v != 0
		} else {
			return zero
		}
	case int64:
		if _, ok := any(zero).(float64); ok {
			result = float64(v)
		} else if _, ok := any(zero).(bool); ok {
			result = v != 0
		} else {
			return zero
		}
	case string:
		if _, ok := any(zero).(bool); ok {
			lowerVal := strings.ToLower(v)
			if lowerVal == "true" || lowerVal == "1" {
				result = true
			} else if lowerVal == "false" || lowerVal == "0" {
				result = false
			} else {
				return zero
			}
		} else {
			result = v
		}
	case bool:
		if _, ok := any(zero).(string); ok {
			result = fmt.Sprintf("%v", v) // 转成 "true" / "false"
		} else {
			result = v
		}
	default:
		return zero
	}

	if finalValue, ok := result.(T); ok {
		return finalValue
	}

	return zero
}
