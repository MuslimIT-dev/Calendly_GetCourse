package review

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type DeleteInput struct {
	ID int64
}

type DeleteOutput struct{}

type DeleteUseCase struct {
	deps Deps
}

func NewDeleteUseCase(deps Deps) *DeleteUseCase {
	return &DeleteUseCase{deps: deps}
}

func (uc *DeleteUseCase) Execute(ctx context.Context, in DeleteInput) (*DeleteOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	card, err := uc.deps.Reviews.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	review := card.Review
	if review.AuthorID != userID && !appcontext.HasRole(ctx, int32(domain.RoleAdmin)) {
		return nil, domain.ErrForbidden
	}
	deleted, err := uc.deps.Reviews.SoftDelete(ctx, in.ID, review.AuthorID)
	if err != nil {
		return nil, err
	}
	invalidateSummary(ctx, uc.deps.Cache, deleted.Target, deleted.TargetID)
	return &DeleteOutput{}, nil
}
