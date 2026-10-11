package service

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type DeleteServiceInput struct {
	ID int32
}

type DeleteServiceOutput struct{}

type DeleteServiceUseCase struct {
	deps Deps
}

func NewDeleteServiceUseCase(deps Deps) *DeleteServiceUseCase {
	return &DeleteServiceUseCase{deps: deps}
}

func (uc *DeleteServiceUseCase) Execute(ctx context.Context, in DeleteServiceInput) (*DeleteServiceOutput, error) {
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

	if err := uc.deps.Services.Delete(ctx, in.ID); err != nil {
		return nil, fmt.Errorf("delete service: %w", err)
	}

	_ = uc.deps.Services.UpdateMasterMinPrice(ctx, master.ID)
	invalidate(ctx, uc.deps.Cache, master.ID)

	return &DeleteServiceOutput{}, nil
}