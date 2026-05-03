package domain

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrResourceNotFound   = errors.New("resource not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
)
