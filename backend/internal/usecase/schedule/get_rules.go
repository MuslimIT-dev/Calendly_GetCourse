package schedule

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type GetRulesInput struct {
	LocationID int32
}

type GetRulesOutput struct {
	Rules []*domain.ScheduleRule
}

type GetRulesUseCase struct {
	deps Deps
}

func NewGetRulesUseCase(deps Deps) *GetRulesUseCase {
	return &GetRulesUseCase{deps: deps}
}

func (uc *GetRulesUseCase) Execute(ctx context.Context, in GetRulesInput) (*GetRulesOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	belongs, err := uc.deps.Locations.BelongsToMaster(ctx, in.LocationID, master.ID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, domain.ErrForbidden
	}

	key := fmt.Sprintf("schedule:rules:%d", in.LocationID)
	if cached, err := uc.deps.Cache.Get(ctx, key); err == nil && cached != nil {
		return &GetRulesOutput{Rules: cached.Rules}, nil
	}

	rules, err := uc.deps.Schedule.GetRulesByLocation(ctx, in.LocationID)
	if err != nil {
		return nil, fmt.Errorf("get rules: %w", err)
	}

	_ = uc.deps.Cache.Set(ctx, key, &domain.CachedScheduleRules{
		LocationID: in.LocationID,
		Rules:      rules,
	}, uc.deps.CacheTTL)

	return &GetRulesOutput{Rules: rules}, nil
}
