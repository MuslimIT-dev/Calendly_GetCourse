package auth

import (
	"time"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

type TokenService interface {
	GeneratePair(userID int32, roles []domain.Role) (access, refresh string, expiresIn int64, err error)
	VerifyRefresh(token string) (userID int32, roles []domain.Role, err error)
}

type SessionValue struct {
	UserID    int32     `json:"user_id"`
	Roles     []int32   `json:"roles"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type VerifyEmailValue struct {
	UserID    int32     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRegisteredEvent struct {
	UserID      int32  `json:"user_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	VerifyToken string `json:"verify_token"`
}

type Deps struct {
	Users          domain.UserRepository
	Sessions       domain.Cache[SessionValue]
	VerifyTokens   domain.Cache[VerifyEmailValue]
	Hasher         PasswordHasher
	Tokens         TokenService
	Events         domain.Publisher[UserRegisteredEvent]
	SessionTTL     time.Duration
	VerifyTokenTTL time.Duration
}
