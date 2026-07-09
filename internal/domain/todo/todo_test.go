package todo_test

import (
	"strings"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTodo_Valid(t *testing.T) {
	got, err := todo.NewTodo("title", "body")
	require.NoError(t, err)
	assert.NotNil(t, got.ID())
	assert.Equal(t, "title", got.Title())
	assert.Equal(t, "body", got.Body())
	assert.False(t, got.Completed())
	assert.Equal(t, 0, got.Version())
}

func TestNewTodo_Validation(t *testing.T) {
	tests := []struct {
		name       string
		inputTitle string
		inputBody  string
		wantErr    bool
	}{
		{name: "正常", inputTitle: "title", inputBody: "body", wantErr: false},
		{name: "日本語", inputTitle: "テストタイトル", inputBody: "テストボディ", wantErr: false},
		{name: "タイトル空文字", inputTitle: "", inputBody: "テストボディ", wantErr: true},
		{name: "ボディ空文字", inputTitle: "テストタイトル", inputBody: "", wantErr: true},
		{name: "タイトル 2 文字以下", inputTitle: "aa", inputBody: "テストボディ", wantErr: true},
		{name: "タイトル 3 文字以上", inputTitle: "aaa", inputBody: "テストボディ", wantErr: false},
		{name: "タイトル 10 文字以下", inputTitle: strings.Repeat("a", 10), inputBody: "テストボディ", wantErr: false},
		{name: "タイトル 11 文字以上", inputTitle: strings.Repeat("a", 11), inputBody: "テストボディ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := todo.NewTodo(tt.inputTitle, tt.inputBody)
			if tt.wantErr {
				require.ErrorIs(t, err, todo.ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReconstruct_Valid(t *testing.T) {
	s := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	tests := []struct {
		name      string
		completed bool
		version   int
	}{
		{name: "未完了", completed: false, version: 3},
		{name: "完了済み", completed: true, version: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := todo.Reconstruct(s, "title", "body", tt.completed, tt.version)
			require.NoError(t, err)
			assert.Equal(t, s, got.ID().String())
			assert.Equal(t, "title", got.Title())
			assert.Equal(t, "body", got.Body())
			assert.Equal(t, tt.completed, got.Completed())
			assert.Equal(t, tt.version, got.Version())
		})
	}
}

func TestReconstruct_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		title string
		body  string
	}{
		{name: "ID 空文字", id: "", title: "title", body: "body"},
		{name: "ID 不正文字", id: "invalid", title: "title", body: "body"},
		{name: "title 空文字", id: "01ARZ3NDEKTSV4RRFFQ69G5FAV", title: "", body: "body"},
		{name: "body 空文字", id: "01ARZ3NDEKTSV4RRFFQ69G5FAV", title: "title", body: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := todo.Reconstruct(tt.id, tt.title, tt.body, false, 1)
			require.ErrorIs(t, err, todo.ErrInvalid)
		})
	}
}

// Reconstruct は業務ルール（todo_title: 3〜10 文字など）を通さない。
// 過去データを今のバリデーションで拒否しないことを保証する。
func TestReconstruct_BypassesBusinessRules(t *testing.T) {
	id := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	body := "body"
	tests := []struct {
		name  string
		title string
	}{
		{name: "1 文字 (NewTodo では 3 文字未満で不可)", title: "a"},
		{name: "20 文字 (NewTodo では 10 文字超で不可)", title: strings.Repeat("a", 20)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := todo.Reconstruct(id, tt.title, body, false, 1)
			require.NoError(t, err)
			assert.Equal(t, tt.title, got.Title())
		})
	}
}

func TestTodo_Complete(t *testing.T) {
	got, err := todo.NewTodo("title", "body")
	require.NoError(t, err)

	err = got.Complete()
	require.NoError(t, err)
	assert.True(t, got.Completed())
}

func TestTodo_Complete_AlreadyCompleted(t *testing.T) {
	got, err := todo.NewTodo("title", "body")
	require.NoError(t, err)

	err = got.Complete()
	require.NoError(t, err)

	err = got.Complete()
	require.ErrorIs(t, err, todo.ErrAlreadyCompleted)
}

func TestTodo_SyncVersion(t *testing.T) {
	got, err := todo.NewTodo("title", "body")
	require.NoError(t, err)
	assert.Equal(t, 0, got.Version())

	got.SyncVersion(5)
	assert.Equal(t, 5, got.Version())
}
