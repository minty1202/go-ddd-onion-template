# TODO: 実装の積み残し

このファイルは、ADR で決定済みだが未実装の項目、および今後の作業をリストする。

## プレゼンテーション層（最優先、一気通貫を目指す）

ADR-0005 で構造と思想は確定（Slack API スタイル + Connect RPC + protovalidate）。残るのは実装。

- [ ] `proto/todo/v1/todo.proto` の作成（`TodoService` + 5 RPC + `Todo` message + protovalidate アノテーション）
- [ ] `buf generate` を実行して `gen/todo/v1/` を生成
- [ ] `internal/presentation/todo/handler.go` 実装（Connect の `TodoServiceHandler` interface 実装）
- [ ] `internal/presentation/todo/mapper.go` 実装（proto Request/Response ↔ usecase Param/Result）
- [ ] `internal/presentation/todo/error.go` 実装（usecase errdefs → connect.Error 変換）
- [ ] protovalidate interceptor の組み込み（境界バリデーション）
- [ ] `cmd/server/main.go` の wiring（HTTP server + Connect ハンドラ登録 + DI）

## 既存 usecase の修正（ADR-0005 から派生）

- [ ] `define` / `view` の `Result` に `Version int` フィールド追加（etag マッピングのため、ADR-0005 帰結）
- [ ] 上記に伴う既存テストの更新

## ドメイン層

- [ ] `Todo.Revise(title, body)` メソッド追加（タイトルとボディの改訂）
- [ ] `Todo.Archive()` 相当の業務動詞メソッド検討（現状は物理削除なので `repo.RemoveByID` で完結、ドメインメソッド要否は再検討）
- [ ] `body` のバリデーションルール詳細決定（文字数の最小・最大、内容ルール等）。現状は `validate:"required"` のみ

## ユースケース層

- [ ] `internal/usecase/todo/revise/`（`ReviseTodo`）新設
- [ ] `internal/usecase/todo/complete/`（`CompleteTodo`）新設
- [ ] `internal/usecase/todo/archive/`（`ArchiveTodo`）新設
- [ ] `errdefs/error.go` に `Aborted` Kind と `NewAborted` ヘルパーを追加（`aggregate.ErrConflict` のマッピング先、ADR-0004 で決定済み）
- [ ] 各新規 usecase のテスト追加

## インフラ層

- [ ] (現時点で大きな積み残しなし)

## テスト

- [ ] 新 usecase のテスト（上記 usecase 追加時）
- [ ] presentation 層のハンドラテスト（追加時）
- [ ] 必要に応じて統合テスト（Connect RPC エンドポイントから DB までの一気通貫）
- [ ] **drift 検知 CI テスト**: `proto/todo/v1/todo.proto` の protovalidate 数値（`min_len` 等）と domain 定数（`TitleMinLen` 等）の一致を検証する Go テスト（ADR-0003 / ADR-0005）

## ADR / ルールの保守

- [ ] 楽観ロックの自動化に踏み込む場合（`GenericRepository[T]` 路線）は別 ADR で再検討（現状は `SaveWithLock` 規約頼り）

## 参考

- `docs/adr/0001-migration-tool.md`（goose 採用）
- `docs/adr/0002-task-based-usecase-and-repository.md`（業務動詞命名、ADR-0004 で部分 supersede）
- `docs/adr/0003-validation-layering-and-rule-export.md`（validation の責務分担）
- `docs/adr/0004-collection-oriented-repository-and-optimistic-lock.md`（Repository + 楽観ロック）
- `docs/adr/0005-presentation-design.md`（presentation 層: Slack API スタイル + Connect RPC + protovalidate + etag）
- `.claude/rules/go/architecture.md`（実用指針）
