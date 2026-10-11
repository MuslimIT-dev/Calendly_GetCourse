package review

import (
	"context"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type ListMyInput struct {
	TargetFilter *domain.ReviewTarget
	PageSize     int32
	PageToken    string
}

type ListMyOutput struct {
	Reviews       []*domain.ReviewCard
	NextPageToken string
}

type ListMyUseCase struct {
	deps Deps
}

func NewListMyUseCase(deps Deps) *ListMyUseCase {
	return &ListMyUseCase{deps: deps}
}

func (uc *ListMyUseCase) Execute(ctx context.Context, in ListMyInput) (*ListMyOutput, error) {
	userID, ok := appcontext.UserID(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	if in.TargetFilter != nil && !validTarget(*in.TargetFilter) {
		return nil, domain.ErrInvalidInput
	}
	cursor, err := decodeCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	pageSize := normalizePageSize(in.PageSize)
	rows, err := uc.deps.Reviews.ListMy(ctx, userID, in.TargetFilter, cursor, pageSize+1)
	if err != nil {
		return nil, err
	}
	var next string
	if len(rows) > int(pageSize) {
		rows = rows[:int(pageSize)]
		last := rows[len(rows)-1]
		next = encodeCursor(&domain.ReviewCursor{CreatedAt: last.Review.CreatedAt, ID: last.Review.ID})
	}
	return &ListMyOutput{Reviews: rows, NextPageToken: next}, nil
}
