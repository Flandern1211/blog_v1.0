package service

import (
	"context"
	"errors"
	g "gin-blog/internal/global"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"
	"gin-blog/internal/utils"
	pkgErrors "gin-blog/pkg/errors"
	"gin-blog/pkg/jwt"
	"time"

	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req request.LoginReq, ipAddress, ipSource string) (*response.LoginVO, error)
	AdminLogin(ctx context.Context, req request.LoginReq, ipAddress, ipSource string) (*response.LoginVO, error)
	Register(ctx context.Context, req request.RegisterReq) error
	VerifyCode(ctx context.Context, code string) error
	Logout(ctx context.Context, authId int, tokenStr string) error
	SendCode(ctx context.Context, email string) error
	GetUserAuthById(ctx context.Context, id int) (*entity.UserAuth, error)
	CheckTokenExists(ctx context.Context, tokenStr string) bool
	GetResource(url, method string) (*entity.Resource, error)
	CheckRoleAuth(roleId int, url, method string) (bool, error)
	CheckUserHasResource(userId int, url, method string) (bool, error)
}

type authService struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
}

func NewAuthService(authRepo repository.AuthRepository, userRepo repository.UserRepository) AuthService {
	return &authService{authRepo: authRepo, userRepo: userRepo}
}

func (s *authService) doLogin(ctx context.Context, req request.LoginReq, ipAddress, ipSource string) (*entity.UserAuth, *entity.UserInfo, []int, string, string, error) {
	userAuth, err := s.authRepo.GetUserAuthInfoByName(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeUserNotFound)
		}
		return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}

	if userAuth.IsDisable {
		return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeUserDisabled)
	}

	if !utils.BcryptCheck(req.Password, userAuth.Password) {
		return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeInvalidCredentials)
	}

	userInfo, err := s.authRepo.GetUserInfoById(userAuth.UserInfoId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeUserNotFound)
		}
		return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}

	roleIds, err := s.authRepo.GetRoleIdsByUserId(userAuth.ID)
	if err != nil {
		return nil, nil, nil, "", "", pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}

	return userAuth, userInfo, roleIds, ipAddress, ipSource, nil
}

func (s *authService) buildLoginVO(ctx context.Context, userAuth *entity.UserAuth, userInfo *entity.UserInfo, roleIds []int, ipAddress, ipSource string) (*response.LoginVO, error) {
	articleLikeSet, _ := s.userRepo.GetArticleLikeSet(ctx, userAuth.ID)
	commentLikeSet, _ := s.userRepo.GetCommentLikeSet(ctx, userAuth.ID)

	conf := g.Conf.JWT
	token, err := jwt.GenerateToken(conf.Secret, conf.Issuer, int(conf.Expire), userAuth.ID, roleIds)
	if err != nil {
		return nil, pkgErrors.NewDefault(pkgErrors.CodeTokenCreateErr)
	}

	tokenKey := utils.MD5(token)
	if err := s.authRepo.SetToken(ctx, tokenKey, userAuth.ID, time.Duration(conf.Expire)*time.Hour); err != nil {
		return nil, pkgErrors.NewDefault(pkgErrors.CodeRedisOpError)
	}

	err = s.authRepo.UpdateUserLoginInfo(userAuth.ID, ipAddress, ipSource)
	if err != nil {
		return nil, pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}

	s.authRepo.DelOfflineMark(ctx, userAuth.ID)

	return &response.LoginVO{
		UserInfo:       *userInfo,
		ArticleLikeSet: articleLikeSet,
		CommentLikeSet: commentLikeSet,
		Token:          token,
		IsSuper:        userAuth.IsSuper,
	}, nil
}

func (s *authService) Login(ctx context.Context, req request.LoginReq, ipAddress, ipSource string) (*response.LoginVO, error) {
	userAuth, userInfo, roleIds, ipAddress, ipSource, err := s.doLogin(ctx, req, ipAddress, ipSource)
	if err != nil {
		return nil, err
	}
	return s.buildLoginVO(ctx, userAuth, userInfo, roleIds, ipAddress, ipSource)
}

func (s *authService) AdminLogin(ctx context.Context, req request.LoginReq, ipAddress, ipSource string) (*response.LoginVO, error) {
	userAuth, userInfo, roleIds, ipAddress, ipSource, err := s.doLogin(ctx, req, ipAddress, ipSource)
	if err != nil {
		return nil, err
	}

	if !userAuth.IsSuper {
		hasResource, err := s.authRepo.CheckUserHasResource(userAuth.ID, g.RESOURCE_BACKEND_LOGIN, g.METHOD_BACKEND_LOGIN)
		if err != nil {
			return nil, pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
		}
		if !hasResource {
			return nil, pkgErrors.NewDefault(pkgErrors.CodeNoAdminAccess)
		}
	}

	return s.buildLoginVO(ctx, userAuth, userInfo, roleIds, ipAddress, ipSource)
}

func (s *authService) Logout(ctx context.Context, authId int, tokenStr string) error {
	s.userRepo.DelOnlineUser(ctx, authId)
	if tokenStr != "" {
		s.authRepo.DelToken(ctx, utils.MD5(tokenStr))
	}
	return nil
}

func (s *authService) GetUserAuthById(ctx context.Context, id int) (*entity.UserAuth, error) {
	return s.authRepo.GetUserAuthInfoById(id)
}

func (s *authService) SendCode(ctx context.Context, email string) error {
	code := utils.RandomCode(6)
	if err := s.authRepo.SetEmailCode(ctx, email, code, 15*time.Minute); err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeRedisOpError)
	}

	err := utils.SendCodeEmail(email, &utils.EmailData{
		UserName: email,
		Subject:  "注册验证码",
		Code:     code,
	})
	if err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeSendEmailErr)
	}

	return nil
}

func (s *authService) Register(ctx context.Context, req request.RegisterReq) error {
	req.Email = utils.Format(req.Email)

	auth, err := s.authRepo.GetUserAuthInfoByName(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}
	if auth != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeEmailExist)
	}

	info := utils.GenEmailVerificationInfo(req.Email, req.Password)
	if err := s.authRepo.SetVerificationInfo(ctx, info, 15*time.Minute); err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeRedisOpError)
	}

	emailData := utils.GetEmailData(req.Email, info)
	err = utils.SendEmail(req.Email, emailData)
	if err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeSendEmailErr)
	}
	return nil
}

func (s *authService) VerifyCode(ctx context.Context, code string) error {
	val, err := s.authRepo.GetVerificationInfo(ctx, code)
	if err != nil || val == "" {
		return pkgErrors.NewDefault(pkgErrors.CodeCodeWrong)
	}
	s.authRepo.DelVerificationInfo(ctx, code)

	username, password, err := utils.ParseEmailVerificationInfo(code)
	if err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeCodeWrong)
	}

	_, _, _, err = s.authRepo.CreateNewUser(username, username, password)
	if err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}

	return nil
}

func (s *authService) CheckTokenExists(ctx context.Context, tokenStr string) bool {
	return s.authRepo.TokenExists(ctx, tokenStr)
}

func (s *authService) GetResource(url, method string) (*entity.Resource, error) {
	return s.authRepo.GetResource(url, method)
}

func (s *authService) CheckRoleAuth(roleId int, url, method string) (bool, error) {
	return s.authRepo.CheckRoleAuth(roleId, url, method)
}

func (s *authService) CheckUserHasResource(userId int, url, method string) (bool, error) {
	return s.authRepo.CheckUserHasResource(userId, url, method)
}
