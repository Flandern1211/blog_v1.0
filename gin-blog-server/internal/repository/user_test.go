package repository

import (
	"gin-blog/internal/model/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
	db, _ := initTestDB()
	repo := NewUserRepository()
	authRepo := NewAuthRepository()

	var auth = entity.UserAuth{
		Username: "test",
		Password: "123456",
	}
	db.Create(&auth)

	// Test GetUserAuthInfoById
	val, err := authRepo.GetUserAuthInfoById(db, auth.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test", val.Username)

	// Test UpdateUserPassword
	err = repo.UpdateUserPassword(db, auth.ID, "654321")
	assert.Nil(t, err)

	// Test UpdateUserInfo
	userInfo := entity.UserInfo{
		Nickname: "nickname",
		Avatar:   "avatar",
		Intro:    "intro",
	}
	db.Create(&userInfo)

	err = repo.UpdateUserInfo(db, userInfo.ID, "update_nickname", "update_avatar", "intro", "website")
	assert.Nil(t, err)

	db.First(&userInfo, userInfo.ID)
	assert.Equal(t, "update_nickname", userInfo.Nickname)
	assert.Equal(t, "update_avatar", userInfo.Avatar)
}
