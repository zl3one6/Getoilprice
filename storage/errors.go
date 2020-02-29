package storage

import "errors"

var (
	ErrAlreadyExists = errors.New("object already exists")
	ErrDoesNotExists = errors.New("object does not exists")
)
