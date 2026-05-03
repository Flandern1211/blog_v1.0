package service

import (
	"context"
	"gin-blog/internal/utils/upload"
	"mime/multipart"
)

type UploadService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error)
}

type uploadService struct{}

func NewUploadService() UploadService {
	return &uploadService{}
}

func (s *uploadService) UploadFile(ctx context.Context, file *multipart.FileHeader) (string, error) {
	oss := upload.NewOSS()
	url, _, err := oss.UploadFile(file)
	if err != nil {
		return "", err
	}
	return url, nil
}
