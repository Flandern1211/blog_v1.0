package repository

import (
	"context"
	"gin-blog/internal/model/entity"
	"strconv"

	global "gin-blog/internal/global"
	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

type InteractionRepository interface {
	// Message
	GetMessageList(page, size int, nickname string, isReview *bool) ([]entity.Message, int64, error)
	DeleteMessages(ids []int) error
	UpdateMessagesReview(ids []int, isReview bool) error
	SaveMessage(message *entity.Message) error

	// Comment
	GetCommentList(page, size, typ int, isReview *bool, nickname string) ([]entity.Comment, int64, error)
	DeleteComments(ids []int) error
	UpdateCommentsReview(ids []int, isReview bool) error
	GetFrontCommentList(page, size, topic, typ int) ([]entity.Comment, map[int][]entity.Comment, int64, error)
	GetCommentReplyList(id, page, size int) ([]entity.Comment, error)
	AddComment(comment *entity.Comment) error
	GetCommentById(id int) (*entity.Comment, error)
	GetArticleCommentCount(articleId int) (int64, error)

	// Redis operations
	LikeComment(ctx context.Context, authId, commentId int) error
	GetCommentLikeCountMap(ctx context.Context) (map[string]string, error)
}

type interactionRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewInteractionRepository(db *gorm.DB, rdb *redis.Client) InteractionRepository {
	return &interactionRepository{db: db, rdb: rdb}
}

// Message implementations
func (r *interactionRepository) GetMessageList(page, size int, nickname string, isReview *bool) ([]entity.Message, int64, error) {
	var list []entity.Message
	var total int64
	query := r.db.Model(&entity.Message{})

	if nickname != "" {
		query = query.Where("nickname LIKE ?", "%"+nickname+"%")
	}
	if isReview != nil {
		query = query.Where("is_review = ?", *isReview)
	}

	err := query.Count(&total).Order("created_at DESC").Scopes(Paginate(page, size)).Find(&list).Error
	return list, total, err
}

func (r *interactionRepository) DeleteMessages(ids []int) error {
	return r.db.Where("id IN ?", ids).Delete(&entity.Message{}).Error
}

func (r *interactionRepository) UpdateMessagesReview(ids []int, isReview bool) error {
	return r.db.Model(&entity.Message{}).Where("id IN ?", ids).Update("is_review", isReview).Error
}

func (r *interactionRepository) SaveMessage(message *entity.Message) error {
	return r.db.Create(message).Error
}

// Comment implementations
func (r *interactionRepository) GetCommentList(page, size, typ int, isReview *bool, nickname string) ([]entity.Comment, int64, error) {
	var list []entity.Comment
	var total int64
	query := r.db.Model(&entity.Comment{})

	if nickname != "" {
		var uid []int
		r.db.Model(&entity.UserInfo{}).Where("nickname LIKE ?", "%"+nickname+"%").Pluck("id", &uid)
		if len(uid) > 0 {
			query = query.Where("user_id IN ?", uid)
		} else {
			query = query.Where("user_id = ?", 0)
		}
	}

	if typ != 0 {
		query = query.Where("type = ?", typ)
	}
	if isReview != nil {
		query = query.Where("is_review = ?", *isReview)
	}

	err := query.Count(&total).
		Preload("User").Preload("User.UserInfo").
		Preload("ReplyUser").Preload("ReplyUser.UserInfo").
		Preload("Article").
		Order("id DESC").
		Scopes(Paginate(page, size)).
		Find(&list).Error
	return list, total, err
}

func (r *interactionRepository) DeleteComments(ids []int) error {
	return r.db.Where("id IN ?", ids).Delete(&entity.Comment{}).Error
}

func (r *interactionRepository) UpdateCommentsReview(ids []int, isReview bool) error {
	return r.db.Model(&entity.Comment{}).Where("id IN ?", ids).Update("is_review", isReview).Error
}

func (r *interactionRepository) GetFrontCommentList(page, size, topic, typ int) ([]entity.Comment, map[int][]entity.Comment, int64, error) {
	var list []entity.Comment
	var total int64

	tx := r.db.Model(&entity.Comment{})
	if typ != 0 {
		tx = tx.Where("type = ?", typ)
	}
	if topic != 0 {
		tx = tx.Where("topic_id = ?", topic)
	}

	err := tx.Where("parent_id = 0").
		Count(&total).
		Preload("User").Preload("User.UserInfo").
		Order("id DESC").
		Scopes(Paginate(page, size)).
		Find(&list).Error

	if err != nil {
		return nil, nil, 0, err
	}

	replyMap := make(map[int][]entity.Comment)
	for i := range list {
		var replyList []entity.Comment
		r.db.Model(&entity.Comment{}).
			Where("parent_id = ?", list[i].ID).
			Preload("User").Preload("User.UserInfo").
			Preload("ReplyUser").Preload("ReplyUser.UserInfo").
			Order("id DESC").
			Find(&replyList)
		replyMap[list[i].ID] = replyList
	}

	return list, replyMap, total, nil
}

func (r *interactionRepository) GetCommentReplyList(id, page, size int) ([]entity.Comment, error) {
	var data []entity.Comment
	err := r.db.Model(&entity.Comment{}).
		Where("parent_id = ?", id).
		Preload("User").Preload("User.UserInfo").
		Preload("ReplyUser").Preload("ReplyUser.UserInfo").
		Order("id DESC").
		Scopes(Paginate(page, size)).
		Find(&data).Error
	return data, err
}

func (r *interactionRepository) AddComment(comment *entity.Comment) error {
	return r.db.Create(comment).Error
}

func (r *interactionRepository) GetCommentById(id int) (*entity.Comment, error) {
	var comment entity.Comment
	err := r.db.Preload("User").Preload("User.UserInfo").
		Preload("ReplyUser").Preload("ReplyUser.UserInfo").
		Preload("Article").
		First(&comment, id).Error
	return &comment, err
}

func (r *interactionRepository) GetArticleCommentCount(articleId int) (int64, error) {
	var count int64
	err := r.db.Model(&entity.Comment{}).
		Where("topic_id = ? AND type = 1 AND is_review = 1", articleId).
		Count(&count).Error
	return count, err
}

// Redis operations

func (r *interactionRepository) LikeComment(ctx context.Context, authId, commentId int) error {
	likeKey := global.COMMENT_USER_LIKE_SET + strconv.Itoa(authId)
	if r.rdb.SIsMember(ctx, likeKey, commentId).Val() {
		r.rdb.SRem(ctx, likeKey, commentId)
		r.rdb.HIncrBy(ctx, global.COMMENT_LIKE_COUNT, strconv.Itoa(commentId), -1)
	} else {
		r.rdb.SAdd(ctx, likeKey, commentId)
		r.rdb.HIncrBy(ctx, global.COMMENT_LIKE_COUNT, strconv.Itoa(commentId), 1)
	}
	return nil
}

func (r *interactionRepository) GetCommentLikeCountMap(ctx context.Context) (map[string]string, error) {
	return r.rdb.HGetAll(ctx, global.COMMENT_LIKE_COUNT).Result()
}
