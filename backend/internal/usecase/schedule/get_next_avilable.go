package schedule

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

const (
	nextAvailableMaxDays = 14
	nextAvailableLimit   = 3
)

type GetNextAvailableInput struct {
	MasterID   int32
	ServiceID  int32
	LocationID int32
	Timezone   string
}

type GetNextAvailableOutput struct {
	Slots []domain.TimeSlot
}

type GetNextAvailableUseCase struct {
	deps Deps
}

func NewGetNextAvailableUseCase(deps Deps) *GetNextAvailableUseCase {
	return &GetNextAvailableUseCase{deps: deps}
}

func (uc *GetNextAvailableUseCase) Execute(ctx context.Context, in GetNextAvailableInput) (*GetNextAvailableOutput, error) {
	loc, err := time.LoadLocation(in.Timezone)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	dayUC := NewGetDayUseCase(uc.deps)
	now := time.Now().In(loc)
	var out []domain.TimeSlot

	for i := 0; i < nextAvailableMaxDays; i++ {
		date := now.AddDate(0, 0, i)
		res, err := dayUC.Execute(ctx, GetDayInput{
			MasterID:   in.MasterID,
			ServiceID:  in.ServiceID,
			LocationID: in.LocationID,
			Date:       date,
			Timezone:   in.Timezone,
		})
		if err != nil {
			return nil, err
		}

		for _, s := range res.Slots {
			if s.Start.After(now) {
				out = append(out, s)
				if len(out) >= nextAvailableLimit {
					return &GetNextAvailableOutput{Slots: out}, nil
				}
			}
		}
	}

	return &GetNextAvailableOutput{Slots: out}, nil
}