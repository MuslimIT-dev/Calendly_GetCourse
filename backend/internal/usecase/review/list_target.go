package review

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type ListTargetInput struct {
	Target    domain.ReviewTarget
	TargetID  int32
	PageSize  int32
	PageToken string
}

type ListTargetOutput struct {
	Reviews       []*domain.ReviewCard
	NextPageToken string
	Summary       *domain.ReviewSummary
}

type ListTargetUseCase struct {
	deps Deps
}

func NewListTargetUseCase(deps Deps) *ListTargetUseCase {
	return &ListTargetUseCase{deps: deps}
}

func (uc *ListTargetUseCase) Execute(ctx context.Context, in ListTargetInput) (*ListTargetOutput, error) {
	if !validTarget(in.Target) || in.TargetID <= 0 {
		return nil, domain.ErrInvalidInput
	}
	pageSize := normalizePageSize(in.PageSize)
	cursor, err := decodeCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	rows, err := uc.deps.Reviews.ListByTarget(ctx, in.Target, in.TargetID, cursor, pageSize+1)
	if err != nil {
		return nil, err
	}
	var next string
	if len(rows) > int(pageSize) {
		rows = rows[:int(pageSize)]
		last := rows[len(rows)-1]
		next = encodeCursor(&domain.ReviewCursor{CreatedAt: last.Review.CreatedAt, ID: last.Review.ID})
	}

	summary, err := getCachedSummary(ctx, uc.deps, in.Target, in.TargetID)
	if err != nil {
		return nil, err
	}
	return &ListTargetOutput{Reviews: rows, NextPageToken: next, Summary: summary}, nil
}

func normalizePageSize(size int32) int32 {
	if size <= 0 {
		return defaultPageSize
	}
	if size > maxPageSize {
		return maxPageSize
	}
	return size
}

func validTarget(target domain.ReviewTarget) bool {
	return target == domain.ReviewTargetMaster || target == domain.ReviewTargetCourse
}

func encodeCursor(cursor *domain.ReviewCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(token string) (*domain.ReviewCursor, error) {
	if token == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var cursor domain.ReviewCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID <= 0 || cursor.CreatedAt.IsZero() {
		return nil, domain.ErrInvalidInput
	}
	return &cursor, nil
}
