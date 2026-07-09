package view

import (
	"errors"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
)

func fromRepoError(err error) *errdefs.UseCaseError {
	switch {
	case errors.Is(err, todo.ErrNotFound):
		return errdefs.NewNotFound(err)
	default:
		return errdefs.NewInternal(err)
	}
}
