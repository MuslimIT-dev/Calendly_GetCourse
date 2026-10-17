package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	"github.com/MuslimIT-dev/zaman/backend/internal/repository/db"
)

type BookingRepo struct {
	q *db.Queries
}

func NewBookingRepo(q *db.Queries) *BookingRepo {
	return &BookingRepo{q: q}
}

var _ domain.BookingRepository = (*BookingRepo)(nil)

func (r *BookingRepo) Create(ctx context.Context, a *domain.Appointment) (*domain.Appointment, error) {
	var createdBy pgtype.Int4
	if a.CreatedBy != nil {
		createdBy = pgtype.Int4{Int32: *a.CreatedBy, Valid: true}
	}

	row, err := r.q.CreateAppointment(ctx, db.CreateAppointmentParams{
		ClientID:    a.ClientID,
		CreatedBy:   createdBy,
		MasterID:    a.MasterID,
		ServiceID:   a.ServiceID,
		LocationID:  a.LocationID,
		StartTime:   pgtype.Timestamptz{Time: a.StartTime, Valid: true},
		EndTime:     pgtype.Timestamptz{Time: a.EndTime, Valid: true},
		Timezone:    a.Timezone,
		Price:       a.Price,
		Currency:    a.Currency,
		IsOnline:    a.IsOnline,
		MeetingUrl:  pgtype.Text{String: a.MeetingURL, Valid: a.MeetingURL != ""},
		ClientNotes: pgtype.Text{String: a.ClientNotes, Valid: a.ClientNotes != ""},
	})
	if err != nil {
		return nil, err
	}
	return mapApptCreate(row), nil
}

func (r *BookingRepo) GetByID(ctx context.Context, id int64) (*domain.AppointmentCard, error) {
	row, err := r.q.GetAppointmentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAppointmentNotFound
		}
		return nil, err
	}
	return mapCardGet(row), nil
}

