package domain

import "errors"

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrResourceNotFound   = errors.New("resource not found")
)
