package menu

import (
	"github.com/gin-gonic/gin"
)

func RegisterMenuRouter(r *gin.RouterGroup, ctrl *MenuController) {
	menu := r.Group("/menu")
	{
		menu.GET("/list", ctrl.GetTreeList)
		menu.POST("", ctrl.SaveOrUpdate)
		menu.DELETE("/:id", ctrl.Delete)
		menu.GET("/user/list", ctrl.GetUserMenu)
		menu.GET("/option", ctrl.GetOption)
	}
}
