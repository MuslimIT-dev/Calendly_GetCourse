package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

    "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
    "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/postgres/db"
)

type UserRepo struct {
	q *db.Queries
}

func NewUserRepo(q *db.Queries) *UserRepo {
	return &UserRepo{q: q}
}

var _ domain.UserRepository = (*UserRepo)(nil)

func (r *UserRepo) Create(ctx context.Context, u *domain.User, roles []domain.Role) (*domain.User, error) {
	roleIDs := make([]int32, len(roles))
	for i, role := range roles {
		roleIDs[i] = int32(role)
	}

	row, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		RoleIDs:      roleIDs,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "email") {
			return nil, domain.ErrEmailAlreadyTaken
		}
		return nil, err
	}

	return &domain.User{
		ID:            row.ID,
		Name:          row.Name,
		Email:         row.Email,
		AvatarURL:     row.AvatarURL,
		Timezone:      row.Timezone,
		EmailVerified: row.EmailVerified,
		Roles:         roles,
	}, nil
}