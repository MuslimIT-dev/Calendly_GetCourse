package schedule

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type DeleteExceptionInput struct {
	ExceptionID int32
}

type DeleteExceptionOutput struct{}

type DeleteExceptionUseCase struct {
	deps Deps
}

func NewDeleteExceptionUseCase(deps Deps) *DeleteExceptionUseCase {
	return &DeleteExceptionUseCase{deps: deps}
}

func (uc *DeleteExceptionUseCase) Execute(ctx context.Context, in DeleteExceptionInput) (*DeleteExceptionOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	_, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := uc.deps.Schedule.DeleteException(ctx, in.ExceptionID); err != nil {
		return nil, fmt.Errorf("delete exception: %w", err)
	}
	return &DeleteExceptionOutput{}, nil
}