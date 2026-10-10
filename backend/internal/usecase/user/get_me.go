package user

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetMeInput struct{}

type GetMeOutput struct {
	User *domain.User
}

type GetMeUseCase struct {
	deps Deps
}

func NewGetMeUseCase(deps Deps) *GetMeUseCase {
	return &GetMeUseCase{deps: deps}
}

func (uc *GetMeUseCase) Execute(ctx context.Context, _ GetMeInput) (*GetMeOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	out, err := NewGetUserUseCase(uc.deps).Execute(ctx, GetUserInput{UserID: userID})
	if err != nil {
		return nil, err
	}
	return &GetMeOutput{User: out.User}, nil
}