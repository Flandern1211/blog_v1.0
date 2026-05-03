package service

import (
	"context"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/repository"
	"strings"
)

type FrontService interface {
	GetHomeInfo(ctx context.Context) (response.FrontHomeVO, error)
	LikeArticle(ctx context.Context, articleId int, authId int) error
	LikeComment(ctx context.Context, commentId int, authId int) error
	SearchArticle(ctx context.Context, keyword string) ([]response.ArticleSearchVO, error)
}

type frontService struct {
	articleRepo  repository.ArticleRepository
	blogInfoRepo repository.BlogInfoRepository
	interactRepo repository.InteractionRepository
}

func NewFrontService(articleRepo repository.ArticleRepository, blogInfoRepo repository.BlogInfoRepository, interactRepo repository.InteractionRepository) FrontService {
	return &frontService{
		articleRepo:  articleRepo,
		blogInfoRepo: blogInfoRepo,
		interactRepo: interactRepo,
	}
}

func (s *frontService) GetHomeInfo(ctx context.Context) (response.FrontHomeVO, error) {
	var data response.FrontHomeVO

	articleCount, userCount, messageCount, err := s.blogInfoRepo.GetBlogStats(ctx)
	if err != nil {
		return data, err
	}

	categoryCount, _ := s.articleRepo.GetCategoryCount(ctx)
	tagCount, _ := s.articleRepo.GetTagCount(ctx)

	configMap, err := s.blogInfoRepo.GetAllConfig(ctx)
	if err == nil {
		data.Config = configMap
	}

	viewCount, err := s.blogInfoRepo.GetViewCount(ctx)
	if err != nil {
		return data, err
	}

	return response.FrontHomeVO{
		ArticleCount:  articleCount,
		UserCount:     userCount,
		MessageCount:  messageCount,
		CategoryCount: categoryCount,
		TagCount:      tagCount,
		Config:        configMap,
		ViewCount:     int64(viewCount),
	}, nil
}

func (s *frontService) LikeArticle(ctx context.Context, articleId int, authId int) error {
	return s.articleRepo.LikeArticle(ctx, authId, articleId)
}

func (s *frontService) LikeComment(ctx context.Context, commentId int, authId int) error {
	return s.interactRepo.LikeComment(ctx, authId, commentId)
}

func (s *frontService) SearchArticle(ctx context.Context, keyword string) ([]response.ArticleSearchVO, error) {
	var result []response.ArticleSearchVO
	if keyword == "" {
		return result, nil
	}

	articles, err := s.articleRepo.SearchArticles(ctx, keyword)
	if err != nil {
		return nil, err
	}

	for _, article := range articles {
		title := strings.ReplaceAll(article.Title, keyword, "<span style='color:#f47466'>"+keyword+"</span>")
		content := s.buildSearchContent(article.Content, keyword)
		result = append(result, response.ArticleSearchVO{
			ID:      article.ID,
			Title:   title,
			Content: content,
		})
	}

	return result, nil
}

func (s *frontService) buildSearchContent(content, keyword string) string {
	keywordStartIndex := unicodeIndex(content, keyword)
	if keywordStartIndex == -1 {
		return content
	}

	preIndex := 0
	if keywordStartIndex > 25 {
		preIndex = keywordStartIndex - 25
	}
	preText := substring(content, preIndex, keywordStartIndex)

	keywordEndIndex := keywordStartIndex + unicodeLen(keyword)
	afterLength := len([]rune(content)) - keywordEndIndex
	afterIndex := keywordEndIndex + 175
	if afterIndex > keywordEndIndex+afterLength {
		afterIndex = keywordEndIndex + afterLength
	}
	afterText := substring(content, keywordStartIndex, afterIndex)

	return strings.ReplaceAll(preText+afterText, keyword, "<span style='color:#f47466'>"+keyword+"</span>")
}

func unicodeIndex(str, substr string) int {
	result := strings.Index(str, substr)
	if result > 0 {
		prefix := []byte(str)[0:result]
		rs := []rune(string(prefix))
		result = len(rs)
	}
	return result
}

func unicodeLen(str string) int {
	var r = []rune(str)
	return len(r)
}

func substring(source string, start int, end int) string {
	var unicodeStr = []rune(source)
	length := len(unicodeStr)
	if start >= end {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start <= 0 && end >= length {
		return source
	}
	var substring = ""
	for i := start; i < end; i++ {
		substring += string(unicodeStr[i])
	}
	return substring
}
