package todorepo_test

import (
	"context"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/aggregate"
	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/db"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/db/dbtest"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/repository/todorepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTodoRepository_Add(t *testing.T) {
	pool := dbtest.Setup(t)
	repo := todorepo.New(db.New(pool))

	created, err := todo.NewTodo("title", "body")
	require.NoError(t, err)

	require.NoError(t, repo.Add(context.Background(), created))

	found, err := repo.FindByID(context.Background(), created.ID())
	require.NoError(t, err)
	assert.Equal(t, created.ID(), found.ID())
	assert.Equal(t, "title", found.Title())
	assert.Equal(t, "body", found.Body())
	assert.False(t, found.Completed())
	assert.Equal(t, 0, found.Version())
}

func TestTodoRepository_FindByID_NotFound(t *testing.T) {
	pool := dbtest.Setup(t)
	repo := todorepo.New(db.New(pool))

	_, err := repo.FindByID(context.Background(), todo.NewID())
	require.ErrorIs(t, err, todo.ErrNotFound)
}

func TestTodoRepository_Save(t *testing.T) {
	pool := dbtest.Setup(t)
	repo := todorepo.New(db.New(pool))

	created, err := todo.NewTodo("title", "body")
	require.NoError(t, err)
	require.NoError(t, repo.Add(context.Background(), created))

	found, err := repo.FindByID(context.Background(), created.ID())
	require.NoError(t, err)
	require.NoError(t, found.Complete())

	require.NoError(t, repo.Save(context.Background(), found))
	assert.Equal(t, 1, found.Version())

	reloaded, err := repo.FindByID(context.Background(), created.ID())
	require.NoError(t, err)
	assert.True(t, reloaded.Completed())
	assert.Equal(t, 1, reloaded.Version())
}

func TestTodoRepository_Save_NonExistent(t *testing.T) {
	pool := dbtest.Setup(t)
	repo := todorepo.New(db.New(pool))

	notExisting, err := todo.NewTodo("title", "body")
	require.NoError(t, err)

	err = repo.Save(context.Background(), notExisting)
	require.ErrorIs(t, err, aggregate.ErrConflict)
}

func TestTodoRepository_Save_OptimisticLock(t *testing.T) {
	pool := dbtest.Setup(t)
	repo := todorepo.New(db.New(pool))

	created, err := todo.NewTodo("title", "body")
	require.NoError(t, err)
	require.NoError(t, repo.Add(context.Background(), created))

	first, err := repo.FindByID(context.Background(), created.ID())
	require.NoError(t, err)
	second, err := repo.FindByID(context.Background(), created.ID())
	require.NoError(t, err)

	require.NoError(t, first.Complete())
	require.NoError(t, repo.Save(context.Background(), first))

	require.NoError(t, second.Complete())
	err = repo.Save(context.Background(), second)
	require.ErrorIs(t, err, aggregate.ErrConflict)
}

func TestTodoRepository_RemoveByID(t *testing.T) {
	pool := dbtest.Setup(t)
	repo := todorepo.New(db.New(pool))

	created, err := todo.NewTodo("title", "body")
	require.NoError(t, err)
	require.NoError(t, repo.Add(context.Background(), created))

	require.NoError(t, repo.RemoveByID(context.Background(), created.ID()))

	_, err = repo.FindByID(context.Background(), created.ID())
	require.ErrorIs(t, err, todo.ErrNotFound)
}
