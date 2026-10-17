package connect

import (
	"context"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	reviewv1 "github.com/MuslimIT-dev/zaman/backend/gen/go/review/v1"
	reviewv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/review/v1/reviewv1connect"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	reviewuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/review"
)

type ReviewHandler struct {
	createUC     *reviewuc.CreateUseCase
	getUC        *reviewuc.GetUseCase
	listTargetUC *reviewuc.ListTargetUseCase
	listMyUC     *reviewuc.ListMyUseCase
	replyUC      *reviewuc.ReplyUseCase
	updateUC     *reviewuc.UpdateUseCase
	deleteUC     *reviewuc.DeleteUseCase
	hideUC       *reviewuc.HideUseCase
	summaryUC    *reviewuc.GetSummaryUseCase
}

func NewReviewHandler(
	createUC *reviewuc.CreateUseCase,
	getUC *reviewuc.GetUseCase,
	listTargetUC *reviewuc.ListTargetUseCase,
	listMyUC *reviewuc.ListMyUseCase,
	replyUC *reviewuc.ReplyUseCase,
	updateUC *reviewuc.UpdateUseCase,
	deleteUC *reviewuc.DeleteUseCase,
	hideUC *reviewuc.HideUseCase,
	summaryUC *reviewuc.GetSummaryUseCase,
) *ReviewHandler {
	return &ReviewHandler{
		createUC: createUC, getUC: getUC, listTargetUC: listTargetUC, listMyUC: listMyUC,
		replyUC: replyUC, updateUC: updateUC, deleteUC: deleteUC, hideUC: hideUC, summaryUC: summaryUC,
	}
}

var _ reviewv1connect.ReviewServiceHandler = (*ReviewHandler)(nil)

