---
paths:
  - "**/*.go"
---

# エラー設計ルール

## ドメイン層

- エラーは各集約パッケージ内の `errors.go` に固定エラー値（英語では sentinel error）として定義する
  - 例: `internal/domain/todo/errors.go` に `var ErrAlreadyCompleted = errors.New("todo: already completed")`
- 命名: `Err<具体違反名>`（例: `ErrInvalid`, `ErrAlreadyCompleted`）
- エラーメッセージの prefix にパッケージ名を付ける（例: `"todo: already completed"`）
- ドメイン層に共通のエラー型（`DomainError` 構造体、`ErrorKind` enum）を置かない
- エラーに詳細情報（違反フィールド、違反値など）を持たせない
- validator などが返す詳細エラーも `%w` でラップせず固定エラー値に畳む。**詳細は presentation 層の境界バリデーション（protovalidate 等）が扱う責務であり、ドメインは「違反した事実」のみを保持する**。ドメインの validator は最後の砦として動作し、通常経路では presentation 層で先に弾かれる前提。「何に違反したか」は固定エラー値の種別（`ErrInvalid` / `ErrAlreadyCompleted` 等）で表現する
- **fail-fast**: 適切に設計されたドメインでは 1 操作 = 1 結果になるため、自然と fail-fast になる
- 複数違反を返したくなった場合は、まず **ドメイン設計の見直し** を推奨する。それでも必要なら例外的に許可
- 詳細な違反情報や複数違反のまとめ返却は presentation 層 / ユースケース層の責務（ドメインの責務ではない）
- ドメインの固定エラー値は 2 種類に分けて運用する:
  - **バリデーション違反**: 集約ごとに汎用の固定エラー値（例: `ErrInvalid`）**1 つに畳む**。詳細フィールド情報は持たせない
  - **業務不変条件違反**: `ErrAlreadyCompleted`、`ErrAlreadyShipped` のように **具体名で個別の固定エラー値** を定義する。usecase 層で個別マッピング（`InvalidArgument` / `FailedPrecondition` 等）に振り分けるため

## ユースケース層

- 共通の `UseCaseError` 型と `Kind` enum を `internal/usecase/errdefs/` に定義する
- `Kind` はプロトコル非依存の抽象分類を使う（HTTP ステータスや特定プロトコル固有の命名を避ける）
- `Kind` の具体値は Google API Design Guide（aip.dev）/ gRPC Code に合わせる:
  - `InvalidArgument` — 入力バリデーション違反
  - `NotFound` — 集約が見つからない
  - `AlreadyExists` — 既に存在する集約を新規追加しようとした
  - `FailedPrecondition` — 業務不変条件違反（例: 既に完了済みの集約を再度完了しようとした）
  - `Aborted` — 並行性のコンフリクト（楽観ロック競合、`ErrConflict`）
  - `Internal` — 未知のインフラエラー、想定外の状態
  - 必要に応じて追加（`Unauthenticated`, `PermissionDenied`, `ResourceExhausted` 等）
- HTTP ステータスなどプロトコル固有の関心はユースケース層に持ち込まない（プレゼン層の責務）

### プロトコル別マッピング（presentation 層の責務）

`Kind` から各プロトコルへのマッピングは presentation 層が担当する。本テンプレートが想定する 2 つのプロトコル（Connect RPC / HTTP REST）の標準的なマッピングは以下:

| Kind | `connect.Code` | HTTP status | body code（AIP-193 流） |
|------|----------------|-------------|----------------------|
| `InvalidArgument` | `CodeInvalidArgument` | `400 Bad Request` | `INVALID_ARGUMENT` |
| `NotFound` | `CodeNotFound` | `404 Not Found` | `NOT_FOUND` |
| `FailedPrecondition` | `CodeFailedPrecondition` | `400 Bad Request` | `FAILED_PRECONDITION` |
| `AlreadyExists` | `CodeAlreadyExists` | `409 Conflict` | `ALREADY_EXISTS` |
| `Aborted` | `CodeAborted` | `409 Conflict` | `ABORTED` |
| `Internal` | `CodeInternal` | `500 Internal Server Error` | `INTERNAL` |

出典: [google.rpc.Code](https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto) と [Google AIP-193](https://google.aip.dev/193) の標準マッピング。

**HTTP の場合、複数 Kind が同じ status になる**（`AlreadyExists` と `Aborted` がどちらも 409）。そのため presentation 層は **response body に `code` フィールド**を埋めて区別する。AIP-193 流のレスポンス形式:

```json
{
  "error": {
    "code": "ABORTED",
    "message": "Concurrent update detected, please retry"
  }
}
```

**Connect の場合**は `connect.Code` が wire 上に乗るため、HTTP status が 409 で潰れていてもクライアントは code フィールドで区別できる。手動の HTTP status マッピングは不要（`connect.NewError(connect.CodeAborted, err)` で完結）。
- ドメインエラーを `UseCaseError` に変換するヘルパーは各ユースケースパッケージ内に置く
  - 例: `internal/usecase/todo/define/errors.go`
- ただし、分岐が 1 ケースで helper の中身が実質 1 行になる場合はインライン化してよい（YAGNI）
- 各ユースケースは自分が扱うドメインエラーだけを知る（共通の変換関数は置かない）
- `errors.Is` でドメインエラーを個別に `UseCaseError` にマッピングし、未知のエラー（リポジトリ実装から素通しで来たインフラエラー）は `errdefs.NewInternal(err)` にまとめる

## リポジトリ実装層（infra）

- **業務上意味のあるインフラエラーは、ドメイン層の固定エラー値に変換する**
  - 例: `pgx.ErrNoRows` → `todo.ErrNotFound`（リポジトリ実装で `errors.Is` で判定して変換）
- それ以外のインフラエラー（接続エラー、SQL 構文エラー、外部キー違反等）は **素通し** する。usecase 側で `Internal` にまとめられる
- 「業務上意味があるかどうか」の判断基準: **usecase 側で個別に識別したいエラーかどうか**
- ドメインの固定エラー値（`todo.ErrNotFound` 等）はドメイン層に置く（インフラ層（リポジトリ実装含む）には置かない。usecase が import すると依存方向違反になるため）
