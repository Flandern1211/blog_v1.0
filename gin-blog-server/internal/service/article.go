package service

import (
	"context"
	"errors"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"

	"gorm.io/gorm"

	bizErr "gin-blog/pkg/errors"
)

type ArticleService interface {
	// Article
	GetList(ctx context.Context, query request.ArticleQuery) ([]response.ArticleVO, int64, error)
	GetById(ctx context.Context, id int) (*response.ArticleVO, error)
	SaveOrUpdate(ctx context.Context, authId int, req request.AddOrEditArticleReq) error
	UpdateTop(ctx context.Context, req request.UpdateArticleTopReq) error
	SoftDelete(ctx context.Context, req request.SoftDeleteReq) error
	Delete(ctx context.Context, ids []int) error

	// Front-end specific Article methods
	GetBlogArticleList(ctx context.Context, query request.FArticleQuery) ([]response.ArticleVO, int64, error)
	GetBlogArticle(ctx context.Context, id int) (*response.BlogArticleVO, error)

	// Category
	GetCategoryList(ctx context.Context, query request.CategoryQuery) ([]response.CategoryVO, int64, error)
	SaveOrUpdateCategory(ctx context.Context, req request.AddOrEditCategoryReq) error
	DeleteCategories(ctx context.Context, ids []int) error
	GetCategoryOption(ctx context.Context) ([]response.OptionVO, error)

	// Tag
	GetTagList(ctx context.Context, query request.TagQuery) ([]response.TagVO, int64, error)
	SaveOrUpdateTag(ctx context.Context, req request.AddOrEditTagReq) error
	DeleteTags(ctx context.Context, ids []int) error
	GetTagOption(ctx context.Context) ([]response.OptionVO, error)
}

type articleService struct {
	repo         repository.ArticleRepository
	interactRepo repository.InteractionRepository
}

func NewArticleService(repo repository.ArticleRepository, interactRepo repository.InteractionRepository) ArticleService {
	return &articleService{
		repo:         repo,
		interactRepo: interactRepo,
	}
}

// Article implementations
func (s *articleService) GetList(ctx context.Context, query request.ArticleQuery) ([]response.ArticleVO, int64, error) {
	list, total, err := s.repo.GetList(query.GetPage(), query.GetSize(), query.Title, query.CategoryId, query.TagId, query.Type, query.Status, query.IsDelete)
	if err != nil {
		return nil, 0, err
	}

	var res []response.ArticleVO
	for _, art := range list {
		vo := response.ArticleVO{Article: art}
		likeCount, _ := s.repo.GetArticleLikeCount(ctx, art.ID)
		viewCount, _ := s.repo.GetArticleViewCount(ctx, art.ID)
		vo.LikeCount = likeCount
		vo.ViewCount = viewCount
		res = append(res, vo)
	}
	return res, total, nil
}

func (s *articleService) GetById(ctx context.Context, id int) (*response.ArticleVO, error) {
	art, err := s.repo.GetById(id)
	if err != nil {
		return nil, err
	}
	vo := &response.ArticleVO{Article: *art}
	likeCount, _ := s.repo.GetArticleLikeCount(ctx, art.ID)
	viewCount, _ := s.repo.GetArticleViewCount(ctx, art.ID)
	vo.LikeCount = likeCount
	vo.ViewCount = viewCount
	return vo, nil
}

func (s *articleService) SaveOrUpdate(ctx context.Context, authId int, req request.AddOrEditArticleReq) error {
	article := &entity.Article{
		Model:       entity.Model{ID: req.ID},
		Title:       req.Title,
		Desc:        req.Desc,
		Content:     req.Content,
		Img:         req.Img,
		Type:        req.Type,
		Status:      req.Status,
		IsTop:       req.IsTop,
		OriginalUrl: req.OriginalUrl,
		UserId:      authId,
	}
	return s.repo.SaveOrUpdate(article, req.CategoryName, req.TagNames)
}

func (s *articleService) UpdateTop(ctx context.Context, req request.UpdateArticleTopReq) error {
	return s.repo.UpdateTop(req.ID, req.IsTop)
}

func (s *articleService) SoftDelete(ctx context.Context, req request.SoftDeleteReq) error {
	return s.repo.SoftDelete(req.Ids, req.IsDelete)
}

func (s *articleService) Delete(ctx context.Context, ids []int) error {
	return s.repo.Delete(ids)
}

func (s *articleService) entityToVO(article entity.Article) response.ArticleVO {
	return response.ArticleVO{
		Article: article,
	}
}

