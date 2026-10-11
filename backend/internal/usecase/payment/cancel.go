package payment

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CancelInput struct {
	ID int64
}

type CancelOutput struct {
	Payment *domain.Payment
}

type CancelUseCase struct {
	deps Deps
}

func NewCancelUseCase(deps Deps) *CancelUseCase {
	return &CancelUseCase{deps: deps}
}

func (uc *CancelUseCase) Execute(ctx context.Context, in CancelInput) (*CancelOutput, error) {
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
	if card.Payment.Status != domain.PaymentPending {
		return nil, domain.ErrPaymentInvalidStatus
	}

	updated, err := uc.deps.Payments.Cancel(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	return &CancelOutput{Payment: updated}, nil
}
