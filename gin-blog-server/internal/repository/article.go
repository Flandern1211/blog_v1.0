package repository

import (
	"context"
	"gin-blog/internal/model/entity"
	"strconv"

	global "gin-blog/internal/global"
	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	// 后台
	GetList(page, size int, title string, categoryId, tagId, artType, status int, isDelete *bool) ([]entity.Article, int64, error)
	GetById(id int) (*entity.Article, error)
	SaveOrUpdate(article *entity.Article, categoryName string, tagNames []string) error
	UpdateTop(id int, isTop bool) error
	SoftDelete(ids []int, isDelete bool) error
	Delete(ids []int) error

	// 前台
	GetBlogArticle(id int) (*entity.Article, error)
	GetBlogArticleList(page, size, categoryId, tagId int) ([]entity.Article, int64, error)
	GetRecommendList(id, n int) ([]entity.RecommendArticleVO, error)
	GetLastArticle(id int) (entity.ArticlePaginationVO, error)
	GetNextArticle(id int) (entity.ArticlePaginationVO, error)
	GetNewestList(n int) ([]entity.RecommendArticleVO, error)
	ImportArticle(userAuthId int, title, content, img, categoryName, tagName string) error

	// Category 后台
	GetCategoryList(page, size int, keyword string) ([]entity.CategoryVO, int64, error)
	SaveOrUpdateCategory(id int, name string) error
	DeleteCategories(ids []int) error
	GetCategoryOption() ([]entity.Category, error)

	// Tag 后台
	GetTagList(page, size int, keyword string) ([]entity.TagVO, int64, error)
	SaveOrUpdateTag(id int, name string) error
	DeleteTags(ids []int) error
	GetTagOption() ([]entity.Tag, error)

	// Count operations
	GetCategoryCount(ctx context.Context) (int64, error)
	GetTagCount(ctx context.Context) (int64, error)

	// Search
	SearchArticles(ctx context.Context, keyword string) ([]entity.Article, error)

	// Redis operations
	GetArticleLikeCount(ctx context.Context, id int) (int, error)
	GetArticleViewCount(ctx context.Context, id int) (int, error)
	IncrArticleView(ctx context.Context, id int) error
	LikeArticle(ctx context.Context, authId, articleId int) error
}

type articleRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewArticleRepository(db *gorm.DB, rdb *redis.Client) ArticleRepository {
	return &articleRepository{db: db, rdb: rdb}
}

// Article implementations
func (r *articleRepository) GetList(page, size int, title string, categoryId, tagId, artType, status int, isDelete *bool) ([]entity.Article, int64, error) {
	var list []entity.Article
	var total int64
	query := r.db.Model(&entity.Article{}).Preload("Category").Preload("Tags")

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if categoryId != 0 {
		query = query.Where("category_id = ?", categoryId)
	}
	if tagId != 0 {
		query = query.Joins("JOIN article_tag ON article.id = article_tag.article_id").Where("article_tag.tag_id = ?", tagId)
	}
	if artType != 0 {
		query = query.Where("type = ?", artType)
	}
	if status != 0 {
		query = query.Where("status = ?", status)
	}
	if isDelete != nil {
		query = query.Where("is_delete = ?", *isDelete)
	}

	err := query.Count(&total).Scopes(Paginate(page, size)).Order("is_top DESC, id DESC").Find(&list).Error
	return list, total, err
}

func (r *articleRepository) GetById(id int) (*entity.Article, error) {
	var article entity.Article
	err := r.db.Preload("Category").Preload("Tags").First(&article, id).Error
	return &article, err
}

func (r *articleRepository) SaveOrUpdate(article *entity.Article, categoryName string, tagNames []string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if categoryName != "" {
			var category entity.Category
			if err := tx.Where("name = ?", categoryName).First(&category).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					category.Name = categoryName
					if err := tx.Create(&category).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			}
			article.CategoryId = category.ID
		}

		if article.ID == 0 {
			if err := tx.Create(article).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(article).Updates(article).Error; err != nil {
				return err
			}
			if err := tx.Delete(&entity.ArticleTag{}, "article_id = ?", article.ID).Error; err != nil {
				return err
			}
		}

		if len(tagNames) > 0 {
			var tags []entity.Tag
			for _, name := range tagNames {
				var tag entity.Tag
				if err := tx.Where("name = ?", name).First(&tag).Error; err != nil {
					if err == gorm.ErrRecordNotFound {
						tag.Name = name
						if err := tx.Create(&tag).Error; err != nil {
							return err
						}
					} else {
						return err
					}
				}
				tags = append(tags, tag)
			}
			var articleTags []entity.ArticleTag
			for _, tag := range tags {
				articleTags = append(articleTags, entity.ArticleTag{ArticleId: article.ID, TagId: tag.ID})
			}
			if err := tx.Create(&articleTags).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *articleRepository) UpdateTop(id int, isTop bool) error {
	return r.db.Model(&entity.Article{Model: entity.Model{ID: id}}).Update("is_top", isTop).Error
}

