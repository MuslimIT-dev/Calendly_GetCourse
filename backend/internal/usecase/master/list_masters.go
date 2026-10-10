package master

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type ListMastersInput struct {
	PageSize  int32
	PageToken string
	Filter    domain.MasterFilter
}

type ListMastersOutput struct {
	Cards         []*domain.MasterCardData
	NextPageToken string
}

type ListMastersUseCase struct {
	deps Deps
}

func NewListMastersUseCase(deps Deps) *ListMastersUseCase {
	return &ListMastersUseCase{deps: deps}
}

func (uc *ListMastersUseCase) Execute(ctx context.Context, in ListMastersInput) (*ListMastersOutput, error) {
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	cursor, err := decodeCursor(in.PageToken)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}

	cards, err := uc.deps.Masters.ListCards(ctx, in.Filter, cursor, pageSize+1)
	if err != nil {
		return nil, fmt.Errorf("list masters: %w", err)
	}

	var nextToken string
	if len(cards) > int(pageSize) {
		cards = cards[:pageSize]
		last := cards[len(cards)-1]
		nextToken = encodeCursor(buildCursor(in.Filter, last))
	}

	return &ListMastersOutput{
		Cards:         cards,
		NextPageToken: nextToken,
	}, nil
}

func buildCursor(f domain.MasterFilter, c *domain.MasterCardData) *domain.ListCursor {
	cursor := &domain.ListCursor{ID: c.Profile.ID}
	switch f.SortBy {
	case domain.SortTypeRating:
		v := float64(c.Profile.AvgRating)
		cursor.TimeValue = &v
	case domain.SortTypePrice:
		v := int32(c.MinPrice)
		cursor.IntValue = &v
	case domain.SortTypeExperience:
		v := int32(c.Profile.YearsOfExperience)
		cursor.IntValue = &v
	}
	return cursor
}

func encodeCursor(c *domain.ListCursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (*domain.ListCursor, error) {
	if s == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c domain.ListCursor
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
