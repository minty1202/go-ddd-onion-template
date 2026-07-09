package view_test

import (
	"context"
	"errors"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/view"
	mocktodo "github.com/minty1202/go-ddd-onion-template/mocks/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestViewUseCase_Execute_Success(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)
	id := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	foundTodo, err := todo.Reconstruct(id, "title", "body", false, 5)
	require.NoError(t, err)
	repo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(foundTodo, nil)

	uc := view.NewUseCase(repo)
	got, err := uc.Execute(context.Background(), view.Param{ID: foundTodo.ID().String()})

	require.NoError(t, err)
	assert.Equal(t, foundTodo.ID().String(), got.ID)
	assert.Equal(t, "title", got.Title)
	assert.Equal(t, "body", got.Body)
	assert.False(t, got.Completed)
	assert.Equal(t, 5, got.Version)
}

func TestViewUseCase_Execute_ValidationError(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)

	uc := view.NewUseCase(repo)
	_, err := uc.Execute(context.Background(), view.Param{})

	var ucErr *errdefs.UseCaseError
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, errdefs.InvalidArgument, ucErr.Kind)
}

func TestViewUseCase_Execute_RepoError(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	id := todo.NewID().String()
	uc := view.NewUseCase(repo)
	_, err := uc.Execute(context.Background(), view.Param{ID: id})

	var ucErr *errdefs.UseCaseError
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, errdefs.Internal, ucErr.Kind)
}

func TestViewUseCase_Execute_NotFound(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)
	repo.EXPECT().FindByID(mock.Anything, mock.Anything).Return(nil, todo.ErrNotFound)

	id := todo.NewID().String()
	uc := view.NewUseCase(repo)
	_, err := uc.Execute(context.Background(), view.Param{ID: id})

	var ucErr *errdefs.UseCaseError
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, errdefs.NotFound, ucErr.Kind)
}
