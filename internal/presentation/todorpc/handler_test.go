package todorpc

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	todov1 "github.com/minty1202/go-ddd-onion-template/gen/todo/v1"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/define"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDefineUseCase struct {
	result *define.Result
	err    error
}

func (f *fakeDefineUseCase) Execute(_ context.Context, _ define.Param) (*define.Result, error) {
	return f.result, f.err
}

type fakeViewUseCase struct {
	result *view.Result
	err    error
}

func (f *fakeViewUseCase) Execute(_ context.Context, _ view.Param) (*view.Result, error) {
	return f.result, f.err
}

func TestHandler_Define_Success(t *testing.T) {
	ulid := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	fakeDefine := &fakeDefineUseCase{
		result: &define.Result{
			ID:        ulid,
			Title:     "title",
			Body:      "body",
			Completed: false,
			Version:   0,
		},
	}
	h := newHandler(fakeDefine, &fakeViewUseCase{}, nil)

	resp, err := h.Define(context.Background(), &todov1.DefineRequest{Title: "title", Body: "body"})

	require.NoError(t, err)
	require.NotNil(t, resp.Todo)
	assert.Equal(t, ulid, resp.Todo.Id)
	assert.Equal(t, "title", resp.Todo.Title)
	assert.Equal(t, "body", resp.Todo.Body)
	assert.False(t, resp.Todo.Completed)
	assert.Equal(t, "0", resp.Todo.Etag)
}

func TestHandler_Define_UseCaseError(t *testing.T) {
	fakeDefine := &fakeDefineUseCase{
		err: errdefs.NewInvalidArgument(errors.New("invalid")),
	}
	h := newHandler(fakeDefine, &fakeViewUseCase{}, nil)

	_, err := h.Define(context.Background(), &todov1.DefineRequest{})

	require.Error(t, err)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestHandler_View_Success(t *testing.T) {
	ulid := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	fakeView := &fakeViewUseCase{
		result: &view.Result{
			ID:        ulid,
			Title:     "title",
			Body:      "body",
			Completed: false,
			Version:   5,
		},
	}
	h := newHandler(&fakeDefineUseCase{}, fakeView, nil)

	resp, err := h.View(context.Background(), &todov1.ViewRequest{Id: ulid})

	require.NoError(t, err)
	require.NotNil(t, resp.Todo)
	assert.Equal(t, ulid, resp.Todo.Id)
	assert.Equal(t, "title", resp.Todo.Title)
	assert.Equal(t, "body", resp.Todo.Body)
	assert.False(t, resp.Todo.Completed)
	assert.Equal(t, "5", resp.Todo.Etag)
}

func TestHandler_View_NotFound(t *testing.T) {
	ulid := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	fakeView := &fakeViewUseCase{
		err: errdefs.NewNotFound(errors.New("not found")),
	}
	h := newHandler(&fakeDefineUseCase{}, fakeView, nil)

	_, err := h.View(context.Background(), &todov1.ViewRequest{Id: ulid})

	require.Error(t, err)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}
