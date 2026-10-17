package user

import (
	"time"

	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
)

type Deps struct {
	Users    domain.UserRepository
	Cache    domain.Cache[domain.CachedUser]
	CacheTTL time.Duration
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}
