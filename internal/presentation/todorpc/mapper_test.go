package todorpc

import (
	"testing"

	todov1 "github.com/minty1202/go-ddd-onion-template/gen/todo/v1"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/define"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/view"
	"github.com/stretchr/testify/assert"
)

func TestToDefineParam(t *testing.T) {
	req := &todov1.DefineRequest{Title: "title", Body: "body"}
	got := toDefineParam(req)
	assert.Equal(t, "title", got.Title)
	assert.Equal(t, "body", got.Body)
}

func TestToViewParam(t *testing.T) {
	ulid := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	req := &todov1.ViewRequest{Id: ulid}
	got := toViewParam(req)

	assert.Equal(t, ulid, got.ID)
}

func TestToDefineResponse(t *testing.T) {
	ulid := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	r := &define.Result{
		ID:        ulid,
		Title:     "title",
		Body:      "body",
		Completed: false,
		Version:   0,
	}
	got := toDefineResponse(r)

	assert.Equal(t, ulid, got.Todo.Id)
	assert.Equal(t, "title", got.Todo.Title)
	assert.Equal(t, "body", got.Todo.Body)
	assert.False(t, got.Todo.Completed)
	assert.Equal(t, "0", got.Todo.Etag)
}

func TestToViewResponse(t *testing.T) {
	ulid := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	r := &view.Result{
		ID:        ulid,
		Title:     "title",
		Body:      "body",
		Completed: true,
		Version:   7,
	}
	got := toViewResponse(r)

	assert.Equal(t, ulid, got.Todo.Id)
	assert.Equal(t, "title", got.Todo.Title)
	assert.Equal(t, "body", got.Todo.Body)
	assert.True(t, got.Todo.Completed)
	assert.Equal(t, "7", got.Todo.Etag)
}
