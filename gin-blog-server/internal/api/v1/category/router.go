package category

import (
	"github.com/gin-gonic/gin"
)

func RegisterCategoryRouter(r *gin.RouterGroup, ctrl *CategoryController) {
	cat := r.Group("/category")
	{
		cat.GET("/list", ctrl.GetList)
		cat.POST("", ctrl.SaveOrUpdate)
		cat.DELETE("", ctrl.Delete)
		cat.GET("/option", ctrl.GetOption)
	}
}
