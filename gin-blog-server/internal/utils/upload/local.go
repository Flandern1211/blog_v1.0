package upload

import (
	"errors"
	g "gin-blog/internal/global"
	"gin-blog/internal/utils"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// 本地文件上传
type Local struct{}

// 文件上传到本地
func (*Local) UploadFile(file *multipart.FileHeader) (filePath, fileName string, err error) {
	ext := path.Ext(file.Filename)                                     // 读取文件后缀
	name := strings.TrimSuffix(file.Filename, ext)                     // 读取文件名
	name = utils.MD5(name)                                             // 加密文件名
	filename := name + "_" + time.Now().Format("20060102150405") + ext // 拼接新文件名

	conf := g.Conf.Upload
	mkdirErr := os.MkdirAll(conf.StorePath, os.ModePerm) // 创建存储路径
	if mkdirErr != nil {
		slog.Error("function os.MkdirAll() Filed", slog.Any("err", mkdirErr.Error()))
		return "", "", errors.New("function os.MkdirAll() Filed, err:" + mkdirErr.Error())
	}

	storePath := conf.StorePath + "/" + filename // 文件存储路径
	filepath := conf.Path + "/" + filename       // 文件展示路径

	f, openError := file.Open() // 读取文件
	if openError != nil {
		slog.Error("function file.Open() Filed", slog.String("err", openError.Error()))
		return "", "", errors.New("function file.Open() Filed, err:" + openError.Error())
	}
	defer func(f multipart.File) {
		closeErr := f.Close()
		if closeErr != nil {
			slog.Error("function file.Close() Filed", slog.String("err", openError.Error()))
			return
		}
	}(f) // 创建文件 defer 关闭

	out, createErr := os.Create(storePath)
	if createErr != nil {
		slog.Error("function os.Create() Filed", slog.String("err", createErr.Error()))
		return "", "", errors.New("function os.Create() Filed, err:" + createErr.Error())
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			slog.Error("function file.Close() Filed", slog.String("err", openError.Error()))
			return
		}
	}(out)

	_, copyErr := io.Copy(out, f) // 拷贝文件
	if copyErr != nil {
		slog.Error("function io.Copy() Filed", slog.String("err", copyErr.Error()))
		return "", "", errors.New("function io.Copy() Filed, err:" + copyErr.Error())
	}
	return filepath, filename, nil
}

// 从本地删除文件
func (*Local) DeleteFile(key string) error {
	conf := g.Conf.Upload
	storePath := conf.StorePath
	p := storePath + "/" + key
	//防止../等会跳出存储路径的非法key
	//filepath.Clean会清除路径中的非法字符
	if !strings.HasPrefix(filepath.Clean(p), filepath.Clean(conf.StorePath)) {
		return errors.New("非法文件路径")
	} else {
		if err := os.Remove(p); err != nil {
			return errors.New("本地文件删除失败, err:" + err.Error())
		}
	}
	return nil
}
