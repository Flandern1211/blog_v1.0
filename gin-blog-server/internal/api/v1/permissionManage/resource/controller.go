package resource

import (
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/service"
	"gin-blog/pkg/errors"
	"gin-blog/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResourceController struct {
	svc service.PermissionService
}

func NewResourceController(svc service.PermissionService) *ResourceController {
	return &ResourceController{svc: svc}
}

func (ctrl *ResourceController) GetTreeList(c *gin.Context) {
	keyword := c.Query("keyword")
	list, err := ctrl.svc.GetResourceTreeList(c.Request.Context(), keyword)
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, list)
}

func (ctrl *ResourceController) GetOption(c *gin.Context) {
	list, err := ctrl.svc.GetResourceOption(c.Request.Context())
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, list)
}

func (ctrl *ResourceController) SaveOrUpdate(c *gin.Context) {
	var req request.AddOrEditResourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.SaveOrUpdateResource(c.Request.Context(), req); err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, nil)
}

func (ctrl *ResourceController) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.DeleteResource(c.Request.Context(), id); err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, nil)
}

func (ctrl *ResourceController) UpdateAnonymous(c *gin.Context) {
	var req request.EditAnonymousReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.UpdateResourceAnonymous(c.Request.Context(), req); err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, nil)
}
