package service

import (
	"context"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"
	g2 "gin-blog/pkg/errors"
	"sort"
)

type PermissionService interface {
	// Role
	GetRoleList(ctx context.Context, query request.PageQuery, keyword string) ([]response.RoleVO, int64, error)
	GetRoleOption(ctx context.Context) ([]response.OptionVO, error)
	SaveOrUpdateRole(ctx context.Context, req request.AddOrEditRoleReq) error
	DeleteRoles(ctx context.Context, ids []int) error

	// Menu
	GetUserMenu(ctx context.Context, authId int, isSuper bool) ([]response.MenuTreeVO, error)
	GetMenuTreeList(ctx context.Context, keyword string) ([]response.MenuTreeVO, error)
	GetMenuOption(ctx context.Context) ([]response.TreeOptionVO, error)
	SaveOrUpdateMenu(ctx context.Context, req request.AddOrEditMenuReq) error
	DeleteMenu(ctx context.Context, id int) error

	// Resource
	GetResourceTreeList(ctx context.Context, keyword string) ([]response.ResourceTreeVO, error)
	GetResourceOption(ctx context.Context) ([]response.TreeOptionVO, error)
	SaveOrUpdateResource(ctx context.Context, req request.AddOrEditResourceReq) error
	DeleteResource(ctx context.Context, id int) error
	UpdateResourceAnonymous(ctx context.Context, req request.EditAnonymousReq) error
}

type permissionService struct {
	repo repository.PermissionRepository
}

func NewPermissionService(repo repository.PermissionRepository) PermissionService {
	return &permissionService{repo: repo}
}

// Role implementations
func (s *permissionService) GetRoleList(ctx context.Context, query request.PageQuery, keyword string) ([]response.RoleVO, int64, error) {
	list, total, err := s.repo.GetRoleList(query.GetPage(), query.GetSize(), keyword)
	if err != nil {
		return nil, 0, err
	}

	var res []response.RoleVO
	for _, role := range list {
		rVO := response.RoleVO{
			ID:        role.ID,
			Name:      role.Name,
			Label:     role.Label,
			IsDisable: role.IsDisable,
			CreatedAt: role.CreatedAt,
		}
		rVO.ResourceIds, _ = s.repo.GetResourceIdsByRoleId(role.ID)
		rVO.MenuIds, _ = s.repo.GetMenuIdsByRoleId(role.ID)
		res = append(res, rVO)
	}
	return res, total, nil
}

func (s *permissionService) GetRoleOption(ctx context.Context) ([]response.OptionVO, error) {
	list, err := s.repo.GetRoleOption()
	if err != nil {
		return nil, err
	}
	var res []response.OptionVO
	for _, role := range list {
		res = append(res, response.OptionVO{ID: role.ID, Label: role.Label})
	}
	return res, nil
}

func (s *permissionService) SaveOrUpdateRole(ctx context.Context, req request.AddOrEditRoleReq) error {
	if req.ID == 0 {
		return s.repo.SaveRole(req.Name, req.Label)
	}
	return s.repo.UpdateRole(req.ID, req.Name, req.Label, req.IsDisable, req.ResourceIds, req.MenuIds)
}

func (s *permissionService) DeleteRoles(ctx context.Context, ids []int) error {
	return s.repo.DeleteRoles(ids)
}

// Menu implementations
func (s *permissionService) GetUserMenu(ctx context.Context, authId int, isSuper bool) ([]response.MenuTreeVO, error) {
	var menus []entity.Menu
	var err error
	if isSuper {
		menus, err = s.repo.GetAllMenuList()
	} else {
		menus, err = s.repo.GetMenuListByUserId(authId)
	}
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus, 0), nil
}

func (s *permissionService) GetMenuTreeList(ctx context.Context, keyword string) ([]response.MenuTreeVO, error) {
	menus, err := s.repo.GetMenuList(keyword)
	if err != nil {
		return nil, err
	}
	return s.buildMenuTree(menus, 0), nil
}

func (s *permissionService) GetMenuOption(ctx context.Context) ([]response.TreeOptionVO, error) {
	menus, err := s.repo.GetMenuList("")
	if err != nil {
		return nil, err
	}
	return s.buildMenuTreeOption(menus, 0), nil
}

