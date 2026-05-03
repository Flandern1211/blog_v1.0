package repository

import (
	"gin-blog/internal/model/entity"

	"gorm.io/gorm"
)

type SystemRepository interface {
	// FriendLink
	GetLinkList(page, size int, keyword string) ([]entity.FriendLink, int64, error)
	SaveOrUpdateLink(link *entity.FriendLink) error
	DeleteLinks(ids []int) error

	// OperationLog
	GetOperationLogList(page, size int, keyword string) ([]entity.OperationLog, int64, error)
	DeleteOperationLogs(ids []int) error
	CreateOperationLog(log *entity.OperationLog) error
}

type systemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) SystemRepository {
	return &systemRepository{db: db}
}

// FriendLink implementations
func (r *systemRepository) GetLinkList(page, size int, keyword string) ([]entity.FriendLink, int64, error) {
	var list []entity.FriendLink
	var total int64
	query := r.db.Model(&entity.FriendLink{})

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%").
			Or("address LIKE ?", "%"+keyword+"%").
			Or("intro LIKE ?", "%"+keyword+"%")
	}

	err := query.Count(&total).Order("created_at DESC").Scopes(Paginate(page, size)).Find(&list).Error
	return list, total, err
}

func (r *systemRepository) SaveOrUpdateLink(link *entity.FriendLink) error {
	if link.ID > 0 {
		return r.db.Updates(link).Error
	}
	return r.db.Create(link).Error
}

func (r *systemRepository) DeleteLinks(ids []int) error {
	return r.db.Where("id IN ?", ids).Delete(&entity.FriendLink{}).Error
}

// OperationLog implementations
func (r *systemRepository) GetOperationLogList(page, size int, keyword string) ([]entity.OperationLog, int64, error) {
	var list []entity.OperationLog
	var total int64
	query := r.db.Model(&entity.OperationLog{})

	if keyword != "" {
		query = query.Where("opt_module LIKE ?", "%"+keyword+"%").
			Or("opt_desc LIKE ?", "%"+keyword+"%")
	}

	err := query.Count(&total).Order("created_at DESC").Scopes(Paginate(page, size)).Find(&list).Error
	return list, total, err
}

func (r *systemRepository) DeleteOperationLogs(ids []int) error {
	return r.db.Where("id IN ?", ids).Delete(&entity.OperationLog{}).Error
}

func (r *systemRepository) CreateOperationLog(log *entity.OperationLog) error {
	return r.db.Create(log).Error
}
