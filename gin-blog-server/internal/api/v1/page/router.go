package page

import (
	"github.com/gin-gonic/gin"
)

func RegisterPageRouter(r *gin.RouterGroup, ctrl *PageController) {
	page := r.Group("/page")
	{
		page.GET("/list", ctrl.GetList)
		page.POST("", ctrl.SaveOrUpdate)
		page.DELETE("", ctrl.Delete)
	}
}