func (s *permissionService) buildMenuTreeOption(menus []entity.Menu, parentId int) []response.TreeOptionVO {
	var tree []response.TreeOptionVO
	for _, m := range menus {
		if m.ParentId == parentId {
			tree = append(tree, response.TreeOptionVO{
				ID:       m.ID,
				Label:    m.Name,
				Children: s.buildMenuTreeOption(menus, m.ID),
			})
		}
	}
	return tree
}

func (s *permissionService) buildMenuTree(menus []entity.Menu, parentId int) []response.MenuTreeVO {
	var tree []response.MenuTreeVO
	for _, m := range menus {
		if m.ParentId == parentId {
			tree = append(tree, response.MenuTreeVO{
				Menu:     m,
				Children: s.buildMenuTree(menus, m.ID),
			})
		}
	}
	sort.Slice(tree, func(i, j int) bool {
		return tree[i].OrderNum < tree[j].OrderNum
	})
	return tree
}

func (s *permissionService) SaveOrUpdateMenu(ctx context.Context, req request.AddOrEditMenuReq) error {
	menu := &entity.Menu{
		Model:        entity.Model{ID: req.ID},
		ParentId:     req.ParentId,
		Name:         req.Name,
		Path:         req.Path,
		Component:    req.Component,
		Icon:         req.Icon,
		OrderNum:     req.OrderNum,
		Redirect:     req.Redirect,
		Catalogue:    req.Catalogue,
		Hidden:       req.Hidden,
		KeepAlive:    req.KeepAlive,
		External:     req.External,
		ExternalLink: req.ExternalLink,
	}
	return s.repo.SaveOrUpdateMenu(menu)
}

func (s *permissionService) DeleteMenu(ctx context.Context, id int) error {
	inUse, _ := s.repo.CheckMenuInUse(id)
	if inUse {
		return g2.NewDefault(g2.CodeMenuUsedByRole)
	}
	hasChild, _ := s.repo.CheckMenuHasChild(id)
	if hasChild {
		return g2.NewDefault(g2.CodeMenuHasChildren)
	}
	return s.repo.DeleteMenu(id)
}

// Resource implementations
func (s *permissionService) GetResourceTreeList(ctx context.Context, keyword string) ([]response.ResourceTreeVO, error) {
	resources, err := s.repo.GetResourceList(keyword)
	if err != nil {
		return nil, err
	}
	return s.buildResourceTree(resources, 0), nil
}

func (s *permissionService) buildResourceTree(resources []entity.Resource, parentId int) []response.ResourceTreeVO {
	var tree []response.ResourceTreeVO
	for _, r := range resources {
		if r.ParentId == parentId {
			tree = append(tree, response.ResourceTreeVO{
				ID:        r.ID,
				CreatedAt: r.CreatedAt,
				Name:      r.Name,
				Url:       r.Url,
				Method:    r.Method,
				Anonymous: r.Anonymous,
				Children:  s.buildResourceTree(resources, r.ID),
			})
		}
	}
	return tree
}

func (s *permissionService) GetResourceOption(ctx context.Context) ([]response.TreeOptionVO, error) {
	resources, err := s.repo.GetResourceList("")
	if err != nil {
		return nil, err
	}
	return s.buildResourceOptionTree(resources, 0), nil
}

func (s *permissionService) buildResourceOptionTree(resources []entity.Resource, parentId int) []response.TreeOptionVO {
	var tree []response.TreeOptionVO
	for _, r := range resources {
		if r.ParentId == parentId {
			tree = append(tree, response.TreeOptionVO{
				ID:       r.ID,
				Label:    r.Name,
				Children: s.buildResourceOptionTree(resources, r.ID),
			})
		}
	}
	return tree
}

func (s *permissionService) SaveOrUpdateResource(ctx context.Context, req request.AddOrEditResourceReq) error {
	return s.repo.SaveOrUpdateResource(req.ID, req.ParentId, req.Name, req.Url, req.Method)
}

func (s *permissionService) DeleteResource(ctx context.Context, id int) error {
	return s.repo.DeleteResource(id)
}

func (s *permissionService) UpdateResourceAnonymous(ctx context.Context, req request.EditAnonymousReq) error {
	return s.repo.UpdateResourceAnonymous(req.ID, req.Anonymous)
}
