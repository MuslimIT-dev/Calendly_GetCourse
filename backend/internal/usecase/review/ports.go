package review

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Reviews  domain.ReviewRepository
	Bookings domain.BookingRepository
	Masters  domain.MasterRepository
	Cache    domain.Cache[domain.CachedReviewSummary]
	CacheTTL time.Duration
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, event *any) error
}

func publishEvent(ctx context.Context, publisher EventPublisher, event any) error {
	return publisher.Publish(ctx, "reviews", &event)
}
