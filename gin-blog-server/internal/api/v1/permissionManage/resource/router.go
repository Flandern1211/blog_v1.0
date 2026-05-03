package resource

import (
	"github.com/gin-gonic/gin"
)

func RegisterResourceRouter(r *gin.RouterGroup, ctrl *ResourceController) {
	resource := r.Group("/resource")
	{
		resource.GET("/list", ctrl.GetTreeList)
		resource.GET("/option", ctrl.GetOption)
		resource.POST("", ctrl.SaveOrUpdate)
		resource.DELETE("/:id", ctrl.Delete)
		resource.PUT("/anonymous", ctrl.UpdateAnonymous)
	}
}
