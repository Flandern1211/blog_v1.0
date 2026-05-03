package service

import (
	"context"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"
)

type SystemService interface {
	// FriendLink
	GetLinkList(ctx context.Context, query request.FriendLinkQuery) ([]entity.FriendLink, int64, error)
	SaveOrUpdateLink(ctx context.Context, req request.AddOrEditLinkReq) (*entity.FriendLink, error)
	DeleteLinks(ctx context.Context, ids []int) error

	// OperationLog
	GetOperationLogList(ctx context.Context, query request.OperationLogQuery) ([]entity.OperationLog, int64, error)
	DeleteOperationLogs(ctx context.Context, ids []int) error
	CreateOperationLog(log *entity.OperationLog) error
}

type systemService struct {
	repo repository.SystemRepository
}

func NewSystemService(repo repository.SystemRepository) SystemService {
	return &systemService{repo: repo}
}

// FriendLink implementations
func (s *systemService) GetLinkList(ctx context.Context, query request.FriendLinkQuery) ([]entity.FriendLink, int64, error) {
	return s.repo.GetLinkList(query.GetPage(), query.GetSize(), query.Keyword)
}

func (s *systemService) SaveOrUpdateLink(ctx context.Context, req request.AddOrEditLinkReq) (*entity.FriendLink, error) {
	link := &entity.FriendLink{
		Model:   entity.Model{ID: req.ID},
		Name:    req.Name,
		Avatar:  req.Avatar,
		Address: req.Address,
		Intro:   req.Intro,
	}
	err := s.repo.SaveOrUpdateLink(link)
	return link, err
}

func (s *systemService) DeleteLinks(ctx context.Context, ids []int) error {
	return s.repo.DeleteLinks(ids)
}

// OperationLog implementations
func (s *systemService) GetOperationLogList(ctx context.Context, query request.OperationLogQuery) ([]entity.OperationLog, int64, error) {
	return s.repo.GetOperationLogList(query.GetPage(), query.GetSize(), query.Keyword)
}

func (s *systemService) DeleteOperationLogs(ctx context.Context, ids []int) error {
	return s.repo.DeleteOperationLogs(ids)
}

func (s *systemService) CreateOperationLog(log *entity.OperationLog) error {
	return s.repo.CreateOperationLog(log)
}
