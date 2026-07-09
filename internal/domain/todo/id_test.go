package todo_test

import (
	"strings"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewID_NotZero(t *testing.T) {
	id := todo.NewID()
	assert.NotEqual(t, todo.ID{}, id)
}

func TestNewID_Uniqueness(t *testing.T) {
	id1 := todo.NewID()
	id2 := todo.NewID()
	assert.NotEqual(t, id1, id2)
}

func TestNewID_RoundTrip(t *testing.T) {
	id := todo.NewID()
	parsed, err := todo.ParseID(id.String())
	require.NoError(t, err)
	assert.Equal(t, id, parsed)
}

func TestParseID_Valid(t *testing.T) {
	s := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	id, err := todo.ParseID(s)
	require.NoError(t, err)
	assert.Equal(t, s, id.String())
}

func TestParseID_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "空文字", input: ""},
		{name: "短すぎ", input: "abc"},
		{name: "長すぎ", input: strings.Repeat("0", 30)},
		{name: "不正な文字", input: strings.Repeat("!", 26)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := todo.ParseID(tt.input)
			require.ErrorIs(t, err, todo.ErrInvalid)
		})
	}
}
