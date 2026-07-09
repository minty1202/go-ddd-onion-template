# 議事録: 2026-04-17 — アーキテクチャ方針決定（layer-first → feature-first 移行の検討）

## 背景・目的

### 発端
現プロジェクトは layer-first 構造で、`internal/domain/todo`, `internal/usecase/todo`, `internal/infrastructure/repository/todo` の 3 箇所が全て `package todo` となっていた。同名パッケージが同一ファイル内で衝突し、import エイリアス（`domainTodo "..."` のような）が常時必要な状態。ユーザーが「shadowing（同名パッケージによる識別子衝突）の DX がここまで悪いとテンプレートとしては微妙」と指摘。

### 目的
このプロジェクトは**今後の Go 新規開発の土台（テンプレート）**にする前提で作られている。単発の不便を解消するのではなく、「**DDD + オニオンアーキテクチャのテンプレートとして実践的に一番良い構成**」を研究しながら決定する必要があった。

### 判断軸（終始貫いた基準）
- Go 公式のスタンスと実運用プロジェクトの実態を優先
- 教育用テンプレートリポジトリの事例は弱い根拠として扱う
- 原典（Evans, Vernon, Martin, Fowler, Palermo）と著名 Gopher（Cheney, Kennedy, Cox）の発言を強い根拠として扱う
- 推測で「標準」「ベストプラクティス」と断定しない

---

## 議論の流れ

### 1. ルール読み込みの調整

.claude/rules/ 配下のルールファイルが自動読み込みされていないことが判明。

- `advice.md`（paths: `**/*`）と `no-code.md`（paths: `**/*.go`, `**/*.proto`, `**/*.sql`）は実質的に全作業で適用されるはずなのに、自動ロードされていない
- `CLAUDE.md` の先頭に `@.claude/rules/advice.md` と `@.claude/rules/no-code.md` を import として追加、常時適用される状態にした

加えて、`no-code.md` と memory `feedback_test_examples.md` に矛盾があったので整理：
- no-code.md：「Go / proto / SQL のコードを書かない、別ドメインで例示する」
- memory：「テストサンプルは別ドメインに置き換えず素直に書く」

**議論の結果**：テストサンプルは写経素材だから別ドメインに置き換えると import パスが架空になり学習にならない、という意図。ただし「写経」だからといって学習用に簡略化するわけではなく、実務で通用する品質で書くべき。さらに、「テストが特別」ではなく「**Go の具体的な書き方を学ぶためのコード**」が広い例外。

no-code.md に例外節を追加し、memory の feedback_test_examples.md は削除（役割がルールに移行したため）。

### 2. 現状構造のパッケージ命名調査

ローカルにあるリファレンスプロジェクトの実装を読んで、layer-first でどう衝突を回避しているか確認。

#### 参照先
- `~/Development/my/go_lang/references/go-clean-arch`
- `~/Development/my/go_lang/references/go-clean-template`
- `~/Development/my/go_lang/references/wild-workouts-go-ddd-example`

#### 見えたパターン
- **go-clean-arch（bxcodec）**：`domain/` に全ドメインを 1 パッケージで集約（`package domain`）、`article/` に service 層を feature 名で
- **go-clean-template（evrone）**：layer-first（`entity/usecase/repo/controller`）だが `usecase/task/` のように feature でサブ分割
- **wild-workouts（ThreeDotsLabs）**：feature-first（`trainings/trainer/users`）、各 feature 内に `adapters/app/domain/ports` のオニオン

ここで「package 名を 3 回繰り返さないために、どの構造を取るか」の選択肢が出た。

### 3. 命名パターンの候補整理（1 回目、推測混入）

Assistant が「参考 3 プロジェクトで採用されていない」「〜が一般的」などの表現で命名案を提示。ユーザーが「それは事実に基づいてる？」と指摘。

**Assistant 側の問題**：ローカル 3 プロジェクトだけを見て「業界全体」の一般化を語っていた。

**修正**：バックグラウンドで調査エージェントを走らせ、Go 公式 / 著名 Gopher / 主要 OSS の pkg.go.dev で一次ソースを確認。

#### 調査結果（1 回目、pkg.go.dev ベース）

