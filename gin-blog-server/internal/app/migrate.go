package app

import (
	"gin-blog/internal/model/entity"

	"gorm.io/gorm"
)

func MakeMigrate(db *gorm.DB) error {
	db.SetupJoinTable(&entity.Role{}, "Menus", &entity.RoleMenu{})
	db.SetupJoinTable(&entity.Role{}, "Resources", &entity.RoleResource{})
	db.SetupJoinTable(&entity.Role{}, "Users", &entity.UserAuthRole{})
	db.SetupJoinTable(&entity.UserAuth{}, "Roles", &entity.UserAuthRole{})

	return db.AutoMigrate(
		&entity.Article{},
		&entity.Category{},
		&entity.Tag{},
		&entity.Comment{},
		&entity.Message{},
		&entity.FriendLink{},
		&entity.Page{},
		&entity.Config{},
		&entity.OperationLog{},
		&entity.UserInfo{},
		&entity.UserAuth{},
		&entity.Role{},
		&entity.Menu{},
		&entity.Resource{},
		&entity.RoleMenu{},
		&entity.RoleResource{},
		&entity.UserAuthRole{},
	)
}