func (h *ReviewHandler) CreateReview(ctx context.Context, req *connect.Request[reviewv1.CreateReviewRequest]) (*connect.Response[reviewv1.CreateReviewResponse], error) {
	targetID, err := parseID(req.Msg.TargetId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.createUC.Execute(ctx, reviewuc.CreateInput{
		Target: domain.ReviewTarget(req.Msg.Target), TargetID: targetID,
		ReferenceID: req.Msg.ReferenceId, Rating: req.Msg.Rating, Comment: req.Msg.Comment,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.CreateReviewResponse{Review: toProtoReview(out.Review)}), nil
}

func (h *ReviewHandler) GetReview(ctx context.Context, req *connect.Request[reviewv1.GetReviewRequest]) (*connect.Response[reviewv1.GetReviewResponse], error) {
	id, err := parseReviewID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.getUC.Execute(ctx, reviewuc.GetInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.GetReviewResponse{Review: toProtoReview(out.Review)}), nil
}

func (h *ReviewHandler) ListTargetReviews(ctx context.Context, req *connect.Request[reviewv1.ListTargetReviewsRequest]) (*connect.Response[reviewv1.ListTargetReviewsResponse], error) {
	targetID, err := parseID(req.Msg.TargetId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.listTargetUC.Execute(ctx, reviewuc.ListTargetInput{
		Target: domain.ReviewTarget(req.Msg.Target), TargetID: targetID,
		PageSize: req.Msg.PageSize, PageToken: req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	reviews := make([]*reviewv1.Review, len(out.Reviews))
	for i, review := range out.Reviews {
		reviews[i] = toProtoReview(review)
	}
	return connect.NewResponse(&reviewv1.ListTargetReviewsResponse{
		Reviews: reviews, NextPageToken: out.NextPageToken, Summary: toProtoSummary(out.Summary),
	}), nil
}

func (h *ReviewHandler) ListMyReviews(ctx context.Context, req *connect.Request[reviewv1.ListMyReviewsRequest]) (*connect.Response[reviewv1.ListMyReviewsResponse], error) {
	var filter *domain.ReviewTarget
	if req.Msg.TargetFilter != reviewv1.ReviewTarget_REVIEW_TARGET_UNSPECIFIED {
		target := domain.ReviewTarget(req.Msg.TargetFilter)
		filter = &target
	}
	out, err := h.listMyUC.Execute(ctx, reviewuc.ListMyInput{
		TargetFilter: filter, PageSize: req.Msg.PageSize, PageToken: req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	reviews := make([]*reviewv1.Review, len(out.Reviews))
	for i, review := range out.Reviews {
		reviews[i] = toProtoReview(review)
	}
	return connect.NewResponse(&reviewv1.ListMyReviewsResponse{Reviews: reviews, NextPageToken: out.NextPageToken}), nil
}

func (h *ReviewHandler) ReplyToReview(ctx context.Context, req *connect.Request[reviewv1.ReplyToReviewRequest]) (*connect.Response[reviewv1.ReplyToReviewResponse], error) {
	id, err := parseReviewID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.replyUC.Execute(ctx, reviewuc.ReplyInput{ID: id, Reply: req.Msg.Reply})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.ReplyToReviewResponse{Review: toProtoReview(out.Review)}), nil
}

func (h *ReviewHandler) UpdateReview(ctx context.Context, req *connect.Request[reviewv1.UpdateReviewRequest]) (*connect.Response[reviewv1.UpdateReviewResponse], error) {
	id, err := parseReviewID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.updateUC.Execute(ctx, reviewuc.UpdateInput{
		ID: id, Rating: req.Msg.Rating, Comment: req.Msg.Comment,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.UpdateReviewResponse{Review: toProtoReview(out.Review)}), nil
}

func (h *ReviewHandler) DeleteReview(ctx context.Context, req *connect.Request[reviewv1.DeleteReviewRequest]) (*connect.Response[reviewv1.DeleteReviewResponse], error) {
	id, err := parseReviewID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if _, err := h.deleteUC.Execute(ctx, reviewuc.DeleteInput{ID: id}); err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.DeleteReviewResponse{}), nil
}

func (h *ReviewHandler) HideReview(ctx context.Context, req *connect.Request[reviewv1.HideReviewRequest]) (*connect.Response[reviewv1.HideReviewResponse], error) {
	id, err := parseReviewID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.hideUC.Execute(ctx, reviewuc.HideInput{ID: id, Reason: req.Msg.Reason})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.HideReviewResponse{Review: toProtoReview(out.Review)}), nil
}

func (h *ReviewHandler) GetReviewSummary(ctx context.Context, req *connect.Request[reviewv1.GetReviewSummaryRequest]) (*connect.Response[reviewv1.GetReviewSummaryResponse], error) {
	targetID, err := parseID(req.Msg.TargetId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	out, err := h.summaryUC.Execute(ctx, reviewuc.GetSummaryInput{
		Target: domain.ReviewTarget(req.Msg.Target), TargetID: targetID,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&reviewv1.GetReviewSummaryResponse{Summary: toProtoSummary(out.Summary)}), nil
}

func parseReviewID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid review ID %q", value)
	}
	return id, nil
}

func toProtoReview(card *domain.ReviewCard) *reviewv1.Review {
	if card == nil || card.Review == nil {
		return nil
	}
	review := card.Review
	var repliedAt *timestamppb.Timestamp
	if review.RepliedAt != nil {
		repliedAt = timestamppb.New(*review.RepliedAt)
	}
	return &reviewv1.Review{
		Id: strconv.FormatInt(review.ID, 10), AuthorId: strconv.FormatInt(int64(review.AuthorID), 10),
		TargetId: strconv.FormatInt(int64(review.TargetID), 10), Target: reviewv1.ReviewTarget(review.Target),
		ReferenceId: review.ReferenceID, Rating: review.Rating, Comment: review.Comment, Reply: review.Reply,
		RepliedAt: repliedAt, Status: reviewv1.ReviewStatus(review.Status), AuthorName: card.AuthorName,
		AuthorAvatar: card.AuthorAvatar, CreatedAt: timestamppb.New(review.CreatedAt), UpdatedAt: timestamppb.New(review.UpdatedAt),
	}
}

func toProtoSummary(summary *domain.ReviewSummary) *reviewv1.ReviewSummary {
	if summary == nil {
		return nil
	}
	return &reviewv1.ReviewSummary{
		TargetId: strconv.FormatInt(int64(summary.TargetID), 10), Target: reviewv1.ReviewTarget(summary.Target),
		AvgRating: summary.AvgRating, TotalReviews: summary.TotalReviews,
		Rating_1Count: summary.Rating1Count, Rating_2Count: summary.Rating2Count,
		Rating_3Count: summary.Rating3Count, Rating_4Count: summary.Rating4Count, Rating_5Count: summary.Rating5Count,
	}
}
