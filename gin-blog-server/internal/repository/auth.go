package repository

import (
	"context"
	"gin-blog/internal/model/entity"
	"strconv"
	"time"

	global "gin-blog/internal/global"
	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

type AuthRepository interface {
	GetUserAuthInfoByName(username string) (*entity.UserAuth, error)
	GetUserInfoById(id int) (*entity.UserInfo, error)
	GetRoleIdsByUserId(userId int) ([]int, error)
	UpdateUserLoginInfo(userId int, ipAddress, ipSource string) error
	CreateNewUser(username, email, password string) (*entity.UserAuth, *entity.UserInfo, *entity.UserAuthRole, error)
	GetUserAuthInfoById(id int) (*entity.UserAuth, error)
	GetResource(url, method string) (*entity.Resource, error)
	CheckRoleAuth(roleId int, url, method string) (bool, error)
	CheckUserHasResource(userId int, url, method string) (bool, error)

	// Redis operations
	SetToken(ctx context.Context, tokenStr string, userId int, expire time.Duration) error
	TokenExists(ctx context.Context, tokenStr string) bool
	DelToken(ctx context.Context, tokenStr string) error
	SetEmailCode(ctx context.Context, email, code string, ttl time.Duration) error
	GetEmailCode(ctx context.Context, email string) (string, error)
	DelEmailCode(ctx context.Context, email string) error
	SetVerificationInfo(ctx context.Context, info string, ttl time.Duration) error
	GetVerificationInfo(ctx context.Context, info string) (string, error)
	DelVerificationInfo(ctx context.Context, info string) error
	DelOfflineMark(ctx context.Context, userId int) error
	SetOfflineMark(ctx context.Context, userId int, ttl time.Duration) error
}

type authRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewAuthRepository(db *gorm.DB, rdb *redis.Client) AuthRepository {
	return &authRepository{db: db, rdb: rdb}
}

func (r *authRepository) GetUserAuthInfoByName(username string) (*entity.UserAuth, error) {
	var userAuth entity.UserAuth
	result := r.db.Where("username = ?", username).First(&userAuth)
	return &userAuth, result.Error
}

func (r *authRepository) GetUserInfoById(id int) (*entity.UserInfo, error) {
	var userInfo entity.UserInfo
	result := r.db.First(&userInfo, id)
	return &userInfo, result.Error
}

func (r *authRepository) GetRoleIdsByUserId(userId int) ([]int, error) {
	var ids []int
	result := r.db.Model(&entity.UserAuthRole{UserAuthId: userId}).Pluck("role_id", &ids)
	return ids, result.Error
}

func (r *authRepository) UpdateUserLoginInfo(userId int, ipAddress, ipSource string) error {
	return r.db.Model(&entity.UserAuth{Model: entity.Model{ID: userId}}).
		Updates(map[string]interface{}{
			"ip_address": ipAddress,
			"ip_source":  ipSource,
		}).Error
}

func (r *authRepository) CreateNewUser(username, email, password string) (*entity.UserAuth, *entity.UserInfo, *entity.UserAuthRole, error) {
	var num int64
	r.db.Model(&entity.UserInfo{}).Count(&num)
	number := strconv.FormatInt(num, 10)

	userInfo := &entity.UserInfo{
		Email:    email,
		Nickname: "游客" + number,
		Avatar:   "https://www.bing.com/rp/ar_9isCNU2Q-VG1yEDDHnx8HAFQ.png",
		Intro:    "我是这个程序的第" + number + "个用户",
	}

	var userAuth *entity.UserAuth
	var userRole *entity.UserAuthRole

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(userInfo).Error; err != nil {
			return err
		}

		userAuth = &entity.UserAuth{
			Username:   username,
			Password:   password,
			UserInfoId: userInfo.ID,
		}
		if err := tx.Create(userAuth).Error; err != nil {
			return err
		}

		userRole = &entity.UserAuthRole{
			UserAuthId: userAuth.ID,
			RoleId:     2,
		}
		if err := tx.Create(userRole).Error; err != nil {
			return err
		}

		return nil
	})

	return userAuth, userInfo, userRole, err
}

func (r *authRepository) GetUserAuthInfoById(id int) (*entity.UserAuth, error) {
	var userAuth entity.UserAuth
	err := r.db.Preload("Roles").Preload("UserInfo").First(&userAuth, id).Error
	return &userAuth, err
}

func (r *authRepository) GetResource(url, method string) (*entity.Resource, error) {
	var resource entity.Resource
	err := r.db.Where("url = ? AND method = ?", url, method).First(&resource).Error
	return &resource, err
}

func (r *authRepository) CheckRoleAuth(roleId int, url, method string) (bool, error) {
	var role entity.Role
	if err := r.db.Preload("Resources").First(&role, roleId).Error; err != nil {
		return false, err
	}

	for _, res := range role.Resources {
		if res.Anonymous || (res.Url == url && res.Method == method) {
			return true, nil
		}
	}

	return false, nil
}

func (r *authRepository) CheckUserHasResource(userId int, url, method string) (bool, error) {
	var count int64
	err := r.db.Table("role_resource").
		Joins("JOIN user_auth_role ON user_auth_role.role_id = role_resource.role_id").
		Joins("JOIN resource ON resource.id = role_resource.resource_id").
		Where("user_auth_role.user_auth_id = ?", userId).
		Where("resource.url = ?", url).
		Where("resource.method = ?", method).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Redis operations

func (r *authRepository) SetToken(ctx context.Context, tokenStr string, userId int, expire time.Duration) error {
	key := global.TOKEN_WHITELIST + tokenStr
	return r.rdb.Set(ctx, key, userId, expire).Err()
}

func (r *authRepository) TokenExists(ctx context.Context, tokenStr string) bool {
	key := global.TOKEN_WHITELIST + tokenStr
	return r.rdb.Exists(ctx, key).Val() == 1
}

func (r *authRepository) DelToken(ctx context.Context, tokenStr string) error {
	key := global.TOKEN_WHITELIST + tokenStr
	return r.rdb.Del(ctx, key).Err()
}

func (r *authRepository) SetEmailCode(ctx context.Context, email, code string, ttl time.Duration) error {
	return r.rdb.Set(ctx, global.EMAIL_CODE+email, code, ttl).Err()
}

func (r *authRepository) GetEmailCode(ctx context.Context, email string) (string, error) {
	return r.rdb.Get(ctx, global.EMAIL_CODE+email).Result()
}

func (r *authRepository) DelEmailCode(ctx context.Context, email string) error {
	return r.rdb.Del(ctx, global.EMAIL_CODE+email).Err()
}

func (r *authRepository) SetVerificationInfo(ctx context.Context, info string, ttl time.Duration) error {
	return r.rdb.Set(ctx, info, info, ttl).Err()
}

func (r *authRepository) GetVerificationInfo(ctx context.Context, info string) (string, error) {
	return r.rdb.Get(ctx, info).Result()
}

func (r *authRepository) DelVerificationInfo(ctx context.Context, info string) error {
	return r.rdb.Del(ctx, info).Err()
}

func (r *authRepository) DelOfflineMark(ctx context.Context, userId int) error {
	return r.rdb.Del(ctx, global.OFFLINE_USER+strconv.Itoa(userId)).Err()
}

func (r *authRepository) SetOfflineMark(ctx context.Context, userId int, ttl time.Duration) error {
	key := global.OFFLINE_USER + strconv.Itoa(userId)
	return r.rdb.Set(ctx, key, "1", ttl).Err()
}
