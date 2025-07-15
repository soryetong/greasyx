package ginaoss

import (
	"fmt"
	"math/rand"
	"mime/multipart"
	"path"
	"strings"
	"time"

	"github.com/soryetong/greasyx/ginahelper"
	"github.com/spf13/viper"
)

type ossType string

const (
	OssTypeLocal  ossType = "local"
	OssTypeQiNiu  ossType = "qiniu"
	OssTypeAliYun ossType = "aliyun"
)

type UploadRet struct {
	Filename string `json:"filename"`
	Hash     string `json:"hash"`
	Url      string `json:"url"`
	Key      string `json:"key"`
}

type oss interface {
	Upload(file multipart.File, fileHeader *multipart.FileHeader, uploadDir ...string) (*UploadRet, error)
}

func New(ossType ossType) oss {
	return start(ossType)
}

func NewAliYun() oss {
	return start(OssTypeAliYun)
}

func NewQiNiu() oss {
	return start(OssTypeQiNiu)
}

func NewLocal() oss {
	return start(OssTypeLocal)
}

func NewByConf() oss {
	return start(ossType(viper.GetString("Oss.Type")))
}

func start(ossType ossType) oss {
	switch ossType {
	//case OssTypeAliYun:
	//	return &aliYun{}
	case OssTypeQiNiu:
		return &qiNiu{}
	default:
		return &local{}
	}
}

func getUploadDirAndFilename(fileHeader *multipart.FileHeader, uploadDir ...string) (string, string) {
	version := time.Now().Format("20060102")
	var dir string
	if len(uploadDir) > 0 {
		dir = uploadDir[0]
	} else {
		dir = viper.GetString("Oss.SavePath")
	}
	if dir == "" {
		dir = "./static/uploads/attach/"
	}
	dir = strings.TrimRight(dir, "/") + "/"

	// 读取文件后缀
	ext := path.Ext(fileHeader.Filename)
	// 读取文件名并加密
	name := ginahelper.Md5Encode(fileHeader.Filename + fmt.Sprintf("%d%04d", time.Now().Unix(), rand.Int31()))
	// 拼接新文件名
	filename := name + ext

	return dir + version + "/", filename
}
