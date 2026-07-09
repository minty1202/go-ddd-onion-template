# ADR-0002: ユースケース命名の業務動詞化とリポジトリ設計の再評価

## ステータス

採択（2026-04-28、修正: 2026-04-29 — `Remove(aggregate)` を今回採用しない方針に修正。その後 ADR-0004 で「リポジトリ interface（`Save` 1 つ Upsert）」と「並行更新整合性スコープ外条項」を supersede）

## コンテキスト

このプロジェクトは「Go × DDD × オニオンアーキテクチャ」のテンプレートを残すことを目的とする。初期実装には CRUD 思考の名残があり、DDD の本流から外れていることが議論で判明した。

### 元々の実装

- **usecase**: `create_todo`, `get_todo` という CRUD 命名のパッケージ
- **Repository interface** (`internal/domain/todo/repository.go`):
  - `Create(ctx, todo) (*Todo, error)`
  - `Update(ctx, todo) (*Todo, error)`
  - `FindByID(ctx, id) (*Todo, error)`
- **Todo モデル**: `id, title, completed` の 3 フィールド

### 問題意識

- ユースケース名が **CRUD（データ操作）に引っ張られている**。業務上の意図を表していない
- リポジトリの `Create / Update` 分離は DDD 原典のコレクションメタファーから外れる
- 直感的にも `repo.Create(todo)` は英語的に違和感がある（「リポジトリが todo を作る」と読める。何を作るのか）

### 経緯: Slack API の構造から DDD への接続点を感じた

CRUD 思考から脱却する糸口として、Slack API のメソッド命名（例: `chat.postMessage`, `chat.update`, `reactions.add`, `conversations.archive`）が DDD のタスクベース設計と思想が近い、という気づきがあった。リソース名 + **業務動詞** を API レベルで直接表現するスタイル（「メッセージを投稿する」「リアクションを付ける」「会話をアーカイブする」）は、Google AIP の Custom Methods や、Vaughn Vernon IDDD の Application Service と通底する。

この観察から「ユースケース命名も、データ操作ではなく **業務上の使われ方** を反映すべき」という確信に至り、本 ADR の検討に入った。

## 検討と判断

### 1. ユースケースという語そのものの意味

「ユースケース（use case）」という言葉は、文字通り「**利用ケース**」、つまり **業務上の使われ方** を意味する。Ivar Jacobson が UML で導入し、Robert C. Martin の Clean Architecture が usecase 層として採用した語で、いずれも「アクター（ユーザー）がシステムをどう使うか」を表す概念。

ユースケース層が **CRUD 名で構成されているのは、層の名前そのものと矛盾する**:

- CRUD（Create / Read / Update / Delete）は **永続化レイヤーの語彙**（DB 操作）
- ユースケースは **業務上の行為の語彙**（アクターの利用ケース）

注文ドメインで対比すると:

- データ操作の語彙（層の名前と矛盾）: `CreateOrder`, `UpdateOrder`, `DeleteOrder`
- 業務動詞（層の名前と整合）: `PlaceOrder`（注文を確定する）、`ShipOrder`（発送する）、`CancelOrder`（キャンセルする）、`ApplyCoupon`（クーポンを適用する）

業務動詞で命名することにより:

- usecase 名を読んだだけで「この操作で何が起こるか」が分かる（Intention-Revealing Interfaces, Eric Evans）
- 不変条件 / エラー条件が業務動詞ごとに自然に閉じる（業務動詞 ≒ 状態遷移単位）
- 同じドメインから生まれる `Result` 型でも、ユースケースごとに意味（セマンティック）が違うため別物として扱える
- `repository.md` で定義する **「リポジトリには業務動詞を持たせない」** 方針と整合する（業務動詞はユースケース層とドメインオブジェクトに閉じる）

### 2. 業務動詞でモデル化することは DDD の本流か

独立調査エージェントによる原典・OSS 横断調査の結果、**5 / 5 で本流** と裏付けられた:

