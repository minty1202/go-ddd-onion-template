---
paths:
  - "sql/migrations/**"
  - "scripts/migrate.sh"
---
# マイグレーションルール

## 命名規則

- `create_xxx` — テーブル作成（例: `create_todos`）
- `add_xxx_to_yyy` — カラム追加（例: `add_email_to_users`）
- `drop_xxx` — テーブル削除（例: `drop_legacy_sessions`）
- `alter_xxx` — テーブル変更（例: `alter_users_add_role`）
- `create_index_xxx` — インデックス追加（例: `create_index_users_email`）

xxx = テーブル名（**複数形**で統一）、yyy = カラム名

## 運用

- タイムスタンプモード（goose デフォルト）
- マイグレーションファイルは `sql/migrations/` に置く
- up/down を必ず両方書く

## タイムスタンプの自動付与

PostgreSQL の慣習に従い、INSERT 時と UPDATE 時を分けて DB 側で自動付与する。

### created_at（INSERT 時）

カラム定義の `DEFAULT NOW()` で付与する。

```sql
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
```

### updated_at（UPDATE 時）

トリガー + 共通 function で付与する。

- 共通 function（例: `set_updated_at`）を**別マイグレーション**で定義し、各テーブルの trigger から `EXECUTE FUNCTION` で呼ぶ
- 関数の中身は `NEW.updated_at = NOW()` のみ
- カラム定義側にも `DEFAULT NOW()` を付けて、INSERT 時の初期値として機能させる
- アプリ側の SQL では `SET updated_at = NOW()` を書かない（書き忘れ防止のため DB に責務を寄せる）

## lint 警告の抑制 (postgres-language-server)

postgres-language-server（pgls）は `DROP TABLE` に対して `lint/safety/banDropTable` という警告を出す。goose の Down セクションでは `DROP TABLE` を書くのが正常運用なので、Down 側に限定して suppression コメントで抑制する。

### 採用している書式（デフォルト: 1 文だけ抑制）

```sql
-- +goose Down
-- pgls-ignore lint/safety/banDropTable
DROP TABLE todos;
```

`-- pgls-ignore <rule>` は**直後の 1 文だけ**に効く。`DROP TABLE` が 1 つだけの素直な Down セクションではこれを使う。

### 他に使える書式

以下も pgls が公式にサポートしている書式なので、状況に応じて使い分けてよい。

| 用途 | 書式 | スコープ |
|---|---|---|
| 次の 1 文だけ抑制 | `-- pgls-ignore lint/safety/banDropTable` | 直後の 1 statement |
| ブロック抑制（複数文） | `-- pgls-ignore-start lint/safety/banDropTable` 〜 `-- pgls-ignore-end lint/safety/banDropTable` | start〜end で囲んだ範囲 |
| ファイル全体抑制 | `-- pgls-ignore-all lint/safety/banDropTable` | ファイル全体（ただし**必ずファイル先頭に置く**、`-- +goose Up` より前） |

ルール指定は階層で広げることもできる。

- `lint/safety/banDropTable` — この 1 ルールだけ
- `lint/safety` — safety カテゴリ全体
- `lint` — lint 機能全体

### 使い分けの指針

- **Down に `DROP TABLE` が 1 つだけ** → インラインの `-- pgls-ignore`（デフォルト）
- **Down に `DROP TABLE` が複数並ぶ**（例: 複合的な巻き戻し） → `pgls-ignore-start` / `pgls-ignore-end` のブロック抑制で Down セクション全体を囲む
- **ファイル全体で banDropTable を無視したい特殊ケース** → ファイル先頭に `-- pgls-ignore-all lint/safety/banDropTable`
  - ただしこの場合、誤って Up セクションに `DROP TABLE` を書いても警告が出なくなるので、基本は使わない

### やってはいけないこと

- `postgres-language-server.jsonc` の `linter.rules.safety.banDropTable` を `"off"` にしてプロジェクト全体で無効化するのは避ける。Up セクションの誤った `DROP TABLE` も検出できなくなる
- suppression コメントに理由のないままルール全体（`lint` のみ指定）を抑制しない。抑制対象は最小スコープに留める
