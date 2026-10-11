package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
)

type ReviewRepo struct {
	q *db.Queries
}

func NewReviewRepo(q *db.Queries) *ReviewRepo {
	return &ReviewRepo{q: q}
}

var _ domain.ReviewRepository = (*ReviewRepo)(nil)

func (r *ReviewRepo) Create(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	row, err := r.q.CreateReview(ctx, db.CreateReviewParams{
		AuthorID:    review.AuthorID,
		TargetID:    review.TargetID,
		Target:      int16(review.Target),
		ReferenceID: review.ReferenceID,
		Rating:      int16(review.Rating),
		Comment:     pgtype.Text{String: review.Comment, Valid: review.Comment != ""},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrReviewAlreadyExists
		}
		return nil, err
	}
	return mapReview(row), nil
}

func (r *ReviewRepo) GetByID(ctx context.Context, id int64) (*domain.ReviewCard, error) {
	row, err := r.q.GetReviewByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, err
	}
	return &domain.ReviewCard{
		Review:       mapReviewFields(row.ID, row.AuthorID, row.TargetID, row.Target, row.ReferenceID, row.Rating, row.Comment, row.Reply, row.RepliedAt, row.Status, row.CreatedAt, row.UpdatedAt),
		AuthorName:   row.AuthorName.String,
		AuthorAvatar: row.AuthorAvatar.String,
	}, nil
}

func (r *ReviewRepo) ExistsForReference(ctx context.Context, authorID int32, referenceID string) (bool, error) {
	return r.q.ReviewExistsForReference(ctx, db.ReviewExistsForReferenceParams{
		AuthorID: authorID, ReferenceID: referenceID,
	})
}

