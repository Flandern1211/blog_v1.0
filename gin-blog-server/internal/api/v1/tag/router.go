package tag

import (
"github.com/gin-gonic/gin"
)

func RegisterTagRouter(r *gin.RouterGroup, ctrl *TagController) {
tag := r.Group("/tag")
{
tag.GET("/list", ctrl.GetList)
tag.POST("", ctrl.SaveOrUpdate)
tag.DELETE("", ctrl.Delete)
tag.GET("/option", ctrl.GetOption)
}
}
