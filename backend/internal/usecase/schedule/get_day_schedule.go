package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetDayInput struct {
	MasterID   int32
	ServiceID  int32
	LocationID int32
	Date       time.Time
	Timezone   string
}

type GetDayOutput struct {
	Slots []domain.TimeSlot
}

type GetDayUseCase struct {
	deps Deps
}

func NewGetDayUseCase(deps Deps) *GetDayUseCase {
	return &GetDayUseCase{deps: deps}
}

func (uc *GetDayUseCase) Execute(ctx context.Context, in GetDayInput) (*GetDayOutput, error) {
	loc, err := time.LoadLocation(in.Timezone)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	svc, err := uc.deps.Services.GetByID(ctx, in.ServiceID)
	if err != nil {
		return nil, err
	}
	if svc.MasterID != in.MasterID {
		return nil, domain.ErrForbidden
	}

	date := time.Date(in.Date.Year(), in.Date.Month(), in.Date.Day(), 0, 0, 0, 0, loc)
	dayEnd := date.AddDate(0, 0, 1)

	rules, err := uc.deps.Schedule.GetRulesByLocation(ctx, in.LocationID)
	if err != nil {
		return nil, err
	}
	var rule *domain.ScheduleRule
	for _, r := range rules {
		if r.DayOfWeek == int32(date.Weekday()) {
			rule = r
			break
		}
	}

	exception, _ := uc.deps.Schedule.GetExceptionByDate(ctx, in.LocationID, date)

	booked, err := uc.deps.Schedule.GetBookedSlots(ctx, in.MasterID, in.LocationID, date, dayEnd)
	if err != nil {
		return nil, fmt.Errorf("get booked: %w", err)
	}

	gen := newSlotGenerator(rule, exception, svc.DurationMins, loc, booked)
	slots := gen.Generate(date)

	return &GetDayOutput{Slots: slots}, nil
}