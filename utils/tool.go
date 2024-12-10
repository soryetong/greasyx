package utils

import (
	"fmt"
	"reflect"
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
