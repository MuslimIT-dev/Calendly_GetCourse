package appcontext

import "context"

type key int

const (
	userIDKey key = iota
	rolesKey
)

func WithUserID(ctx context.Context, id int32) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserID(ctx context.Context) (int32, bool) {
	id, ok := ctx.Value(userIDKey).(int32)
	return id, ok
}

func WithRoles(ctx context.Context, roles []int32) context.Context {
	return context.WithValue(ctx, rolesKey, roles)
}

func Roles(ctx context.Context) ([]int32, bool) {
	roles, ok := ctx.Value(rolesKey).([]int32)
	return roles, ok
}

func HasRole(ctx context.Context, role int32) bool {
	roles, ok := Roles(ctx)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}