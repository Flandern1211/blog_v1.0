package menu

import (
	"gin-blog/internal/middleware"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/service"
	"gin-blog/pkg/errors"
	"gin-blog/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MenuController struct {
	svc service.PermissionService
}

func NewMenuController(svc service.PermissionService) *MenuController {
	return &MenuController{svc: svc}
}

func (ctrl *MenuController) GetUserMenu(c *gin.Context) {
	authId := middleware.GetUserID(c)
	if authId == 0 {
		response.Error(c, errors.CodeNoLogin, errors.GetMessage(errors.CodeNoLogin))
		return
	}

	isSuper := middleware.IsSuper(c)

	list, err := ctrl.svc.GetUserMenu(c.Request.Context(), authId, isSuper)
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, list)
}

func (ctrl *MenuController) GetTreeList(c *gin.Context) {
	keyword := c.Query("keyword")
	list, err := ctrl.svc.GetMenuTreeList(c.Request.Context(), keyword)
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, list)
}

func (ctrl *MenuController) GetOption(c *gin.Context) {
	list, err := ctrl.svc.GetMenuOption(c.Request.Context())
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, list)
}

func (ctrl *MenuController) SaveOrUpdate(c *gin.Context) {
	var req request.AddOrEditMenuReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.SaveOrUpdateMenu(c.Request.Context(), req); err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, nil)
}

func (ctrl *MenuController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.DeleteMenu(c.Request.Context(), id); err != nil {
		response.BizError(c, err)
		return
	}
	response.Success(c, nil)
}
