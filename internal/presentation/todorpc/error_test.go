package todorpc

import (
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToConnectError_Mapping(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		wantCode connect.Code
	}{
		{name: "InvalidArgument", input: errdefs.NewInvalidArgument(errors.New("invalid")), wantCode: connect.CodeInvalidArgument},
		{name: "NotFound", input: errdefs.NewNotFound(errors.New("not found")), wantCode: connect.CodeNotFound},
		{name: "FailedPrecondition", input: errdefs.NewFailedPrecondition(errors.New("conflict")), wantCode: connect.CodeFailedPrecondition},
		{name: "AlreadyExists", input: errdefs.NewAlreadyExists(errors.New("exists")), wantCode: connect.CodeAlreadyExists},
		{name: "Internal", input: errdefs.NewInternal(errors.New("internal")), wantCode: connect.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := toConnectError(tt.input)
			require.Error(t, err)
			var connectErr *connect.Error
			require.ErrorAs(t, err, &connectErr)
			assert.Equal(t, tt.wantCode, connectErr.Code())
		})
	}
}

func TestToConnectError_Nil(t *testing.T) {
	assert.NoError(t, toConnectError(nil))
}

func TestToConnectError_NonUseCaseError(t *testing.T) {
	err := toConnectError(errors.New("unknown"))
	require.Error(t, err)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInternal, connectErr.Code())
}

func TestToConnectError_InternalDoesNotLeakOriginalMessage(t *testing.T) {
	original := errors.New("sensitive db error: SELECT * FROM users WHERE password='secret'")
	err := toConnectError(errdefs.NewInternal(original))
	require.Error(t, err)
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.NotContains(t, connectErr.Message(), "sensitive")
	assert.NotContains(t, connectErr.Message(), "password")
	assert.Equal(t, "internal server error", connectErr.Message())
}
