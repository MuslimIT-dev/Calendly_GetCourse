package schedule

import (
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type slotGenerator struct {
	rule      *domain.ScheduleRule
	exception *domain.Exception
	duration  time.Duration
	location  *time.Location
	booked    []domain.TimeSlot
}

func newSlotGenerator(
	rule *domain.ScheduleRule,
	exception *domain.Exception,
	durationMins int32,
	loc *time.Location,
	booked []domain.TimeSlot,
) *slotGenerator {
	return &slotGenerator{
		rule:      rule,
		exception: exception,
		duration:  time.Duration(durationMins) * time.Minute,
		location:  loc,
		booked:    booked,
	}
}

func (g *slotGenerator) Generate(date time.Time) []domain.TimeSlot {
	workStart, workEnd, restStart, restEnd, ok := g.resolveHours()
	if !ok {
		return nil
	}

	dayStart := g.atTime(date, workStart)
	dayEnd := g.atTime(date, workEnd)

	step := g.duration
	if g.rule != nil && g.rule.SlotStepMins > 0 {
		step = time.Duration(g.rule.SlotStepMins) * time.Minute
	}
	if step <= 0 {
		step = 30 * time.Minute
	}

	buffer := time.Duration(0)
	if g.rule != nil {
		buffer = time.Duration(g.rule.BufferBetweenSlotsMin) * time.Minute
	}

	var restStartT, restEndT *time.Time
	if restStart != nil && restEnd != nil {
		s := g.atTime(date, restStart)
		e := g.atTime(date, restEnd)
		restStartT = &s
		restEndT = &e
	}

	var slots []domain.TimeSlot
	for cur := dayStart; !cur.Add(g.duration).After(dayEnd); cur = cur.Add(step + buffer) {
		slotEnd := cur.Add(g.duration)

		if restStartT != nil && restEndT != nil && overlaps(cur, slotEnd, *restStartT, *restEndT) {
			continue
		}
		if g.isBooked(cur, slotEnd) {
			continue
		}

		slots = append(slots, domain.TimeSlot{Start: cur, End: slotEnd})
	}

	return slots
}

func (g *slotGenerator) resolveHours() (*domain.TimeOfDay, *domain.TimeOfDay, *domain.TimeOfDay, *domain.TimeOfDay, bool) {
	if g.exception != nil {
		switch g.exception.Type {
		case domain.ExceptionDayOff, domain.ExceptionVacation:
			return nil, nil, nil, nil, false
		case domain.ExceptionCustomHours:
			if g.exception.CustomStart == nil || g.exception.CustomEnd == nil {
				return nil, nil, nil, nil, false
			}
			return g.exception.CustomStart, g.exception.CustomEnd, g.exception.RestStart, g.exception.RestEnd, true
		}
	}

	if g.rule == nil || !g.rule.IsWorkingDay {
		return nil, nil, nil, nil, false
	}
	if g.rule.WorkStart == nil || g.rule.WorkEnd == nil {
		return nil, nil, nil, nil, false
	}

	return g.rule.WorkStart, g.rule.WorkEnd, g.rule.RestStart, g.rule.RestEnd, true
}

func (g *slotGenerator) atTime(date time.Time, t *domain.TimeOfDay) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		int(t.Hours), int(t.Minutes), int(t.Seconds), 0,
		g.location,
	)
}

func (g *slotGenerator) isBooked(start, end time.Time) bool {
	for _, b := range g.booked {
		if overlaps(start, end, b.Start, b.End) {
			return true
		}
	}
	return false
}

func overlaps(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}