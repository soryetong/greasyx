package utils

import (
	"encoding/json"
	"gopkg.in/natefinch/lumberjack.v2"
	"net"
	"os"
)

type tool struct {
}

func Tool() *tool {
	return &tool{}
}

func (self *tool) GetLocalIP() string {
	addrList, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrList {
		if ip, ok := addr.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP.String()
			}
		}
	}

	return ""
}

func (self *tool) LoadJsonConfig(_filename string, _config interface{}) (err error) {
	f, err := os.Open(_filename)
	if err == nil {
		defer f.Close()
		var fileInfo os.FileInfo
		fileInfo, err = f.Stat()
		if err == nil {
			bytes := make([]byte, fileInfo.Size())
			_, err = f.Read(bytes)
			if err == nil {
				BOM := []byte{0xEF, 0xBB, 0xBF} // remove windows text file BOM
				if bytes[0] == BOM[0] && bytes[1] == BOM[1] && bytes[2] == BOM[2] {
					bytes = bytes[3:]
				}
				err = json.Unmarshal(bytes, _config)
			}
		}
	}

	return
}

func (self *tool) NewLumberjack(logFileName string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   logFileName,
		MaxSize:    2,    // 单文件最大容量, 单位是MB
		MaxBackups: 3,    // 最大保留过期文件个数
		MaxAge:     1,    // 保留过期文件的最大时间间隔, 单位是天
		Compress:   true, // 是否需要压缩滚动日志, 使用的gzip压缩
		LocalTime:  true, // 是否使用计算机的本地时间, 默认UTC
	}
}