| パターン | 代表例 |
|---|---|
| フィーチャー優先 + 各レイヤーはジェネリック名（`package usecase`, `package repository` 等） | bxcodec/go-clean-arch, go-kit |
| レイヤー優先 + ドメイン別サブディレクトリ | evrone/go-clean-template |
| ドメイン + レイヤー suffix（`userbus`, `userapp`, `userdb`） | ardanlabs/service |
| `tododomain` / `todousecase` のようなプレフィックス単体 | 目立った star 1k+ プロジェクトで確認できず |

**Go 公式の立場**（[go.dev/blog/package-names](https://go.dev/blog/package-names)）：
- 同名パッケージは許容（`runtime/pprof` と `net/http/pprof` が並存）
- ただし「頻繁に一緒に使うパッケージは別名にすべし」と明記
- import エイリアスは「衝突回避のためだけに使う」[go.dev/wiki/CodeReviewComments](https://go.dev/wiki/CodeReviewComments)

### 4. feature-first vs layer-first の方向性

ユーザーの初期感覚は「B（layer-first）の方が馴染み深い。オニオンアーキテクチャの同心円イラストとそのまま一致する」。この背景に Rails 長年経験があり、「設定より規約」の世界観で**ディレクトリ構造を見ればアーキテクチャが即わかる**ことに慣れている、と自己分析された（user memory に追加）。

一方、feature-first（A）は shadowing の心配がほぼない。ただし「A は Go ハックじゃないか？アーキテクチャとして通用しているのか？」という懸念。

### 5. feature-first の正統性調査

バックグラウンドで原典調査を実施。

#### 調査結果（原典レベル）
- **Palermo "The Onion Architecture"（2008）**（[jeffreypalermo.com](https://jeffreypalermo.com/2008/07/the-onion-architecture-part-1/)）：単一アプリを想定した記述。複数 BC への明示的言及なし。1 アプリ = 1 オニオンを断言せず、複数オニオンを禁じてもいない
- **Evans 『Domain-Driven Design』(2003)**：Bounded Context は「モデルが一貫性を持つ境界」、各 BC のアーキテクチャは縛らない（[Fowler bliki](https://martinfowler.com/bliki/BoundedContext.html) で要点確認）
- **Vernon 『Implementing Domain-Driven Design』(2013) 第 4 章**：「DDD は特定のアーキテクチャを強制しない。各 Bounded Context が Layered / Hexagonal / SOA / CQRS / Event-Driven などから選べる」
- **Fowler "PresentationDomainDataLayering"**（[martinfowler.com](https://martinfowler.com/bliki/PresentationDomainDataLayering.html)）：**最重要の一次根拠**。「once any of these layers gets too big you should split your top level into domain oriented modules which are internally layered」→ まさに feature-first + 内部オニオン
- **Martin "Screaming Architecture"（2011）**（[blog.cleancoder.com](https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html)）：トップレベルはフレームワークではなく業務ドメインを叫ぶべし
- **Bogard "Vertical Slice Architecture"（2018）**（[jimmybogard.com](https://www.jimmybogard.com/vertical-slice-architecture/)）：リクエスト（機能）単位で垂直に束ね、スライス間の結合は最小、スライス内の結合は最大

#### 結論
「BC ごとに独立したオニオン」「feature 単位で内部に層を持つ」という考え方は **Go 以前から存在する一般的な設計思想**。feature-first + 内部オニオンは Fowler / Uncle Bob / Vernon / Bogard の系譜に位置づけられる。**A は Go 独自のハックではない**。

### 6. Bounded Context の定義に関する認識合わせ

Assistant が BC 概念を少し広く使いすぎていた箇所がユーザーから指摘された：「Todo アプリでスケジュール機能を追加するのは BC ではなく、同じ BC 内の別アグリゲート／サブドメイン」。

**整理**：
- BC は「モデルが一貫性を持つ境界」であって、Todo と Schedule のように**同じアプリ・同じコンテキスト内**で追加される機能は**1 BC 内の別アグリゲート**に過ぎない
- したがって A の正統性は「BC ごとにオニオン」ではなく、「**1 BC の中で、アグリゲート／機能ごとにモジュール化して、各モジュール内部にレイヤーを持つ**」という設計として理解すべき
- 根拠は Fowler（機能で分けて内部で層化）、Uncle Bob（ドメインが叫ぶ）、Bogard（機能で縦切り）

### 7. Go 実運用プロジェクトでの feature-first 採用率

次にユーザーから「feature-first はどこまで一般的か？Go の有名プロジェクトで採用されているか？」という質問。テンプレートレベルではなく、大規模実運用の実態を調べる必要があった。

#### 調査結果（大規模業務アプリ／API サーバー系、一次ソース確認）

| プロジェクト | 構造 | 参照 |
|---|---|---|
| Terraform | `internal/` 直下に feature 分割 | [internal/](https://github.com/hashicorp/terraform/tree/main/internal) |
| Vault | リポジトリ直下に feature 分割 | [vault root](https://github.com/hashicorp/vault) |
| Consul | `internal/` 直下に feature 分割 | [internal/](https://github.com/hashicorp/consul/tree/main/internal) |
| Caddy | リポジトリ直下 | [caddy root](https://github.com/caddyserver/caddy) |
| MinIO | `internal/` 直下 | [internal/](https://github.com/minio/minio/tree/master/internal) |
| Gitea | **layer-first（`models/modules/routers/services`）** 唯一の例外 | [gitea root](https://github.com/go-gitea/gitea) |
| CockroachDB | `pkg/` 直下に feature 分割 | [pkg/](https://github.com/cockroachdb/cockroach/tree/master/pkg) |
| Grafana | `pkg/` 直下に feature 中心 | [pkg/](https://github.com/grafana/grafana/tree/main/pkg) |
| Kubernetes | `pkg/` 直下に feature 分割 | [pkg/](https://github.com/kubernetes/kubernetes/tree/master/pkg) |

#### 判明事項
- **layer-first を採る有名 Go プロジェクトは Gitea のみ**。他はほぼ全て feature-first または feature-first に近い独自構造
- Go 中核メンバーの立場：
  - Russ Cox（Go tech lead）：`golang-standards/project-layout` を「**Go の標準ではない**」と公に否定 [issue #117](https://github.com/golang-standards/project-layout/issues/117)
  - Dave Cheney：「util / common / helpers / base のような汎用名（層名もこれに含まれる）を避け、**behavior（振る舞い）で切れ**」[原文](https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common)
  - Bill Kennedy（ardanlabs）：Package-Oriented Design、feature で割ることを推奨 [原文](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)

#### 判断
**テンプレートとして Go で自然・本番で通用する構成を選ぶなら A（feature-first）** が明確な答え。B の魅力の正体は Rails / MVC 経験の親しみやすさで、アーキテクチャ上の優位ではない。

### 8. feature-first（A）で確定

ユーザーから「一旦 feature first で組み直すことは決定させましょうか」で方針確定。

### 9. feature 内部の構造決定

#### 論点 1: presentation 層の構造
- サブ分割するか、フラットか
- HTTP のみ想定なら YAGNI（必要になるまで作らない原則）でフラット
- **決定：フラット（`presentation/` 単体、`package presentation`）**

#### 論点 2: server / http の配置
ユーザーから「`server/` が `todo/` と同じ階層にあるのは毛色が違って気持ち悪い」との指摘。

選択肢：
1. `features/` ラッパーで business feature を括る
2. `platform/` のようなインフラ集約ディレクトリ
3. `cmd/server/` に server 組み立てを移す

議論：
- 1 は Go で `features/` ラッパーを見かけない
- 2 は platform が抽象的な汎用名で Dave Cheney の批判対象に近い
- 3 は「cmd/ にあるコードは `internal/` の可視性保護を受けないので、ルート routes.go / middleware.go は置きづらい」

さらにテラフォーム等の大規模プロジェクトを確認すると**フラットで business feature と infrastructure を混在させている**。つまり「気持ち悪い」は Go 的には許容される。

**決定：Go 慣習に従いフラット**。`internal/` 直下に `todo/`, `user/`, `http/`, `db/`, `config/`, `logger/` などを並列配置。

#### 論点 3: ID の typed / untyped 判断

Assistant が「DDD ベストプラクティスとして typed ID per feature」と言及したが、ユーザーから「事実に基づいてる？」と指摘。

調査：
- ローカル参照プロジェクトでは typed ID は**一切採用なし**。全て `string` / `int64` / `uuid.UUID` 直接
- さらに実運用プロジェクトを調査：

| プロジェクト | ID の扱い | 参照 |
|---|---|---|
| ardanlabs/service | `uuid.UUID` 直接 | [model.go](https://raw.githubusercontent.com/ardanlabs/service/master/business/domain/userbus/model.go) |
| stripe/stripe-go | `ID string` | [customer.go](https://raw.githubusercontent.com/stripe/stripe-go/master/customer.go) |
| wild-workouts | `UUID string` | [training.go](https://raw.githubusercontent.com/ThreeDotsLabs/wild-workouts-go-ddd-example/master/internal/trainings/domain/training/training.go) |
| sklinkert/go-ddd | `uuid.UUID` 直接 | [go-ddd](https://github.com/sklinkert/go-ddd) |
| marcusolsson/goddd（Evans Cargo 例の Go 移植） | `type TrackingID string` **採用** | [cargo.go](https://raw.githubusercontent.com/marcusolsson/goddd/master/cargo.go) |

Three Dots Labs の記事群：「enum や値オブジェクトには typed を推奨」だが「**ID は例外的に string/uuid のまま**」の立場を記事実装で示している：
- [DDD Lite](https://threedots.tech/post/ddd-lite-in-go-introduction/)
- [Repository pattern](https://threedots.tech/post/repository-pattern-in-go/)
- [Safer enums in Go](https://threedots.tech/post/safer-enums-in-go/)

ユーザーから鋭い仮説：「Web CRUD ではエンティティ同士の比較を同じ関数でやらないから、typed ID の型安全メリットが発動する場面が少ないんじゃないか？」

整理：
- typed ID の価値は「**複数の ID 型を同じ関数に並べて渡す瞬間**」にだけ発動
- Web CRUD は「1 エンティティの ID で 1 操作」が大半で、恩恵発動頻度が低い
- 境界（DB、JSON、HTTP、外部 API）で変換コストが出る一方、安全性の恩恵は薄い
- Bill Kennedy / Stripe / Three Dots Labs が揃って plain primitive を選んでいるのはこのトレードオフの判断結果

**決定：`ulid.ULID` を直接使う**。typed ID ラッパーは作らない。共通の `internal/id/` パッケージは廃止。

#### 論点 4: errdefs の扱い

ユーザーから「errdefs ってどっかからパクってきた名前なので、どう実装しているか見たい」との要望。

#### Moby / containerd の errdefs 調査（[moby/moby/errdefs](https://github.com/moby/moby/tree/master/errdefs) / [containerd/errdefs](https://github.com/containerd/errdefs)）

- **Moby**：単一パッケージ、フラット構造。`defs.go` に interface 定義（`ErrNotFound`, `ErrInvalidParameter`, `ErrConflict`, `ErrUnauthorized`, `ErrForbidden` など）、`helpers.go` に wrapper 関数
- **containerd**：同様に単一パッケージ、フラット。gRPC ステータスコードにマップ
- 共通：**層（domain / usecase）で分けず、エラーの「カテゴリ」で分類**。どの層からも同じ interface を使える

ユーザーの洞察：「Moby の errdefs は『共通する層』という考え方ではなく、**さらに外側の全てに依存される何か**」＝「**層の内側・外側という概念を超えた、横断的関心事（cross-cutting concern）**」。

さらにユーザー提案：「内側の詳細を外に漏らさないを原則にするなら、interface だけ共通に置いて、具体エラーは各 feature で作る」。これが DDD 原則と Go の interface-based polymorphism を噛み合わせた最も綺麗な形。

**決定：
- `internal/errdefs/` には interface 定義だけを置く（`ErrNotFound`, `ErrConflict` 等のカテゴリ契約）
- 各 feature の domain で**具体エラー型**を定義し、`errdefs/` の interface を実装する
- 層（domain / usecase）で errdefs を分割しない**

### 10. ディレクトリ構造の最終形（決定事項の集約）

```
cmd/
  main.go                         薄いエントリポイント
internal/
  todo/                           feature
    domain/                       package domain（エンティティ、値オブジェクト、集約、domain エラー）
    usecase/                      package usecase
    infrastructure/               package infrastructure（DB 実装等）
    presentation/                 package presentation（HTTP ハンドラ、DTO、エラーマッピング）
    todo.go                       package todo（組み立て関数 New()）
  user/                           別 feature（同じ内部構造）
  http/                           HTTP サーバー、ルーティング、ミドルウェア
  db/                             DB 接続プール
  config/                         設定読み込み
  logger/                         ロガー
  errdefs/                        エラー interface のみ定義（横断契約）
```

---

## 決定事項まとめ

1. **feature-first アーキテクチャ**（`internal/<feature>/` 配下にオニオン層を持つ）
2. **各 feature の内部**：`domain/`, `usecase/`, `infrastructure/`, `presentation/`（フラット、YAGNI）、組み立て関数を feature ルートに置く
3. **`internal/` 直下はフラット**。business feature と横断的関心事（http, db, config, logger, errdefs）を並列配置。`platform/` / `foundation/` のようなラッパーは作らない
4. **ID は `ulid.ULID` を直接使用**。typed ID ラッパーは作らない
5. **errdefs は interface のみ定義する横断契約パッケージ**。具体エラー型は各 feature の domain に置く

---

## 未確定事項

### 1. `apperr` の扱い
現プロジェクトの `internal/apperr/violation.go` との役割が errdefs と被る可能性。要調査・判断。

### 2. feature 間の依存ルール
feature A が feature B を直接 import していいかの方針が未確定。`depguard` 等の静的検証も含めて別途検討。

### 3. 横断処理（orchestration）の置き場
独立した複数 feature が連携する処理の置き場。現時点では発生していないため、必要になったときに判断。

---

## このセッションで改定したルール・メモリ

### プロジェクト CLAUDE.md
- 先頭に `## 前提` 節を追加（テンプレート目的・研究方針・ルール化方針を明記）
- `@.claude/rules/advice.md` と `@.claude/rules/no-code.md` の常時 import を追加
- `## 進め方` 節を追加（対話的に 1 論点ずつ、調査結果を網羅列挙しない）

### プロジェクト `.claude/rules/`
- `no-code.md` に例外節を追加：「Go の具体的な書き方を学ぶためのコード」は Todo ドメインで実務品質で書く
- `go/architecture.md` を全面書き換え：feature-first 構造、各層の責務、横断的関心事、未確定事項を明記

### グローバル `~/.claude/settings.json`
- `permissions.allow` に `WebFetch` と `WebSearch` を追加（サブエージェントも含めて調査可能にするため）

### グローバル `~/.claude/CLAUDE.md`
- `## 権限開放の方針（厳守）` 節を追加：`permissions.allow` に入れてよいのは read-only の操作のみ、副作用を伴う操作は絶対に入れない

### memory（`~/.claude/projects/-Users-aki-Development-my-go-lang-todo-app/memory/`）
- `feedback_test_examples.md` を削除（役割が no-code.md の例外節に移行）
- `feedback_interactive_decisions.md` を新規作成、後に「次の論点も AI が決めない」を追記
- `feedback_explain_english_terms.md` を新規作成（専門用語は初出で日本語併記）
- `feedback_evidence_weight.md` を新規作成（テンプレートは弱い証拠、実運用・原典を優先）
- `user_rails_background.md` を新規作成（Rails 経験者、構造から一目のアーキテクチャを好む）

---

## 参照した情報源

### Go 公式
- [Package names — Go Blog](https://go.dev/blog/package-names)
- [Go Code Review Comments (Imports, Package Names)](https://go.dev/wiki/CodeReviewComments)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Modules Layout](https://go.dev/doc/modules/layout)

### Go コミュニティ / 著名 Gopher
- [Dave Cheney - Avoid package names like base, util, or common](https://dave.cheney.net/2019/01/08/avoid-package-names-like-base-util-or-common)
- [Bill Kennedy - Package Oriented Design (ardanlabs)](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)
- [Russ Cox on golang-standards/project-layout issue #117](https://github.com/golang-standards/project-layout/issues/117)
- [golang-standards/project-layout README](https://github.com/golang-standards/project-layout)

### アーキテクチャ原典
- [Jeffrey Palermo - The Onion Architecture: Part 1](https://jeffreypalermo.com/2008/07/the-onion-architecture-part-1/)
- [Jeffrey Palermo - Onion Architecture tag](https://jeffreypalermo.com/tag/onion-architecture/)
- [Martin Fowler - BoundedContext](https://martinfowler.com/bliki/BoundedContext.html)
- [Martin Fowler - PresentationDomainDataLayering](https://martinfowler.com/bliki/PresentationDomainDataLayering.html)
- [Robert C. Martin - Screaming Architecture](https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html)
- [Jimmy Bogard - Vertical Slice Architecture](https://www.jimmybogard.com/vertical-slice-architecture/)
- [Vaughn Vernon - Implementing Domain-Driven Design, Ch.4 Architecture](https://www.oreilly.com/library/view/implementing-domain-driven-design/9780133039900/ch04lev1sec3.html)

### Three Dots Labs の DDD 記事群
- [DDD Lite in Go - Introduction](https://threedots.tech/post/ddd-lite-in-go-introduction/)
- [Repository pattern in Go](https://threedots.tech/post/repository-pattern-in-go/)
- [Safer enums in Go](https://threedots.tech/post/safer-enums-in-go/)

### Go DDD 実装記事
- [pkritiotis - DDD Entity in Go](https://pkritiotis.io/ddd-entity-in-go/)
- [Ompluscator - Practical DDD in Golang: Entity](https://www.ompluscator.com/article/golang/practical-ddd-entity/)

### 大規模実運用 Go プロジェクト
- [Terraform internal/](https://github.com/hashicorp/terraform/tree/main/internal)
- [Vault](https://github.com/hashicorp/vault)
- [Consul internal/](https://github.com/hashicorp/consul/tree/main/internal)
- [Caddy](https://github.com/caddyserver/caddy)
- [MinIO internal/](https://github.com/minio/minio/tree/master/internal)
- [CockroachDB pkg/](https://github.com/cockroachdb/cockroach/tree/master/pkg)
- [Grafana pkg/](https://github.com/grafana/grafana/tree/main/pkg)
- [Kubernetes pkg/](https://github.com/kubernetes/kubernetes/tree/master/pkg)
- [Gitea](https://github.com/go-gitea/gitea)

### Go DDD / Clean Architecture テンプレ（参考度：低〜中、弱い証拠として扱う）
- [bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch)（[pkg.go.dev](https://pkg.go.dev/github.com/bxcodec/go-clean-arch)）
- [evrone/go-clean-template](https://github.com/evrone/go-clean-template/tree/master/internal)（[pkg.go.dev](https://pkg.go.dev/github.com/evrone/go-clean-template)）
- [ThreeDotsLabs/wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/tree/master/internal)
- [amitshekhariitbhu/go-backend-clean-architecture](https://github.com/amitshekhariitbhu/go-backend-clean-architecture)
- [qiangxue/go-rest-api](https://github.com/qiangxue/go-rest-api/tree/master/internal/album)
- [AleksK1NG/Go-Clean-Architecture-REST-API](https://github.com/AleksK1NG/Go-Clean-Architecture-REST-API/tree/master/internal)
- [sklinkert/go-ddd](https://github.com/sklinkert/go-ddd/tree/main/internal)
- [go-kit/kit](https://pkg.go.dev/github.com/go-kit/kit)

### 実運用 Go プロジェクトの ID 実装
- [ardanlabs/service - userbus model.go](https://raw.githubusercontent.com/ardanlabs/service/master/business/domain/userbus/model.go)
- [ardanlabs/service - userbus.go](https://raw.githubusercontent.com/ardanlabs/service/master/business/domain/userbus/userbus.go)
- [stripe/stripe-go - customer.go](https://raw.githubusercontent.com/stripe/stripe-go/master/customer.go)
- [wild-workouts - training.go](https://raw.githubusercontent.com/ThreeDotsLabs/wild-workouts-go-ddd-example/master/internal/trainings/domain/training/training.go)
- [marcusolsson/goddd - cargo.go](https://raw.githubusercontent.com/marcusolsson/goddd/master/cargo.go)
- [GORM - model.go](https://raw.githubusercontent.com/go-gorm/gorm/master/model.go)

### errdefs 実装
- [moby/moby errdefs](https://github.com/moby/moby/tree/master/errdefs)
- [containerd/errdefs](https://github.com/containerd/errdefs)

### ardanlabs/service リポジトリ構造
- [ardanlabs/service business/domain](https://github.com/ardanlabs/service/tree/master/business/domain)
