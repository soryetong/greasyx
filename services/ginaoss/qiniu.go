package ginaoss

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"

	"github.com/qiniu/go-sdk/v7/storagev2/credentials"
	"github.com/qiniu/go-sdk/v7/storagev2/http_client"
	"github.com/qiniu/go-sdk/v7/storagev2/uploader"
	"github.com/spf13/viper"
)

type qiNiu struct{}

func (*qiNiu) Upload(file multipart.File, fileHeader *multipart.FileHeader, uploadDir ...string) (*UploadRet, error) {
	// 获取上传目录和文件名
	savePathUri, filename := getUploadDirAndFilename(fileHeader, uploadDir...)

	accessKey := viper.GetString("Oss.AccessKey")
	secretKey := viper.GetString("Oss.SecretKey")
	bucket := viper.GetString("Oss.Bucket")
	ossUrl := viper.GetString("Oss.Url")
	if accessKey == "" || secretKey == "" || ossUrl == "" || bucket == "" {
		return nil, errors.New("config has empty value")
	}
	mac := credentials.NewCredentials(accessKey, secretKey)
	uploadManager := uploader.NewUploadManager(&uploader.UploadManagerOptions{
		Options: http_client.Options{
			Credentials: mac,
		},
	})

	ret := new(UploadRet)
	objectName := savePathUri + filename
	err := uploadManager.UploadReader(context.Background(), file, &uploader.ObjectOptions{
		BucketName: bucket,
		ObjectName: &objectName,
		FileName:   filename,
	}, &ret)

	return &UploadRet{
		Hash:     ret.Hash,
		Filename: filename,
		Url:      strings.Trim(ossUrl, "/") + "/" + objectName,
	}, err
}
