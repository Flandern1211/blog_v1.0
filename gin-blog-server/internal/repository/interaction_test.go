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

func initTestDB() (*gorm.DB, error) {
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

func TestInteractionRepository_GetCommentList(t *testing.T) {
	db, _ := initTestDB()
	repo := NewInteractionRepository()

	user := entity.UserAuth{
		Username: "username",
		Password: "123456",
		UserInfo: &entity.UserInfo{
			Nickname: "nickname",
		},
	}
	db.Create(&user)

	article := entity.Article{Title: "title", Content: "content"}
	db.Create(&article)

	// Add parent comment
	comment := entity.Comment{
		UserId:   user.ID,
		Type:     1, // TYPE_ARTICLE
		TopicId:  article.ID,
		Content:  "content",
		IsReview: true,
	}
	db.Create(&comment)

	// Add reply comment
	reply := entity.Comment{
		UserId:      user.ID,
		ReplyUserId: user.ID,
		ParentId:    comment.ID,
		Type:        1, // TYPE_ARTICLE
		TopicId:     article.ID,
		Content:     "reply_content",
		IsReview:    true,
	}
	db.Create(&reply)

	data, total, err := repo.GetCommentList(db, 1, 10, 1, nil, "")
	assert.Nil(t, err)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "reply_content", data[0].Content)
	assert.Equal(t, "content", data[1].Content)

	v1 := data[0]
	assert.Equal(t, "reply_content", v1.Content)
	assert.Equal(t, "username", v1.User.Username)
	assert.Equal(t, "nickname", v1.User.UserInfo.Nickname)
	assert.Equal(t, "username", v1.ReplyUser.Username)
	assert.Equal(t, "nickname", v1.ReplyUser.UserInfo.Nickname)
	assert.Equal(t, "title", v1.Article.Title)
}

func TestInteractionRepository_Messages(t *testing.T) {
	db, _ := initTestDB()
	repo := NewInteractionRepository()

	msg := &entity.Message{
		Nickname: "test",
		Content:  "content",
		IsReview: true,
	}
	err := repo.SaveMessage(db, msg)
	assert.Nil(t, err)

	list, total, err := repo.GetMessageList(db, 1, 10, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, "test", list[0].Nickname)

	err = repo.UpdateMessagesReview(db, []int{msg.ID}, false)
	assert.Nil(t, err)

	isReview := false
	list, total, err = repo.GetMessageList(db, 1, 10, "", &isReview)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), total)
	assert.False(t, list[0].IsReview)

	err = repo.DeleteMessages(db, []int{msg.ID})
	assert.Nil(t, err)

	list, total, err = repo.GetMessageList(db, 1, 10, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, int64(0), total)
}
