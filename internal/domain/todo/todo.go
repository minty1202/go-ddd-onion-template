package todo

import (
	"github.com/minty1202/go-ddd-onion-template/internal/domain/aggregate"
	"github.com/minty1202/go-ddd-onion-template/internal/domain/validation"
)

type Todo struct {
	lock      aggregate.Lock
	id        ID
	title     string `validate:"required,todo_title"`
	body      string `validate:"required"`
	completed bool
}

func (t *Todo) ID() ID          { return t.id }
func (t *Todo) Title() string   { return t.title }
func (t *Todo) Body() string    { return t.body }
func (t *Todo) Completed() bool { return t.completed }

func (t *Todo) Version() int      { return t.lock.Version() }
func (t *Todo) SyncVersion(v int) { t.lock.SyncVersion(v) }

func NewTodo(title, body string) (*Todo, error) {
	newID := NewID()
	t := &Todo{
		lock:      aggregate.NewLock(),
		id:        newID,
		title:     title,
		body:      body,
		completed: false,
	}

	err := validation.Validate.Struct(t)
	if err != nil {
		return nil, ErrInvalid
	}

	return t, nil
}

// Reconstruct は永続化された状態から Todo を組み立てる。Repository 実装専用。
func Reconstruct(id, title, body string, completed bool, version int) (*Todo, error) {
	parsedID, err := ParseID(id)
	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, ErrInvalid
	}

	if body == "" {
		return nil, ErrInvalid
	}

	t := &Todo{
		lock:      aggregate.ReconstructLock(version),
		id:        parsedID,
		title:     title,
		body:      body,
		completed: completed,
	}
	return t, nil
}

func (t *Todo) Complete() error {
	if t.completed {
		return ErrAlreadyCompleted
	}
	t.completed = true

	return nil
}
