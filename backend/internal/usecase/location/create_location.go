package location

import (
	"context"
	"errors"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type CreateLocationInput struct {
	Name       string
	Address    string
	Timezone   string
	IsOnline   bool
	MeetingURL string
}

type CreateLocationOutput struct {
	Location *domain.Location
}

type CreateLocationUseCase struct {
	deps Deps
}

func NewCreateLocationUseCase(deps Deps) *CreateLocationUseCase {
	return &CreateLocationUseCase{deps: deps}
}

func (uc *CreateLocationUseCase) Execute(ctx context.Context, in CreateLocationInput) (*CreateLocationOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if in.Name == "" || in.Timezone == "" {
		return nil, domain.ErrInvalidInput
	}

	if in.IsOnline && in.MeetingURL == "" {
		return nil, errors.New("meeting_url required for online location")
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc, err := uc.deps.Locations.Create(ctx, &domain.Location{
		MasterID:   master.ID,
		Name:       in.Name,
		Address:    in.Address,
		Timezone:   in.Timezone,
		IsOnline:   in.IsOnline,
		MeetingURL: in.MeetingURL,
		IsActive:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("create location: %w", err)
	}

	invalidate(ctx, uc.deps.Cache, master.ID)

	return &CreateLocationOutput{Location: loc}, nil
}