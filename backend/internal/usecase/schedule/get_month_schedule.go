package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type GetMonthInput struct {
	MasterID   int32
	ServiceID  int32
	LocationID int32
	Year       int32
	Month      int32
	Timezone   string
}

type GetMonthOutput struct {
	Days []*domain.DaySchedule
}

type GetMonthUseCase struct {
	deps Deps
}

func NewGetMonthUseCase(deps Deps) *GetMonthUseCase {
	return &GetMonthUseCase{deps: deps}
}

func (uc *GetMonthUseCase) Execute(ctx context.Context, in GetMonthInput) (*GetMonthOutput, error) {
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

	firstDay := time.Date(int(in.Year), time.Month(in.Month), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	rules, err := uc.deps.Schedule.GetRulesByLocation(ctx, in.LocationID)
	if err != nil {
		return nil, fmt.Errorf("get rules: %w", err)
	}
	rulesByDay := make(map[int32]*domain.ScheduleRule, 7)
	for _, r := range rules {
		rulesByDay[r.DayOfWeek] = r
	}

	exceptions, err := uc.deps.Schedule.ListExceptions(ctx, in.LocationID, firstDay, lastDay.AddDate(0, 0, 1))
	if err != nil {
		return nil, fmt.Errorf("list exceptions: %w", err)
	}
	exceptionByDate := make(map[string]*domain.Exception, len(exceptions))
	for _, e := range exceptions {
		exceptionByDate[e.Date.Format("2006-01-02")] = e
	}

	booked, err := uc.deps.Schedule.GetBookedSlots(ctx, in.MasterID, in.LocationID, firstDay, lastDay.AddDate(0, 0, 1))
	if err != nil {
		return nil, fmt.Errorf("get booked: %w", err)
	}

	days := make([]*domain.DaySchedule, 0, lastDay.Day())
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		rule := rulesByDay[int32(d.Weekday())]
		exception := exceptionByDate[d.Format("2006-01-02")]

		gen := newSlotGenerator(rule, exception, svc.DurationMins, loc, booked)
		slots := gen.Generate(d)

		days = append(days, &domain.DaySchedule{
			Date:       d,
			HasSlots:   len(slots) > 0,
			SlotsCount: int32(len(slots)),
		})
	}

	return &GetMonthOutput{Days: days}, nil
}