func (r *articleRepository) SoftDelete(ids []int, isDelete bool) error {
	return r.db.Model(&entity.Article{}).Where("id IN ?", ids).Update("is_delete", isDelete).Error
}

func (r *articleRepository) Delete(ids []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id IN ?", ids).Delete(&entity.ArticleTag{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entity.Article{}, ids).Error
	})
}

func (r *articleRepository) GetBlogArticle(id int) (*entity.Article, error) {
	var data entity.Article
	result := r.db.Preload("Category").Preload("Tags").
		Where(entity.Article{Model: entity.Model{ID: id}}).
		Where("is_delete = 0 AND status = 1").
		First(&data)
	return &data, result.Error
}

func (r *articleRepository) GetBlogArticleList(page, size, categoryId, tagId int) ([]entity.Article, int64, error) {
	var data []entity.Article
	var total int64
	query := r.db.Model(&entity.Article{}).Where("is_delete = 0 AND status = 1")

	if categoryId != 0 {
		query = query.Where("category_id = ?", categoryId)
	}
	if tagId != 0 {
		query = query.Where("id IN (SELECT article_id FROM article_tag WHERE tag_id = ?)", tagId)
	}

	query.Count(&total)
	result := query.Preload("Tags").Preload("Category").
		Order("is_top DESC, id DESC").
		Scopes(Paginate(page, size)).
		Find(&data)

	return data, total, result.Error
}

func (r *articleRepository) GetRecommendList(id, n int) ([]entity.RecommendArticleVO, error) {
	var list []entity.RecommendArticleVO
	sub1 := r.db.Table("article_tag").Select("tag_id").Where("article_id = ?", id)
	sub2 := r.db.Table("(?) t1", sub1).
		Select("DISTINCT article_id").
		Joins("JOIN article_tag t ON t.tag_id = t1.tag_id").
		Where("article_id != ?", id)
	result := r.db.Table("(?) t2", sub2).
		Select("id, title, img, created_at").
		Joins("JOIN article a ON t2.article_id = a.id").
		Where("a.is_delete = 0 AND a.status = 1").
		Order("is_top DESC, id DESC").
		Limit(n).
		Find(&list)
	return list, result.Error
}

func (r *articleRepository) GetLastArticle(id int) (entity.ArticlePaginationVO, error) {
	var val entity.ArticlePaginationVO
	sub := r.db.Table("article").Select("max(id)").Where("id < ?", id)
	result := r.db.Table("article").
		Select("id, title, img").
		Where("is_delete = 0 AND status = 1 AND id = (?)", sub).
		Limit(1).
		Find(&val)
	return val, result.Error
}

func (r *articleRepository) GetNextArticle(id int) (entity.ArticlePaginationVO, error) {
	var data entity.ArticlePaginationVO
	result := r.db.Model(&entity.Article{}).
		Select("id, title, img").
		Where("is_delete = 0 AND status = 1 AND id > ?", id).
		Limit(1).
		Find(&data)
	return data, result.Error
}

func (r *articleRepository) GetNewestList(n int) ([]entity.RecommendArticleVO, error) {
	var data []entity.RecommendArticleVO
	result := r.db.Model(&entity.Article{}).
		Select("id, title, img, created_at").
		Where("is_delete = 0 AND status = 1").
		Order("created_at DESC, id ASC").
		Limit(n).
		Find(&data)
	return data, result.Error
}

func (r *articleRepository) ImportArticle(userAuthId int, title, content, img, categoryName, tagName string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		article := entity.Article{
			Title:   title,
			Content: content,
			Img:     img,
			Status:  entity.ARTICLE_STATUS_DRAFT,
			Type:    entity.ARTICLE_TYPE_ORIGINAL,
			UserId:  userAuthId,
		}

		var category entity.Category
		if err := tx.Where("name = ?", categoryName).FirstOrCreate(&category, entity.Category{Name: categoryName}).Error; err != nil {
			return err
		}
		article.CategoryId = category.ID

		if err := tx.Create(&article).Error; err != nil {
			return err
		}

		var tag entity.Tag
		if err := tx.Where("name = ?", tagName).FirstOrCreate(&tag, entity.Tag{Name: tagName}).Error; err != nil {
			return err
		}

		return tx.Create(&entity.ArticleTag{
			ArticleId: article.ID,
			TagId:     tag.ID,
		}).Error
	})
}

