package service

import (
	"context"
	global "gin-blog/internal/global"
	"gin-blog/internal/model/dto/request"
	"gin-blog/internal/model/dto/response"
	"gin-blog/internal/model/entity"
	"gin-blog/internal/repository"
	"html/template"
	"strconv"
)

type InteractionService interface {
	// Message
	GetMessageList(ctx context.Context, query request.MessageQuery) ([]entity.Message, int64, error)
	DeleteMessages(ctx context.Context, ids []int) error
	UpdateMessagesReview(ctx context.Context, req request.UpdateReviewReq) error

	// Comment
	GetCommentList(ctx context.Context, query request.CommentQuery) ([]entity.Comment, int64, error)
	DeleteComments(ctx context.Context, ids []int) error
	UpdateCommentsReview(ctx context.Context, req request.UpdateReviewReq) error

	// Front
	AddMessage(ctx context.Context, authId int, req request.FAddMessageReq, ipAddress, ipSource, nickname, avatar string) error
	GetFrontCommentList(ctx context.Context, query request.FCommentQuery) ([]response.CommentVO, int64, error)
	GetCommentReplyList(ctx context.Context, id int, page, size int) ([]response.CommentVO, error)
	AddComment(ctx context.Context, authId int, req request.FAddCommentReq) error
}

type interactionService struct {
	repo         repository.InteractionRepository
	blogInfoRepo repository.BlogInfoRepository
}

func NewInteractionService(repo repository.InteractionRepository, blogInfoRepo repository.BlogInfoRepository) InteractionService {
	return &interactionService{
		repo:         repo,
		blogInfoRepo: blogInfoRepo,
	}
}

// Message implementations
func (s *interactionService) GetMessageList(ctx context.Context, query request.MessageQuery) ([]entity.Message, int64, error) {
	return s.repo.GetMessageList(query.GetPage(), query.GetSize(), query.Nickname, query.IsReview)
}

func (s *interactionService) DeleteMessages(ctx context.Context, ids []int) error {
	return s.repo.DeleteMessages(ids)
}

func (s *interactionService) UpdateMessagesReview(ctx context.Context, req request.UpdateReviewReq) error {
	return s.repo.UpdateMessagesReview(req.Ids, req.IsReview)
}

// Comment implementations
func (s *interactionService) GetCommentList(ctx context.Context, query request.CommentQuery) ([]entity.Comment, int64, error) {
	return s.repo.GetCommentList(query.GetPage(), query.GetSize(), query.Type, query.IsReview, query.Nickname)
}

func (s *interactionService) DeleteComments(ctx context.Context, ids []int) error {
	return s.repo.DeleteComments(ids)
}

func (s *interactionService) UpdateCommentsReview(ctx context.Context, req request.UpdateReviewReq) error {
	return s.repo.UpdateCommentsReview(req.Ids, req.IsReview)
}

// Front implementations

func (s *interactionService) AddMessage(ctx context.Context, authId int, req request.FAddMessageReq, ipAddress, ipSource, nickname, avatar string) error {
	isReview := s.blogInfoRepo.GetConfigBool(global.CONFIG_IS_COMMENT_REVIEW)

	message := &entity.Message{
		Nickname:  nickname,
		Avatar:    avatar,
		Content:   template.HTMLEscapeString(req.Content),
		IpAddress: ipAddress,
		IpSource:  ipSource,
		Speed:     req.Speed,
		IsReview:  isReview,
	}

	return s.repo.SaveMessage(message)
}

func (s *interactionService) GetFrontCommentList(ctx context.Context, query request.FCommentQuery) ([]response.CommentVO, int64, error) {
	comments, replyMap, total, err := s.repo.GetFrontCommentList(query.GetPage(), query.GetSize(), query.TopicId, query.Type)
	if err != nil {
		return nil, 0, err
	}

	likeCountMap, _ := s.repo.GetCommentLikeCountMap(ctx)

	var res []response.CommentVO
	for _, comment := range comments {
		vo := response.CommentVO{
			Comment:    comment,
			LikeCount:  0,
			ReplyCount: len(replyMap[comment.ID]),
		}
		if count, ok := likeCountMap[strconv.Itoa(comment.ID)]; ok {
			vo.LikeCount, _ = strconv.Atoi(count)
		}

		var replies []response.CommentVO
		for _, reply := range replyMap[comment.ID] {
			replyVO := response.CommentVO{
				Comment:   reply,
				LikeCount: 0,
			}
			if count, ok := likeCountMap[strconv.Itoa(reply.ID)]; ok {
				replyVO.LikeCount, _ = strconv.Atoi(count)
			}
			replies = append(replies, replyVO)
		}
		vo.ReplyList = replies
		res = append(res, vo)
	}

	return res, total, nil
}

func (s *interactionService) GetCommentReplyList(ctx context.Context, id int, page, size int) ([]response.CommentVO, error) {
	replies, err := s.repo.GetCommentReplyList(id, page, size)
	if err != nil {
		return nil, err
	}

	likeCountMap, _ := s.repo.GetCommentLikeCountMap(ctx)

	var res []response.CommentVO
	for _, reply := range replies {
		vo := response.CommentVO{
			Comment:   reply,
			LikeCount: 0,
		}
		if count, ok := likeCountMap[strconv.Itoa(reply.ID)]; ok {
			vo.LikeCount, _ = strconv.Atoi(count)
		}
		res = append(res, vo)
	}

	return res, nil
}

func (s *interactionService) AddComment(ctx context.Context, authId int, req request.FAddCommentReq) error {
	isReview := s.blogInfoRepo.GetConfigBool(global.CONFIG_IS_COMMENT_REVIEW)

	comment := &entity.Comment{
		UserId:      authId,
		ReplyUserId: req.ReplyUserId,
		TopicId:     req.TopicId,
		ParentId:    req.ParentId,
		Content:     template.HTMLEscapeString(req.Content),
		Type:        req.Type,
		IsReview:    isReview,
	}

	return s.repo.AddComment(comment)
}
