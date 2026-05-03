package service

import (
	"context"
	"errors"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"
	"gin-blog/internal/utils"
	pkgErrors "gin-blog/pkg/errors"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

type UserService interface {
	GetInfo(ctx context.Context, authId int) (*response.UserInfoVO, error)
	UpdateCurrent(ctx context.Context, authId int, req request.UpdateCurrentUserReq) error
	Update(ctx context.Context, req request.UpdateUserReq) error
	UpdateDisable(ctx context.Context, req request.UpdateUserDisableReq) error
	GetList(ctx context.Context, query request.UserQuery) ([]response.UserVO, int64, error)
	GetOnlineList(ctx context.Context, keyword string) ([]*entity.UserAuth, error)
	ForceOffline(ctx context.Context, currentAuthId, targetUserId int) error
	UpdatePasswordByCode(ctx context.Context, authId int, req request.UpdatePasswordByCodeReq) error
	CheckUserOffline(ctx context.Context, userId int) (bool, error)
	SetOnlineUser(ctx context.Context, auth *entity.UserAuth) error
}

type userService struct {
	userRepo repository.UserRepository
	authRepo repository.AuthRepository
}

func NewUserService(userRepo repository.UserRepository, authRepo repository.AuthRepository) UserService {
	return &userService{userRepo: userRepo, authRepo: authRepo}
}

func (s *userService) GetInfo(ctx context.Context, authId int) (*response.UserInfoVO, error) {
	userAuth, err := s.userRepo.GetInfoById(authId)
	if err != nil {
		return nil, err
	}

	articleLikeSet, _ := s.userRepo.GetArticleLikeSet(ctx, authId)
	commentLikeSet, _ := s.userRepo.GetCommentLikeSet(ctx, authId)

	return &response.UserInfoVO{
		UserInfo:       *userAuth.UserInfo,
		ArticleLikeSet: articleLikeSet,
		CommentLikeSet: commentLikeSet,
	}, nil
}

func (s *userService) UpdateCurrent(ctx context.Context, authId int, req request.UpdateCurrentUserReq) error {
	userAuth, err := s.userRepo.GetInfoById(authId)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateUserInfo(userAuth.UserInfoId, req.Nickname, req.Avatar, req.Intro, req.Website)
}

func (s *userService) Update(ctx context.Context, req request.UpdateUserReq) error {
	return s.userRepo.UpdateUserNicknameAndRole(req.UserAuthId, req.Nickname, req.RoleIds)
}

func (s *userService) UpdateDisable(ctx context.Context, req request.UpdateUserDisableReq) error {
	if err := s.userRepo.UpdateUserDisable(req.UserAuthId, req.IsDisable); err != nil {
		return err
	}

	if req.IsDisable {
		s.userRepo.DelOnlineUser(ctx, req.UserAuthId)
		s.userRepo.SetOfflineMark(ctx, req.UserAuthId, time.Hour)
	}

	return nil
}

func (s *userService) GetList(ctx context.Context, query request.UserQuery) ([]response.UserVO, int64, error) {
	list, total, err := s.userRepo.GetList(query.GetPage(), query.GetSize(), query.LoginType, query.Nickname, query.Username)
	if err != nil {
		return nil, 0, err
	}

	var res []response.UserVO
	for _, user := range list {
		res = append(res, response.UserVO{
			ID:            user.ID,
			UserInfoId:    user.UserInfoId,
			Info:          user.UserInfo,
			Roles:         user.Roles,
			LoginType:     user.LoginType,
			IpAddress:     user.IpAddress,
			IpSource:      user.IpSource,
			CreatedAt:     user.CreatedAt,
			LastLoginTime: user.LastLoginTime,
			IsDisable:     user.IsDisable,
		})
	}
	return res, total, nil
}

func (s *userService) GetOnlineList(ctx context.Context, keyword string) ([]*entity.UserAuth, error) {
	onlineList, err := s.userRepo.GetOnlineUsers(ctx, keyword)
	if err != nil {
		return nil, err
	}

	if keyword != "" {
		var filtered []*entity.UserAuth
		for _, auth := range onlineList {
			if strings.Contains(auth.Username, keyword) ||
				(auth.UserInfo != nil && strings.Contains(auth.UserInfo.Nickname, keyword)) {
				filtered = append(filtered, auth)
			}
		}
		onlineList = filtered
	}

	sort.Slice(onlineList, func(i, j int) bool {
		if onlineList[i].LastLoginTime == nil || onlineList[j].LastLoginTime == nil {
			return false
		}
		return onlineList[i].LastLoginTime.Unix() > onlineList[j].LastLoginTime.Unix()
	})

	return onlineList, nil
}

func (s *userService) ForceOffline(ctx context.Context, currentAuthId, targetUserId int) error {
	s.userRepo.DelOnlineUser(ctx, targetUserId)
	s.userRepo.SetOfflineMark(ctx, targetUserId, time.Hour)
	return nil
}

func (s *userService) UpdatePasswordByCode(ctx context.Context, authId int, req request.UpdatePasswordByCodeReq) error {
	userAuth, err := s.userRepo.GetInfoById(authId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkgErrors.NewDefault(pkgErrors.CodeUserNotFound)
		}
		return pkgErrors.NewDefault(pkgErrors.CodeDbOpError)
	}

	if userAuth.UserInfo == nil || userAuth.UserInfo.Email != req.Email {
		return pkgErrors.NewDefault(pkgErrors.CodeBadRequest)
	}

	storedCode, err := s.authRepo.GetEmailCode(ctx, req.Email)
	if err != nil || storedCode == "" {
		return pkgErrors.NewDefault(pkgErrors.CodeCodeWrong)
	}
	if storedCode != req.Code {
		return pkgErrors.NewDefault(pkgErrors.CodeCodeWrong)
	}

	s.authRepo.DelEmailCode(ctx, req.Email)

	hashedPassword, err := utils.BcryptHash(req.Password)
	if err != nil {
		return pkgErrors.NewDefault(pkgErrors.CodeInternalError)
	}

	return s.userRepo.UpdateUserPassword(authId, hashedPassword)
}

func (s *userService) CheckUserOffline(ctx context.Context, userId int) (bool, error) {
	return s.userRepo.CheckOffline(ctx, userId)
}

func (s *userService) SetOnlineUser(ctx context.Context, auth *entity.UserAuth) error {
	return s.userRepo.SetOnlineUser(ctx, auth, 10*time.Minute)
}
