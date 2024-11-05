package repository

import (
	"errors"
)

var (
	ErrResourceNotFound      = errors.New("resource not found")
	ErrBadRequest            = errors.New("bad request")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrUnknown               = errors.New("unknown error")
)
