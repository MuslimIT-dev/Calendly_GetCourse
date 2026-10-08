package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

    "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
    "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/repository/db"
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
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}, nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id int32) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &domain.User{
		ID:            row.ID,
		Name:          row.Name.String,
		Email:         row.Email,
		AvatarURL:     row.AvatarUrl.String,
		Timezone:      row.Timezone.String,
		EmailVerified: row.EmailVerified.Bool,
		Roles:         convertRoleIDsToRoles(row.RoleIds),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func convertRoleIDsToRoles(roleIDs []int32) []domain.Role {
	roles := make([]domain.Role, len(roleIDs))
	for i, roleID := range roleIDs {
		roles[i] = domain.Role(roleID)
	}
	return roles
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &domain.User{
		ID:            row.ID,
		Name:          row.Name.String,
		Email:         row.Email,
		AvatarURL:     row.AvatarUrl.String,
		Timezone:      row.Timezone.String,
		EmailVerified: row.EmailVerified.Bool,
		Roles:         convertRoleIDsToRoles(row.RoleIds),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	row, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:        u.ID,
		AvatarUrl  u.AvatarURL,
		Name:	   u.Name,
		Timezone:  u.Timezone,
	})
	if err != nil {
		return nil, err	
	}

	return &domain.User{
		ID:            row.ID,
		Name:          row.Name.String,
		Email:         row.Email,
		AvatarURL:     row.AvatarUrl.String,
		Timezone:      row.Timezone.String,
		EmailVerified: row.EmailVerified.Bool,
		Roles:         convertRoleIDsToRoles(row.RoleIds),
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}