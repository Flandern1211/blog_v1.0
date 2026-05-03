package operation_log

import (
"github.com/gin-gonic/gin"
)

func RegisterOperationLogRouter(r *gin.RouterGroup, ctrl *OperationLogController) {
opt := r.Group("/operation/log")
{
opt.GET("/list", ctrl.GetList)
opt.DELETE("", ctrl.Delete)
}
}
