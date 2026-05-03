package repository

import (
	"context"
	"gin-blog/internal/model/entity"
	"strconv"
	"strings"

	global "gin-blog/internal/global"
	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

type BlogInfoRepository interface {
	// Config
	GetConfigMap() (map[string]string, error)
	UpdateConfigMap(m map[string]string) error
	GetConfig(key string) (string, error)
	GetConfigBool(key string) bool
	GetConfigInt(key string) int
	UpdateConfig(key, value string) error

	// Page
	GetPageList() ([]entity.Page, int64, error)
	SaveOrUpdatePage(page *entity.Page) error
	DeletePages(ids []int) error

	// Redis operations
	GetViewCount(ctx context.Context) (int64, error)
	IncrViewCount(ctx context.Context) error
	IsUniqueVisitor(ctx context.Context, uuid string) (bool, error)
	AddUniqueVisitor(ctx context.Context, uuid string) error
	IncrVisitorArea(ctx context.Context, ipSource string)
	GetBlogStats(ctx context.Context) (int64, int64, int64, error)
	GetAllConfig(ctx context.Context) (map[string]string, error)
}

type blogInfoRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewBlogInfoRepository(db *gorm.DB, rdb *redis.Client) BlogInfoRepository {
	return &blogInfoRepository{db: db, rdb: rdb}
}

// Config implementations
func (r *blogInfoRepository) GetConfigMap() (map[string]string, error) {
	var configs []entity.Config
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, config := range configs {
		m[config.Key] = config.Value
	}
	return m, nil
}

func (r *blogInfoRepository) UpdateConfigMap(m map[string]string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for k, v := range m {
			if err := tx.Model(&entity.Config{}).Where("`key` = ?", k).Update("value", v).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *blogInfoRepository) GetConfig(key string) (string, error) {
	var config entity.Config
	if err := r.db.Where("`key` = ?", key).First(&config).Error; err != nil {
		return "", err
	}
	return config.Value, nil
}

func (r *blogInfoRepository) GetConfigBool(key string) bool {
	val, err := r.GetConfig(key)
	if err != nil {
		return false
	}
	return val == "true"
}

func (r *blogInfoRepository) GetConfigInt(key string) int {
	val, err := r.GetConfig(key)
	if err != nil {
		return 0
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}
	return result
}

func (r *blogInfoRepository) UpdateConfig(key, value string) error {
	return r.db.Where(&entity.Config{Key: key}).Assign(&entity.Config{Value: value}).FirstOrCreate(&entity.Config{}).Error
}

// Page implementations
func (r *blogInfoRepository) GetPageList() ([]entity.Page, int64, error) {
	var pages []entity.Page
	var total int64
	if err := r.db.Model(&entity.Page{}).Count(&total).Find(&pages).Error; err != nil {
		return nil, 0, err
	}
	return pages, total, nil
}

func (r *blogInfoRepository) SaveOrUpdatePage(page *entity.Page) error {
	if page.ID > 0 {
		return r.db.Updates(page).Error
	}
	return r.db.Create(page).Error
}

func (r *blogInfoRepository) DeletePages(ids []int) error {
	return r.db.Delete(&entity.Page{}, ids).Error
}

// Redis operations

func (r *blogInfoRepository) GetViewCount(ctx context.Context) (int64, error) {
	val, err := r.rdb.Get(ctx, global.VIEW_COUNT).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *blogInfoRepository) IncrViewCount(ctx context.Context) error {
	_, err := r.rdb.Incr(ctx, global.VIEW_COUNT).Result()
	return err
}

func (r *blogInfoRepository) IsUniqueVisitor(ctx context.Context, uuid string) (bool, error) {
	return r.rdb.SIsMember(ctx, global.KEY_UNIQUE_VISITOR_SET, uuid).Result()
}

func (r *blogInfoRepository) AddUniqueVisitor(ctx context.Context, uuid string) error {
	return r.rdb.SAdd(ctx, global.KEY_UNIQUE_VISITOR_SET, uuid).Err()
}

func (r *blogInfoRepository) IncrVisitorArea(ctx context.Context, ipSource string) {
	address := strings.Split(ipSource, "|")
	if len(address) > 2 {
		province := strings.ReplaceAll(address[2], "省", "")
		r.rdb.HIncrBy(ctx, global.VISITOR_AREA, province, 1)
	} else {
		r.rdb.HIncrBy(ctx, global.VISITOR_AREA, "未知", 1)
	}
}

func (r *blogInfoRepository) GetBlogStats(ctx context.Context) (int64, int64, int64, error) {
	var articleCount int64
	if err := r.db.Model(&entity.Article{}).Where("status = ? AND is_delete = ?", entity.ARTICLE_STATUS_PUBLIC, false).Count(&articleCount).Error; err != nil {
		return 0, 0, 0, err
	}

	var userCount int64
	if err := r.db.Model(&entity.UserInfo{}).Count(&userCount).Error; err != nil {
		return 0, 0, 0, err
	}

	var messageCount int64
	if err := r.db.Table("message").Count(&messageCount).Error; err != nil {
		return 0, 0, 0, err
	}

	return articleCount, userCount, messageCount, nil
}

func (r *blogInfoRepository) GetAllConfig(ctx context.Context) (map[string]string, error) {
	var configs []entity.Config
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, conf := range configs {
		m[conf.Key] = conf.Value
	}
	return m, nil
}
