# Connect RPC ブートストラップ進行プラン

Connect RPC 標準スタックを実装するための一時プラン。
Phase 5 完了後にこのファイルは削除する（ルール化が必要な部分は別途 `.claude/rules/` に切り出す）。

## 全体像

```
Phase 1: Connect 最小経路        ← 中  ✅ 完了（2026-05-01）
Phase 2: A 層（Interceptor 群）        ✅ 完了（2026-05-03）
Phase 3: B 層（HTTP Middleware）
Phase 4: C 層（運用エンドポイント）
Phase 5: 業務要件次第の拡張        ← 外
```

各 Phase で handler 本体は変更しない（decorator として外側を足すだけ）。

---

## Phase 1: Connect 最小経路 ✅ 完了

**目的**: proto から handler まで 1 本通して `buf curl` で疎通させる。

### 実装内容

- 1.1 ✅ `proto/todo/v1/todo.proto`（Define / View RPC、Todo リソース型、protovalidate annotations）
- 1.2 ✅ `go tool buf generate` で `gen/todo/v1/` を生成
- 1.3 ✅ `internal/presentation/todorpc/{handler,mapper,error}.go` 実装 + テスト
- 1.4 ✅ `cmd/server/main.go` で Connect サーバを起動、handler 登録（`internal/server/` の Wire / Module / Router / Server を経由）
- 1.5 ✅ `just curl-define` で疎通確認

### Phase 1 で確定した追加事項

- **infrastructure → infra リネーム**（ADR-0007）
- **repository todo → todorepo リネーム + ファイル名変更**（domain 名予約ルール適用）
- **集約ごとの module pattern**（`server/modules.go` に `newXxxModule` を集約）
- **Mounter interface パターン**（presentation 側に `Mount(mux)` メソッドを生やして自己登録、server 側で interface 定義）
- **composition root を main.go に**（`Wire → New(deps, port) → Run`）
- **errgroup ベースの graceful shutdown**
- **`net.JoinHostPort("", port)` で IPv6 対応**
- **`forbidigo` linter で `SyncVersion` / `Reconstruct` の呼び出し元を構造的に制限**
- **air + direnv で開発環境セットアップ**（`just dev` + `.envrc.example`）

---

## Phase 2: A 層（Interceptor 群）✅ 完了（2026-05-03）

**目的**: A 層を全部入れる。**注**: `error_tracking interceptor` は別途廃止され、error 通知は `otel` に統合。A 層は **6 個構成**（旧 7 個）。

並び順（外 → 内）:

```
recovery → request_id → logging → otel → normalize → validate → handler
```

### 実装ステップ（達成済）

- **2.1 recovery** ✅ — panic を error に変換、`http.ErrAbortHandler` 再 panic、unary + streaming 両対応
- **2.2 request_id** ✅ — UUIDv7 発行 / 伝播、クライアント送信値の検証付き（最大 128 バイト + ASCII printable）
- **2.3 logging** ✅ — slog 構造化アクセスログ。本体は出さない方針（機密漏洩 + サイズ抑制）。logger は `log/slog` 確定
- **2.4 otel** ✅ — `connectrpc.com/otelconnect`。SDK 未設定で no-op だが wire 済み。error 通知も担う
- **2.5 normalize** ✅ — protoreflect で全 string を NFC 化、再帰走査
- **2.6 validate** ✅ — `connectrpc.com/validate`、proto の制約を自動実行

### 配置

- `internal/server/interceptor/`
- 各 interceptor は **1 ファイル 1 責務**（`recovery.go` / `request_id.go` / `logging.go` / `normalize.go`、各 `_test.go` 付き）

### Phase 2 で確定した追加事項

- **`Wire` → `NewDependencies` リネーム + ファイル名 `dependencies.go`**: google/wire との混同回避
- **`Mounter` interface を `mount.go` に独立配置**
- **`Dependencies.Close()` メソッド導入**: pool / OTel SDK 等の長命リソース解放を一元化
- **`Config.ShutdownTimeout`**: 環境変数 `SHUTDOWN_TIMEOUT_SECONDS`（既定 30 秒）で制御可能化
- **`errdefs.Aborted` Kind + `NewAborted` 追加**: 楽観ロック競合 → `connect.CodeAborted` のマッピング完成
- **godoc 充足**: 公開識別子全般に日本語 godoc を追加
- **`errors.AsType` モダナイズ**: Go 1.26 の generic API を全面採用
- **テスト方針**: 各 interceptor のユニットテスト + クロス（`request_id × logging`）の 1 ケースのみ。chain test（リクエストテスト相当）は不採用

### 完了条件（達成済）

- ✅ 全 interceptor が起動時に組み立てられ、build / test / lint が clean
- ✅ Phase 1 の RPC が引き続き通る
- ✅ 動作確認: `just curl-define` でレスポンスヘッダに `X-Request-Id` (UUIDv7) が載る

---

## Phase 3: B 層（HTTP Middleware）

**目的**: B 層を入れる。

- 3.1 CORS（`github.com/rs/cors`、既に go.mod に入っている）
- 3.2 security headers
- 3.3 max body size
- 3.4 h2c handler（HTTP/2 over cleartext、gRPC クライアント対応）
- 3.5 graceful shutdown 強化（既に Phase 1 で基本対応済み、必要に応じて拡張）

### 配置

- `internal/server/middleware/`

### 完了条件

- ブラウザクライアントから RPC を叩いて通る
- gRPC クライアント（grpcurl 等）からも叩ける（h2c 経由）

---

## Phase 4: C 層（運用エンドポイント）

**目的**: C 層を入れる。

- 4.1 `/healthz`（liveness）
- 4.2 `/readyz`（DB 接続を含む readiness）
- 4.3 grpchealth（`connectrpc.com/grpchealth`）
- 4.4 grpcreflect（`connectrpc.com/grpcreflect`、`buf curl` でデバッグ可能になる）
- 4.5 `/version`（ビルド情報、`-ldflags` で埋め込み）
- 4.6 `/debug/pprof/*`
  - **着手前判断**: 公開範囲制御の実装方式 → 別 ADR

### 配置

- `internal/server/ops/`

### 完了条件

- 各 endpoint に curl で疎通する
- DB 停止時に `/readyz` が 503 を返す

---

## Phase 5: 業務要件次第の拡張

業務要件が確定してから着手:

- auth / authz（認証 / 認可）
- rate limit
- idempotency
- audit log
- usecase 拡張（revise / complete / archive）

それぞれ別 ADR で個別に判断する。

---

## 完了後の片付け

- このファイル（`docs/connect-rpc-bootstrap-plan.md`）を削除
- 実装中に確定したルール（並び順、ディレクトリ構造、運用上の取り決め）を `.claude/rules/go/` に反映

---

## 動作確認のクイックリファレンス

開発環境のセットアップ:

```sh
cp .envrc.example .envrc
direnv allow .
just db-up
just migrate up
just dev          # air + hot reload
```

別ターミナルで:

```sh
just curl-define                                    # Define RPC
just curl-view id=<返ってきた id>                    # View RPC
```

検証:

```sh
just check        # lint + sqlc compile + buf lint
just test         # 全テスト
```
