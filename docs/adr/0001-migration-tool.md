# ADR-0001: マイグレーションツールに goose を継続採用

## ステータス

採択（2026-04-26）

## コンテキスト

プロジェクトは Go + PostgreSQL + sqlc + pgx の構成で、マイグレーションツールに goose v3 を採用していた。

採用当初の根拠は「**goose の Go migrations を使って開発時の seed も書ける**」だったが、後にこれは誤った理解と判明した:

- goose の Go migrations は migration として `goose_db_version` テーブルで追跡される
- 一度適用すると再実行されない
- **開発時にリフレッシュしたい seed には不向き**

このため、seed は別 Go プログラム（`cmd/seed/main.go`）で書く方針に変更した。選定の本来の根拠が崩れたため、ツールを再評価した。

## 検討した代替案

| ツール | 特徴 |
|---|---|
| **tern v2** (jackc/tern) | pgx 作者製、PostgreSQL 専用、pgx native の Go API |
| **golang-migrate** | 最大コミュニティ (Star 約 18.4k)、多 DB 対応、pgx v5 native driver あり |
| **Atlas** (ariga/atlas) | declarative + versioned 両対応、sqlc 連携、destructive 検出 |
| **dbmate** | schema.sql の自動 dump 機能、Rails 的アプローチ |

## 判断の経緯

### 当初の方向: tern への切替を検討

- pgx native で dbtest が綺麗になると判断
- pgx エコシステムの純正であり思想が一貫すると判断

### 評価エージェントによる訂正

独立評価で以下の**事実誤認**が判明:

1. **「dbtest で 2 接続必要」は誇張**
   - goose は `stdlib.OpenDBFromPool(pool)` で同じ pgxpool から `*sql.DB` を生成可能
   - 「pool が分離する」というのは実装次第
2. **golang-migrate の `database/sql` 依存も事実誤認**
   - pgx v5 native driver が公式提供されている
3. **Atlas が検討漏れ**
   - sqlc との連携機能（共通 schema.sql を介した疎結合）を持つ

### Atlas の深掘り評価

- Atlas は `sqlc.yaml` を読まない。共通の `schema.sql` を介した**疎結合連携**
- 自動生成は**単純なスキーマ追加・削除**のみ強力
- データ移行（UPDATE）や複雑な型変換（`USING` 句が必要なケース）は**手書き必須**
- Terraform 的なメンタルモデル、Rails 出身者には**癖**がある（schema.sql を編集対象、dev-url 必須、down が薄い等）

### dbtest 観点の再評価

goose と tern どちらも `pgxpool.Pool` から「何かを取り出す」ラップが必要:

- goose: `stdlib.OpenDBFromPool(pool) → *sql.DB`
- tern: `pool.Acquire() → *pgx.Conn`

**行数差は小さく、決定要因にならない**。

### schema.sql 出力の検討

- `docker compose exec -T db pg_dump` でクリーンな schema.sql 出力可能（約 73 行）
- ただし**Go プロジェクトでは schema.sql 管理は少数派**
- 主要 OSS（kubernetes、prometheus、cockroachdb 等）に schema.sql 文化なし
- migrations + `psql` で十分という慣習が一般的

## 決定

**goose v3 を継続採用する。**

- migration ツール変更しない
- schema.sql 自動 dump も採用しない
- seed は別 Go プログラム（`cmd/seed/main.go`）で書く方針（migration ツールとは独立）

## 帰結

### 利点

- 既存資産（マイグレーションファイル、`scripts/migrate.sh`）をそのまま使える
- 切替コストゼロ
- Go コミュニティで広く使われている（Star 約 10.6k）
- 将来 Go migrations が必要になった時に使える保険

### 欠点

- pgx native ではない（`OpenDBFromPool` で同 pool 共有は可能だが、思想的には完全一致ではない）
- Atlas の自動生成 / destructive lint は享受できない
- 当初の選定根拠（seed もできる）は無効化された

### スコープの確定

このテンプレートのスコープは **「Go + PostgreSQL」** に限定する:

- pgx, sqlc, PG 固有 SQL（trigger、TIMESTAMPTZ 等）を多用
- DB 非依存性は実質的に成立しない
- migration ツールの多 DB 対応は実用上のメリットなし

## 参考リンク

- [pressly/goose](https://github.com/pressly/goose)
- [jackc/tern](https://github.com/jackc/tern)
- [golang-migrate/migrate](https://github.com/golang-migrate/migrate)
- [ariga/atlas](https://github.com/ariga/atlas)
- [amacneil/dbmate](https://github.com/amacneil/dbmate)
- [pgx stdlib OpenDBFromPool](https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib#OpenDBFromPool)
- [Atlas sqlc-versioned](https://atlasgo.io/guides/frameworks/sqlc-versioned)
