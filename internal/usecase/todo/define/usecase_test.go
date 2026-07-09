package define_test

import (
	"context"
	"errors"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/define"
	mocktodo "github.com/minty1202/go-ddd-onion-template/mocks/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDefineUseCase_Execute_Success(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)

	var captured *todo.Todo

	repo.EXPECT().Add(mock.Anything, mock.Anything).
		Run(func(_ context.Context, agg *todo.Todo) {
			captured = agg
		}).
		Return(nil)

	uc := define.NewUseCase(repo)
	got, err := uc.Execute(context.Background(), define.Param{Title: "title", Body: "body"})

	require.NoError(t, err)
	require.NotNil(t, captured)

	assert.Equal(t, captured.ID().String(), got.ID)

	assert.Equal(t, "title", got.Title)
	assert.Equal(t, "body", got.Body)
	assert.False(t, got.Completed)
	assert.Equal(t, 0, got.Version)
}

func TestDefineUseCase_Execute_ValidationError(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)

	uc := define.NewUseCase(repo)
	_, err := uc.Execute(context.Background(), define.Param{})

	var ucErr *errdefs.UseCaseError
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, errdefs.InvalidArgument, ucErr.Kind)
}

func TestDefineUseCase_Execute_RepoError(t *testing.T) {
	repo := mocktodo.NewMockRepository(t)
	repo.EXPECT().Add(mock.Anything, mock.Anything).Return(errors.New("db error"))

	uc := define.NewUseCase(repo)
	_, err := uc.Execute(context.Background(), define.Param{Title: "title", Body: "body"})

	var ucErr *errdefs.UseCaseError
	require.ErrorAs(t, err, &ucErr)
	assert.Equal(t, errdefs.Internal, ucErr.Kind)
}
