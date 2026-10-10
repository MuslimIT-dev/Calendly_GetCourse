package location

import (
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type Deps struct {
	Locations domain.LocationRepository
	Masters   domain.MasterRepository
	Cache     domain.Cache[domain.CachedLocations]
	CacheTTL  time.Duration
}