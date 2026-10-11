package review

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type UpdateInput struct {
	ID      int64
	Rating  int32
	Comment string
}

type UpdateOutput struct {
	Review *domain.ReviewCard
}

type UpdateUseCase struct {
	deps Deps
}

func NewUpdateUseCase(deps Deps) *UpdateUseCase {
	return &UpdateUseCase{deps: deps}
}

func (uc *UpdateUseCase) Execute(ctx context.Context, in UpdateInput) (*UpdateOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	if in.Rating < 1 || in.Rating > 5 {
		return nil, domain.ErrInvalidInput
	}
	card, err := uc.deps.Reviews.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	review := card.Review
	if review.AuthorID != userID && !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}
	if review.Status != domain.ReviewPublished {
		return nil, domain.ErrInvalidStatus
	}
	review.Rating = in.Rating
	review.Comment = in.Comment
	updated, err := uc.deps.Reviews.Update(ctx, review)
	if err != nil {
		return nil, err
	}
	invalidateSummary(ctx, uc.deps.Cache, updated.Target, updated.TargetID)
	updatedCard, err := uc.deps.Reviews.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}
	return &UpdateOutput{Review: updatedCard}, nil
}
