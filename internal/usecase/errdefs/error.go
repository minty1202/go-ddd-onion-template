package errdefs

import "fmt"

type Kind int

const (
	InvalidArgument    Kind = iota // 入力不正
	NotFound                       // リソースが見つからない
	FailedPrecondition             // 状態衝突（すでに完了済み等）
	AlreadyExists                  // 重複
	Aborted                        // 並行性のコンフリクト（楽観ロック競合等）
	Internal                       // 内部エラー
)

func (k Kind) String() string {
	switch k {
	case InvalidArgument:
		return "InvalidArgument"
	case NotFound:
		return "NotFound"
	case FailedPrecondition:
		return "FailedPrecondition"
	case AlreadyExists:
		return "AlreadyExists"
	case Aborted:
		return "Aborted"
	case Internal:
		return "Internal"
	default:
		panic("unreachable")
	}
}

type UseCaseError struct {
	Kind Kind
	Err  error
}

func (e *UseCaseError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("usecase error: %s", e.Kind)
	}
	return fmt.Sprintf("usecase error: %s: %v", e.Kind, e.Err)
}

func (e *UseCaseError) Unwrap() error {
	return e.Err
}

func NewInvalidArgument(err error) *UseCaseError {
	return &UseCaseError{Kind: InvalidArgument, Err: err}
}

func NewNotFound(err error) *UseCaseError {
	return &UseCaseError{Kind: NotFound, Err: err}
}

func NewFailedPrecondition(err error) *UseCaseError {
	return &UseCaseError{Kind: FailedPrecondition, Err: err}
}

func NewAlreadyExists(err error) *UseCaseError {
	return &UseCaseError{Kind: AlreadyExists, Err: err}
}

func NewAborted(err error) *UseCaseError {
	return &UseCaseError{Kind: Aborted, Err: err}
}

func NewInternal(err error) *UseCaseError {
	return &UseCaseError{Kind: Internal, Err: err}
}
