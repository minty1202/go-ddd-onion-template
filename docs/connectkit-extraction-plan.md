# connectkit 切り出し進行プラン

ADR-0008 で確定した「Connect RPC ベースの個人共通基盤を別 private リポジトリ `connectkit` に切り出す」方針の実行プラン。一時ドキュメントで、切り出し完了後に削除する。

## 全体像

```
Phase A: connectkit リポジトリ作成 + 基本構造
Phase B: 既存コードの move
Phase C: Todo プロジェクトを connectkit 利用形に refactor
Phase D: Phase 3 / 4 を connectkit 内で実装
Phase E: 新サービス開始 (本リポジトリ外)
```

各 Phase で既存テストが落ちないことを完了条件とする。

## 進捗 (2026-05-04)

- Phase A: 完了
- Phase B: 完了 (errdefs / codemap は todo_app に残す方針に変更、副次として連結を **単一 root パッケージ** にも flat 化)
- Phase C: skip (`connectkit.NewServer` の引数渡しで完結、connectkit に config を持たせない)
- Phase D: 完了 (flat 構造により middleware / ops は connectkit root に直接配置)
- Phase E: 未着手

各 Phase の本文ステップは当時の計画スナップショットで、実際の path は connectkit の flat 構造（`connectkit/X.go`）に合わせて読み替える。

---

## Phase A: connectkit リポジトリ作成

### 目的

新リポジトリの骨格を用意し、`go.work` で Todo プロジェクトと横断開発できる状態にする。

### 実装ステップ

- A.1 GitHub に private repo `github.com/minty1202/connectkit` を作成（.gitignore: Go 用、LICENSE は当面なし、README は最小限）
- A.2 ローカルに clone、`go mod init github.com/minty1202/connectkit`
- A.3 `GOPRIVATE=github.com/minty1202/*` を `~/.zshenv` 等に追加（既に設定済みなら不要）
- A.4 ディレクトリスケルトン作成:
  ```
  connectkit/
  ├── go.mod
  ├── README.md
  ├── server/
  ├── interceptor/
  └── config/
  ```
- A.5 Todo プロジェクトのリポジトリ親ディレクトリに `go.work` を置く（todo_app + connectkit を `use` で含める）
- A.6 connectkit を最初の commit + push

### 完了条件

- `go.work` 経由で connectkit のパッケージを Todo プロジェクトから import 可能（空 package でも import できる状態）

---

## Phase B: 既存コードの move

### 目的

Todo プロジェクトの `internal/server/` / `internal/server/interceptor/` を connectkit に移動。Todo プロジェクト側は connectkit を `require` する形に置き換え（`internal/usecase/errdefs/` と `internal/presentation/todorpc/error.go` の Kind → connect.Code マッピングは todo_app に残す）。

### 実装ステップ

- B.1 **interceptor 群の move**
  - `internal/server/interceptor/recovery.go` → `connectkit/interceptor/recovery.go`
  - `internal/server/interceptor/request_id.go` → `connectkit/interceptor/request_id.go`
  - `internal/server/interceptor/logging.go` → `connectkit/interceptor/logging.go`
  - `internal/server/interceptor/normalize.go` → `connectkit/interceptor/normalize.go`
  - 各 `_test.go` も move
  - `package interceptor` のままで OK
- B.2 **server lifecycle の move**
  - `internal/server/server.go` → `connectkit/server/server.go`
  - `internal/server/mount.go` → `connectkit/server/mount.go`
  - `internal/server/router.go` → `connectkit/server/router.go`
  - `dependencies.go` の generic 化（後述 B.3）
- B.3 **`Dependencies` の generic 化**
  - 現状の `NewDependencies(ctx, dsn)` は `db.NewPool` を呼び、`newTodoModule` を呼んでいる → Todo 固有
  - connectkit 側の `Dependencies` は **`Mounters []Mounter` と `Interceptors []connect.Interceptor` を外部から受け取る generic 構造**にする
  - Todo プロジェクト側で具体的な Mounter / Interceptor を組み立てて渡す
  - 案:
    ```go
    // connectkit/server/dependencies.go
    type Dependencies struct {
        Mounters     []Mounter
        Interceptors []connect.Interceptor
    }
    ```
    - DB pool やアプリ固有リソースの管理は Todo プロジェクト側に移す
- B.4 **import 更新**
  - Todo 側で:
    - `github.com/minty1202/todo_app/internal/server` → `github.com/minty1202/connectkit/server`
    - `github.com/minty1202/todo_app/internal/server/interceptor` → `github.com/minty1202/connectkit/interceptor`
  - `goimports` で一括整理
- B.5 **Todo 側の `internal/server/dependencies.go` を refactor**
  - connectkit の `Dependencies` を組む helper を Todo 側に置く（例: `internal/app/wire.go`）
  - DB pool 生成 + Todo 用 Mounter / Interceptor 構築 + connectkit の `Dependencies` への詰め込み
  - `cmd/server/main.go` の呼び出し先を更新
- B.6 **Todo 側の不要コード削除**
  - `internal/server/dependencies.go` / `internal/server/modules.go` を削除 or `internal/app/` に整理移動
  - `internal/server/` ディレクトリは空になるので削除

### 完了条件

- Todo プロジェクトで `go build ./...` 通過
- `go test ./...` 全通過
- `golangci-lint run ./...` 0 issues
- `just curl-define` で動作確認（X-Request-ID がレスポンスに乗る）

---

## Phase C: BaseConfig の整理

### 目的

`Config` の universal 部分を connectkit に移動、Todo 側はアプリ固有部分だけ残す。

