package upload

import "github.com/gin-gonic/gin"

func RegisterUploadRouter(r *gin.RouterGroup, ctrl *UploadController) {
	r.POST("/upload", ctrl.UploadFile)
}
