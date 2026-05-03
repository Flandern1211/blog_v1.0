package middleware

import (
	"fmt"
	"gin-blog/internal/service"
	pkgErrors "gin-blog/pkg/errors"
	"gin-blog/pkg/response"
	"log/slog"

	"github.com/gin-gonic/gin"
)

// 监听在线状态中间件
// 登录时: 移除用户的强制下线标记
// 退出登录时: 添加用户的在线标记
func ListenOnline(authSvc service.AuthService, userSvc service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, err := CurrentUserAuth(c, authSvc)
		if err != nil {
			response.BizError(c, pkgErrors.NewWithErr(pkgErrors.CodeUserAuthError, pkgErrors.GetMessage(pkgErrors.CodeUserAuthError), err))
			return
		}

		// 判断当前用户是否被强制下线
		offline, err := userSvc.CheckUserOffline(c.Request.Context(), auth.ID)
		if err != nil {
			response.BizError(c, pkgErrors.NewWithErr(pkgErrors.CodeDbOpError, pkgErrors.GetMessage(pkgErrors.CodeDbOpError), err))
			c.Abort()
			return
		}
		if offline {
			fmt.Println("用户被强制下线")
			response.Error(c, pkgErrors.CodeForceOffline, pkgErrors.GetMessage(pkgErrors.CodeForceOffline))
			c.Abort()
			return
		}

		// 每次发送请求会更新 Redis 中的在线状态: 重新计算 10 分钟
		if err := userSvc.SetOnlineUser(c.Request.Context(), auth); err != nil {
			slog.Error("更新在线状态失败", "err", err)
		}
		c.Next()
	}
}
