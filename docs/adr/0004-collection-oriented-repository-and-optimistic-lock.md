# ADR-0004: Collection-Oriented Repository への切替と楽観ロック採用

## ステータス

採択（2026-04-29、修正: 2026-04-30 — `IncrementVersion()` を `SyncVersion(v int)` に統一。Hibernate の `incrementConcurrencyVersion()` 流儀（+1 を予測）ではなく、DB の `RETURNING version` で取得した値を反映する流儀（DB 真実主義、ent / Hibernate Hook と等価）に揃えた。`SaveWithLock` の persistFn シグネチャも `(newVersion int, err error)` に明示）

## コンテキスト

ADR-0002 で **Persistence-Oriented Repository（`Save` 1 つで Upsert）** を採用した。その後、楽観ロックの採用を検討する中で、構造的な制約が判明した。

### 判明した問題

楽観ロックは UPDATE で `WHERE id = ? AND version = ?` を使う仕組み。`Save` 1 つで Upsert を実現する場合、UPSERT の内部で「新規 INSERT」と「既存 UPDATE」を区別する必要が出る:

- **新規 INSERT**: 楽観ロック不要（新規挿入なので競合しようがない）
- **既存 UPDATE**: 楽観ロックチェック必要（`WHERE version = ?` で取得時 version との一致を確認、影響行数 0 なら conflict）

### 判別方法の検討と問題

| 候補 | 問題 |
|---|---|
| UPSERT + `xmax = 0` で判別 | PostgreSQL の hidden column（MVCC 内部実装）に依存、SQL ハック |
| 集約に `persisted` フラグ | ドメインに永続化の関心を持ち込む |
| 主キー（ID）で判別（GORM 流儀） | ulid を Go 側で生成しているため使えない（新規でも非 zero） |
| Repository が SELECT で存在確認 | クエリ重複、レース条件 |
| Save 内部で SELECT → INSERT/UPDATE 分岐 | usecase が事前に `FindByID` しているケースで二重 SELECT、無駄 |
| **`Add` / `Save` 分離（Collection-Oriented）** | **usecase 側で「新規 → Add」「既存 → Save」を明示的に呼び分け、判別ロジック不要** |

### 維持したい制約

- **DDD オニオンの維持**: ドメインモデルとデータモデルの分離、private フィールド + Always Valid 原則
- **ドメインに永続化の関心を持ち込まない**: `persisted` フラグは入れない
- **SQL ハックを避ける**: `xmax` のような PostgreSQL 内部実装への依存は避ける
- **ulid を Go 側で生成**: 集約生成時に ID を確定（Always Valid 原則と整合）

これらすべてを満たすには、`Save` 1 つで Upsert を実現する道は塞がれる。Vernon IDDD のもう 1 つの流儀である **Collection-Oriented Repository（`Add` 新規 + `Save` 既存更新）** に切り替えるのが筋となった。

## 検討と判断

### 1. Collection-Oriented Repository への切替

Vernon IDDD では Persistence-Oriented（`Save` 1 つ Upsert）と Collection-Oriented（`Add` 新規 + `Save` 既存更新）の両方が「正解」とされる。

Go + 生 SQL + ulid Go 側生成という前提下では:

- **Persistence-Oriented**: `Save` 内部で「新規 vs 既存」の判別が構造的に必要 → 各種ハック・暗黙ルール・クエリ重複のいずれかを伴う
- **Collection-Oriented**: usecase 側で「新規 → Add」「既存 → Save」を明示的に呼び分け → 判別ロジック不要

usecase の典型フロー:

- **新規生成**: `NewTodo → repo.Add(todo)`
- **既存更新**: `repo.FindByID → todo.メソッド → repo.Save(todo)`

usecase 側は「新規か既存か」を元から知っている（`FindByID` を呼んでいるかで決まる）ため、呼び分けは自然で読み手にも明示的。

### 2. 楽観ロック採用（ADR-0002 のスコープ外条項を覆す）

ADR-0002 のスコープ外で「採用しない」とした楽観ロックを採用する:

- 並行更新の競合検知は本番 Go プロジェクトの標準（Vernon IDDD でも Persistence-Oriented Repository は通常 `version` を伴う）
- ADR-0002 の理由文「単一クライアント前提のサンプル実装」は **practical_mindset 違反**（プロジェクト固有の状況を盾にした簡略化）で、再検討に値した

