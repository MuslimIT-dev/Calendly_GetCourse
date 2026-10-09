package jwt

import (
	"errors"
	"fmt"
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

func (s *JWTService) VerifyAccess(token string) (int32, []domain.Role, error) {
	return s.verify(token, "access")
}

func (s *JWTService) VerifyRefresh(token string) (int32, []domain.Role, error) {
	return s.verify(token, "refresh")
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

func (s *JWTService) verify(tokenStr, typ string) (int32, []domain.Role, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return 0, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, nil, errors.New("invalid token")
	}

	if t, _ := claims["type"].(string); t != typ {
		return 0, nil, errors.New("wrong token type")
	}

	sub, ok := claims["sub"].(float64)
	if !ok {
		return 0, nil, errors.New("invalid sub claim")
	}

	rawRoles, _ := claims["roles"].([]any)
	roles := make([]domain.Role, 0, len(rawRoles))
	for _, r := range rawRoles {
		if f, ok := r.(float64); ok {
			roles = append(roles, domain.Role(int32(f)))
		}
	}

	return int32(sub), roles, nil
}