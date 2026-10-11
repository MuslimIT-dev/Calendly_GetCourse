package domain

import (
	"context"
	"time"
)

type ExceptionType int32

const (
	ExceptionDayOff      ExceptionType = 1
	ExceptionCustomHours ExceptionType = 2
	ExceptionVacation    ExceptionType = 3
)

type TimeOfDay struct {
	Hours   int32
	Minutes int32
	Seconds int32
}

type ScheduleRule struct {
	ID                    int32
	LocationID            int32
	DayOfWeek             int32
	IsWorkingDay          bool
	WorkStart             *TimeOfDay
	WorkEnd               *TimeOfDay
	RestStart             *TimeOfDay
	RestEnd               *TimeOfDay
	BufferBetweenSlotsMin int32
	SlotStepMins          int32
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type Exception struct {
	ID          int32
	LocationID  int32
	Date        time.Time
	Type        ExceptionType
	IsFullDay   bool
	CustomStart *TimeOfDay
	CustomEnd   *TimeOfDay
	RestStart   *TimeOfDay
	RestEnd     *TimeOfDay
	Reason      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TimeSlot struct {
	Start time.Time
	End   time.Time
}

type DaySchedule struct {
	Date       time.Time
	HasSlots   bool
	SlotsCount int32
}

type CachedScheduleRules struct {
	LocationID int32           `json:"location_id"`
	Rules      []*ScheduleRule `json:"rules"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type ScheduleRepository interface {
	GetRulesByLocation(ctx context.Context, locationID int32) ([]*ScheduleRule, error)
	UpsertRule(ctx context.Context, r *ScheduleRule) (*ScheduleRule, error)
	DeleteRulesByLocation(ctx context.Context, locationID int32) error

	ListExceptions(ctx context.Context, locationID int32, from, to time.Time) ([]*Exception, error)
	GetExceptionByDate(ctx context.Context, locationID int32, date time.Time) (*Exception, error)
	CreateException(ctx context.Context, e *Exception) (*Exception, error)
	DeleteException(ctx context.Context, id int32) error
	DeleteExceptionByDate(ctx context.Context, locationID int32, date time.Time) error

	GetBookedSlots(ctx context.Context, masterID, locationID int32, from, to time.Time) ([]TimeSlot, error)
}