package domain

import (
	"context"
	"time"
)

type ReviewTarget int32

const (
	ReviewTargetMaster ReviewTarget = 1
	ReviewTargetCourse ReviewTarget = 2
)

type ReviewStatus int32

const (
	ReviewPublished ReviewStatus = 1
	ReviewHidden    ReviewStatus = 2
	ReviewDeleted   ReviewStatus = 3
)

type Review struct {
	ID          int64
	AuthorID    int32
	TargetID    int32
	Target      ReviewTarget
	ReferenceID string
	Rating      int32
	Comment     string
	Reply       string
	RepliedAt   *time.Time
	Status      ReviewStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ReviewCard struct {
	Review       *Review
	AuthorName   string
	AuthorAvatar string
}

type ReviewSummary struct {
	TargetID     int32
	Target       ReviewTarget
	AvgRating    float32
	TotalReviews int32
	Rating1Count int32
	Rating2Count int32
	Rating3Count int32
	Rating4Count int32
	Rating5Count int32
}

type CachedReviewSummary struct {
	Summary   *ReviewSummary `json:"summary"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type ReviewCursor struct {
	CreatedAt time.Time
	ID        int64
}

type ReviewRepository interface {
	Create(ctx context.Context, review *Review) (*Review, error)
	GetByID(ctx context.Context, id int64) (*ReviewCard, error)
	ExistsForReference(ctx context.Context, authorID int32, referenceID string) (bool, error)
	ListByTarget(ctx context.Context, target ReviewTarget, targetID int32, cursor *ReviewCursor, limit int32) ([]*ReviewCard, error)
	ListMy(ctx context.Context, authorID int32, targetFilter *ReviewTarget, cursor *ReviewCursor, limit int32) ([]*ReviewCard, error)
	ListMyTarget(ctx context.Context, target ReviewTarget, targetID int32, onlyUnanswered bool, cursor *ReviewCursor, limit int32) ([]*ReviewCard, error)
	Update(ctx context.Context, review *Review) (*Review, error)
	Reply(ctx context.Context, id int64, reply string) (*Review, error)
	SoftDelete(ctx context.Context, id int64, authorID int32) (*Review, error)
	Hide(ctx context.Context, id int64) (*Review, error)
	GetSummary(ctx context.Context, target ReviewTarget, targetID int32) (*ReviewSummary, error)
	BelongsToAuthor(ctx context.Context, id int64, authorID int32) (bool, error)
}
