package repository

import (
	"gin-blog/internal/model/entity"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	// Role
	GetRoleList(page, size int, keyword string) ([]entity.Role, int64, error)
	GetRoleOption() ([]entity.Role, error)
	SaveRole(name, label string) error
	UpdateRole(id int, name, label string, isDisable bool, resourceIds, menuIds []int) error
	DeleteRoles(ids []int) error
	GetResourceIdsByRoleId(roleId int) ([]int, error)
	GetMenuIdsByRoleId(roleId int) ([]int, error)

	// Menu
	GetMenuList(keyword string) ([]entity.Menu, error)
	GetMenuListByUserId(userId int) ([]entity.Menu, error)
	GetAllMenuList() ([]entity.Menu, error)
	SaveOrUpdateMenu(menu *entity.Menu) error
	DeleteMenu(id int) error
	GetMenuById(id int) (*entity.Menu, error)
	CheckMenuInUse(id int) (bool, error)
	CheckMenuHasChild(id int) (bool, error)

	// Resource
	GetResourceList(keyword string) ([]entity.Resource, error)
	SaveOrUpdateResource(id, parentId int, name, url, method string) error
	DeleteResource(id int) error
	UpdateResourceAnonymous(id int, anonymous bool) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// Role implementations
func (r *permissionRepository) GetRoleList(page, size int, keyword string) ([]entity.Role, int64, error) {
	var list []entity.Role
	var total int64
	db := r.db.Model(&entity.Role{})
	if keyword != "" {
		db = db.Where("name LIKE ? OR label LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	err := db.Count(&total).Scopes(Paginate(page, size)).Find(&list).Error
	return list, total, err
}

func (r *permissionRepository) GetRoleOption() ([]entity.Role, error) {
	var list []entity.Role
	err := r.db.Model(&entity.Role{}).Select("id", "label").Find(&list).Error
	return list, err
}

func (r *permissionRepository) SaveRole(name, label string) error {
	return r.db.Create(&entity.Role{Name: name, Label: label}).Error
}

func (r *permissionRepository) UpdateRole(id int, name, label string, isDisable bool, resourceIds, menuIds []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Role{Model: entity.Model{ID: id}}).Updates(entity.Role{
			Name:      name,
			Label:     label,
			IsDisable: isDisable,
		}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&entity.RoleResource{}, "role_id = ?", id).Error; err != nil {
			return err
		}
		if len(resourceIds) > 0 {
			var roleResources []entity.RoleResource
			for _, rid := range resourceIds {
				roleResources = append(roleResources, entity.RoleResource{RoleId: id, ResourceId: rid})
			}
			if err := tx.Create(&roleResources).Error; err != nil {
				return err
			}
		}

		if err := tx.Delete(&entity.RoleMenu{}, "role_id = ?", id).Error; err != nil {
			return err
		}
		if len(menuIds) > 0 {
			var roleMenus []entity.RoleMenu
			for _, mid := range menuIds {
				roleMenus = append(roleMenus, entity.RoleMenu{RoleId: id, MenuId: mid})
			}
			if err := tx.Create(&roleMenus).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *permissionRepository) DeleteRoles(ids []int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&entity.Role{}, ids).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entity.RoleResource{}, "role_id IN ?", ids).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entity.RoleMenu{}, "role_id IN ?", ids).Error; err != nil {
			return err
		}
		if err := tx.Delete(&entity.UserAuthRole{}, "role_id IN ?", ids).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *permissionRepository) GetResourceIdsByRoleId(roleId int) ([]int, error) {
	var ids []int
	err := r.db.Model(&entity.RoleResource{}).Where("role_id = ?", roleId).Pluck("resource_id", &ids).Error
	return ids, err
}

func (r *permissionRepository) GetMenuIdsByRoleId(roleId int) ([]int, error) {
	var ids []int
	err := r.db.Model(&entity.RoleMenu{}).Where("role_id = ?", roleId).Pluck("menu_id", &ids).Error
	return ids, err
}

// Menu implementations
func (r *permissionRepository) GetMenuList(keyword string) ([]entity.Menu, error) {
	var list []entity.Menu
	db := r.db.Model(&entity.Menu{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	err := db.Order("order_num").Find(&list).Error
	return list, err
}

func (r *permissionRepository) GetMenuListByUserId(userId int) ([]entity.Menu, error) {
	var list []entity.Menu
	err := r.db.Table("menu").
		Joins("JOIN role_menu ON menu.id = role_menu.menu_id").
		Joins("JOIN user_auth_role ON role_menu.role_id = user_auth_role.role_id").
		Where("user_auth_role.user_auth_id = ?", userId).
		Distinct("menu.*").
		Order("order_num").
		Find(&list).Error
	return list, err
}

func (r *permissionRepository) GetAllMenuList() ([]entity.Menu, error) {
	var list []entity.Menu
	err := r.db.Model(&entity.Menu{}).Order("order_num").Find(&list).Error
	return list, err
}

func (r *permissionRepository) SaveOrUpdateMenu(menu *entity.Menu) error {
	if menu.ID == 0 {
		return r.db.Create(menu).Error
	}
	return r.db.Model(menu).Updates(menu).Error
}

func (r *permissionRepository) DeleteMenu(id int) error {
	return r.db.Delete(&entity.Menu{}, id).Error
}

func (r *permissionRepository) GetMenuById(id int) (*entity.Menu, error) {
	var menu entity.Menu
	err := r.db.First(&menu, id).Error
	return &menu, err
}

func (r *permissionRepository) CheckMenuInUse(id int) (bool, error) {
	var count int64
	err := r.db.Model(&entity.RoleMenu{}).Where("menu_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *permissionRepository) CheckMenuHasChild(id int) (bool, error) {
	var count int64
	err := r.db.Model(&entity.Menu{}).Where("parent_id = ?", id).Count(&count).Error
	return count > 0, err
}

// Resource implementations
func (r *permissionRepository) GetResourceList(keyword string) ([]entity.Resource, error) {
	var list []entity.Resource
	db := r.db.Model(&entity.Resource{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	err := db.Find(&list).Error
	return list, err
}

func (r *permissionRepository) SaveOrUpdateResource(id, parentId int, name, url, method string) error {
	if id == 0 {
		return r.db.Create(&entity.Resource{
			ParentId: parentId,
			Name:     name,
			Url:      url,
			Method:   method,
		}).Error
	}
	return r.db.Model(&entity.Resource{Model: entity.Model{ID: id}}).Updates(entity.Resource{
		ParentId: parentId,
		Name:     name,
		Url:      url,
		Method:   method,
	}).Error
}

func (r *permissionRepository) DeleteResource(id int) error {
	return r.db.Delete(&entity.Resource{}, id).Error
}

func (r *permissionRepository) UpdateResourceAnonymous(id int, anonymous bool) error {
	return r.db.Model(&entity.Resource{Model: entity.Model{ID: id}}).Update("anonymous", anonymous).Error
}