// Category implementations
func (r *articleRepository) GetCategoryList(page, size int, keyword string) ([]entity.CategoryVO, int64, error) {
	var list []entity.CategoryVO
	var total int64

	query := r.db.Table("category c").
		Select("c.id", "c.name", "COUNT(a.id) AS article_count", "c.created_at", "c.updated_at").
		Joins("LEFT JOIN article a ON c.id = a.category_id AND a.is_delete = 0 AND a.status = 1")

	if keyword != "" {
		query = query.Where("c.name LIKE ?", "%"+keyword+"%")
	}

	err := query.Group("c.id").
		Order("c.updated_at DESC").
		Count(&total).
		Scopes(Paginate(page, size)).
		Find(&list).Error

	return list, total, err
}

func (r *articleRepository) SaveOrUpdateCategory(id int, name string) error {
	if id == 0 {
		return r.db.Create(&entity.Category{Name: name}).Error
	}
	return r.db.Model(&entity.Category{Model: entity.Model{ID: id}}).Update("name", name).Error
}

func (r *articleRepository) DeleteCategories(ids []int) error {
	return r.db.Delete(&entity.Category{}, ids).Error
}

func (r *articleRepository) GetCategoryOption() ([]entity.Category, error) {
	var list []entity.Category
	err := r.db.Model(&entity.Category{}).Select("id", "name").Find(&list).Error
	return list, err
}

// Tag implementations
func (r *articleRepository) GetTagList(page, size int, keyword string) ([]entity.TagVO, int64, error) {
	var list []entity.TagVO
	var total int64

	query := r.db.Table("tag t").
		Joins("LEFT JOIN article_tag at ON t.id = at.tag_id").
		Select("t.id", "t.name", "COUNT(at.article_id) AS article_count", "t.created_at", "t.updated_at")

	if keyword != "" {
		query = query.Where("t.name LIKE ?", "%"+keyword+"%")
	}

	err := query.Group("t.id").
		Order("t.updated_at DESC").
		Count(&total).
		Scopes(Paginate(page, size)).
		Find(&list).Error

	return list, total, err
}

func (r *articleRepository) SaveOrUpdateTag(id int, name string) error {
	if id == 0 {
		return r.db.Create(&entity.Tag{Name: name}).Error
	}
	return r.db.Model(&entity.Tag{Model: entity.Model{ID: id}}).Update("name", name).Error
}

func (r *articleRepository) DeleteTags(ids []int) error {
	return r.db.Delete(&entity.Tag{}, ids).Error
}

func (r *articleRepository) GetTagOption() ([]entity.Tag, error) {
	var list []entity.Tag
	err := r.db.Model(&entity.Tag{}).Select("id", "name").Find(&list).Error
	return list, err
}

// Redis operations

func (r *articleRepository) GetCategoryCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Category{}).Count(&count).Error
	return count, err
}

func (r *articleRepository) GetTagCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Tag{}).Count(&count).Error
	return count, err
}

func (r *articleRepository) SearchArticles(ctx context.Context, keyword string) ([]entity.Article, error) {
	var articles []entity.Article
	err := r.db.Where("is_delete = ? AND status = ? AND (title LIKE ? OR content LIKE ?)",
		false, entity.ARTICLE_STATUS_PUBLIC, "%"+keyword+"%", "%"+keyword+"%").Find(&articles).Error
	return articles, err
}

func (r *articleRepository) GetArticleLikeCount(ctx context.Context, id int) (int, error) {
	return r.rdb.HGet(ctx, global.ARTICLE_LIKE_COUNT, strconv.Itoa(id)).Int()
}

func (r *articleRepository) GetArticleViewCount(ctx context.Context, id int) (int, error) {
	val, err := r.rdb.ZScore(ctx, global.ARTICLE_VIEW_COUNT, strconv.Itoa(id)).Result()
	return int(val), err
}

func (r *articleRepository) IncrArticleView(ctx context.Context, id int) error {
	return r.rdb.ZIncrBy(ctx, global.ARTICLE_VIEW_COUNT, 1, strconv.Itoa(id)).Err()
}

func (r *articleRepository) LikeArticle(ctx context.Context, authId, articleId int) error {
	likeKey := global.ARTICLE_USER_LIKE_SET + strconv.Itoa(authId)
	if r.rdb.SIsMember(ctx, likeKey, articleId).Val() {
		r.rdb.SRem(ctx, likeKey, articleId)
		r.rdb.HIncrBy(ctx, global.ARTICLE_LIKE_COUNT, strconv.Itoa(articleId), -1)
	} else {
		r.rdb.SAdd(ctx, likeKey, articleId)
		r.rdb.HIncrBy(ctx, global.ARTICLE_LIKE_COUNT, strconv.Itoa(articleId), 1)
	}
	return nil
}
