package location

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ListLocationsInput struct {
	MasterID   int32
	OnlyActive bool
}

type ListLocationsOutput struct {
	Locations []*domain.Location
}

type ListLocationsUseCase struct {
	deps Deps
}

func NewListLocationsUseCase(deps Deps) *ListLocationsUseCase {
	return &ListLocationsUseCase{deps: deps}
}

func (uc *ListLocationsUseCase) Execute(ctx context.Context, in ListLocationsInput) (*ListLocationsOutput, error) {
	if in.MasterID == 0 {
		return nil, domain.ErrInvalidInput
	}

	if in.OnlyActive {
		key := fmt.Sprintf("locations:master:%d:active", in.MasterID)

		if cached, err := uc.deps.Cache.Get(ctx, key); err == nil && cached != nil {
			return &ListLocationsOutput{Locations: cached.Locations}, nil
		}

		list, err := uc.deps.Locations.ListByMaster(ctx, in.MasterID, true)
		if err != nil {
			return nil, err
		}

		_ = uc.deps.Cache.Set(ctx, key, &domain.CachedLocations{
			MasterID:  in.MasterID,
			Locations: list,
		}, uc.deps.CacheTTL)

		return &ListLocationsOutput{Locations: list}, nil
	}

	list, err := uc.deps.Locations.ListByMaster(ctx, in.MasterID, false)
	if err != nil {
		return nil, err
	}
	return &ListLocationsOutput{Locations: list}, nil
}