### 3. version の保持先

楽観ロックは「取得時の version」を UPDATE 時に WHERE で使う仕組みのため、version をどこかで保持する必要がある:

| 候補 | 評価 |
|---|---|
| **ドメインに version を持つ（Vernon IDDD 流儀）** | Repository interface はシンプル、ただしドメインに永続化の関心が入る |
| Repository interface に version を露出（返り値・引数） | ドメイン綺麗、Repository / usecase が複雑化 |
| Repository が状態保持 | リクエストまたぎ困難、複数集約扱うと混乱 |
| 永続化コンテキスト（Hibernate 相当） | 自前実装は複雑、Go で再発明することになる |

Vernon IDDD は「**version は集約の責務 = 楽観ロックを集約の関心事として扱う**」と整理し、ドメインに持たせる流儀。本プロジェクトもこれに従う。

### 4. 共通化（Aggregate interface + Lock struct）

各集約で version 管理を毎回書くと冗長。実装漏れをコンパイル時に検知できる仕組みが望ましい。

設計:

- `internal/domain/aggregate/` 共通パッケージに `Aggregate` interface と `Lock` struct を定義
- `Aggregate` interface: `Version() int`, `SyncVersion(v int)` を要求
- `Lock` struct: `version int` を持ち、上記メソッドを提供
- 集約は `Lock` を **private 名前付きフィールド**（`lock aggregate.Lock`）で保持し、必要なメソッド（`Version()`, `SyncVersion(v int)`）を集約側で公開する（カプセル化を強くする、IDDD_Samples の `protected` 相当）
- `internal/infra/repository/optimistic.go` に `SaveWithLock(ctx, agg aggregate.Aggregate, persistFn ...) error` ヘルパーを置く
- 各 Repository 実装の `Save` は `SaveWithLock` を呼ぶ
- `SaveWithLock` の引数 `aggregate.Aggregate` 型が、各集約に `Aggregate` interface の実装を **コンパイル時に強制** する

### 5. version 同期（Save 成功後）

Save 成功時、DB 側は `SET version = version + 1` で更新され、`RETURNING version` で新しい version が返る。メモリ上の集約の version もこの値に同期する必要がある（さもないと、同じリクエスト内で再 Save する際に conflict が起きる）:

- Vernon IDDD の Java 実装では Hibernate が自動で同期する（`@Version` フィールドを Hibernate が更新）
- Go では Hibernate 相当がないため、**自前で `SyncVersion(newVersion)` を呼ぶ**（DB から `RETURNING version` で取得した値を集約に反映）
- `SaveWithLock` ヘルパーが Save 成功時に `agg.SyncVersion(newVersion)` を呼ぶ

評価エージェントの調査で、Vernon IDDD の `ConcurrencySafeEntity` には version の自動インクリメントが無い（Hibernate 任せ）ことが確認された。Go で生 SQL ベースの場合、`SyncVersion(v)` を自前で呼ぶのは「ORM がない世界での妥当な選択」となる（DB の真実を反映する形でメモリと同期、ent や Hibernate の Hook と等価な役割）。

## 決定

### Repository interface

旧（ADR-0002）: `Save / RemoveByID / FindByID`

新:

```
Add(ctx, todo) error                       - 新規挿入（INSERT）
Save(ctx, todo) error                      - 既存更新（楽観ロック付き UPDATE）
FindByID(ctx, id) (*Todo, error)
RemoveByID(ctx, id) error
```

- **`Add`**: 集約をリポジトリに新規追加。SQL は INSERT
- **`Save`**: 既存集約の状態を更新。SQL は楽観ロック付き UPDATE（`WHERE id = ? AND version = ?` + `SET version = version + 1`）、影響行数 0 なら `ErrConflict`
- **`FindByID`**: 集約取得（version 含む）
- **`RemoveByID`**: 単純削除（ADR-0002 の決定を維持）

### usecase の流れ

- **新規生成**: `NewTodo(...) → repo.Add(todo)`
- **既存更新**: `repo.FindByID(id) → todo.メソッド → repo.Save(todo)`

### 楽観ロック実装

