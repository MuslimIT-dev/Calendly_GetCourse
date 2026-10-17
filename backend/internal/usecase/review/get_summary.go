package review

import (
	"context"
	"log"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type GetSummaryInput struct {
	Target   domain.ReviewTarget
	TargetID int32
}

type GetSummaryOutput struct {
	Summary *domain.ReviewSummary
}

type GetSummaryUseCase struct {
	deps Deps
}

func NewGetSummaryUseCase(deps Deps) *GetSummaryUseCase {
	return &GetSummaryUseCase{deps: deps}
}

func (uc *GetSummaryUseCase) Execute(ctx context.Context, in GetSummaryInput) (*GetSummaryOutput, error) {
	if !validTarget(in.Target) || in.TargetID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	summary, err := getCachedSummary(ctx, uc.deps, in.Target, in.TargetID)
	if err != nil {
		return nil, err
	}
	return &GetSummaryOutput{Summary: summary}, nil
}

func getCachedSummary(ctx context.Context, deps Deps, target domain.ReviewTarget, targetID int32) (*domain.ReviewSummary, error) {
	key := summaryKey(target, targetID)
	cached, err := deps.Cache.Get(ctx, key)
	if err == nil && cached != nil {
		return cached.Summary, nil
	}
	if err != nil {
		log.Printf("read review summary cache: %v", err)
	}
	summary, err := deps.Reviews.GetSummary(ctx, target, targetID)
	if err != nil {
		return nil, err
	}
	if err := deps.Cache.Set(ctx, key, &domain.CachedReviewSummary{Summary: summary}, deps.CacheTTL); err != nil {
		log.Printf("write review summary cache: %v", err)
	}
	return summary, nil
}
