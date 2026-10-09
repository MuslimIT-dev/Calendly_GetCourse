package middleware

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

type TokenVerifier interface {
	VerifyAccess(token string) (int32, []domain.Role, error)
}

var publicMethods = map[string]struct{}{
	"/auth.v1.AuthService/Register":       {},
	"/auth.v1.AuthService/Login":          {},
	"/auth.v1.AuthService/RefreshToken":   {},
	"/auth.v1.AuthService/VerifyEmail":    {},
	"/auth.v1.AuthService/ForgotPassword": {},
	"/auth.v1.AuthService/ResetPassword":  {},

	"/master.v1.MasterService/ListMasters": {},
	"/master.v1.MasterService/GetMaster":   {},

	"/catalog.v1.CatalogService/ListCategories":      {},
	"/catalog.v1.CatalogService/ListDefaultServices": {},
	"/catalog.v1.CatalogService/GetDefaultService":   {},
}

func NewAuthInterceptor(verifier TokenVerifier) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if _, ok := publicMethods[req.Spec().Procedure]; ok {
				return next(ctx, req)
			}

			header := req.Header().Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing bearer token"))
			}

			token := strings.TrimPrefix(header, "Bearer ")
			userID, roles, err := verifier.VerifyAccess(token)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			ctx = appcontext.WithUserID(ctx, userID)
			roleIDs := make([]int32, len(roles))
			for i, role := range roles {
				roleIDs[i] = int32(role)
			}
			ctx = appcontext.WithRoles(ctx, roleIDs)

			return next(ctx, req)
		}
	}
}
