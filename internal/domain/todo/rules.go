package todo

import (
	"github.com/go-playground/validator/v10"
	"github.com/minty1202/go-ddd-onion-template/internal/domain/validation"
)

const (
	TitleMinLen = 3
	TitleMaxLen = 10
)

func init() {
	if err := validation.Validate.RegisterValidation("todo_title", validateTitle); err != nil {
		panic(err)
	}
}

func validateTitle(fl validator.FieldLevel) bool {
	n := len([]rune(fl.Field().String()))
	return n >= TitleMinLen && n <= TitleMaxLen
}
