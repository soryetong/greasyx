package ginaoss

import (
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/soryetong/greasyx/ginahelper"
)

type local struct{}

func (*local) Upload(file multipart.File, fileHeader *multipart.FileHeader, uploadDir ...string) (*UploadRet, error) {
	// 获取上传目录和文件名
	savePathUri, filename := getUploadDirAndFilename(fileHeader, uploadDir...)

	// 检查并创建上传目录
	isPath, mkdirErr := ginahelper.FileIsExist(savePathUri)
	if mkdirErr != nil {
		return nil, mkdirErr
	}
	if !isPath {
		_ = os.MkdirAll(savePathUri, os.ModePerm)
	}

	// 保存文件
	filePath := filepath.Join(savePathUri, filename)
	if err := ginahelper.SaveFile(fileHeader, filePath); err != nil {
		return nil, err
	}

	return &UploadRet{
		Hash:     ginahelper.Md5Encode(filePath),
		Filename: filename,
		Url:      ginahelper.AssembleServerPath(filePath),
	}, nil
}
