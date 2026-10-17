package review

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type HideInput struct {
	ID     int64
	Reason string
}

type HideOutput struct {
	Review *domain.ReviewCard
}

type HideUseCase struct {
	deps Deps
}

func NewHideUseCase(deps Deps) *HideUseCase {
	return &HideUseCase{deps: deps}
}

func (uc *HideUseCase) Execute(ctx context.Context, in HideInput) (*HideOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}
	updated, err := uc.deps.Reviews.Hide(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	invalidateSummary(ctx, uc.deps.Cache, updated.Target, updated.TargetID)
	card, err := uc.deps.Reviews.GetByID(ctx, updated.ID)
	if err != nil {
		return nil, err
	}
	return &HideOutput{Review: card}, nil
}
