package oss

import (
	"context"
	"fmt"
	"github.com/flipped-aurora/gf-vue-admin/interfaces"
	"github.com/flipped-aurora/gf-vue-admin/library/global"
	"github.com/pkg/errors"
	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

var _ interfaces.Oss = (*tencent)(nil)

var Tencent = new(tencent)

type tencent struct {
	filename string
}

func NewTencentClient() (*cos.Client, error) {
	_url, err := url.Parse("https://" + global.Config.Tencent.Bucket + ".cos." + global.Config.Tencent.Region + ".myqcloud.com")
	if err != nil {
		return nil, errors.Wrap(err, "url 拼接失败!")
	}
	baseURL := &cos.BaseURL{BucketURL: _url}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  global.Config.Tencent.SecretID,
			SecretKey: global.Config.Tencent.SecretKey,
		},
	})
	return client, nil
}

func (t *tencent) DeleteByKey(key string) error {
	client, err := NewTencentClient()
	if err != nil {
		return err
	}
	name := global.Config.Tencent.PathPrefix + "/" + key
	if _, err = client.Object.Delete(context.Background(), name); err != nil {
		return errors.Wrap(err, "文件删除失败!")
	}
	return nil
}

func (t *tencent) UploadByFile(file multipart.File) (filepath string, filename string, err error) {
	var client *cos.Client
	client, err = NewTencentClient()
	if err != nil {
		return filepath, filename, err
	}

	defer func() {
		if err = file.Close(); err != nil {
			zap.L().Error("文件关闭失败!", zap.Error(err))
		}
	}() // 关闭文件流

	filename = global.Config.Tencent.Filename(t.filename)
	filepath = global.Config.Tencent.Filepath(filename)

	_, err = client.Object.Put(context.Background(), filename, file, nil)
	if err != nil {
		return filepath, t.filename, errors.Wrap(err, "文件上传失败!")
	}
	return filepath, t.filename, nil
}

func (t *tencent) UploadByFileHeader(file *multipart.FileHeader) (filepath string, filename string, err error) {
	var open multipart.File
	open, err = file.Open()
	if err != nil {
		return filepath, filename, err
	}
	t.filename = fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename)
	return t.UploadByFile(open)
}
