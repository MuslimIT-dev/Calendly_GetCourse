package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type AddExceptionInput struct {
	LocationID  int32
	Date        time.Time
	Type        domain.ExceptionType
	IsFullDay   bool
	CustomStart *domain.TimeOfDay
	CustomEnd   *domain.TimeOfDay
	RestStart   *domain.TimeOfDay
	RestEnd     *domain.TimeOfDay
	Reason      string
}

type AddExceptionOutput struct {
	Exception *domain.Exception
}

type AddExceptionUseCase struct {
	deps Deps
}

func NewAddExceptionUseCase(deps Deps) *AddExceptionUseCase {
	return &AddExceptionUseCase{deps: deps}
}

func (uc *AddExceptionUseCase) Execute(ctx context.Context, in AddExceptionInput) (*AddExceptionOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if in.Type == domain.ExceptionType(0) {
		return nil, domain.ErrInvalidInput
	}
	if in.Type == domain.ExceptionCustomHours && (in.CustomStart == nil || in.CustomEnd == nil) {
		return nil, domain.ErrInvalidInput
	}

	master, err := uc.deps.Masters.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	belongs, err := uc.deps.Locations.BelongsToMaster(ctx, in.LocationID, master.ID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, domain.ErrForbidden
	}

	if err := uc.deps.Schedule.DeleteExceptionByDate(ctx, in.LocationID, in.Date); err != nil {
		return nil, fmt.Errorf("delete existing exception: %w", err)
	}

	exc, err := uc.deps.Schedule.CreateException(ctx, &domain.Exception{
		LocationID:  in.LocationID,
		Date:        in.Date,
		Type:        in.Type,
		IsFullDay:   in.IsFullDay,
		CustomStart: in.CustomStart,
		CustomEnd:   in.CustomEnd,
		RestStart:   in.RestStart,
		RestEnd:     in.RestEnd,
		Reason:      in.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("create exception: %w", err)
	}

	invalidateRules(ctx, uc.deps.Cache, in.LocationID)

	return &AddExceptionOutput{Exception: exc}, nil
}