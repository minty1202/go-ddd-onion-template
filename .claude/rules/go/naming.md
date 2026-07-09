---
paths:
  - "**/*.go"
---
# 命名ルール

## パッケージ名

- `common`, `util`, `helper`, `shared` のような汎用的な名前は避ける（Effective Go, Go Code Review Comments）
- パッケージの責務がわかる名前にする
- 共通化するものは `common` にまとめず、責務ごとにディレクトリまたはファイルを作る
  - 例: ID → `id/id.go`
- **ドメインモデル名（`todo`, `user` 等）のパッケージは domain 層だけが名乗る**。他層で同名パッケージを作らない（import alias 必須化と shadowing 事故を避けるため）
  - presentation 層は `todohandler` / `todorpc` / `todoapi` のように層 / プロトコルを名前に含めて衝突回避する（例: `internal/presentation/todorpc/` → `package todorpc`）
  - usecase 層は対象別ディレクトリ配下で動詞パッケージ（`usecase/todo/define/` → `package define`）として衝突回避済み
  - infra 層は対象 + 役割の合成命名で衝突回避する（例: `internal/infra/repository/todorepo/` → `package todorepo`、`internal/infra/repository/userrepo/` → `package userrepo`）

## 変数名

- **Usecase 層で** Repository の戻り値を受ける変数は、実装詳細ではなく、**発生した事象**で命名する
  - 例: `saved`, `persisted`, `found`
  - NG: `row`, `dbTodo` などの永続化層を匂わせる名前 — Usecase 層が Repository の実装（DB かメモリかファイルか）を前提にしてしまう
  - 注: `Add` / `Save` を呼び分ける構成でも、Repository の戻り値を受ける変数名は `saved` / `persisted` / `found` で共通でよい（`created` / `updated` のように区別しない）。区別する必要があれば呼び出し側のドメインオブジェクトで判別する
  - 適用範囲: この指針は **usecase 層** に対するもの。リポジトリ実装層では実装依存の命名（sqlc 慣習の `row` など）が逆に望ましい（永続化層由来の生データを表すため）
