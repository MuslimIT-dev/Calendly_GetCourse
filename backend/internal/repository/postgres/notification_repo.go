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

type NotificationRepo struct {
	q *db.Queries
}

func NewNotificationRepo(q *db.Queries) *NotificationRepo {
	return &NotificationRepo{q: q}
}

var _ domain.NotificationRepository = (*NotificationRepo)(nil)

func (r *NotificationRepo) Create(ctx context.Context, notification *domain.Notification) (*domain.Notification, error) {
	row, err := r.q.CreateNotification(ctx, db.CreateNotificationParams{
		UserID: notification.UserID, Type: int16(notification.Type), Channel: int16(notification.Channel),
		Status: int16(notification.Status), Title: notification.Title, Body: notification.Body,
		Link:     pgtype.Text{String: notification.Link, Valid: notification.Link != ""},
		Metadata: notification.Metadata,
	})
	if err != nil {
		return nil, err
	}
	return mapNotification(row), nil
}

func (r *NotificationRepo) GetByID(ctx context.Context, id int64) (*domain.Notification, error) {
	row, err := r.q.GetNotificationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, err
	}
	return mapNotification(row), nil
}

func (r *NotificationRepo) ListByUser(ctx context.Context, userID int32, status *domain.NotificationStatus, cursor *domain.NotificationCursor, limit int32) ([]*domain.Notification, error) {
	var statusFilter pgtype.Int2
	if status != nil {
		statusFilter = pgtype.Int2{Int16: int16(*status), Valid: true}
	}
	var cursorTime pgtype.Timestamptz
	var cursorID pgtype.Int8
	if cursor != nil {
		cursorTime = pgtype.Timestamptz{Time: cursor.CreatedAt, Valid: true}
		cursorID = pgtype.Int8{Int64: cursor.ID, Valid: true}
	}
	rows, err := r.q.ListMyNotifications(ctx, db.ListMyNotificationsParams{
		UserID: userID, StatusFilter: statusFilter, CursorTime: cursorTime, CursorID: cursorID, PageSize: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		out[i] = mapNotificationFields(row.ID, row.UserID, row.Type, row.Channel, row.Status, row.Title, row.Body,
			row.Link, row.Metadata, row.SentAt, row.ReadAt, row.CreatedAt)
	}
	return out, nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id int64, userID int32) (int64, error) {
	return r.q.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{ID: id, UserID: userID})
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID int32) (int64, error) {
	return r.q.MarkAllNotificationsAsRead(ctx, userID)
}

func (r *NotificationRepo) Delete(ctx context.Context, id int64, userID int32) (int64, error) {
	return r.q.DeleteNotification(ctx, db.DeleteNotificationParams{ID: id, UserID: userID})
}

func (r *NotificationRepo) CountUnread(ctx context.Context, userID int32) (int32, error) {
	return r.q.GetUnreadCount(ctx, userID)
}

func (r *NotificationRepo) ListPending(ctx context.Context, limit int32) ([]*domain.Notification, error) {
	rows, err := r.q.GetPendingNotifications(ctx, limit)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Notification, len(rows))
	for i, row := range rows {
		out[i] = mapNotification(row)
	}
	return out, nil
}

func (r *NotificationRepo) MarkSent(ctx context.Context, id int64) error {
	return r.q.MarkNotificationSent(ctx, id)
}

func (r *NotificationRepo) MarkFailed(ctx context.Context, id int64) error {
	return r.q.MarkNotificationFailed(ctx, id)
}

func (r *NotificationRepo) GetSettings(ctx context.Context, userID int32) (*domain.NotificationSettings, error) {
	row, err := r.q.GetNotificationSettings(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		row, err = r.q.CreateDefaultNotificationSettings(ctx, userID)
		if errors.Is(err, pgx.ErrNoRows) {
			row, err = r.q.GetNotificationSettings(ctx, userID)
		}
	}
	if err != nil {
		return nil, err
	}
	return mapNotificationSettings(row), nil
}

func (r *NotificationRepo) UpsertSettings(ctx context.Context, settings *domain.NotificationSettings) (*domain.NotificationSettings, error) {
	mutedTypes := make([]int16, len(settings.MutedTypes))
	for i, notificationType := range settings.MutedTypes {
		mutedTypes[i] = int16(notificationType)
	}
	row, err := r.q.UpdateNotificationSettings(ctx, db.UpdateNotificationSettingsParams{
		UserID: settings.UserID, EmailEnabled: settings.EmailEnabled,
		TelegramEnabled: settings.TelegramEnabled, SmsEnabled: settings.SMSEnabled,
		PushEnabled: settings.PushEnabled, MutedTypes: mutedTypes,
		Reminder24h: settings.Reminder24h, Reminder2h: settings.Reminder2h, Reminder30m: settings.Reminder30m,
	})
	if err != nil {
		return nil, err
	}
	return mapNotificationSettings(row), nil
}

func mapNotification(row db.Notification) *domain.Notification {
	return mapNotificationFields(row.ID, row.UserID, row.Type, row.Channel, row.Status, row.Title, row.Body,
		row.Link, row.Metadata, row.SentAt, row.ReadAt, row.CreatedAt)
}

func mapNotificationFields(id int64, userID int32, notificationType, channel, status int16, title, body string,
	link pgtype.Text, metadata []byte, sentAt, readAt, createdAt pgtype.Timestamptz) *domain.Notification {
	var sent, read *time.Time
	if sentAt.Valid {
		sent = &sentAt.Time
	}
	if readAt.Valid {
		read = &readAt.Time
	}
	return &domain.Notification{
		ID: id, UserID: userID, Type: domain.NotificationType(notificationType),
		Channel: domain.NotificationChannel(channel), Status: domain.NotificationStatus(status),
		Title: title, Body: body, Link: link.String, Metadata: metadata, SentAt: sent, ReadAt: read, CreatedAt: createdAt.Time,
	}
}

func mapNotificationSettings(row db.NotificationSetting) *domain.NotificationSettings {
	muted := make([]int32, len(row.MutedTypes))
	for i, notificationType := range row.MutedTypes {
		muted[i] = int32(notificationType)
	}
	return &domain.NotificationSettings{
		UserID: row.UserID, EmailEnabled: row.EmailEnabled, TelegramEnabled: row.TelegramEnabled,
		SMSEnabled: row.SmsEnabled, PushEnabled: row.PushEnabled, MutedTypes: muted,
		Reminder24h: row.Reminder24h, Reminder2h: row.Reminder2h, Reminder30m: row.Reminder30m,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
