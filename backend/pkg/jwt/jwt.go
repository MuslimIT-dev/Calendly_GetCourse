package jwt

import (
    "time"

    "github.com/golang-jwt/jwt/v5"

    "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type JWTService struct {
    secret []byte
    issuer string
}

func NewJWTService(secret, issuer string) *JWTService {
    return &JWTService{secret: []byte(secret), issuer: issuer}
}

func (s *JWTService) GeneratePair(userID int32, roles []domain.Role) (string, string, int64, error) {
    expiresIn := int64(15 * 60)

    access, err := s.sign(userID, roles, 15*time.Minute, "access")
    if err != nil {
        return "", "", 0, err
    }

    refresh, err := s.sign(userID, roles, 30*24*time.Hour, "refresh")
    if err != nil {
        return "", "", 0, err
    }

    return access, refresh, expiresIn, nil
}

func (s *JWTService) sign(userID int32, roles []domain.Role, ttl time.Duration, typ string) (string, error) {
    roleInts := make([]int, len(roles))
    for i, r := range roles {
        roleInts[i] = int(r)
    }

    claims := jwt.MapClaims{
        "sub":   userID,
        "roles": roleInts,
        "type":  typ,
        "iss":   s.issuer,
        "exp":   time.Now().Add(ttl).Unix(),
        "iat":   time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secret)
}