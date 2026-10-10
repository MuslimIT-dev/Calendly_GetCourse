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
	case errors.Is(err, domain.ErrUnauthorized):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, domain.ErrWeakPassword):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, domain.ErrTokenExpired):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, domain.ErrTokenNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, domain.ErrSessionNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, domain.ErrEmailNotVerified):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, domain.ErrPwnedPassword):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, domain.ErrInvalidInput):
		return connect.NewError(connect.CodeInvalidArgument, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}