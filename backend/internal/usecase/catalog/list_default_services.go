package catalog

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type ListDefaultServicesInput struct {
	CategoryID *int32
}

type ListDefaultServicesOutput struct {
	Services []*domain.DefaultService
}

type ListDefaultServicesUseCase struct {
	deps Deps
}

func NewListDefaultServicesUseCase(deps Deps) *ListDefaultServicesUseCase {
	return &ListDefaultServicesUseCase{deps: deps}
}

func (uc *ListDefaultServicesUseCase) Execute(ctx context.Context, in ListDefaultServicesInput) (*ListDefaultServicesOutput, error) {
	key := "catalog:default_services:all"
	if in.CategoryID != nil {
		key = fmt.Sprintf("catalog:default_services:%d", *in.CategoryID)
	}

	if cached, err := uc.deps.ServicesCache.Get(ctx, key); err == nil && cached != nil {
		return &ListDefaultServicesOutput{Services: cached.Services}, nil
	}

	list, err := uc.deps.Catalog.ListDefaultServices(ctx, in.CategoryID)
	if err != nil {
		return nil, err
	}

	_ = uc.deps.ServicesCache.Set(ctx, key, &domain.CachedDefaultServices{
		CategoryID: in.CategoryID,
		Services:   list,
	}, uc.deps.CacheTTL)

	return &ListDefaultServicesOutput{Services: list}, nil
}
