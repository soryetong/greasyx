package helper

import (
	"math/rand"
	"strings"
	"time"
	"unsafe"

	uuid "github.com/satori/go.uuid"
)

// 生成uuid
func GenerateUuid() string {
	return uuid.NewV4().String()
}

// 生成不带横杠的32位uuid
func GenerateNoWhippletreeUuid() string {
	uuidStr := uuid.NewV4().String()
	uuidStr = strings.ReplaceAll(uuidStr, "-", "")

	return uuidStr
}

var src = rand.NewSource(time.Now().UnixNano())

const (
	letters = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// 6 bits to represent a letter index
	letterIdBits = 6
	// All 1-bits as many as letterIdBits
	letterIdMask = 1<<letterIdBits - 1
	letterIdMax  = 63 / letterIdBits
)

func RandString(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdMax letters!
	for i, cache, remain := n-1, src.Int63(), letterIdMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdMax
		}
		if idx := int(cache & letterIdMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func Interval64(min, max int64) int64 {
	if min == max {
		return min
	}

	if min < 0 {
		min = 0
	}

	if min > max {
		min, max = max, min
	}

	return rand.Int63n(max-min) + min
}
