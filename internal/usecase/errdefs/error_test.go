package errdefs_test

import (
	"errors"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseError_Error_WithWrappedError(t *testing.T) {
	inner := errors.New("db connection failed")
	ucErr := errdefs.NewInternal(inner)

	assert.Equal(t, errdefs.Internal, ucErr.Kind)
	assert.Contains(t, ucErr.Error(), "Internal")
	assert.Contains(t, ucErr.Error(), "db connection failed")
}

func TestUseCaseError_Error_WithoutWrappedError(t *testing.T) {
	ucErr := &errdefs.UseCaseError{Kind: errdefs.InvalidArgument}

	assert.Equal(t, "usecase error: InvalidArgument", ucErr.Error())
}

func TestUseCaseError_Unwrap(t *testing.T) {
	inner := errors.New("original")
	ucErr := errdefs.NewFailedPrecondition(inner)

	assert.Equal(t, inner, errors.Unwrap(ucErr))
	assert.True(t, errors.Is(ucErr, inner))
}

func TestKind_String(t *testing.T) {
	tests := []struct {
		kind errdefs.Kind
		want string
	}{
		{errdefs.InvalidArgument, "InvalidArgument"},
		{errdefs.NotFound, "NotFound"},
		{errdefs.FailedPrecondition, "FailedPrecondition"},
		{errdefs.AlreadyExists, "AlreadyExists"},
		{errdefs.Internal, "Internal"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.kind.String())
		})
	}
}

func TestKind_String_Unknown_Panics(t *testing.T) {
	assert.Panics(t, func() {
		_ = errdefs.Kind(999).String()
	})
}
