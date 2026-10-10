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
    ErrUnauthorized              = errors.New("unauthorized")
    ErrWeakPassword              = errors.New("password is too weak")
    ErrTokenExpired              = errors.New("token has expired")
    ErrTokenNotFound             = errors.New("token not found")
    ErrSessionNotFound           = errors.New("session not found")
    ErrEmailNotVerified          = errors.New("email not verified")
    ErrPwnedPassword             = errors.New("password has been exposed in a data breach")
)