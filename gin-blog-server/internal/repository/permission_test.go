package repository

import (
	"gin-blog/internal/app"
	"gin-blog/internal/model/entity"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func initPermissionTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	app.MakeMigrate(db)
	return db, nil
}

func TestPermissionRepository(t *testing.T) {
	db, _ := initPermissionTestDB()
	repo := NewPermissionRepository()
	authRepo := NewAuthRepository()

	// Add Resources
	r1 := &entity.Resource{Name: "api_1", Url: "/v1/api_1", Method: "GET"}
	r2 := &entity.Resource{Name: "api_2", Url: "/v1/api_2", Method: "POST"}
	db.Create(r1)
	db.Create(r2)

	// Add Role with Resources
	role := &entity.Role{Name: "admin", Label: "管理员"}
	db.Create(role)
	err := repo.UpdateRole(db, role.ID, role.Name, role.Label, false, []int{r1.ID, r2.ID}, nil)
	assert.Nil(t, err)

	resources, err := repo.GetResourceIdsByRoleId(db, role.ID)
	assert.Nil(t, err)
	assert.Len(t, resources, 2)

	// Update Role Resources
	err = repo.UpdateRole(db, role.ID, role.Name, role.Label, false, []int{r1.ID}, nil)
	assert.Nil(t, err)
	resources, err = repo.GetResourceIdsByRoleId(db, role.ID)
	assert.Nil(t, err)
	assert.Len(t, resources, 1)

	// Test Role Auth
	flag, err := authRepo.CheckRoleAuth(db, role.ID, "/v1/api_1", "GET")
	assert.Nil(t, err)
	assert.True(t, flag)

	flag, err = authRepo.CheckRoleAuth(db, role.ID, "/v1/api_99", "POST")
	assert.Nil(t, err)
	assert.False(t, flag)
}
