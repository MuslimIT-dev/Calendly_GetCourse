package service

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type GetMyServicesInput struct{}

type GetMyServicesOutput struct {
	Services []*domain.Service
}

type GetMyServicesUseCase struct {
	deps Deps
}

func NewGetMyServicesUseCase(deps Deps) *GetMyServicesUseCase {
	return &GetMyServicesUseCase{deps: deps}
}

func (uc *GetMyServicesUseCase) Execute(ctx context.Context, _ GetMyServicesInput) (*GetMyServicesOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("services:master:%d", master.ID)

	if cached, err := uc.deps.Cache.Get(ctx, key); err == nil && cached != nil {
		return &GetMyServicesOutput{Services: cached.Services}, nil
	}

	list, err := uc.deps.Services.ListByMaster(ctx, master.ID)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}

	_ = uc.deps.Cache.Set(ctx, key, &domain.CachedServices{
		MasterID: master.ID,
		Services: list,
	}, uc.deps.CacheTTL)

	return &GetMyServicesOutput{Services: list}, nil
}