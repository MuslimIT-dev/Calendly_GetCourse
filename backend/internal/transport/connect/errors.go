package connect

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailAlreadyTaken):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, domain.ErrUserNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, domain.ErrInvalidToken):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, domain.ErrInvalidCredentials):
		return connect.NewError(connect.CodeUnauthenticated, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}