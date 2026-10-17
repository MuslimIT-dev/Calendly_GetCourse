package booking

import (
	"context"

	"github.com/MuslimIT-dev/zaman/backend/internal/appcontext"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type ListMasterInput struct {
	MasterID     int32
	LocationID   *int32
	StatusFilter *domain.AppointmentStatus
	FromDate     string
	ToDate       string
	PageSize     int32
	PageToken    string
}

type ListMasterOutput struct {
	Appointments  []*domain.AppointmentCard
	NextPageToken string
}

type ListMasterUseCase struct {
	deps Deps
}

func NewListMasterUseCase(deps Deps) *ListMasterUseCase {
	return &ListMasterUseCase{deps: deps}
}

func (uc *ListMasterUseCase) Execute(ctx context.Context, in ListMasterInput) (*ListMasterOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}

	if !isMasterOwner(ctx, uc.deps.Masters, userID, in.MasterID) {
		return nil, domain.ErrForbidden
	}

	from, err := parseDate(in.FromDate)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	to, err := parseDate(in.ToDate)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	if to != nil {
		next := to.AddDate(0, 0, 1)
		to = &next
	}

	pageSize := normalizePageSize(in.PageSize)

	cursor, err := decodeCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	rows, err := uc.deps.Bookings.ListByMaster(ctx, in.MasterID, in.LocationID, in.StatusFilter, from, to, cursor, pageSize+1)
	if err != nil {
		return nil, err
	}

	var next string
	if len(rows) > int(pageSize) {
		rows = rows[:pageSize]
		last := rows[len(rows)-1]
		next = encodeCursor(&domain.AppointmentCursor{
			StartTime: last.Appointment.StartTime,
			ID:        last.Appointment.ID,
		})
	}

	return &ListMasterOutput{Appointments: rows, NextPageToken: next}, nil
}