func (r *BookingRepo) ListByClient(
	ctx context.Context,
	clientID int32,
	status *domain.AppointmentStatus,
	cursor *domain.AppointmentCursor,
	limit int32,
) ([]*domain.AppointmentCard, error) {
	var st pgtype.Int2
	if status != nil {
		st = pgtype.Int2{Int16: int16(*status), Valid: true}
	}
	var ct pgtype.Timestamptz
	var cid pgtype.Int8
	if cursor != nil {
		ct = pgtype.Timestamptz{Time: cursor.StartTime, Valid: true}
		cid = pgtype.Int8{Int64: cursor.ID, Valid: true}
	}

	rows, err := r.q.ListAppointmentsByClient(ctx, db.ListAppointmentsByClientParams{
		ClientID:     clientID,
		StatusFilter: st,
		CursorTime:   ct,
		CursorID:     cid,
		Lim:          limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AppointmentCard, len(rows))
	for i, row := range rows {
		out[i] = mapCardClient(row)
	}
	return out, nil
}

func (r *BookingRepo) ListByMaster(
	ctx context.Context,
	masterID int32,
	locationID *int32,
	status *domain.AppointmentStatus,
	from, to *time.Time,
	cursor *domain.AppointmentCursor,
	limit int32,
) ([]*domain.AppointmentCard, error) {
	var locID pgtype.Int4
	if locationID != nil {
		locID = pgtype.Int4{Int32: *locationID, Valid: true}
	}
	var st pgtype.Int2
	if status != nil {
		st = pgtype.Int2{Int16: int16(*status), Valid: true}
	}
	var ft, tt pgtype.Timestamptz
	if from != nil {
		ft = pgtype.Timestamptz{Time: *from, Valid: true}
	}
	if to != nil {
		tt = pgtype.Timestamptz{Time: *to, Valid: true}
	}
	var ct pgtype.Timestamptz
	var cid pgtype.Int8
	if cursor != nil {
		ct = pgtype.Timestamptz{Time: cursor.StartTime, Valid: true}
		cid = pgtype.Int8{Int64: cursor.ID, Valid: true}
	}

	rows, err := r.q.ListAppointmentsByMaster(ctx, db.ListAppointmentsByMasterParams{
		MasterID:     masterID,
		LocationID:   locID,
		StatusFilter: st,
		FromTime:     ft,
		ToTime:       tt,
		CursorTime:   ct,
		CursorID:     cid,
		Lim:          limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AppointmentCard, len(rows))
	for i, row := range rows {
		out[i] = mapCardMaster(row)
	}
	return out, nil
}

func (r *BookingRepo) Confirm(ctx context.Context, id int64) (*domain.Appointment, error) {
	row, err := r.q.ConfirmAppointment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapApptConfirm(row), nil
}

func (r *BookingRepo) Cancel(
	ctx context.Context,
	id int64,
	status domain.AppointmentStatus,
	byRole domain.CancelledByRole,
	byUserID *int32,
	reason string,
) (*domain.Appointment, error) {
	var byUID pgtype.Int4
	if byUserID != nil {
		byUID = pgtype.Int4{Int32: *byUserID, Valid: true}
	}

	row, err := r.q.CancelAppointment(ctx, db.CancelAppointmentParams{
		ID:              id,
		Status:          int16(status),
		CancelledBy:     byUID,
		CancelledByRole: pgtype.Int2{Int16: int16(byRole), Valid: true},
		CancelReason:    pgtype.Text{String: reason, Valid: reason != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapApptCancel(row), nil
}

func (r *BookingRepo) Reschedule(ctx context.Context, id int64, newStart, newEnd time.Time, locationID int32) (*domain.Appointment, error) {
	row, err := r.q.RescheduleAppointment(ctx, db.RescheduleAppointmentParams{
		ID:         id,
		StartTime:  pgtype.Timestamptz{Time: newStart, Valid: true},
		EndTime:    pgtype.Timestamptz{Time: newEnd, Valid: true},
		LocationID: locationID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapApptReschedule(row), nil
}

func (r *BookingRepo) Complete(ctx context.Context, id int64, masterNotes string) (*domain.Appointment, error) {
	row, err := r.q.CompleteAppointment(ctx, db.CompleteAppointmentParams{
		ID:          id,
		MasterNotes: pgtype.Text{String: masterNotes, Valid: masterNotes != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapApptComplete(row), nil
}

func (r *BookingRepo) MarkNoShow(ctx context.Context, id int64) (*domain.Appointment, error) {
	row, err := r.q.MarkAppointmentNoShow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapApptNoShow(row), nil
}

func (r *BookingRepo) UpdatePaymentStatus(ctx context.Context, id int64, status domain.PaymentStatus) error {
	return r.q.UpdateAppointmentPaymentStatus(ctx, db.UpdateAppointmentPaymentStatusParams{
		ID:            id,
		PaymentStatus: int16(status),
	})
}

func int2ToCancelledRolePtr(v pgtype.Int2) *domain.CancelledByRole {
	if !v.Valid {
		return nil
	}
	r := domain.CancelledByRole(v.Int16)
	return &r
}

func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	return &v.Time
}

func mapApptCreate(row db.CreateAppointmentRow) *domain.Appointment {
	return &domain.Appointment{
		ID:              row.ID,
		ClientID:        row.ClientID,
		CreatedBy:       int4Ptr(row.CreatedBy),
		MasterID:        row.MasterID,
		ServiceID:       row.ServiceID,
		LocationID:      row.LocationID,
		StartTime:       row.StartTime.Time,
		EndTime:         row.EndTime.Time,
		Timezone:        row.Timezone,
		Status:          domain.AppointmentStatus(row.Status),
		Price:           row.Price,
		Currency:        row.Currency,
		PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
		IsOnline:        row.IsOnline,
		MeetingURL:      row.MeetingUrl.String,
		ClientNotes:     row.ClientNotes.String,
		MasterNotes:     row.MasterNotes.String,
		CancelledAt:     timePtr(row.CancelledAt),
		CancelledBy:     int4Ptr(row.CancelledBy),
		CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
		CancelReason:    row.CancelReason.String,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func mapApptConfirm(row db.ConfirmAppointmentRow) *domain.Appointment {
	return &domain.Appointment{
		ID:              row.ID,
		ClientID:        row.ClientID,
		CreatedBy:       int4Ptr(row.CreatedBy),
		MasterID:        row.MasterID,
		ServiceID:       row.ServiceID,
		LocationID:      row.LocationID,
		StartTime:       row.StartTime.Time,
		EndTime:         row.EndTime.Time,
		Timezone:        row.Timezone,
		Status:          domain.AppointmentStatus(row.Status),
		Price:           row.Price,
		Currency:        row.Currency,
		PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
		IsOnline:        row.IsOnline,
		MeetingURL:      row.MeetingUrl.String,
		ClientNotes:     row.ClientNotes.String,
		MasterNotes:     row.MasterNotes.String,
		CancelledAt:     timePtr(row.CancelledAt),
		CancelledBy:     int4Ptr(row.CancelledBy),
		CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
		CancelReason:    row.CancelReason.String,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func mapApptCancel(row db.CancelAppointmentRow) *domain.Appointment {
	return &domain.Appointment{
		ID:              row.ID,
		ClientID:        row.ClientID,
		CreatedBy:       int4Ptr(row.CreatedBy),
		MasterID:        row.MasterID,
		ServiceID:       row.ServiceID,
		LocationID:      row.LocationID,
		StartTime:       row.StartTime.Time,
		EndTime:         row.EndTime.Time,
		Timezone:        row.Timezone,
		Status:          domain.AppointmentStatus(row.Status),
		Price:           row.Price,
		Currency:        row.Currency,
		PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
		IsOnline:        row.IsOnline,
		MeetingURL:      row.MeetingUrl.String,
		ClientNotes:     row.ClientNotes.String,
		MasterNotes:     row.MasterNotes.String,
		CancelledAt:     timePtr(row.CancelledAt),
		CancelledBy:     int4Ptr(row.CancelledBy),
		CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
		CancelReason:    row.CancelReason.String,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func mapApptReschedule(row db.RescheduleAppointmentRow) *domain.Appointment {
	return &domain.Appointment{
		ID:              row.ID,
		ClientID:        row.ClientID,
		CreatedBy:       int4Ptr(row.CreatedBy),
		MasterID:        row.MasterID,
		ServiceID:       row.ServiceID,
		LocationID:      row.LocationID,
		StartTime:       row.StartTime.Time,
		EndTime:         row.EndTime.Time,
		Timezone:        row.Timezone,
		Status:          domain.AppointmentStatus(row.Status),
		Price:           row.Price,
		Currency:        row.Currency,
		PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
		IsOnline:        row.IsOnline,
		MeetingURL:      row.MeetingUrl.String,
		ClientNotes:     row.ClientNotes.String,
		MasterNotes:     row.MasterNotes.String,
		CancelledAt:     timePtr(row.CancelledAt),
		CancelledBy:     int4Ptr(row.CancelledBy),
		CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
		CancelReason:    row.CancelReason.String,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func mapApptComplete(row db.CompleteAppointmentRow) *domain.Appointment {
	return &domain.Appointment{
		ID:              row.ID,
		ClientID:        row.ClientID,
		CreatedBy:       int4Ptr(row.CreatedBy),
		MasterID:        row.MasterID,
		ServiceID:       row.ServiceID,
		LocationID:      row.LocationID,
		StartTime:       row.StartTime.Time,
		EndTime:         row.EndTime.Time,
		Timezone:        row.Timezone,
		Status:          domain.AppointmentStatus(row.Status),
		Price:           row.Price,
		Currency:        row.Currency,
		PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
		IsOnline:        row.IsOnline,
		MeetingURL:      row.MeetingUrl.String,
		ClientNotes:     row.ClientNotes.String,
		MasterNotes:     row.MasterNotes.String,
		CancelledAt:     timePtr(row.CancelledAt),
		CancelledBy:     int4Ptr(row.CancelledBy),
		CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
		CancelReason:    row.CancelReason.String,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func mapApptNoShow(row db.MarkAppointmentNoShowRow) *domain.Appointment {
	return &domain.Appointment{
		ID:              row.ID,
		ClientID:        row.ClientID,
		CreatedBy:       int4Ptr(row.CreatedBy),
		MasterID:        row.MasterID,
		ServiceID:       row.ServiceID,
		LocationID:      row.LocationID,
		StartTime:       row.StartTime.Time,
		EndTime:         row.EndTime.Time,
		Timezone:        row.Timezone,
		Status:          domain.AppointmentStatus(row.Status),
		Price:           row.Price,
		Currency:        row.Currency,
		PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
		IsOnline:        row.IsOnline,
		MeetingURL:      row.MeetingUrl.String,
		ClientNotes:     row.ClientNotes.String,
		MasterNotes:     row.MasterNotes.String,
		CancelledAt:     timePtr(row.CancelledAt),
		CancelledBy:     int4Ptr(row.CancelledBy),
		CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
		CancelReason:    row.CancelReason.String,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func mapCardGet(row db.GetAppointmentByIDRow) *domain.AppointmentCard {
	return &domain.AppointmentCard{
		Appointment: &domain.Appointment{
			ID:              row.ID,
			ClientID:        row.ClientID,
			CreatedBy:       int4Ptr(row.CreatedBy),
			MasterID:        row.MasterID,
			ServiceID:       row.ServiceID,
			LocationID:      row.LocationID,
			StartTime:       row.StartTime.Time,
			EndTime:         row.EndTime.Time,
			Timezone:        row.Timezone,
			Status:          domain.AppointmentStatus(row.Status),
			Price:           row.Price,
			Currency:        row.Currency,
			PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
			IsOnline:        row.IsOnline,
			MeetingURL:      row.MeetingUrl.String,
			ClientNotes:     row.ClientNotes.String,
			MasterNotes:     row.MasterNotes.String,
			CancelledAt:     timePtr(row.CancelledAt),
			CancelledBy:     int4Ptr(row.CancelledBy),
			CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
			CancelReason:    row.CancelReason.String,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		},
		MasterSlug:   row.MasterSlug,
		MasterName:   row.MasterName.String,
		ServiceName:  row.ServiceName,
		ServiceSlug:  row.ServiceSlug,
		LocationName: row.LocationName,
		Address:      row.LocationAddress,
	}
}

func mapCardClient(row db.ListAppointmentsByClientRow) *domain.AppointmentCard {
	return &domain.AppointmentCard{
		Appointment: &domain.Appointment{
			ID:              row.ID,
			ClientID:        row.ClientID,
			CreatedBy:       int4Ptr(row.CreatedBy),
			MasterID:        row.MasterID,
			ServiceID:       row.ServiceID,
			LocationID:      row.LocationID,
			StartTime:       row.StartTime.Time,
			EndTime:         row.EndTime.Time,
			Timezone:        row.Timezone,
			Status:          domain.AppointmentStatus(row.Status),
			Price:           row.Price,
			Currency:        row.Currency,
			PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
			IsOnline:        row.IsOnline,
			MeetingURL:      row.MeetingUrl.String,
			ClientNotes:     row.ClientNotes.String,
			MasterNotes:     row.MasterNotes.String,
			CancelledAt:     timePtr(row.CancelledAt),
			CancelledBy:     int4Ptr(row.CancelledBy),
			CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
			CancelReason:    row.CancelReason.String,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		},
		MasterSlug:   row.MasterSlug,
		MasterName:   row.MasterName.String,
		ServiceName:  row.ServiceName,
		ServiceSlug:  row.ServiceSlug,
		LocationName: row.LocationName,
		Address:      row.LocationAddress,
	}
}

func mapCardMaster(row db.ListAppointmentsByMasterRow) *domain.AppointmentCard {
	return &domain.AppointmentCard{
		Appointment: &domain.Appointment{
			ID:              row.ID,
			ClientID:        row.ClientID,
			CreatedBy:       int4Ptr(row.CreatedBy),
			MasterID:        row.MasterID,
			ServiceID:       row.ServiceID,
			LocationID:      row.LocationID,
			StartTime:       row.StartTime.Time,
			EndTime:         row.EndTime.Time,
			Timezone:        row.Timezone,
			Status:          domain.AppointmentStatus(row.Status),
			Price:           row.Price,
			Currency:        row.Currency,
			PaymentStatus:   domain.PaymentStatus(row.PaymentStatus),
			IsOnline:        row.IsOnline,
			MeetingURL:      row.MeetingUrl.String,
			ClientNotes:     row.ClientNotes.String,
			MasterNotes:     row.MasterNotes.String,
			CancelledAt:     timePtr(row.CancelledAt),
			CancelledBy:     int4Ptr(row.CancelledBy),
			CancelledByRole: int2ToCancelledRolePtr(row.CancelledByRole),
			CancelReason:    row.CancelReason.String,
			CreatedAt:       row.CreatedAt.Time,
			UpdatedAt:       row.UpdatedAt.Time,
		},
		MasterSlug:   row.MasterSlug,
		MasterName:   row.MasterName.String,
		ServiceName:  row.ServiceName,
		ServiceSlug:  row.ServiceSlug,
		LocationName: row.LocationName,
		Address:      row.LocationAddress,
	}
}
