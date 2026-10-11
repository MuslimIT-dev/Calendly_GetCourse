package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
)

type PaymentRepo struct {
	q *db.Queries
}

func NewPaymentRepo(q *db.Queries) *PaymentRepo {
	return &PaymentRepo{q: q}
}

var _ domain.PaymentRepository = (*PaymentRepo)(nil)

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) (*domain.Payment, error) {
	row, err := r.q.CreatePayment(ctx, db.CreatePaymentParams{
		UserID:      p.UserID,
		Amount:      p.Amount,
		Currency:    p.Currency,
		Provider:    int16(p.Provider),
		Purpose:     int16(p.Purpose),
		ReferenceID: p.ReferenceID,
		ExternalID:  pgtype.Text{String: p.ExternalID, Valid: p.ExternalID != ""},
		PaymentUrl:  pgtype.Text{String: p.PaymentURL, Valid: p.PaymentURL != ""},
		ExpiresAt:   pgtype.Timestamptz{Time: p.ExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) SetExternal(ctx context.Context, id int64, externalID, paymentURL string) (*domain.Payment, error) {
	row, err := r.q.SetPaymentExternal(ctx, db.SetPaymentExternalParams{
		ID:         id,
		ExternalID: pgtype.Text{String: externalID, Valid: externalID != ""},
		PaymentUrl: pgtype.Text{String: paymentURL, Valid: paymentURL != ""},
	})
	if err != nil {
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) GetByID(ctx context.Context, id int64) (*domain.PaymentCard, error) {
	row, err := r.q.GetPaymentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return mapPaymentCardByID(row), nil
}

func (r *PaymentRepo) GetByExternalID(ctx context.Context, provider domain.PaymentProvider, externalID string) (*domain.PaymentCard, error) {
	row, err := r.q.GetPaymentByExternalID(ctx, db.GetPaymentByExternalIDParams{
		Provider:   int16(provider),
		ExternalID: pgtype.Text{String: externalID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return mapPaymentCardExternalID(row), nil
}

func (r *PaymentRepo) GetPaidByReference(ctx context.Context, purpose domain.PaymentPurpose, referenceID string) (*domain.Payment, error) {
	row, err := r.q.GetPaidPaymentByReference(ctx, db.GetPaidPaymentByReferenceParams{
		Purpose:     int16(purpose),
		ReferenceID: referenceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) ListByUser(ctx context.Context, userID int32, status *domain.PaymentStatus, purpose *domain.PaymentPurpose, cursor *domain.PaymentCursor, limit int32) ([]*domain.PaymentCard, error) {
	var st, pp pgtype.Int2
	if status != nil {
		st = pgtype.Int2{Int16: int16(*status), Valid: true}
	}
	if purpose != nil {
		pp = pgtype.Int2{Int16: int16(*purpose), Valid: true}
	}
	var ct pgtype.Timestamptz
	var cid pgtype.Int8
	if cursor != nil {
		ct = pgtype.Timestamptz{Time: cursor.CreatedAt, Valid: true}
		cid = pgtype.Int8{Int64: cursor.ID, Valid: true}
	}
	rows, err := r.q.ListMyPayments(ctx, db.ListMyPaymentsParams{
		UserID:        userID,
		StatusFilter:  st,
		PurposeFilter: pp,
		CursorTime:    ct,
		CursorID:      cid,
		PageSize:      limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.PaymentCard, len(rows))
	for i, row := range rows {
		out[i] = mapPaymentCardUser(row)
	}
	return out, nil
}

func (r *PaymentRepo) ListAll(ctx context.Context, status *domain.PaymentStatus, userID *int32, from, to *time.Time, cursor *domain.PaymentCursor, limit int32) ([]*domain.PaymentCard, error) {
	var st pgtype.Int2
	if status != nil {
		st = pgtype.Int2{Int16: int16(*status), Valid: true}
	}
	var uid pgtype.Int4
	if userID != nil {
		uid = pgtype.Int4{Int32: *userID, Valid: true}
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
		ct = pgtype.Timestamptz{Time: cursor.CreatedAt, Valid: true}
		cid = pgtype.Int8{Int64: cursor.ID, Valid: true}
	}
	rows, err := r.q.ListAllPayments(ctx, db.ListAllPaymentsParams{
		StatusFilter: st,
		UserID:       uid,
		FromTime:     ft,
		ToTime:       tt,
		CursorTime:   ct,
		CursorID:     cid,
		PageSize:     limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.PaymentCard, len(rows))
	for i, row := range rows {
		out[i] = mapPaymentCardAll(row)
	}
	return out, nil
}

func (r *PaymentRepo) MarkPaid(ctx context.Context, id int64, externalID, receiptURL, paymentURL string) (*domain.Payment, error) {
	row, err := r.q.MarkPaymentPaid(ctx, db.MarkPaymentPaidParams{
		ID:         id,
		ExternalID: pgtype.Text{String: externalID, Valid: externalID != ""},
		ReceiptUrl: pgtype.Text{String: receiptURL, Valid: receiptURL != ""},
		PaymentUrl: pgtype.Text{String: paymentURL, Valid: paymentURL != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentAlreadyPaid
		}
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) MarkFailed(ctx context.Context, id int64, reason string) (*domain.Payment, error) {
	row, err := r.q.MarkPaymentFailed(ctx, db.MarkPaymentFailedParams{
		ID:            id,
		FailureReason: pgtype.Text{String: reason, Valid: reason != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentInvalidStatus
		}
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) MarkRefunded(ctx context.Context, id int64) (*domain.Payment, error) {
	row, err := r.q.MarkPaymentRefunded(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentInvalidStatus
		}
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) Cancel(ctx context.Context, id int64) (*domain.Payment, error) {
	row, err := r.q.CancelPayment(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentInvalidStatus
		}
		return nil, err
	}
	return mapPaymentFull(row), nil
}

func (r *PaymentRepo) ExpirePending(ctx context.Context) (int64, error) {
	return r.q.ExpirePendingPayments(ctx)
}

func mapPaymentBasic(row db.Payment) *domain.Payment {
	return &domain.Payment{
		ID:            row.ID,
		UserID:        row.UserID,
		Amount:        row.Amount,
		Currency:      row.Currency,
		Status:        domain.PaymentStatus(row.Status),
		Provider:      domain.PaymentProvider(row.Provider),
		Purpose:       domain.PaymentPurpose(row.Purpose),
		ReferenceID:   row.ReferenceID,
		ExternalID:    row.ExternalID.String,
		PaymentURL:    row.PaymentUrl.String,
		FailureReason: row.FailureReason.String,
		ReceiptURL:    row.ReceiptUrl.String,
		PaidAt:        timePtr(row.PaidAt),
		RefundedAt:    timePtr(row.RefundedAt),
		ExpiresAt:     row.ExpiresAt.Time,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}

func mapPaymentFull(row db.Payment) *domain.Payment {
	return mapPaymentBasic(row)
}

func mapPaymentCardByID(row db.GetPaymentByIDRow) *domain.PaymentCard {
	return &domain.PaymentCard{
		Payment: &domain.Payment{
			ID:            row.ID,
			UserID:        row.UserID,
			Amount:        row.Amount,
			Currency:      row.Currency,
			Status:        domain.PaymentStatus(row.Status),
			Provider:      domain.PaymentProvider(row.Provider),
			Purpose:       domain.PaymentPurpose(row.Purpose),
			ReferenceID:   row.ReferenceID,
			ExternalID:    row.ExternalID.String,
			PaymentURL:    row.PaymentUrl.String,
			FailureReason: row.FailureReason.String,
			ReceiptURL:    row.ReceiptUrl.String,
			PaidAt:        timePtr(row.PaidAt),
			RefundedAt:    timePtr(row.RefundedAt),
			ExpiresAt:     row.ExpiresAt.Time,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		},
		UserName:   row.UserName.String,
		UserEmail:  row.UserEmail,
		UserAvatar: row.UserAvatarUrl.String,
	}
}

func mapPaymentCardExternalID(row db.GetPaymentByExternalIDRow) *domain.PaymentCard {
	return &domain.PaymentCard{
		Payment: &domain.Payment{
			ID:            row.ID,
			UserID:        row.UserID,
			Amount:        row.Amount,
			Currency:      row.Currency,
			Status:        domain.PaymentStatus(row.Status),
			Provider:      domain.PaymentProvider(row.Provider),
			Purpose:       domain.PaymentPurpose(row.Purpose),
			ReferenceID:   row.ReferenceID,
			ExternalID:    row.ExternalID.String,
			PaymentURL:    row.PaymentUrl.String,
			FailureReason: row.FailureReason.String,
			ReceiptURL:    row.ReceiptUrl.String,
			PaidAt:        timePtr(row.PaidAt),
			RefundedAt:    timePtr(row.RefundedAt),
			ExpiresAt:     row.ExpiresAt.Time,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		},
		UserName:   row.UserName.String,
		UserEmail:  row.UserEmail,
		UserAvatar: row.UserAvatarUrl.String,
	}
}

func mapPaymentCardUser(row db.ListMyPaymentsRow) *domain.PaymentCard {
	return &domain.PaymentCard{
		Payment: &domain.Payment{
			ID:            row.ID,
			UserID:        row.UserID,
			Amount:        row.Amount,
			Currency:      row.Currency,
			Status:        domain.PaymentStatus(row.Status),
			Provider:      domain.PaymentProvider(row.Provider),
			Purpose:       domain.PaymentPurpose(row.Purpose),
			ReferenceID:   row.ReferenceID,
			ExternalID:    row.ExternalID.String,
			PaymentURL:    row.PaymentUrl.String,
			FailureReason: row.FailureReason.String,
			ReceiptURL:    row.ReceiptUrl.String,
			PaidAt:        timePtr(row.PaidAt),
			RefundedAt:    timePtr(row.RefundedAt),
			ExpiresAt:     row.ExpiresAt.Time,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		},
		UserName:   row.UserName.String,
		UserEmail:  row.UserEmail,
		UserAvatar: row.UserAvatarUrl.String,
	}
}

func mapPaymentCardAll(row db.ListAllPaymentsRow) *domain.PaymentCard {
	return &domain.PaymentCard{
		Payment: &domain.Payment{
			ID:            row.ID,
			UserID:        row.UserID,
			Amount:        row.Amount,
			Currency:      row.Currency,
			Status:        domain.PaymentStatus(row.Status),
			Provider:      domain.PaymentProvider(row.Provider),
			Purpose:       domain.PaymentPurpose(row.Purpose),
			ReferenceID:   row.ReferenceID,
			ExternalID:    row.ExternalID.String,
			PaymentURL:    row.PaymentUrl.String,
			FailureReason: row.FailureReason.String,
			ReceiptURL:    row.ReceiptUrl.String,
			PaidAt:        timePtr(row.PaidAt),
			RefundedAt:    timePtr(row.RefundedAt),
			ExpiresAt:     row.ExpiresAt.Time,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		},
		UserName:   row.UserName.String,
		UserEmail:  row.UserEmail,
		UserAvatar: row.UserAvatarUrl.String,
	}
}
