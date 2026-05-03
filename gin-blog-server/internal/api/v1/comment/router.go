package comment

import (
	"github.com/gin-gonic/gin"
)

func RegisterCommentRouter(r *gin.RouterGroup, ctrl *CommentController) {
	cmt := r.Group("/comment")
	{
		cmt.GET("/list", ctrl.GetList)
		cmt.PUT("/review", ctrl.UpdateReview)
		cmt.DELETE("", ctrl.Delete)
	}
}
