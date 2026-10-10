package master

import (
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type Deps struct {
	Masters   domain.MasterRepository
	Users     domain.UserRepository
	Cache     domain.Cache[domain.CachedMaster]
	CacheTTL  time.Duration
}