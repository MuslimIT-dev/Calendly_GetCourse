package notification

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type DeleteInput struct{ ID int64 }
type DeleteOutput struct{}
type DeleteUseCase struct{ deps Deps }

func NewDeleteUseCase(deps Deps) *DeleteUseCase { return &DeleteUseCase{deps: deps} }

func (uc *DeleteUseCase) Execute(ctx context.Context, in DeleteInput) (*DeleteOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	rows, err := uc.deps.Notifications.Delete(ctx, in.ID, userID)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		return nil, domain.ErrNotificationNotFound
	}
	return &DeleteOutput{}, nil
}