func (r *ReviewRepo) ListByTarget(ctx context.Context, target domain.ReviewTarget, targetID int32, cursor *domain.ReviewCursor, limit int32) ([]*domain.ReviewCard, error) {
	cursorTime, cursorID := reviewCursorParams(cursor)
	rows, err := r.q.ListTargetReviews(ctx, db.ListTargetReviewsParams{
		Target: int16(target), TargetID: targetID, CursorTime: cursorTime, CursorID: cursorID, PageSize: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.ReviewCard, len(rows))
	for i, row := range rows {
		out[i] = &domain.ReviewCard{
			Review:       mapReviewFields(row.ID, row.AuthorID, row.TargetID, row.Target, row.ReferenceID, row.Rating, row.Comment, row.Reply, row.RepliedAt, row.Status, row.CreatedAt, row.UpdatedAt),
			AuthorName:   row.AuthorName.String,
			AuthorAvatar: row.AuthorAvatar.String,
		}
	}
	return out, nil
}

func (r *ReviewRepo) ListMy(ctx context.Context, authorID int32, targetFilter *domain.ReviewTarget, cursor *domain.ReviewCursor, limit int32) ([]*domain.ReviewCard, error) {
	var filter pgtype.Int2
	if targetFilter != nil {
		filter = pgtype.Int2{Int16: int16(*targetFilter), Valid: true}
	}
	cursorTime, cursorID := reviewCursorParams(cursor)
	rows, err := r.q.ListMyReviews(ctx, db.ListMyReviewsParams{
		AuthorID: authorID, TargetFilter: filter, CursorTime: cursorTime, CursorID: cursorID, PageSize: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.ReviewCard, len(rows))
	for i, row := range rows {
		out[i] = &domain.ReviewCard{
			Review:       mapReviewFields(row.ID, row.AuthorID, row.TargetID, row.Target, row.ReferenceID, row.Rating, row.Comment, row.Reply, row.RepliedAt, row.Status, row.CreatedAt, row.UpdatedAt),
			AuthorName:   row.AuthorName.String,
			AuthorAvatar: row.AuthorAvatar.String,
		}
	}
	return out, nil
}

func (r *ReviewRepo) ListMyTarget(ctx context.Context, target domain.ReviewTarget, targetID int32, onlyUnanswered bool, cursor *domain.ReviewCursor, limit int32) ([]*domain.ReviewCard, error) {
	var only pgtype.Bool
	if onlyUnanswered {
		only = pgtype.Bool{Bool: true, Valid: true}
	}
	cursorTime, cursorID := reviewCursorParams(cursor)
	rows, err := r.q.ListMyTargetReviews(ctx, db.ListMyTargetReviewsParams{
		Target: int16(target), TargetID: targetID, OnlyUnanswered: only,
		CursorTime: cursorTime, CursorID: cursorID, PageSize: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.ReviewCard, len(rows))
	for i, row := range rows {
		out[i] = &domain.ReviewCard{
			Review:       mapReviewFields(row.ID, row.AuthorID, row.TargetID, row.Target, row.ReferenceID, row.Rating, row.Comment, row.Reply, row.RepliedAt, row.Status, row.CreatedAt, row.UpdatedAt),
			AuthorName:   row.AuthorName.String,
			AuthorAvatar: row.AuthorAvatar.String,
		}
	}
	return out, nil
}

func (r *ReviewRepo) Update(ctx context.Context, review *domain.Review) (*domain.Review, error) {
	row, err := r.q.UpdateReview(ctx, db.UpdateReviewParams{
		ID: review.ID, AuthorID: review.AuthorID,
		Rating:  pgtype.Int2{Int16: int16(review.Rating), Valid: true},
		Comment: pgtype.Text{String: review.Comment, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, err
	}
	return mapReview(row), nil
}

func (r *ReviewRepo) Reply(ctx context.Context, id int64, reply string) (*domain.Review, error) {
	row, err := r.q.ReplyToReview(ctx, db.ReplyToReviewParams{
		ID: id, Reply: pgtype.Text{String: reply, Valid: reply != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapReview(row), nil
}

func (r *ReviewRepo) SoftDelete(ctx context.Context, id int64, authorID int32) (*domain.Review, error) {
	row, err := r.q.DeleteReview(ctx, db.DeleteReviewParams{ID: id, AuthorID: authorID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrReviewNotFound
		}
		return nil, err
	}
	review, err := r.q.GetReviewByID(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	return mapReviewFields(review.ID, review.AuthorID, review.TargetID, review.Target, review.ReferenceID, review.Rating, review.Comment, review.Reply, review.RepliedAt, row.Status, review.CreatedAt, review.UpdatedAt), nil
}

func (r *ReviewRepo) Hide(ctx context.Context, id int64) (*domain.Review, error) {
	row, err := r.q.HideReview(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidStatus
		}
		return nil, err
	}
	return mapReview(row), nil
}

func (r *ReviewRepo) GetSummary(ctx context.Context, target domain.ReviewTarget, targetID int32) (*domain.ReviewSummary, error) {
	row, err := r.q.GetReviewSummary(ctx, db.GetReviewSummaryParams{Target: int16(target), TargetID: targetID})
	if err != nil {
		return nil, err
	}
	return &domain.ReviewSummary{
		TargetID: row.TargetID, Target: domain.ReviewTarget(row.Target),
		AvgRating: float32(row.AvgRating), TotalReviews: row.TotalReviews,
		Rating1Count: row.Rating1Count, Rating2Count: row.Rating2Count,
		Rating3Count: row.Rating3Count, Rating4Count: row.Rating4Count, Rating5Count: row.Rating5Count,
	}, nil
}

func (r *ReviewRepo) BelongsToAuthor(ctx context.Context, id int64, authorID int32) (bool, error) {
	return r.q.ReviewBelongsToAuthor(ctx, db.ReviewBelongsToAuthorParams{ID: id, AuthorID: authorID})
}

func reviewCursorParams(cursor *domain.ReviewCursor) (pgtype.Timestamptz, pgtype.Int8) {
	if cursor == nil {
		return pgtype.Timestamptz{}, pgtype.Int8{}
	}
	return pgtype.Timestamptz{Time: cursor.CreatedAt, Valid: true},
		pgtype.Int8{Int64: cursor.ID, Valid: true}
}

func mapReview(row db.Review) *domain.Review {
	return mapReviewFields(row.ID, row.AuthorID, row.TargetID, row.Target, row.ReferenceID, row.Rating, row.Comment, row.Reply, row.RepliedAt, row.Status, row.CreatedAt, row.UpdatedAt)
}

func mapReviewFields(id int64, authorID, targetID int32, target int16, referenceID string, rating int16, comment, reply pgtype.Text, repliedAt pgtype.Timestamptz, status int16, createdAt, updatedAt pgtype.Timestamptz) *domain.Review {
	var replied *time.Time
	if repliedAt.Valid {
		replied = &repliedAt.Time
	}
	return &domain.Review{
		ID: id, AuthorID: authorID, TargetID: targetID, Target: domain.ReviewTarget(target),
		ReferenceID: referenceID, Rating: int32(rating), Comment: comment.String, Reply: reply.String,
		RepliedAt: replied, Status: domain.ReviewStatus(status), CreatedAt: createdAt.Time, UpdatedAt: updatedAt.Time,
	}
}
