package todorpc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	todov1 "github.com/minty1202/go-ddd-onion-template/gen/todo/v1"
	"github.com/minty1202/go-ddd-onion-template/gen/todo/v1/todov1connect"
	"github.com/minty1202/go-ddd-onion-template/internal/domain/todo"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/define"
	"github.com/minty1202/go-ddd-onion-template/internal/usecase/todo/view"
)

type defineUseCase interface {
	Execute(ctx context.Context, p define.Param) (*define.Result, error)
}

type viewUseCase interface {
	Execute(ctx context.Context, p view.Param) (*view.Result, error)
}

// Handler は Todo サービスの Connect RPC ハンドラ。生成された
// TodoServiceHandler interface を満たし、各 RPC で usecase 層に処理を委譲、
// エラーを Connect 形式に変換して返す。
type Handler struct {
	define       defineUseCase
	view         viewUseCase
	interceptors []connect.Interceptor
}

// New は Repository と Connect Interceptor 群から Handler を組み立てる。
// usecase 層のインスタンスを内部で生成するため、外からは Repository だけ
// 渡せばよい。interceptors は Mount 時に connect.WithInterceptors に展開
// される。
func New(repo todo.Repository, interceptors []connect.Interceptor) *Handler {
	return newHandler(
		define.NewUseCase(repo),
		view.NewUseCase(repo),
		interceptors,
	)
}

var _ todov1connect.TodoServiceHandler = (*Handler)(nil)

func newHandler(d defineUseCase, v viewUseCase, interceptors []connect.Interceptor) *Handler {
	return &Handler{
		define:       d,
		view:         v,
		interceptors: interceptors,
	}
}

// Define は Todo を新規定義する RPC のハンドラ。
func (h *Handler) Define(ctx context.Context, req *todov1.DefineRequest) (*todov1.DefineResponse, error) {
	result, err := h.define.Execute(ctx, toDefineParam(req))
	if err != nil {
		return nil, toConnectError(err)
	}
	return toDefineResponse(result), nil
}

// View は Todo を 1 件取得する RPC のハンドラ。
func (h *Handler) View(ctx context.Context, req *todov1.ViewRequest) (*todov1.ViewResponse, error) {
	result, err := h.view.Execute(ctx, toViewParam(req))
	if err != nil {
		return nil, toConnectError(err)
	}
	return toViewResponse(result), nil
}

// Mount は Handler を http.ServeMux に登録する Mounter インタフェースの
// 実装。Connect の生成コードが返すパス + ハンドラを mux に紐づけ、構築
// 時に渡された interceptors をチェーンとして適用する。
func (h *Handler) Mount(mux *http.ServeMux) {
	path, handler := todov1connect.NewTodoServiceHandler(
		h,
		connect.WithInterceptors(h.interceptors...),
	)
	mux.Handle(path, handler)
}
