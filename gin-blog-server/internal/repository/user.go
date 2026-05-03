package repository

import (
	"context"
	"encoding/json"
	"gin-blog/internal/model/entity"
	"strconv"
	"time"

	global "gin-blog/internal/global"
	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetInfoById(id int) (*entity.UserAuth, error)
	UpdateUserInfo(id int, nickname, avatar, intro, website string) error
	UpdateUserPassword(id int, password string) error
	UpdateUserNicknameAndRole(authId int, nickname string, roleIds []int) error
	UpdateUserDisable(id int, isDisable bool) error
	GetList(page, size int, loginType int8, nickname, username string) ([]entity.UserAuth, int64, error)

	// Redis operations
	GetArticleLikeSet(ctx context.Context, authId int) ([]string, error)
	GetCommentLikeSet(ctx context.Context, authId int) ([]string, error)
	GetOnlineUsers(ctx context.Context, keyword string) ([]*entity.UserAuth, error)
	SetOnlineUser(ctx context.Context, auth *entity.UserAuth, ttl time.Duration) error
	DelOnlineUser(ctx context.Context, userId int) error
	SetOfflineMark(ctx context.Context, userId int, ttl time.Duration) error
	CheckOffline(ctx context.Context, userId int) (bool, error)
}

type userRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewUserRepository(db *gorm.DB, rdb *redis.Client) UserRepository {
	return &userRepository{db: db, rdb: rdb}
}

func (r *userRepository) GetInfoById(id int) (*entity.UserAuth, error) {
	var userAuth entity.UserAuth
	err := r.db.Model(&userAuth).
		Preload("Roles").Preload("UserInfo").
		First(&userAuth, id).Error
	return &userAuth, err
}

func (r *userRepository) UpdateUserInfo(id int, nickname, avatar, intro, website string) error {
	return r.db.Model(&entity.UserInfo{Model: entity.Model{ID: id}}).
		Select("nickname", "avatar", "intro", "website").
		Updates(entity.UserInfo{
			Nickname: nickname,
			Avatar:   avatar,
			Intro:    intro,
			Website:  website,
		}).Error
}

func (r *userRepository) UpdateUserPassword(id int, password string) error {
	return r.db.Model(&entity.UserAuth{Model: entity.Model{ID: id}}).
		Update("password", password).Error
}

func (r *userRepository) UpdateUserNicknameAndRole(authId int, nickname string, roleIds []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var userAuth entity.UserAuth
		if err := tx.First(&userAuth, authId).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.UserInfo{Model: entity.Model{ID: userAuth.UserInfoId}}).
			Update("nickname", nickname).Error; err != nil {
			return err
		}

		if len(roleIds) > 0 {
			if err := tx.Delete(&entity.UserAuthRole{}, "user_auth_id = ?", authId).Error; err != nil {
				return err
			}
			var userRoles []entity.UserAuthRole
			for _, rid := range roleIds {
				userRoles = append(userRoles, entity.UserAuthRole{
					UserAuthId: authId,
					RoleId:     rid,
				})
			}
			if err := tx.Create(&userRoles).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *userRepository) UpdateUserDisable(id int, isDisable bool) error {
	return r.db.Model(&entity.UserAuth{Model: entity.Model{ID: id}}).
		Update("is_disable", isDisable).Error
}

func (r *userRepository) GetList(page, size int, loginType int8, nickname, username string) ([]entity.UserAuth, int64, error) {
	var list []entity.UserAuth
	var total int64

	query := r.db.Model(&entity.UserAuth{}).
		Joins("LEFT JOIN user_info ON user_info.id = user_auth.user_info_id")

	if loginType != 0 {
		query = query.Where("user_auth.login_type = ?", loginType)
	}
	if username != "" {
		query = query.Where("user_auth.username LIKE ?", "%"+username+"%")
	}
	if nickname != "" {
		query = query.Where("user_info.nickname LIKE ?", "%"+nickname+"%")
	}

	err := query.Count(&total).
		Preload("UserInfo").
		Preload("Roles").
		Scopes(Paginate(page, size)).
		Find(&list).Error

	return list, total, err
}

// Redis operations

func (r *userRepository) GetArticleLikeSet(ctx context.Context, authId int) ([]string, error) {
	return r.rdb.SMembers(ctx, global.ARTICLE_USER_LIKE_SET+strconv.Itoa(authId)).Result()
}

func (r *userRepository) GetCommentLikeSet(ctx context.Context, authId int) ([]string, error) {
	return r.rdb.SMembers(ctx, global.COMMENT_USER_LIKE_SET+strconv.Itoa(authId)).Result()
}

func (r *userRepository) GetOnlineUsers(ctx context.Context, keyword string) ([]*entity.UserAuth, error) {
	onlineList := make([]*entity.UserAuth, 0)
	keys := r.rdb.Keys(ctx, global.ONLINE_USER+"*").Val()

	for _, key := range keys {
		val, err := r.rdb.Get(ctx, key).Result()
		if err != nil || val == "" {
			continue
		}
		var auth entity.UserAuth
		if err := json.Unmarshal([]byte(val), &auth); err != nil {
			continue
		}
		onlineList = append(onlineList, &auth)
	}

	return onlineList, nil
}

func (r *userRepository) SetOnlineUser(ctx context.Context, auth *entity.UserAuth, ttl time.Duration) error {
	authJson, _ := json.Marshal(auth)
	return r.rdb.Set(ctx, global.ONLINE_USER+strconv.Itoa(auth.ID), string(authJson), ttl).Err()
}

func (r *userRepository) DelOnlineUser(ctx context.Context, userId int) error {
	return r.rdb.Del(ctx, global.ONLINE_USER+strconv.Itoa(userId)).Err()
}

func (r *userRepository) SetOfflineMark(ctx context.Context, userId int, ttl time.Duration) error {
	return r.rdb.Set(ctx, global.OFFLINE_USER+strconv.Itoa(userId), "1", ttl).Err()
}

func (r *userRepository) CheckOffline(ctx context.Context, userId int) (bool, error) {
	key := global.OFFLINE_USER + strconv.Itoa(userId)
	val, err := r.rdb.Exists(ctx, key).Result()
	return val == 1, err
}
