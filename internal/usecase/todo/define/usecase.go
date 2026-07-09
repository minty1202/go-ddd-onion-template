package define

import (
	"context"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
)

type UseCase struct {
	repo todo.Repository
}

type Param struct {
	Title string
	Body  string
}

type Result struct {
	ID        string
	Title     string
	Body      string
	Completed bool
	Version   int
}

func toResult(t *todo.Todo) *Result {
	return &Result{
		ID:        t.ID().String(),
		Title:     t.Title(),
		Body:      t.Body(),
		Completed: t.Completed(),
		Version:   t.Version(),
	}
}

func NewUseCase(repo todo.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(ctx context.Context, p Param) (*Result, error) {
	t, err := todo.NewTodo(p.Title, p.Body)
	if err != nil {
		return nil, errdefs.NewInvalidArgument(err)
	}

	err = uc.repo.Add(ctx, t)
	if err != nil {
		return nil, errdefs.NewInternal(err)
	}

	return toResult(t), nil
}
