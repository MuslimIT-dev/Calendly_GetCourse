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

func (r *UserRepo) SetEmailVerified(ctx context.Context, id int32) (error) {
	err := r.q.SetEmailVerified(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int32, passwordHash string) error {
	err := r.q.UpdatePassword(ctx, db.UpdatePasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, id int32) error {
	err := r.q.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) AssignRole(ctx context.Context, userID int32, roleID domain.Role) error {
	err := r.q.AssignRole(ctx, db.AssignRoleParams{
		UserID: userID,
		RoleID: int32(roleID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) RemoveRole(ctx context.Context, userID int32, roleID domain.Role) error {
	err := r.q.RemoveRole(ctx, db.RemoveRoleParams{
		UserID: userID,
		RoleID: int32(roleID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetUserRoles(ctx context.Context, userID int32) ([]domain.Role, error) {
	rows, err := r.q.GetUserRoles(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	roles := make([]domain.Role, len(rows))
	for i, row := range rows {
		roles[i] = domain.Role(row.RoleID)
	}
	return roles, nil
}

func (r *UserRepo) DeleteUserRoles(ctx context.Context, userID int32) error {
	err := r.q.SetUserRoles(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) AddUserRoles(ctx context.Context, userID int32, roleIDs []domain.Role) error {
	intRoleIDs := make([]int32, len(roleIDs))
	for i, roleID := range roleIDs {
		intRoleIDs[i] = int32(roleID)
	}

	err := r.q.AddUserRoles(ctx, db.AddUserRolesParams{
		UserID:  userID,
		RoleIDs: intRoleIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *UserRepo) HasRole(ctx context.Context, userID int32, roleID domain.Role) (bool, error) {
	hasRole, err := r.q.HasRole(ctx, db.HasRoleParams{
		UserID: userID,
		RoleID: int32(roleID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, domain.ErrUserNotFound
		}
		return false, err
	}
	return hasRole, nil
}