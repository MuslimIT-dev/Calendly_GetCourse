package master

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetMyProfileInput struct{}

type GetMyProfileOutput struct {
	Profile *domain.MasterProfile
}

type GetMyProfileUseCase struct {
	deps Deps
}

func NewGetMyProfileUseCase(deps Deps) *GetMyProfileUseCase {
	return &GetMyProfileUseCase{deps: deps}
}

func (uc *GetMyProfileUseCase) Execute(ctx context.Context, _ GetMyProfileInput) (*GetMyProfileOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	key := fmt.Sprintf("master:user:%d", userID)

	if cached, err := uc.deps.Cache.Get(ctx, key); err == nil && cached != nil {
		return &GetMyProfileOutput{Profile: cached.Profile}, nil
	}

	profile, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = uc.deps.Cache.Set(ctx, key, &domain.CachedMaster{Profile: profile}, uc.deps.CacheTTL)

	return &GetMyProfileOutput{Profile: profile}, nil
}