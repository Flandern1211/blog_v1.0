package role

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoleRouter(r *gin.RouterGroup, ctrl *RoleController) {
	role := r.Group("/role")
	{
		role.GET("/option", ctrl.GetOption)
		role.GET("/list", ctrl.GetTreeList)
		role.POST("", ctrl.SaveOrUpdate)
		role.DELETE("", ctrl.Delete)
	}
}
