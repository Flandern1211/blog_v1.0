package repository

import (
	"gin-blog/internal/model/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRepository(t *testing.T) {
	db, _ := initTestDB()
	repo := NewAuthRepository()

	// Test CreateNewUser
	userAuth, userInfo, _, err := repo.CreateNewUser(db, "admin", "123456")
	assert.Nil(t, err)
	assert.Equal(t, "admin", userAuth.Username)
	assert.Equal(t, "admin", userInfo.Email)

	// Test GetUserAuthInfoById
	val, err := repo.GetUserAuthInfoById(db, userAuth.ID)
	assert.Nil(t, err)
	assert.Equal(t, userAuth.Username, val.Username)
	assert.Equal(t, userAuth.UserInfoId, val.UserInfoId)
	assert.Equal(t, userInfo.Nickname, val.UserInfo.Nickname)
}

func TestAuthRepository_GetMenuListByUserId(t *testing.T) {
	db, _ := initTestDB()
	repo := NewPermissionRepository()

	user := entity.UserAuth{
		Username: "user",
		Password: "password",
		UserInfo: &entity.UserInfo{
			Nickname: "nickname",
		},
		Roles: []*entity.Role{
			{
				Name:  "role_1",
				Label: "label_1",
				Menus: []entity.Menu{
					{Name: "menu1", Path: "/menu1"},
					{Name: "menu2", Path: "/menu2"},
				},
			},
			{
				Name:  "role_2",
				Label: "label_2",
				Menus: []entity.Menu{
					{Name: "menu3", Path: "/menu3"},
					{Name: "menu4", Path: "/menu4"},
				},
			},
		},
	}

	db.Create(&user)

	menus, err := repo.GetMenuListByUserId(db, user.ID)
	assert.Nil(t, err)
	assert.Len(t, menus, 4)
}
