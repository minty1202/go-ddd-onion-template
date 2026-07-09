package view

import (
	"context"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
)

type UseCase struct {
	repo todo.Repository
}

type Param struct {
	ID string
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
	id, err := todo.ParseID(p.ID)
	if err != nil {
		return nil, errdefs.NewInvalidArgument(err)
	}

	t, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fromRepoError(err)
	}

	return toResult(t), nil
}
