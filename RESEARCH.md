# Connect RPC + Go エコシステム調査

## コアツールチェーン

- **buf CLI** — protoc の代替。lint / breaking change 検出 / コード生成を一括管理
- **protoc-gen-go** — .proto から Go のメッセージ型を生成
- **protoc-gen-connect-go** — .proto から Connect 用の handler / client スタブを生成
- **connect-go** — Connect RPC のコアランタイム。gRPC / gRPC-Web / Connect プロトコルをサポート

## 公式サテライトライブラリ

- **connectrpc.com/validate** — protovalidate ベースのバリデーション interceptor
- **connectrpc.com/authn** — 認証ミドルウェア
- **connectrpc.com/cors** — CORS 設定ヘルパー
- **connectrpc.com/otelconnect** — OpenTelemetry 連携

## DB ライブラリ（PostgreSQL）

**sqlc + pgx が最も相性が良い。**

- **pgx** — Go の PostgreSQL ドライバのデファクト。lib/pq はメンテナンスモード
- **sqlc** — SQL を書くと型安全な Go コードを生成。pgx/v5 をサポート
- buf でコード生成、sqlc でコード生成、という思想が一貫する

その他: GORM、ent、sqlx

## マイグレーション

- **golang-migrate** — 最もポピュラー。シンプルな up/down SQL ファイル方式
- **goose** — 軽量。Go コードでのマイグレーションも書ける
- **Atlas** — 宣言的スキーマ管理。差分を自動計算

## Interceptor パターン

- 認証: authn-go
- バリデーション: validate-go
- ロギング: slog 等
- トレース: otelconnect
- 実行順序はリクエスト時 first-to-last、レスポンス時 last-to-first
