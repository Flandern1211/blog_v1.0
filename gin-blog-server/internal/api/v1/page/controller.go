package page

import (
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/service"
	"gin-blog/pkg/errors"
	"gin-blog/pkg/response"

	"github.com/gin-gonic/gin"
)

type PageController struct {
	svc service.BlogInfoService
}

func NewPageController(svc service.BlogInfoService) *PageController {
	return &PageController{svc: svc}
}

func (ctrl *PageController) GetList(c *gin.Context) {
	data, _, err := ctrl.svc.GetPageList(c.Request.Context())
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, data)
}

func (ctrl *PageController) SaveOrUpdate(c *gin.Context) {
	var req request.AddOrEditPageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	page, err := ctrl.svc.SaveOrUpdatePage(c.Request.Context(), req)
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, page)
}

func (ctrl *PageController) Delete(c *gin.Context) {
	var ids []int
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.DeletePages(c.Request.Context(), ids); err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, nil)
}