- 原典:
  - Eric Evans (Intention-Revealing Interfaces、命名はユビキタス言語の一部)
  - Vaughn Vernon (IDDD: Application Service の Command 設計)
  - Greg Young ("Task-Based UI"、CRUD-DTO 批判: "the intent of the user was lost")
  - Udi Dahan ("From CRUD to Domain-Driven Fluency"、ライフサイクル動詞への置き換え)
  - Robert C. Martin (Clean Architecture: usecase = アプリ固有の業務ルール)
- 公式サンプル: VaughnVernon/IDDD_Samples（DDD 原典の付属コード）。`ForumApplicationService` 等で Application Service の Command 命名を例示
- Go OSS の補助例: ThreeDotsLabs/wild-workouts-go-ddd-example（star 6.3k、教育用に簡略化されているため補助）。`internal/trainings/app/command/` に `schedule_training.go`, `cancel_training.go`, `reschedule_training.go` のような業務動詞ベース命名
- 反対論: 「単純な管理画面は CRUD で良い」（程度問題、Mathias Verraes ら）。Greg Young 自身も "CQRS does not require a task based UI" と認めるが、業務動詞化方針自体を否定する原典は無し

### 3. 業務動詞を見つけるプロセス（議論の経緯）

業務動詞を見つけるのは、抽象的なドメインだと意外と難しい。今回の議論で得た気づき:

- **抽象的に捉えると業務動詞は浮かばない**: 「Todo を作る」と CRUD で考えると CRUD の引力に負ける
- **具体化すると業務動詞が見える**: スケジュール（「予定を入れる」「予定を確認する」）や GitHub issue（「起票する」「着手する」「完了する」「クローズする」）など、業務イメージが定着している近接ドメインを借りると自然に出てくる
- **ユーザー自身の感覚で「これをやる」という意図を言語化することが鍵**:
  - 「具現化させる / 象らせる / 曖昧なものをタスクとして切り出す」 → `Define`（定義する、輪郭を与える）
  - 「やる / やらない を決めて、未来のために残す」 → `Define`
  - 「内容を見直して直す（修正する）」 → `Revise`（改訂する。`Modify` / `Change` のように汎用的な語ではなく、「改訂する」という業務的ニュアンスのある語を選ぶ）
  - 「削除」より「アーカイブ」のほうが業務的意図が強い → `Archive`（Gmail / Slack の archive 概念と同じ。「もう触らないけど残す」という業務上の状態遷移）

途中で検討した `Open` / `Capture` / `Decide` などはニュアンスが合わずに却下:

- `Open`: issue は「open / closed」状態を持つモノなので自然だが、Todo は「開く」モノではない
- `Capture`: GTD 的な「思いつきを取り込む」意図、ユーザーの感覚（「決める / 残す」）と合わない
- `Decide`: 意味は近いが英語として `decide a todo` が硬く / 不自然

最終的に、ユーザーの「**曖昧なものをタスクとして切り出す**」という言語化から `Define`（曖昧 → 具体に形を与える）に着地した。

### 4. ユースケース粒度の判断基準

業務動詞ベースで命名するとき、粒度の判断は **業務的意図** で決める:

- **業務的意図が違う → 分ける**: 注文の `ChangeShippingAddress` vs `ChangeQuantity`（配送に影響 / 金額に影響、それぞれ別の不変条件）
- **業務的意図が同じ → まとめる**: GitHub issue の title + body は「内容を改訂する」共通意図 → `ReviseTodo` 1 つで十分
- **「フィールドごとに分ける」は過剰**。判断基準は業務的意図、フィールドではない

### 5. リポジトリは業務動詞を持つべきか

`repo.Complete(id)` のような業務動詞メソッドを持たせるのは **アンチパターン**:

- 業務ロジック（不変条件判定）がリポジトリに漏れる
- ドメインオブジェクトが薄くなり、**Anemic Domain Model（貧血ドメインモデル、Martin Fowler）** になる
- 責務の境界が壊れる（リポジトリ = 永続化、ドメイン = 業務ロジック、usecase = 組み立て）

正しい流れ: `repo.FindByID(id)` → `todo.Complete()` → `repo.Save(todo)`

