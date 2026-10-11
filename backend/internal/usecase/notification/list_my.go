package notification

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ListMyInput struct {
	StatusFilter *domain.NotificationStatus
	PageSize     int32
	PageToken    string
}

type ListMyOutput struct {
	Notifications []*domain.Notification
	NextPageToken string
	UnreadCount   int32
}

type ListMyUseCase struct{ deps Deps }

func NewListMyUseCase(deps Deps) *ListMyUseCase { return &ListMyUseCase{deps: deps} }

func (uc *ListMyUseCase) Execute(ctx context.Context, in ListMyInput) (*ListMyOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	if in.StatusFilter != nil && (*in.StatusFilter < domain.NotifStatusPending || *in.StatusFilter > domain.NotifStatusRead) {
		return nil, domain.ErrInvalidInput
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	cursor, err := decodeNotificationCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	rows, err := uc.deps.Notifications.ListByUser(ctx, userID, in.StatusFilter, cursor, pageSize+1)
	if err != nil {
		return nil, err
	}
	var next string
	if len(rows) > int(pageSize) {
		rows = rows[:int(pageSize)]
		last := rows[len(rows)-1]
		next = encodeNotificationCursor(&domain.NotificationCursor{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	unread, err := uc.deps.Notifications.CountUnread(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread notifications: %w", err)
	}
	return &ListMyOutput{Notifications: rows, NextPageToken: next, UnreadCount: unread}, nil
}

func encodeNotificationCursor(cursor *domain.NotificationCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeNotificationCursor(token string) (*domain.NotificationCursor, error) {
	if token == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var cursor domain.NotificationCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID <= 0 || cursor.CreatedAt.IsZero() || cursor.CreatedAt.After(time.Now()) {
		return nil, fmt.Errorf("invalid notification cursor")
	}
	return &cursor, nil
}
