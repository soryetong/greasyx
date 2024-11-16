package utils

import (
	"crypto/md5"
	"encoding/hex"
)

type helper struct {
}

func Helper() *helper {
	return &helper{}
}

// ValidatePasswd 校验密码是否一致
func (self *helper) ValidatePasswd(pwd, salt, passwd string) bool {
	return self.Md5Encode(pwd+salt) == passwd
}

// MakePasswd 生成密码
func (self *helper) MakePasswd(pwd, salt string) string {
	return self.Md5Encode(pwd + salt)
}

// Md5Encode md5处理
func (self *helper) Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}