### 6. Create / Update 分離 vs Save 1 つ

Vaughn Vernon が示す 2 スタイルのうち、**Persistence-Oriented Repository**（`Save` / `Remove` / `FindByID`）を採用:

- DDD 原典に忠実
- usecase 側のコードが簡潔（新規 / 既存の分岐不要、`define` も `revise` も `complete` も最後は `repo.Save(todo)` を呼ぶだけ）
- PostgreSQL の `INSERT ... ON CONFLICT DO UPDATE` で素直に実装可能
- **英語的にも整合**: Repository = 倉庫 / 保管庫のメタファー（語源: ラテン語 *repositorium* = 「物を置き戻す場所」）に対し、`Save`（ラテン語 *salvare* = "to keep safe"）が意味的に整合する。`Create` は「リポジトリが todo を作る」と読めて違和感

### 7. 状態を 2 値 vs 3 値（StartTodo の採否）

- 3 値（未着手 / 着手中 / 完了）にすると状態遷移サンプルは豊かになる
- しかし、このプロジェクトの目的はテンプレートとしての規範であり、Todo の細部は本質ではない
- ユーザーの業務感覚として「着手」を明示しないため、2 値のままが自然
- 既存の `Done()` + `ErrAlreadyCompleted` で、状態遷移の不変条件サンプルとしては十分

### 8. 読み取り系（Query）の業務動詞化

書き込み系（Command）を業務動詞化する一方で、読み取り系は業界慣例として CRUD 命名（`Get` / `List`）でも違和感が少ない:

- GitHub API: `list`, `get`（CRUD 的）
- Slack API: `conversations.list`, `users.info`（CRUD 的）
- DDD 原典でも、書き込み系は業務動詞化、読み取り系は `Find` / `List` が一般的（Vaughn Vernon IDDD）

ただし、本プロジェクトでは方針の一貫性のため **読み取り系にも業務動詞化を適用** する:

- `GetTodo` ではなく `ViewTodo`（閲覧する）— 1 件詳細閲覧
- ユーザーの業務的意図（「一覧見て、終わったかチェックする」）として「閲覧する」が自然な業務動詞であり、`Get`（取得する）よりも業務上の意図が表れる
- 一覧閲覧が必要になれば `ListTodos` または `ViewTodos`（複数形）を別ユースケースとして追加（業務的意図が違うため別ユースケース）

なお、`Confirm`（確認する）は英語的に「確定する / 認証する」のニュアンスが強く（例: confirm an order = 注文を確定する）、Todo の単純閲覧とは合わないため不採用。

## 決定

### ユースケース層

業務動詞で命名する（書き込み系・読み取り系ともに）:

- `DefineTodo`（定義する）
- `ReviseTodo`（内容を改訂する: title + body をまとめて更新）
- `CompleteTodo`（完了する）
- `ArchiveTodo`（アーカイブする → 実装は物理削除）
- `ViewTodo`（閲覧する: 1 件詳細閲覧。読み取り系の業務動詞化）

**ディレクトリ構造とパッケージ命名:**

**ユースケースの対象別に名前空間を切り**、その下にユースケースごとのパッケージを置く（イメージは **Slack API の名前空間構造**: `chat.postMessage`, `users.info`, `conversations.archive` のような `chat.*` / `users.*` / `conversations.*`。kgrzybek の `Application/<対象>/<ユースケース>/` を Go に翻訳した形でもある）:

```
internal/usecase/
  todo/                     ← ユースケースの対象別の名前空間
    define/                 ← パッケージ define
      usecase.go            ← define.UseCase 型、define.NewUseCase() 関数
    revise/
      usecase.go
    complete/
      usecase.go
    archive/
      usecase.go
    view/
      usecase.go
```

呼び出し側: `uc := define.NewUseCase(repo); result, err := uc.Execute(ctx, param)`

パッケージ名を `todo`（対象名）にしない理由: `domain/todo` との衝突回避 + ユースケース単位の境界を明示するため。

