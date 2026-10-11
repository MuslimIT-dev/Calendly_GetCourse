package service

import (
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type Deps struct {
	Services domain.ServiceRepository
	Masters  domain.MasterRepository
	Cache    domain.Cache[domain.CachedServices]
	CacheTTL time.Duration
}