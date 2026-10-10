package location

import (
	"context"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type UpdateLocationInput struct {
	ID         int32
	Name       *string
	Address    *string
	Timezone   *string
	IsOnline   *bool
	MeetingURL *string
	IsActive   *bool
}

type UpdateLocationOutput struct {
	Location *domain.Location
}

type UpdateLocationUseCase struct {
	deps Deps
}

func NewUpdateLocationUseCase(deps Deps) *UpdateLocationUseCase {
	return &UpdateLocationUseCase{deps: deps}
}

func (uc *UpdateLocationUseCase) Execute(ctx context.Context, in UpdateLocationInput) (*UpdateLocationOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	loc, err := uc.deps.Locations.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if loc.MasterID != master.ID {
		return nil, domain.ErrForbidden
	}

	if in.Name != nil {
		if *in.Name == "" {
			return nil, domain.ErrInvalidInput
		}
		loc.Name = *in.Name
	}
	if in.Address != nil {
		loc.Address = *in.Address
	}
	if in.Timezone != nil {
		loc.Timezone = *in.Timezone
	}
	if in.IsOnline != nil {
		loc.IsOnline = *in.IsOnline
	}
	if in.MeetingURL != nil {
		loc.MeetingURL = *in.MeetingURL
	}
	if in.IsActive != nil {
		loc.IsActive = *in.IsActive
	}

	updated, err := uc.deps.Locations.Update(ctx, loc)
	if err != nil {
		return nil, fmt.Errorf("update location: %w", err)
	}

	invalidate(ctx, uc.deps.Cache, master.ID)

	return &UpdateLocationOutput{Location: updated}, nil
}