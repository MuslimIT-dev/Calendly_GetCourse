package domain

import "context"

type User struct {
    ID            int32
    Name          string
    Email         string
    PasswordHash  string
    AvatarURL     string
    Timezone      string
    EmailVerified bool
    Roles         []Role

	CreatedAt     time.Time
    UpdatedAt     time.Time
}

type Role int32

const (
	RoleUnspecified Role = 0
	RoleAdmin		Role = 1
	RoleMaster		Role = 2
	RoleUser		Role = 3
)

type UserRepository interface {
	Create(ctx context.Context, u *User, roleIDs []Role) (*User, error)
	GetUserByID(ctx context.Context, id int32) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, u *User) (*User, error)
	SetEmailVerified(ctx context.Context, id int32) error
	UpdatePassword(ctx context.Context, id int32, passwordHash string) error
	DeleteUser(ctx context.Context, id int32) error

	AssignRole(ctx context.Context, userID int32, roleID Role) error
	RemoveRole(ctx context.Context, userID int32, roleID Role) error
	GetUserRoles(ctx context.Context, userID int32) ([]Role, error)
	DeleteUserRoles(ctx context.Context, userID int32) error
	AddUserRoles(ctx context.Context, userID int32, roleIDs []Role) error
	HasRole(ctx context.Context, userID int32, roleID Role) (bool, error)
}