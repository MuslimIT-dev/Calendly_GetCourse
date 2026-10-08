package domain

import "errors"

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrEmailAlreadyTaken = errors.New("email already taken")
    ErrInvalidToken      = errors.New("invalid token")
)