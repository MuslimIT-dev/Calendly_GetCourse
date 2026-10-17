package booking

import (
	"context"
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Bookings  domain.BookingRepository
	Locations domain.LocationRepository
	Services  domain.ServiceRepository
	Masters   domain.MasterRepository
	Users     domain.UserRepository
	Cache     domain.Cache[domain.CachedUser]
	CacheTTL  time.Duration
}

type EventPublisher interface {
	Publish(ctx context.Context, topic string, event *any) error
}

func publishEvent(ctx context.Context, publisher EventPublisher, event any) error {
	return publisher.Publish(ctx, "appointments", &event)
}
