package service

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type AddServiceInput struct {
	DefaultID      int32
	NewDescription string
	DurationMins   int32
	Price          int32
	Locations      []domain.ServiceLocation
}

type AddServiceOutput struct {
	Service *domain.Service
}

type AddServiceUseCase struct {
	deps Deps
}

func NewAddServiceUseCase(deps Deps) *AddServiceUseCase {
	return &AddServiceUseCase{deps: deps}
}

func (uc *AddServiceUseCase) Execute(ctx context.Context, in AddServiceInput) (*AddServiceOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if in.DefaultID == 0 {
		return nil, domain.ErrInvalidInput
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	exists, err := uc.deps.Services.ExistsByDefaultID(ctx, master.ID, in.DefaultID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrServiceAlreadyAdded
	}

	tpl, err := uc.deps.Services.GetDefaultTemplate(ctx, in.DefaultID)
	if err != nil {
		return nil, err
	}

	price := tpl.Price
	if in.Price > 0 {
		price = in.Price
	}

	duration := tpl.DurationMins
	if in.DurationMins > 0 {
		duration = in.DurationMins
	}

	description := tpl.Description
	if in.NewDescription != "" {
		description = in.NewDescription
	}

	defaultID := in.DefaultID

	svc, err := uc.deps.Services.Create(ctx, &domain.Service{
		MasterID:         master.ID,
		DefaultServiceID: &defaultID,
		Slug:             tpl.Slug,
		Name:             tpl.Name,
		Description:      description,
		Price:            price,
		Currency:         tpl.Currency,
		DurationMins:     duration,
		IsActive:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	if len(in.Locations) > 0 {
		if err := uc.deps.Services.ReplaceLocations(ctx, svc.ID, in.Locations); err != nil {
			return nil, fmt.Errorf("replace locations: %w", err)
		}
		svc.Locations = in.Locations
	}

	_ = uc.deps.Services.UpdateMasterMinPrice(ctx, master.ID)
	invalidate(ctx, uc.deps.Cache, master.ID)

	return &AddServiceOutput{Service: svc}, nil
}