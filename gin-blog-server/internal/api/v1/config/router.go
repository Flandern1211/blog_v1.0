package config

import (
	"github.com/gin-gonic/gin"
)

func RegisterConfigRouter(r *gin.RouterGroup, ctrl *ConfigController) {
	r.GET("/config", ctrl.GetConfigMap)
	r.PATCH("/config", ctrl.UpdateConfigMap)
}
