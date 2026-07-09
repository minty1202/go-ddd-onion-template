package todorepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/minty1202/go-ddd-onion-template/internal/domain/aggregate"
	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/db"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/repository"
)

type repo struct {
	queries *db.Queries
}

var _ todo.Repository = (*repo)(nil)

func New(queries *db.Queries) *repo {
	return &repo{queries: queries}
}

func (r *repo) Add(ctx context.Context, t *todo.Todo) error {
	params := db.InsertTodoParams{
		ID:        t.ID().String(),
		Title:     t.Title(),
		Body:      t.Body(),
		Completed: t.Completed(),
		Version:   int32(t.Version()),
	}

	return r.queries.InsertTodo(ctx, params)
}

func (r *repo) Save(ctx context.Context, t *todo.Todo) error {
	return repository.SaveWithLock(ctx, t, func(ctx context.Context, expectedVersion int) (int, error) {
		params := db.UpdateTodoParams{
			ID:        t.ID().String(),
			Title:     t.Title(),
			Body:      t.Body(),
			Completed: t.Completed(),
			Version:   int32(expectedVersion),
		}

		newVersion, err := r.queries.UpdateTodo(ctx, params)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return 0, aggregate.ErrConflict
			}

			return 0, err
		}
		return int(newVersion), nil
	})
}

func (r *repo) FindByID(ctx context.Context, id todo.ID) (*todo.Todo, error) {
	row, err := r.queries.GetTodo(ctx, id.String())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, todo.ErrNotFound
		}

		return nil, err
	}

	return todo.Reconstruct(row.ID, row.Title, row.Body, row.Completed, int(row.Version))
}

func (r *repo) RemoveByID(ctx context.Context, id todo.ID) error {
	return r.queries.DeleteTodoByID(ctx, id.String())
}
