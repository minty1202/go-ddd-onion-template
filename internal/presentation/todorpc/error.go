package todorpc

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/errdefs"
)

func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	ucErr, ok := errors.AsType[*errdefs.UseCaseError](err)
	if !ok {
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	}

	switch ucErr.Kind {
	case errdefs.InvalidArgument:
		return connect.NewError(connect.CodeInvalidArgument, ucErr.Err)
	case errdefs.NotFound:
		return connect.NewError(connect.CodeNotFound, ucErr.Err)
	case errdefs.FailedPrecondition:
		return connect.NewError(connect.CodeFailedPrecondition, ucErr.Err)
	case errdefs.AlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, ucErr.Err)
	case errdefs.Aborted:
		return connect.NewError(connect.CodeAborted, ucErr.Err)
	case errdefs.Internal:
		return connect.NewError(connect.CodeInternal, errors.New("internal server error"))
	default:
		panic("unreachable")
	}
}
