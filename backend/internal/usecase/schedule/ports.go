package schedule

import (
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Schedule domain.ScheduleRepository
	Locations domain.LocationRepository
	Masters  domain.MasterRepository
	Services domain.ServiceRepository
	Cache    domain.Cache[domain.CachedScheduleRules]
	CacheTTL time.Duration
}