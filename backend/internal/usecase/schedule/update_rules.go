package schedule

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type UpdateRulesInput struct {
	LocationID int32
	Rules      []*domain.ScheduleRule
}

type UpdateRulesOutput struct {
	Rules []*domain.ScheduleRule
}

type UpdateRulesUseCase struct {
	deps Deps
}

func NewUpdateRulesUseCase(deps Deps) *UpdateRulesUseCase {
	return &UpdateRulesUseCase{deps: deps}
}

func (uc *UpdateRulesUseCase) Execute(ctx context.Context, in UpdateRulesInput) (*UpdateRulesOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if len(in.Rules) != 7 {
		return nil, domain.ErrInvalidInput
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

	if err := uc.deps.Schedule.DeleteRulesByLocation(ctx, in.LocationID); err != nil {
		return nil, fmt.Errorf("delete rules: %w", err)
	}

	saved := make([]*domain.ScheduleRule, 0, 7)
	for _, r := range in.Rules {
		r.LocationID = in.LocationID
		row, err := uc.deps.Schedule.UpsertRule(ctx, r)
		if err != nil {
			return nil, fmt.Errorf("upsert rule: %w", err)
		}
		saved = append(saved, row)
	}

	invalidateRules(ctx, uc.deps.Cache, in.LocationID)

	return &UpdateRulesOutput{Rules: saved}, nil
}