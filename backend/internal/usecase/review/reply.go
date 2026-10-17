package review

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type ReplyInput struct {
	ID    int64
	Reply string
}

type ReplyOutput struct {
	Review *domain.ReviewCard
}

type ReplyUseCase struct {
	deps   Deps
	events EventPublisher
}

func NewReplyUseCase(deps Deps, events EventPublisher) *ReplyUseCase {
	return &ReplyUseCase{deps: deps, events: events}
}

func (uc *ReplyUseCase) Execute(ctx context.Context, in ReplyInput) (*ReplyOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	reply := strings.TrimSpace(in.Reply)
	if reply == "" {
		return nil, domain.ErrInvalidInput
	}
	card, err := uc.deps.Reviews.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	review := card.Review
	if review.Status != domain.ReviewPublished {
		return nil, domain.ErrInvalidStatus
	}
	if review.Target != domain.ReviewTargetMaster {
		return nil, domain.ErrForbidden
	}

	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		master, err := uc.deps.Masters.GetByUserID(ctx, userID)
		if errors.Is(err, domain.ErrMasterNotFound) {
			return nil, domain.ErrForbidden
		}
		if err != nil {
			return nil, err
		}
		if master.ID != review.TargetID {
			return nil, domain.ErrForbidden
		}
	}
	updatedReview, err := uc.deps.Reviews.Reply(ctx, in.ID, reply)
	if err != nil {
		return nil, err
	}
	updated, err := uc.deps.Reviews.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if uc.events != nil {
		if err := publishEvent(ctx, uc.events, ReviewRepliedEvent{
			ReviewID: updatedReview.ID,
			AuthorID: updatedReview.AuthorID,
			TargetID: updatedReview.TargetID,
			Reply:    reply,
		}); err != nil {
			log.Printf("publish review-replied event: %v", err)
		}
	}
	return &ReplyOutput{Review: updated}, nil
}
