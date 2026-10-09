package domain

import "errors"

var (
    ErrUserNotFound              = errors.New("user not found")
    ErrEmailAlreadyTaken         = errors.New("email already taken")
    ErrInvalidToken              = errors.New("invalid token")
    ErrDatabaseConnectionFailed  = errors.New("database connection failed")
    ErrDatabaseOperationFailed   = errors.New("database operation failed")
    ErrRepositoryOperationFailed = errors.New("repository operation failed")
    ErrInvalidCredentials        = errors.New("invalid credentials")
)