なお「ユースケースの対象」は多くの場合ドメインの集約名と一致するが、概念的には集約そのものではなく、**Slack API の `chat` / `users` のような「対象別の名前空間」** として捉える。これにより、複数集約をまたぐユースケースも「主たる対象」のディレクトリに置く判断が自然にできる。

### ドメインモデル

`Todo { id, title, body, completed }` の 4 フィールド:

- `body` フィールドを追加
- `completed bool` のまま（3 値拡張せず）

### リポジトリ interface

`Create / Update / FindByID` から **`Save` / `RemoveByID` / `FindByID`** に変更:

- `Save(ctx, todo) error` — 新規・既存両対応（Upsert）
- `RemoveByID(ctx, id) error` — 単純削除。取得 + 削除の 2 クエリを避ける
- `FindByID(ctx, id) (*Todo, error)`

architecture.md の基本 API は `Save / Remove / RemoveByID / FindByID` の 4 つで、デフォルトは `Remove(aggregate)`（集約引数の削除）、`RemoveByID` は性能最適化のための例外 API として位置付けられている。今回 `Remove(aggregate)` は採用しない。Todo の `ArchiveTodo` は業務ルール判定が構造的に存在しない単純削除で、architecture.md の使い分け指針が `RemoveByID` を許容する例外条件に該当するため。詳細な使い分け方針は architecture.md 参照。

リポジトリに業務動詞メソッド（`Complete`, `Archive` など）は持たせない。

## 帰結

### 利点

- DDD の本流に沿った設計（原典・著名 OSS で裏付け済み）
- ユースケース名が業務上の意図を直接表現する（ユビキタス言語の体現）
- リポジトリの責務が永続化に閉じ、業務ロジックがドメインオブジェクトに集約される
- usecase 層が簡潔（`Save` 統合で新規 / 既存の分岐不要）
- 「Repository = 倉庫」のメタファーが英語的にも整合（`Save / Remove`）

### 欠点・コスト

- 既存ユースケース（`create_todo`, `get_todo`）の業務動詞リネーム作業
- DB スキーマ変更（`body` カラム追加）
- リポジトリ実装変更（PostgreSQL の `INSERT ... ON CONFLICT DO UPDATE` を使う Upsert 実装）
- sqlc クエリの書き換え

### スコープ外（今回扱わない）

- `StartTodo`（着手中）状態 → 採用しない
- Issue tracker 的な担当者割り当て（assign）→ 採用しない
- **`Remove(aggregate)` メソッド** → 今回採用しない
  - 現状の `ArchiveTodo` は業務ルール判定が構造的に存在しない単純削除で、`RemoveByID` で足りる
  - 削除前の業務ルール判定 / ドメインイベント発行 / 不変条件チェックが必要になった時点で追加する（指針は architecture.md 参照）
- **並行更新の整合性（楽観ロック / `version` 列）** → 採用しない
  - `Save` Upsert は last-write-wins で動作する（並行更新は呼び出し側責務）
  - 単一クライアント前提のサンプル実装。Vernon IDDD の Persistence-Oriented Repository は通常 `version` を伴うが、テンプレートとしての規範性を高める段階で別 ADR で `version` 列の追加と楽観ロックを検討
- これらが必要になったら別 ADR で再評価

## 参考リンク

- Eric Evans, "Domain-Driven Design" (青本, 2003)
- Vaughn Vernon, "Implementing Domain-Driven Design" (赤本, 2013)
- [Greg Young, "Task-Based UI"](https://cqrs.wordpress.com/documents/task-based-ui/)
- [Udi Dahan, "From CRUD to Domain-Driven Fluency"](https://udidahan.com/2008/02/15/from-crud-to-domain-driven-fluency/)
- [Robert C. Martin, "The Clean Architecture"](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [ThreeDotsLabs/wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [VaughnVernon/IDDD_Samples](https://github.com/VaughnVernon/IDDD_Samples)
- [Mathias Verraes, "CRUD is an Anti-pattern"](https://verraes.net/2013/04/crud-is-an-anti-pattern/)
- [Slack API: chat methods](https://api.slack.com/methods?filter=chat)
- [Google AIP-136: Custom Methods](https://google.aip.dev/136)