- 集約は `lock aggregate.Lock` フィールドを private で保持
- 集約は `Version() int` / `SyncVersion(v int)` を公開メソッドとして持つ（embed ではなく手動公開）
- `internal/domain/aggregate/` に `Aggregate` interface（`Version()`, `SyncVersion(v int)`）と `Lock` struct
- `internal/infra/repository/optimistic.go` に `SaveWithLock(ctx, aggregate.Aggregate, persistFn) error` ヘルパー
- Save 成功時、`SaveWithLock` が副作用で `agg.SyncVersion(newVersion)` を呼ぶ（DB から `RETURNING version` で取得した値を反映）
- 競合時は `ErrConflict` を返す（infra 層エラー、usecase 層で `errdefs.NewAborted` にマッピング。gRPC Code の `Aborted` = 並行性コンフリクト。自動リトライはせず、エラーを上層に返してクライアント / ユーザーに判断を委ねる）

### `Add` の楽観ロック扱い

`Add` は新規 INSERT なので楽観ロックは効かせない（version = 0 で挿入、PK 重複なら DB エラー）。

## 帰結

### 利点

- DDD オニオンの「ドメインとデータの分離」を維持
- ドメインに永続化の関心を持ち込まない（`persisted` フラグなし）
- ulid Go 側生成と整合（Always Valid 原則と整合）
- SQL は標準の INSERT / UPDATE で素直、PostgreSQL 内部実装（`xmax`）に依存しない
- 楽観ロック実装が素直（`Save` のみで version 一致をチェック）
- usecase 側で新規 / 既存を明示的に呼び分け、コードの意図が読める
- Vernon IDDD の Collection-Oriented Repository に整合
- 共通化（`Aggregate` interface + `Lock`）で実装漏れをコンパイル時に検知可能

### 欠点・コスト

- ADR-0002 の「`Save` 1 つで簡潔」というメリットを諦める
- usecase 側で新規 / 既存を判別し、`Add` / `Save` を呼び分ける
  - ただし usecase は元から `FindByID` してから操作する流れが多いので、判別自体は自然
- `SyncVersion(v int)` を集約に公開する必要がある（Hibernate がない世界での補完、ent / Hibernate Hook と等価な役割）

### スコープ外（今回扱わない）

- **悲観ロック（`SELECT FOR UPDATE`）**: 採用しない
  - 楽観ロックで足りる前提
  - 競合が頻発する局所場面で必要になれば別 ADR
- **`Remove(aggregate)` メソッド**: ADR-0002 のスコープ外を継続（現状の `RemoveByID` で足りる）
- **`Aggregate` interface への `Equals` / `HashCode` 等の追加**: 必要になった時点で検討

### ADR-0002 との関係（supersedes）

ADR-0004 は ADR-0002 の以下の決定を supersede する:

- **リポジトリ interface**（`Save / RemoveByID / FindByID` → `Add / Save / RemoveByID / FindByID`）
- **並行更新の整合性スコープ外**（採用しない → 採用、楽観ロックを導入）

ADR-0002 の以下の決定は維持する:

- ユースケース命名の業務動詞化（`DefineTodo`, `ReviseTodo`, `CompleteTodo`, `ArchiveTodo`, `ViewTodo`）
- ドメインモデル `Todo { id, title, body, completed }`
- リポジトリに業務動詞メソッドを持たせない
- `Save` の戻り値は `error` のみ

## 参考リンク

- Vaughn Vernon, "Implementing Domain-Driven Design"（赤本, 2013）, ch.10 Aggregates, ch.12 Repositories
- [VaughnVernon/IDDD_Samples](https://github.com/VaughnVernon/IDDD_Samples) — `ConcurrencySafeEntity` 等の実装
- [Khorikov, "Always-Valid Domain Model"](https://enterprisecraftsmanship.com/posts/always-valid-domain-model/)
- [Kamil Grzybek, "Domain Model Validation"](https://www.kamilgrzybek.com/blog/posts/domain-model-validation)
- Eric Evans, "Domain-Driven Design"（青本, 2003）, ch.6 Repositories
- [An Introduction to Generics (Go Blog)](https://go.dev/blog/intro-generics) — interface 引数で済むならジェネリクスを使わない指針
