package message

import (
"github.com/gin-gonic/gin"
)

func RegisterMessageRouter(r *gin.RouterGroup, ctrl *MessageController) {
msg := r.Group("/message")
{
msg.GET("/list", ctrl.GetList)
msg.PUT("/review", ctrl.UpdateReview)
msg.DELETE("", ctrl.Delete)
}
}
