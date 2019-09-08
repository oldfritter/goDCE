package utils

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"gopkg.in/yaml.v2"
)

var QiniuConfig map[string]string

type MyPutRet struct {
	Key    string
	Hash   string
	Fsize  int
	Bucket string
	Name   string
}

func InitQiniuConfig() {
	path_str, _ := filepath.Abs("config/qiniu.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		fmt.Printf("error (%v)", err)
		return
	}
	yaml.Unmarshal(content, &QiniuConfig)
}

func UploadFileToQiniu(bucket, key, filePath string) error {
	putPolicy := storage.PutPolicy{
		Scope:      fmt.Sprintf("%s:%s", bucket, key),
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"bucket":"$(bucket)","name":"$(x:name)"}`,
	}
	mac := qbox.NewMac(QiniuConfig["access_key"], QiniuConfig["secret_key"])
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	formUploader := storage.NewFormUploader(&cfg)
	ret := MyPutRet{}
	putExtra := storage.PutExtra{
		Params: map[string]string{
			"x:name": "panama logo",
		},
	}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, filePath, &putExtra)
	if err != nil {
		return err
	}
	exec.Command("sh", "-c", "rm -rf "+filePath).Output()
	return nil
}
