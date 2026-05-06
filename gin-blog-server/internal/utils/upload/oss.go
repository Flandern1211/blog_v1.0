package upload

import (
	g "gin-blog/internal/global"
	"mime/multipart"
)

// OSS 对象存储接口
type OSS interface {
	UploadFile(file *multipart.FileHeader) (string, string, error)
	DeleteFile(key string) error
}

// 根据配置文件中的配置判断文件上传实例
func NewOSS() OSS {
	switch g.GetConfig().Upload.OssType {
	//暂时只有本地存储方式
	case "local":
		return &Local{}
	default:
		return &Local{}
	}
}
