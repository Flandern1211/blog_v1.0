package repository

import (
	"gin-blog/internal/model/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlogInfoRepository_Config(t *testing.T) {
	db, _ := initTestDB()
	repo := NewBlogInfoRepository()

	configs := []entity.Config{
		{Key: "name", Value: "Blog", Desc: "姓名"},
		{Key: "age", Value: "12", Desc: "年龄"},
		{Key: "enabled", Value: "true", Desc: "是否可用"},
	}
	db.Create(&configs)

	data, err := repo.GetConfigMap(db)
	assert.Nil(t, err)
	assert.Len(t, data, 3)
	assert.Equal(t, "Blog", data["name"])
	assert.Equal(t, "12", data["age"])
	assert.Equal(t, "true", data["enabled"])

	err = repo.UpdateConfigMap(db, map[string]string{
		"name": "Alice",
		"age":  "15",
	})
	assert.Nil(t, err)

	val, _ := repo.GetConfig(db, "name")
	assert.Equal(t, "Alice", val)

	assert.True(t, repo.GetConfigBool(db, "enabled"))
	assert.Equal(t, 15, repo.GetConfigInt(db, "age"))
}

func TestBlogInfoRepository_Page(t *testing.T) {
	db, _ := initTestDB()
	repo := NewBlogInfoRepository()

	db.Create(&entity.Page{Name: "name1", Label: "label1", Cover: "cover1"})
	db.Create(&entity.Page{Name: "name2", Label: "label2", Cover: "cover2"})

	list, total, err := repo.GetPageList(db)
	assert.Nil(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, list, 2)

	page := &entity.Page{Name: "name3", Label: "label3", Cover: "cover3"}
	err = repo.SaveOrUpdatePage(db, page)
	assert.Nil(t, err)
	assert.NotZero(t, page.ID)

	page.Name = "name3_updated"
	err = repo.SaveOrUpdatePage(db, page)
	assert.Nil(t, err)

	var val entity.Page
	db.First(&val, page.ID)
	assert.Equal(t, "name3_updated", val.Name)

	err = repo.DeletePages(db, []int{page.ID})
	assert.Nil(t, err)
	db.Model(&entity.Page{}).Count(&total)
	assert.Equal(t, int64(2), total)
}
