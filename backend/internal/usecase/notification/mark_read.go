package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type MarkReadInput struct{ ID int64 }
type MarkReadOutput struct{}
type MarkReadUseCase struct{ deps Deps }

func NewMarkReadUseCase(deps Deps) *MarkReadUseCase { return &MarkReadUseCase{deps: deps} }

func (uc *MarkReadUseCase) Execute(ctx context.Context, in MarkReadInput) (*MarkReadOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	rows, err := uc.deps.Notifications.MarkRead(ctx, in.ID, userID)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		notification, err := uc.deps.Notifications.GetByID(ctx, in.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotificationNotFound) {
				return nil, domain.ErrNotificationNotFound
			}
			return nil, fmt.Errorf("verify notification ownership: %w", err)
		}
		if notification == nil {
			return nil, domain.ErrNotificationNotFound
		}
		if notification.UserID != userID {
			return nil, domain.ErrNotificationNotFound
		}
		if notification.Status != domain.NotifStatusRead {
			return nil, domain.ErrNotificationNotFound
		}
	}
	return &MarkReadOutput{}, nil
}

type MarkAllReadInput struct{}
type MarkAllReadOutput struct{ UpdatedCount int32 }
type MarkAllReadUseCase struct{ deps Deps }

func NewMarkAllReadUseCase(deps Deps) *MarkAllReadUseCase {
	return &MarkAllReadUseCase{deps: deps}
}

func (uc *MarkAllReadUseCase) Execute(ctx context.Context, _ MarkAllReadInput) (*MarkAllReadOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	count, err := uc.deps.Notifications.MarkAllRead(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &MarkAllReadOutput{UpdatedCount: int32(count)}, nil
}

type UnreadCountInput struct{}
type UnreadCountOutput struct{ Count int32 }
type UnreadCountUseCase struct{ deps Deps }

func NewUnreadCountUseCase(deps Deps) *UnreadCountUseCase {
	return &UnreadCountUseCase{deps: deps}
}

func (uc *UnreadCountUseCase) Execute(ctx context.Context, _ UnreadCountInput) (*UnreadCountOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	count, err := uc.deps.Notifications.CountUnread(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread notifications: %w", err)
	}
	return &UnreadCountOutput{Count: count}, nil
}
