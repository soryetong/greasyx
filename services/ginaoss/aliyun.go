package ginaoss

import "mime/multipart"

type aliYun struct{}

func (*aliYun) Upload(file *multipart.FileHeader, savePath ...string) (string, string, error) {
	return "", "", nil
}