func (s *articleService) GetBlogArticleList(ctx context.Context, query request.FArticleQuery) ([]response.ArticleVO, int64, error) {
	list, total, err := s.repo.GetBlogArticleList(query.GetPage(), query.GetSize(), query.CategoryId, query.TagId)
	if err != nil {
		return nil, 0, err
	}
	var voList []response.ArticleVO
	for _, article := range list {
		voList = append(voList, s.entityToVO(article))
	}
	return voList, total, nil
}

func (s *articleService) GetBlogArticle(ctx context.Context, id int) (*response.BlogArticleVO, error) {
	article, err := s.repo.GetBlogArticle(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr.ErrNotFound
		}
		return nil, err
	}

	vo := &response.BlogArticleVO{
		Article:           *article,
		RecommendArticles: make([]response.RecommendArticleVO, 0),
		NewestArticles:    make([]response.RecommendArticleVO, 0),
	}

	recommendList, _ := s.repo.GetRecommendList(id, 6)
	for _, v := range recommendList {
		vo.RecommendArticles = append(vo.RecommendArticles, response.RecommendArticleVO{
			ID:        v.ID,
			Img:       v.Img,
			Title:     v.Title,
			CreatedAt: v.CreatedAt,
		})
	}

	newestList, _ := s.repo.GetNewestList(5)
	for _, v := range newestList {
		vo.NewestArticles = append(vo.NewestArticles, response.RecommendArticleVO{
			ID:        v.ID,
			Img:       v.Img,
			Title:     v.Title,
			CreatedAt: v.CreatedAt,
		})
	}

	lastArt, _ := s.repo.GetLastArticle(id)
	vo.LastArticle = response.ArticlePaginationVO{
		ID:    lastArt.ID,
		Img:   lastArt.Img,
		Title: lastArt.Title,
	}

	nextArt, _ := s.repo.GetNextArticle(id)
	vo.NextArticle = response.ArticlePaginationVO{
		ID:    nextArt.ID,
		Img:   nextArt.Img,
		Title: nextArt.Title,
	}

	s.repo.IncrArticleView(ctx, id)

	viewCnt, _ := s.repo.GetArticleViewCount(ctx, id)
	vo.ViewCount = int64(viewCnt)
	likeCount, _ := s.repo.GetArticleLikeCount(ctx, id)
	vo.LikeCount = int64(likeCount)
	vo.CommentCount, _ = s.interactRepo.GetArticleCommentCount(id)

	return vo, nil
}

// Category implementations
func (s *articleService) GetCategoryList(ctx context.Context, query request.CategoryQuery) ([]response.CategoryVO, int64, error) {
	list, total, err := s.repo.GetCategoryList(query.GetPage(), query.GetSize(), query.Keyword)
	if err != nil {
		return nil, 0, err
	}
	var res []response.CategoryVO
	for _, cat := range list {
		res = append(res, response.CategoryVO{
			Category:     cat.Category,
			ArticleCount: cat.ArticleCount,
		})
	}
	return res, total, nil
}

func (s *articleService) SaveOrUpdateCategory(ctx context.Context, req request.AddOrEditCategoryReq) error {
	return s.repo.SaveOrUpdateCategory(req.ID, req.Name)
}

func (s *articleService) DeleteCategories(ctx context.Context, ids []int) error {
	return s.repo.DeleteCategories(ids)
}

func (s *articleService) GetCategoryOption(ctx context.Context) ([]response.OptionVO, error) {
	list, err := s.repo.GetCategoryOption()
	if err != nil {
		return nil, err
	}
	var res []response.OptionVO
	for _, cat := range list {
		res = append(res, response.OptionVO{ID: cat.ID, Label: cat.Name})
	}
	return res, nil
}

// Tag implementations
func (s *articleService) GetTagList(ctx context.Context, query request.TagQuery) ([]response.TagVO, int64, error) {
	list, total, err := s.repo.GetTagList(query.GetPage(), query.GetSize(), query.Keyword)
	if err != nil {
		return nil, 0, err
	}
	var res []response.TagVO
	for _, tag := range list {
		res = append(res, response.TagVO{
			Tag:          tag.Tag,
			ArticleCount: tag.ArticleCount,
		})
	}
	return res, total, nil
}

func (s *articleService) SaveOrUpdateTag(ctx context.Context, req request.AddOrEditTagReq) error {
	return s.repo.SaveOrUpdateTag(req.ID, req.Name)
}

func (s *articleService) DeleteTags(ctx context.Context, ids []int) error {
	return s.repo.DeleteTags(ids)
}

func (s *articleService) GetTagOption(ctx context.Context) ([]response.OptionVO, error) {
	list, err := s.repo.GetTagOption()
	if err != nil {
		return nil, err
	}
	var res []response.OptionVO
	for _, tag := range list {
		res = append(res, response.OptionVO{ID: tag.ID, Label: tag.Name})
	}
	return res, nil
}
