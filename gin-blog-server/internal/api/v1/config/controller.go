package config

import (
	"gin-blog/internal/service"
	"gin-blog/pkg/errors"
	"gin-blog/pkg/response"

	"github.com/gin-gonic/gin"
)

type ConfigController struct {
	svc service.BlogInfoService
}

func NewConfigController(svc service.BlogInfoService) *ConfigController {
	return &ConfigController{svc: svc}
}

func (ctrl *ConfigController) GetConfigMap(c *gin.Context) {
	data, err := ctrl.svc.GetConfigMap(c.Request.Context())
	if err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, data)
}

func (ctrl *ConfigController) UpdateConfigMap(c *gin.Context) {
	var m map[string]string
	if err := c.ShouldBindJSON(&m); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}
	if err := ctrl.svc.UpdateConfigMap(c.Request.Context(), m); err != nil {
		response.Error(c, errors.CodeDbOpError, errors.GetMessage(errors.CodeDbOpError))
		return
	}
	response.Success(c, nil)
}
