package payment

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type GetInput struct {
	ID int64
}

type GetOutput struct {
	Payment *domain.Payment
}

type GetUseCase struct {
	deps Deps
}

func NewGetUseCase(deps Deps) *GetUseCase {
	return &GetUseCase{deps: deps}
}

func (uc *GetUseCase) Execute(ctx context.Context, in GetInput) (*GetOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	card, err := uc.deps.Payments.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if card.Payment.UserID != userID && !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}

	return &GetOutput{Payment: card.Payment}, nil
}
