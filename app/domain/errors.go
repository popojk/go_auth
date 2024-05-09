package domain

import "errors"

var (
	ErrInternalServerError = errors.New("internal Server Error")
	ErrNotFound            = errors.New("your request item is not found")
	ErrConflict            = errors.New("your item already exist")
	ErrBadParamInput       = errors.New("givenParam is not Valid")
)
