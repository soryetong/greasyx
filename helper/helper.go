package helper

import (
	"crypto/md5"
	"encoding/hex"
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
