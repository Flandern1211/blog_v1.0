package auth

import (
	global "gin-blog/internal/global"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/service"
	"gin-blog/internal/utils"
	"gin-blog/pkg/errors"
	"gin-blog/pkg/response"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	svc service.AuthService
}

func NewAuthController(svc service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req request.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}

	ipAddress := utils.IP.GetIpAddress(c)
	ipSource := utils.IP.GetIpSourceSimpleIdle(ipAddress)
	vo, err := ctrl.svc.Login(c.Request.Context(), req, ipAddress, ipSource)
	if err != nil {
		response.BizError(c, err)
		return
	}

	session := sessions.Default(c)
	session.Set(global.CTX_USER_AUTH, vo.ID)
	session.Set(global.CTX_IS_SUPER, vo.IsSuper)
	session.Save()

	response.Success(c, vo)
}

func (ctrl *AuthController) AdminLogin(c *gin.Context) {
	var req request.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}

	ipAddress := utils.IP.GetIpAddress(c)
	ipSource := utils.IP.GetIpSourceSimpleIdle(ipAddress)
	vo, err := ctrl.svc.AdminLogin(c.Request.Context(), req, ipAddress, ipSource)
	if err != nil {
		response.BizError(c, err)
		return
	}

	session := sessions.Default(c)
	session.Set(global.CTX_USER_AUTH, vo.ID)
	session.Set(global.CTX_IS_SUPER, vo.IsSuper)
	session.Save()

	response.Success(c, vo)
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	val := session.Get(global.CTX_USER_AUTH)
	if authId, ok := val.(int); ok {
		session.Delete(global.CTX_USER_AUTH)
		session.Save()

		tokenStr := ""
		auth := c.Request.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = auth[7:]
		}
		ctrl.svc.Logout(c.Request.Context(), authId, tokenStr)
	}

	response.Success(c, nil)
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req request.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}

	if err := ctrl.svc.Register(c.Request.Context(), req); err != nil {
		response.BizError(c, err)
		return
	}

	response.Success(c, nil)
}

func (ctrl *AuthController) SendCode(c *gin.Context) {
	var req request.SendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.CodeRequestError, errors.GetMessage(errors.CodeRequestError))
		return
	}

	if err := ctrl.svc.SendCode(c.Request.Context(), req.Email); err != nil {
		response.BizError(c, err)
		return
	}

	response.Success(c, nil)
}
