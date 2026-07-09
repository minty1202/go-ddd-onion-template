package todo

import "errors"

var (
	ErrInvalid          = errors.New("todo: invalid")
	ErrAlreadyCompleted = errors.New("todo: already completed")
	ErrNotFound         = errors.New("todo: not found")
)
