package catalog

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type DeleteDefaultServiceInput struct {
	ID int32
}

type DeleteDefaultServiceOutput struct{}

type DeleteDefaultServiceUseCase struct {
	deps Deps
}

func NewDeleteDefaultServiceUseCase(deps Deps) *DeleteDefaultServiceUseCase {
	return &DeleteDefaultServiceUseCase{deps: deps}
}

func (uc *DeleteDefaultServiceUseCase) Execute(ctx context.Context, in DeleteDefaultServiceInput) (*DeleteDefaultServiceOutput, error) {
	if !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	s, err := uc.deps.Catalog.GetDefaultService(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.deps.Catalog.DeleteDefaultService(ctx, in.ID); err != nil {
		return nil, err
	}

	invalidateServices(ctx, uc.deps.ServicesCache, &s.CategoryID)

	return &DeleteDefaultServiceOutput{}, nil
}
