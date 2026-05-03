package service

import (
	"context"
	global "gin-blog/internal/global"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"
	"gin-blog/internal/utils"
)

type BlogInfoService interface {
	// BlogInfo
	GetHomeInfo(ctx context.Context) (response.BlogHomeVO, error)
	GetAbout(ctx context.Context) (string, error)
	UpdateAbout(ctx context.Context, req request.AboutReq) error
	Report(ctx context.Context, ipAddress, userAgent string) error

	// Config
	GetConfigMap(ctx context.Context) (map[string]string, error)
	UpdateConfigMap(ctx context.Context, m map[string]string) error

	// Page
	GetPageList(ctx context.Context) ([]entity.Page, int64, error)
	SaveOrUpdatePage(ctx context.Context, req request.AddOrEditPageReq) (*entity.Page, error)
	DeletePages(ctx context.Context, ids []int) error
}

type blogInfoService struct {
	repo repository.BlogInfoRepository
}

func NewBlogInfoService(repo repository.BlogInfoRepository) BlogInfoService {
	return &blogInfoService{repo: repo}
}

// BlogInfo implementations
func (s *blogInfoService) GetHomeInfo(ctx context.Context) (response.BlogHomeVO, error) {
	articleCount, userCount, messageCount, err := s.repo.GetBlogStats(ctx)
	if err != nil {
		return response.BlogHomeVO{}, err
	}

	viewCount, err := s.repo.GetViewCount(ctx)
	if err != nil {
		return response.BlogHomeVO{}, err
	}

	return response.BlogHomeVO{
		ArticleCount: int(articleCount),
		UserCount:    int(userCount),
		MessageCount: int(messageCount),
		ViewCount:    int(viewCount),
	}, nil
}

func (s *blogInfoService) GetAbout(ctx context.Context) (string, error) {
	return s.repo.GetConfig(global.CONFIG_ABOUT)
}

func (s *blogInfoService) UpdateAbout(ctx context.Context, req request.AboutReq) error {
	return s.repo.UpdateConfig(global.CONFIG_ABOUT, req.Content)
}

func (s *blogInfoService) Report(ctx context.Context, ipAddress, userAgent string) error {
	var uuid string
	if userAgent != "" {
		uuid = utils.MD5(ipAddress + userAgent)
	} else {
		uuid = utils.MD5(ipAddress)
	}

	isNew, _ := s.repo.IsUniqueVisitor(ctx, uuid)
	if !isNew {
		ipSource := utils.IP.GetIpSource(ipAddress)
		s.repo.IncrVisitorArea(ctx, ipSource)
		s.repo.IncrViewCount(ctx)
		s.repo.AddUniqueVisitor(ctx, uuid)
	}

	return nil
}

// Config implementations
func (s *blogInfoService) GetConfigMap(ctx context.Context) (map[string]string, error) {
	return s.repo.GetConfigMap()
}

func (s *blogInfoService) UpdateConfigMap(ctx context.Context, m map[string]string) error {
	return s.repo.UpdateConfigMap(m)
}

// Page implementations
func (s *blogInfoService) GetPageList(ctx context.Context) ([]entity.Page, int64, error) {
	return s.repo.GetPageList()
}

func (s *blogInfoService) SaveOrUpdatePage(ctx context.Context, req request.AddOrEditPageReq) (*entity.Page, error) {
	page := &entity.Page{
		Model: entity.Model{ID: req.ID},
		Name:  req.Name,
		Label: req.Label,
		Cover: req.Cover,
	}
	err := s.repo.SaveOrUpdatePage(page)
	return page, err
}

func (s *blogInfoService) DeletePages(ctx context.Context, ids []int) error {
	return s.repo.DeletePages(ids)
}
