package service

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type UpdateServiceInput struct {
	ID           int32
	Description  *string
	DurationMins *int32
	Price        *int32
	Locations    []domain.ServiceLocation
	HasLocations bool
}

type UpdateServiceOutput struct {
	Service *domain.Service
}

type UpdateServiceUseCase struct {
	deps Deps
}

func NewUpdateServiceUseCase(deps Deps) *UpdateServiceUseCase {
	return &UpdateServiceUseCase{deps: deps}
}

func (uc *UpdateServiceUseCase) Execute(ctx context.Context, in UpdateServiceInput) (*UpdateServiceOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	svc, err := uc.deps.Services.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if svc.MasterID != master.ID {
		return nil, domain.ErrForbidden
	}

	if in.Description != nil {
		svc.Description = *in.Description
	}
	if in.DurationMins != nil {
		if *in.DurationMins <= 0 {
			return nil, domain.ErrInvalidInput
		}
		svc.DurationMins = *in.DurationMins
	}
	if in.Price != nil {
		if *in.Price < 0 {
			return nil, domain.ErrInvalidInput
		}
		svc.Price = *in.Price
	}

	updated, err := uc.deps.Services.Update(ctx, svc)
	if err != nil {
		return nil, fmt.Errorf("update service: %w", err)
	}

	if in.HasLocations {
		if err := uc.deps.Services.ReplaceLocations(ctx, svc.ID, in.Locations); err != nil {
			return nil, fmt.Errorf("replace locations: %w", err)
		}
		updated.Locations = in.Locations
	}

	_ = uc.deps.Services.UpdateMasterMinPrice(ctx, master.ID)
	invalidate(ctx, uc.deps.Cache, master.ID)

	return &UpdateServiceOutput{Service: updated}, nil
}