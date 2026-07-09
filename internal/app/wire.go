package app

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minty1202/connectkit"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/db"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/repository/todorepo"
	"github.com/minty1202/go-ddd-onion-template/internal/presentation/todorpc"
)

// App はアプリ起動に必要な全リソースの保持者。長命リソース (DB pool 等) の
// オーナーシップを持ち、Close で一括解放する。Server は connectkit が提供する
// もので、`a.Server.Run(ctx)` で起動する。
type App struct {
	pool   *pgxpool.Pool
	Server *connectkit.Server
}

// New はアプリの composition root。DB pool の生成、A 層 interceptor の構成、
// 各集約モジュールの組み立て、connectkit Server の構築までを行う。
//
// 失敗時はそれまでに確保したリソースを解放してから error を返す。成功時は
// 呼び出し側が必ず Close を呼ぶこと。
func New(ctx context.Context, dsn, port string, shutdownTimeout time.Duration) (*App, error) {
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		return nil, err
	}

	queries := db.New(pool)

	interceptors, err := connectkit.Default()
	if err != nil {
		pool.Close()
		return nil, err
	}

	deps := connectkit.Dependencies{
		Mounters: []connectkit.Mounter{
			newTodoModule(queries, interceptors),
			connectkit.NewHealthz(),
			connectkit.NewReadyz(pool.Ping),
		},
		Middlewares: []func(http.Handler) http.Handler{
			connectkit.NewH2C(),               // HTTP/2 (h2c) wrapper、最外側
			connectkit.NewSecurityHeaders(),   // HSTS / X-Frame-Options / CSP 等の防御ヘッダ
			connectkit.NewMaxBody(1 << 20),    // 1 MiB を超える body を 413 で弾く (DoS 防御)
			connectkit.NewMetrics(),           // HTTP-level 観測 (otelhttp)
		},
	}

	return &App{
		pool:   pool,
		Server: connectkit.NewServer(deps, port, shutdownTimeout),
	}, nil
}

// Close は App が保持する長命リソースを解放する。現状は DB プールのみだが、
// 将来 OTel SDK Shutdown / 外部 API クライアント Close 等を順序付きで呼ぶ
// 責務を負う。
//
// 二重 Close は安全 (pgxpool.Pool.Close は idempotent)。
func (a *App) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}

// newTodoModule は todo 集約のプレゼンテーション handler を Mounter として
// 構築する。
func newTodoModule(queries *db.Queries, interceptors []connect.Interceptor) connectkit.Mounter {
	return todorpc.New(todorepo.New(queries), interceptors)
}
