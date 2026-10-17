package master

import (
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Masters   domain.MasterRepository
	Users     domain.UserRepository
	Cache     domain.Cache[domain.CachedMaster]
	CacheTTL  time.Duration
}