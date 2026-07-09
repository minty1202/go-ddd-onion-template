package todo

import (
	"context"
)

type Repository interface {
	Add(ctx context.Context, todo *Todo) error
	Save(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id ID) (*Todo, error)
	RemoveByID(ctx context.Context, id ID) error
}