### 実装ステップ

- C.1 connectkit に `config/config.go` を作成、以下を持つ `BaseConfig` を定義:
  - `AppEnv string`
  - `DatabaseURL string`
  - `Port string`
  - `ShutdownTimeout time.Duration`
  - 上記を env から読み込む `LoadBase() (*BaseConfig, error)` 関数
- C.2 Todo 側の `internal/config/config.go` を refactor:
  ```go
  type Config struct {
      connectkit.BaseConfig
      // 将来アプリ固有の env キーをここに追加
  }
  func Load() (*Config, error) {
      base, err := connectkit.LoadBase()
      if err != nil { return nil, err }
      return &Config{BaseConfig: *base}, nil
  }
  ```
- C.3 `cmd/server/main.go` の Config 利用箇所を確認（embed 経由でアクセスできるはず）

### 完了条件

- 既存と同じ env 読み取り挙動
- `go test ./...` 全通過

---

## Phase D: Phase 3 / 4 を connectkit 内で実装

### 目的

B 層 HTTP middleware と C 層 運用 endpoint を connectkit 内で実装。

### 実装ステップ

#### D.1 B 層 middleware

connectkit の root に直接 `func(http.Handler) http.Handler` を返す関数として配置:

- `cors.go` — `github.com/rs/cors` をラップ (`NewCORS`)
- `security_headers.go` — HSTS / X-Content-Type-Options / X-Frame-Options / Referrer-Policy (`NewSecurityHeaders`)
- `max_body.go` — `http.MaxBytesHandler` ラップ (`NewMaxBody`)
- `h2c.go` — `golang.org/x/net/http2/h2c` ラップ (`NewH2C`)

`Dependencies.Middlewares []func(http.Handler) http.Handler` を追加し、`NewServer` 内で外 → 内順で chain を組む（slice 先頭が最外）。chain helper は不要（標準 `func` 合成で十分）。

#### D.2 C 層 ops endpoints

connectkit の root に `Mounter` を返す関数として配置:

- `health.go` — `/healthz` (`NewHealthz`) と `/readyz` (`NewReadyz`、`ReadinessChecker` 関数を可変長で受け取る)
- `version.go` — `/version` (`NewVersion`、`VersionInfo` を JSON で返す)
- `grpchealth.go` — `connectrpc.com/grpchealth` ラップ (`NewGRPCHealth`)
- `grpcreflect.go` — `connectrpc.com/grpcreflect` ラップ (`NewGRPCReflect`)
- `pprof.go` — `net/http/pprof`、access control は呼び出し側に委ねる (`NewPprof`)

#### D.3 Todo プロジェクトでの利用

`internal/app/wire.go` で connectkit の middleware / ops を組み込む:

```go
deps := &connectkit.Dependencies{
    Mounters: []connectkit.Mounter{
        newTodoModule(queries, interceptors),
        connectkit.NewHealthz(),
        connectkit.NewReadyz(pool.Ping),
        // 必要に応じて NewVersion / NewGRPCHealth / NewGRPCReflect / NewPprof を追加
    },
    Middlewares: []func(http.Handler) http.Handler{
        connectkit.NewH2C(),
        // 必要に応じて NewSecurityHeaders / NewCORS / NewMaxBody を追加
    },
}
```

todo_app は最低限 `NewH2C` middleware と `NewHealthz` / `NewReadyz` を有効化済み。

### 完了条件

- `just curl-define` 通過
- `curl http://localhost:8080/healthz` が 200 返す
- `curl http://localhost:8080/readyz` が DB 接続可なら 200、不可なら 503
- `buf curl --reflect ...` で reflection 経由の呼び出しが通る

---

## Phase E: 新サービス開始

### 目的

connectkit を実際に再利用するシナリオを動かす。

### 実装ステップ

- E.1 新サービスのリポジトリ作成
- E.2 `go.work` に新サービスも追加
- E.3 `go get github.com/minty1202/connectkit@latest`
- E.4 新サービスの `cmd/server/main.go` で connectkit を利用、Mounter として新サービスのドメインを wire
- E.5 新サービスのドメイン実装を進める（DDD + オニオン構造に従う）

### 完了条件

- 新サービスが `go run ./cmd/server` で起動
- recovery / request_id / logging / normalize / validate / otel が全て動く
- 必要に応じて connectkit 側で発見した不足を修正、Todo 側にも反映

---

## 完了後の片付け

- 本ファイル（`docs/connectkit-extraction-plan.md`）を削除
- `docs/connect-rpc-bootstrap-plan.md` を更新（Phase 3 / 4 が connectkit 側で完了済みの旨を記録、または同様に削除）
- `.claude/rules/go/` のうち、connectkit に移したコードに関するルール（命名 / アーキテクチャ）を refactor
- `project_stack.md` メモリの更新（ディレクトリ構造 + 切り出し済みである旨）

---

## リスクと対応

### リスク

1. **Todo 側で動いていた挙動が connectkit move で壊れる**
   - 対応: 各 Phase 完了時に `go build` / `go test` / `golangci-lint` / 手動 curl 確認の 4 点セット
2. **`Dependencies` の generic 化で API 設計を間違える**
   - 対応: 新サービスを実際に立ち上げて連携するまで最終形として確定させない。Phase E 終了まで API は流動的に保つ
3. **`go.work` のローカル設定が壊れる**
   - 対応: `GOWORK=off` で各モジュール単独で build できることを CI 等で担保

### スコープ外

- connectkit の OSS 化
- multi-module 化
- Gin / Echo 等の他フレームワーク対応
- pgxpool / sqlc の wrapping
