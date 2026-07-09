# go-ddd-onion-template

DDD + オニオンアーキテクチャによる Go アプリケーションの実践的な構成を、Todo アプリを題材に研究・構築しているリポジトリです。今後の Go 新規開発の土台（テンプレート）にすることを目的にしており、設計判断はすべて [ADR](docs/adr/) として記録しています。

> **Note**
> 本リポジトリは個人の学習・研究プロジェクトであり、外部利用向けのメンテナンスは行いません。
> また、private モジュール `github.com/minty1202/connectkit`（サーバ基盤の薄い convention layer、[ADR-0008](docs/adr/0008-extract-server-foundation.md) で切り出し）に依存しているため、**アクセス権のない環境では `go build` できません**。コードとドキュメントを読む用途を想定しています（[ADR-0009](docs/adr/0009-todo-app-publication-policy.md)）。

## 技術スタック

| 領域 | 採用ツール / ライブラリ |
|---|---|
| 言語 | Go |
| RPC | [Connect RPC](https://connectrpc.com/)（buf / protoc-gen-connect-go / protovalidate） |
| DB | PostgreSQL 17（pgx） |
| SQL | [sqlc](https://sqlc.dev/)（SQL から型安全な Go コードを生成） |
| マイグレーション | [goose](https://github.com/pressly/goose)（embed.FS 同梱、[ADR-0001](docs/adr/0001-migration-tool.md)） |
| テスト | testing + testify / mockery / [testcontainers-go](https://golang.testcontainers.org/)（実 PostgreSQL での統合テスト） |
| 静的解析 | golangci-lint v2 / buf lint / sqlc compile |
| タスクランナー | [just](https://github.com/casey/just) |

## アーキテクチャ

オニオンアーキテクチャの 4 層 + composition root で構成しています。依存の向きは常に内側（domain）に向かいます。

```
cmd/server/        エントリポイント
internal/
  app/             composition root（DI とサーバの組み立て）
  config/          環境変数の読み込み
  domain/          エンティティ・値オブジェクト・リポジトリ interface・検証ルール
  usecase/         ユースケース（タスクベースに分割、ADR-0002）
  infra/           domain が定義した interface の実装（リポジトリ・sqlc 生成コード）
  presentation/    外部からの入口（Connect RPC handler・mapper・エラー変換）
proto/             proto 定義（buf 管理）
sql/               goose マイグレーションと sqlc クエリ
docs/adr/          Architecture Decision Records
docs/history/      superseded になった過去の検討メモ
```

### 主な設計判断

ユースケースは CRUD 単位ではなく「ユーザーがやりたい操作」の単位でパッケージを分割しています（[ADR-0002](docs/adr/0002-task-based-usecase-and-repository.md)）。リポジトリはコレクション指向で設計し、集約に持たせた version を使って `SaveWithLock` が楽観ロックの競合を検出します。競合時の挙動は testcontainers で実際の PostgreSQL を立てる統合テストまで書いて検証しています（[ADR-0004](docs/adr/0004-collection-oriented-repository-and-optimistic-lock.md)）。

バリデーションは proto スキーマ（protovalidate）と domain 層の検証ルールで責務を分離しています（[ADR-0003](docs/adr/0003-validation-layering-and-rule-export.md)）。usecase 層のエラーは Kind 付きの `errdefs` に集約し、presentation 層で `connect.Code` へ網羅的に変換することで、エラー変換の知識が 1 箇所に閉じるようにしています（[ADR-0005](docs/adr/0005-presentation-design.md)）。

こうした設計上の制約は口約束にせず、lint で機械的に守らせています。golangci-lint の `exhaustive` が switch の網羅を強制し、`forbidigo` が「リポジトリ実装以外から `Reconstruct` / `SyncVersion` を呼べない」ことを保証します（[.golangci.yml](.golangci.yml)）。

## 実装ステータス

| RPC | 内容 | 状態 |
|---|---|---|
| `Define` | Todo の新規作成 | 実装済み |
| `View` | Todo の 1 件取得 | 実装済み |
| `Revise` / `Complete` / `Archive` | 更新 / 完了 / 削除 | 未実装（domain・repository は一部実装済み） |

残タスクは [docs/TODO.md](docs/TODO.md) で管理しています。

## 開発方法

前提: Go / Docker / just / direnv（推奨）がインストール済みで、`minty1202/connectkit` へのアクセス権があること。本リポジトリは兄弟ディレクトリの connectkit と合わせて `go.work` で解決する前提です。

```console
# 環境変数の用意
$ cp .env.example .env
$ cp .envrc.example .envrc && direnv allow .

# セットアップ
$ just setup

# DB 起動とマイグレーション
$ just db-up
$ just migrate up

# 開発サーバ起動（air による hot reload）
$ just dev
```

疎通確認は `buf curl` を使ったレシピで行えます。

```console
$ just curl-define                 # Todo を作成
$ just curl-view <作成された ULID>  # Todo を取得
```

その他の主なレシピ:

| コマンド | 内容 |
|---|---|
| `just check` | golangci-lint + sqlc compile + buf lint |
| `just fmt` | goimports + gofumpt + buf format |
| `just test` | テスト実行（統合テストは Docker が必要） |
| `just gen-proto` / `just gen-sqlc` / `just gen-mock` | 各種コード生成 |
| `just migrate <up\|down\|status\|create>` | マイグレーション操作 |

## ドキュメント

- [docs/adr/](docs/adr/) — 設計判断の記録（マイグレーションツール選定、ユースケース分割、楽観ロック、レイヤ命名など全 9 本）
- [docs/TODO.md](docs/TODO.md) — 残タスク
- [TODO_APP.md](TODO_APP.md) — 当初の目標とスコープ
- [RESEARCH.md](RESEARCH.md) — 研究テーマのメモ
- [docs/history/](docs/history/) — superseded になった過去のアーキテクチャ検討メモ
