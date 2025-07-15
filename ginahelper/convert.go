package ginahelper

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
)

func StringToInt64(s string) (i int64) {
	i, _ = strconv.ParseInt(s, 10, 64)

	return
}

func Int64ToString(i int64) (s string) {
	s = strconv.FormatInt(i, 10)

	return
}

func InterfaceToString(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

func InterfaceToInt64(inVal interface{}) int64 {
	if inVal == nil {
		return 0
	}

	refValue := reflect.ValueOf(inVal)
	if refValue.Kind() == reflect.Ptr {
		refValue = refValue.Elem()
	}
	refType := reflect.TypeOf(inVal)
	if refType.Kind() == reflect.Ptr {
		refType = refType.Elem()
	}

	switch refType.Kind() {
	case reflect.Bool:
		if refValue.Bool() {
			return 1
		} else {
			return 0
		}
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		return int64(refValue.Uint())
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		return refValue.Int()
	case reflect.Float32, reflect.Float64:
		return int64(refValue.Float())
	case reflect.Complex64, reflect.Complex128:
		retValue, _ := strconv.ParseFloat(strconv.FormatComplex(refValue.Complex(), 'f', -1, 128), 64)
		return int64(retValue)
	}

	// 转换为字符串，在其中找数字
	re := regexp.MustCompile("-?[0-9]+")
	valueString := InterfaceToString(inVal)
	numberList := re.FindAllString(valueString, -1)
	if len(numberList) > 0 {
		rVal, err := strconv.ParseInt(numberList[0], 10, 64)
		if err == nil {
			return rVal
		}
	}

	return 0
}
