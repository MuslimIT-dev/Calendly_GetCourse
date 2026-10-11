package review

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetInput struct {
	ID int64
}

type GetOutput struct {
	Review *domain.ReviewCard
}

type GetUseCase struct {
	deps Deps
}

func NewGetUseCase(deps Deps) *GetUseCase {
	return &GetUseCase{deps: deps}
}

func (uc *GetUseCase) Execute(ctx context.Context, in GetInput) (*GetOutput, error) {
	card, err := uc.deps.Reviews.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if card == nil || card.Review == nil || card.Review.Status != domain.ReviewPublished {
		return nil, domain.ErrReviewNotFound
	}
	return &GetOutput{Review: card}, nil
}
