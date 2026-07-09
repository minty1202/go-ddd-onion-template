---
paths:
  - "internal/**/*.go"
---

# アーキテクチャルール（DDD + オニオンアーキテクチャ）

> **本ファイルの方針**: 実用指針を主体とするが、原典系（Vernon IDDD、Evans 青本など）との関係を読み取れるよう、必要に応じて出自注記（「原典準拠」「本プロジェクト固有」「原典から派生」など）を併記する。他のルールファイルとは異なり、本ファイルは出自注記を許容する。

## 用語の前提

**ドメインモデル**: 業務知識をコードで表現したもの。本プロジェクトでは以下の構成要素で表現する。

- **Entity（エンティティ）**: 同一性（ID）で区別されるオブジェクト。フィールドとして Value Object や他の Entity を持つ
- **Value Object（値オブジェクト）**: Entity が値として保持する、不変で等価性で比較されるオブジェクト
- **Aggregate Root（集約ルート）**: ルート Entity（必ず 1 つ）。外部からのエントリポイント
- **Aggregate（集約）**: Aggregate Root を頂点とするオブジェクトグラフ全体（配下 Entity と Value Object を含む）。一貫性境界
- **Domain Service（ドメインサービス）**: 単一の Entity / Value Object に属さない振る舞い
- **Domain Event（ドメインイベント）**: 業務上意味のある出来事
- **Factory（ファクトリ）**: 生成が複雑な場合の生成役
- **Repository（リポジトリ、interface）**: 集約の保管場所の抽象

**Bounded Context（境界づけられたコンテキスト、BC）**: ドメインモデルが一貫して意味を持つ境界。1 つの BC 内では用語と概念がブレずに通用する。

本プロジェクトでは **1 サービス = 1 BC** を前提とする。**以降のアーキテクチャルールは、この前提の上で書かれている**。複数 BC を扱う必要が出た場合は、本ルール自体を見直す。

## ディレクトリ構造

本ルールは `internal/` 配下のレイヤー設計に適用する。`cmd/main.go` は依存の組み立てとアプリケーション起動を担う場所であり、本ルールの対象外。

```
cmd/
  main.go                         ← 薄いエントリポイント
internal/
  domain/                         ← ドメイン層（核、誰にも依存しない）
    todo/
    user/
    schedule/
  usecase/                        ← ユースケース層（Application 層相当）
    todo/                         ← ユースケースの対象別の名前空間（Slack API の chat.* / users.* と同じイメージ）
      define/                     ← 1 ユースケース 1 パッケージ（業務動詞で命名）
        usecase.go
      revise/
        usecase.go
      complete/
        usecase.go
      archive/
        usecase.go
      view/
        usecase.go
    ...
  infra/                          ← インフラ層
    repository/
      todorepo/                   ← Todo の Repository 実装（package todorepo）
      userrepo/                   ← package userrepo
    db/                           ← DB 接続プール
  presentation/                   ← プレゼンテーション層
    todorpc/                      ← Todo の RPC handler（package todorpc）
    ...
```

### Application 層の名称

DDD / オニオンの一般的呼称は「Application 層」だが、本プロジェクトでは既存コードに合わせて **`usecase/`** で統一する。意味は同じ。

## 各層の責務

### domain 層

#### 指針

- ドメインモデルの知識を対応するオブジェクトに書く
- 常に正しいインスタンスしか存在させない

#### Go での体現

- パッケージ名 = ドメインモデル名（`internal/domain/todo/` → `package todo`）
- 型は `todo.Todo`, `todo.ID`, `todo.Status` のように外から参照する
- **`todo.Todo` 型は Aggregate Root を指す。Aggregate（一貫性境界の中身全体）に対応する独立した型は作らず、`todo.Todo` インスタンス + 配下に持つ Entity / Value Object のグラフが概念上の Aggregate になる**（パッケージ構成は本層末尾「ドメイン層は `internal/domain/` 以下に...」の段落を参照）
- コンストラクタ名: パッケージ同名型は `New`、それ以外は `New<型名>`（例: `todo.New` → `todo.Todo`、`todo.NewID` → `todo.ID`）
- 非公開フィールド + getter で外から壊せないようにする
- コンストラクタでバリデーションし、不正なオブジェクトを作れないようにする（**Always Valid 原則**、Vernon IDDD 由来）
- Always Valid 原則の **本プロジェクトでの解釈・運用**:
  - 注: **以下の二分法・復元方針は本プロジェクト独自の運用ルール**。Vernon IDDD では「Always Valid（不正な状態でインスタンスを存在させない）」自体は確立した原則だが、**過去データの validation を集約ごとに「緩い / 厳格」で選ぶ派生解釈は本プロジェクト固有**
  - 解釈の出発点: 「常に正しい」とは「過去のルールで正しかったものを解釈に含めるか」が派生問題になる。集約によって「過去データ保持を優先したい」場合と「業務不正排除を優先したい」場合があるため、集約ごとに方針を選ぶ二分法を採る
  - **構造的不正と業務的不正で分けて適用する**:
    - 構造的不正（ID 空、必須フィールド欠損等）→ 全てのコンストラクタで常に弾く
    - 業務的不正（文字数・金額レンジ等の業務ルール違反）→ 通常コンストラクタは常に弾く（最新ルール準拠）。復元コンストラクタは過去のルールで作られたデータがあるため、集約ごとに方針を選ぶ（下記参照）
- 永続化からの復元コンストラクタ（例: `Reconstruct`）は、集約ごとに以下の 2 つから方針を選ぶ:
  - 背景: 業務ルールは時間とともに変わり得る。一律に現行ルールで全データを検証すると過去データが読み込めなくなる（例: 本のタイトル制限を 100 → 50 文字に変えると、過去の 80 文字の本が読めなくなる）。一方、業務不正データの存在自体を許せない集約もある（例: 計算ミスのある決済記録を残してはいけない）。**集約ごとに優先度（過去データ保持 vs 業務不正排除）を選ぶ**
  - **緩い側**: 構造的チェック（ID 空、必須欠損等）のみ。業務ルールは通す。過去データを読み込み続けられる
  - **厳格側**: 業務ルールも検証する。常に最新ルール準拠の状態だけ許す。ルール変更時はデータ移行が必須
  - 目安:
    - 厳格: 歴史的に不正データが存在することを**許容できない**集約。例: 決済金額、在庫数、ポイント残高（金銭・数量に関わるもの）
    - 緩い: 許容**できる**集約。例: 本のタイトル、ブログのタグ、ユーザーの自己紹介（表示用テキスト）
- 厳格側を採る場合の実装方針:
  - **基本（データ移行）**: 過去データを現行ルールに合わせて書き換える
    - 例: タイトル文字数制限を「100 文字以下」→「50 文字以下」に変えるとき、過去の長いタイトルを切り詰める migration を流す
  - **例外（型を分ける）**: 履歴・監査ログなど書き換えるべきでないデータは、現行用と過去用の型を別にする
    - 例: 経理の帳簿、医療カルテ、取引履歴。現行用の `Transaction`（編集可、現行ルール準拠）と過去用の `HistoricalTransaction`（読み取り専用、当時の状態）を別の型として扱う
- 状態変更メソッドにドメインルールを持たせる（例: 完了にできる、既に完了済みならエラー）
- **1 集約 = 1 Repository**（Vernon IDDD Ch.12「There is a one-to-one relationship between an Aggregate type and a Repository」）。集約を跨いだ操作は Repository には持たせず、別途ドメインサービスやアプリケーションサービスで扱う
- Repository は interface だけこの層で定義し、実装は infra 層
- Repository の API 設計:
  - メタファー: 集約の **保管場所（倉庫 / 保管庫）**（Eric Evans 青本）。メモリ上のコレクションのように振る舞い、内部の永続化手段（DB / メモリ / ファイル）は隠蔽する
  - 基本 API は **`Add` / `Save` / `RemoveByID` / `FindByID`**（Vernon IDDD の Collection-Oriented と Persistence-Oriented を組み合わせた本プロジェクト独自の運用。IDDD に直接の前例はなく、`Add` で挿入、`Save` で更新を明示的に分ける）:
    - `Add(ctx, aggregate) error` — 新規挿入。DB 実装は INSERT
    - `Save(ctx, aggregate) error` — 既存更新（楽観ロック付き、`WHERE id = ? AND version = ?` + `SET version = version + 1`、影響行数 0 で `ErrConflict`）
    - `RemoveByID(ctx, id) error` — 単純削除
    - `FindByID(ctx, id) (*Aggregate, error)` — 取得（version 含む）
  - **`Add` / `Save` の使い分け:**
    - **新規生成**: usecase が `New<集約>(...)` で集約を作って `repo.Add(aggregate)` を呼ぶ
    - **既存更新**: usecase が `repo.FindByID(id)` で取得 → ドメインメソッドで状態変更 → `repo.Save(aggregate)`
    - usecase 側が「新規 vs 既存」を明示的に呼び分ける（Repository 内部に判別ロジックを持たせない）
  - 採用しない API:
    - **`Remove(aggregate)`**: 現状の単純削除では `RemoveByID` で足りる
    - **`Save` 1 つでの Upsert（Persistence-Oriented）**
  - `Create` / `Update` のような CRUD 分離も採らない（英語的に `Create(aggregate)` は「リポジトリが集約を作る」と読めて違和感、DDD 原典のコレクションメタファーから外れる。`Add` を使う）
  - **業務動詞メソッド（`repo.Complete(id)`, `repo.Archive(id)` 等）は持たせない**:
    - 業務ロジック（不変条件判定）がリポジトリに漏れ、ドメインオブジェクトが薄くなる（**Anemic Domain Model / 貧血ドメインモデル**。Vaughn Vernon IDDD 第 12 章 Repositories でも、Repository に業務メソッドを持たせず Application Service + ドメインオブジェクトに業務ロジックを閉じる方針が示される）
    - 正しい流れ: `repo.FindByID(id)` → `aggregate.Complete()` → `repo.Save(aggregate)`
    - 業務動詞はドメインオブジェクトのメソッド + ユースケースに閉じる
  - 読み取り系（Query）の命名は、書き込み系の業務動詞化と異なり **`Find...` / `List...` の汎用語** で命名する:
    - 1 件取得: `FindByID(ctx, id)`、特定条件で 1 件は `FindBy<条件>(ctx, ...)`
    - 複数件: `List(ctx, ...)` または `ListBy<条件>(ctx, ...)`
    - これは Vernon IDDD でも認められた非対称扱い（書き込み系 = 業務動詞、読み取り系 = CRUD 寄りの汎用語）。Slack API / GitHub API も同じ流儀（`conversations.list`, `users.info` 等）
- 楽観ロック（Optimistic Concurrency、Vernon IDDD ch.10）:
  - すべての集約は楽観ロック対応とする
  - 集約は `version int` を **private** で持つ
  - 共通化: `internal/domain/aggregate/` パッケージで `Aggregate` interface と `Lock` struct を提供
    - `Aggregate` interface: `Version() int`, `SyncVersion(v int)` を要求
    - `Lock` struct: `version int` を保持、`Aggregate` を満たす実装を提供
  - 集約の保持方法:
    - **private 名前付きフィールド `lock aggregate.Lock`** で `Lock` を保持（anonymous embed ではなくカプセル化を強くする、IDDD_Samples の `protected` 相当）
    - 集約側で `Version()` / `SyncVersion(v int)` を手動公開（パッケージ外の Repository 実装が呼ぶため）
  - infra 層に共通ヘルパー: `internal/infra/repository/optimistic.go` の `SaveWithLock(ctx, agg aggregate.Aggregate, persistFn func(ctx, expectedVersion int) (newVersion int, err error)) error`
    - `aggregate.Aggregate` の interface 引数で「集約は `Aggregate` を満たすこと」をコンパイル時に強制
    - 各 Repository 実装の `Save` は `SaveWithLock` を経由する
    - Save 成功時、副作用で `agg.SyncVersion(newVersion)` を呼ぶ（DB から `RETURNING version` で取得した最新 version をメモリの集約に反映する）
    - 影響行数 0 で `ErrConflict`（infra 層エラー）を返す
  - usecase 層は `ErrConflict` を `errdefs.NewAborted` にマッピング（エラー設計は `errors.md` 参照）
  - **`SyncVersion(v int)` と `Reconstruct(...)` / `ReconstructLock(...)` は Repository 実装のみが呼ぶ**。usecase / presentation 層から呼ぶと楽観ロックの不変条件破壊や業務ルール検証バイパスが起きるため禁止。Go の言語機能では強制できないため `forbidigo` linter で構造的に検出する（`.golangci.yml` に設定）。テストファイルは内部状態の組み立てが必要なため lint 除外
  - 採用しない: 悲観ロック（`SELECT FOR UPDATE`）
- ドメイン層は原則として標準ライブラリのみで実装する。一部のライブラリは例外として許容するが、新規追加は都度議論し、許容理由 + 適用範囲 + 適用外を下記に残す（抽象基準は設けない）
- 許容ライブラリ:
  - `go-playground/validator` — 値オブジェクトのコンストラクタでの構造的バリデーション用途
    - **適用範囲**: 文字数 / 必須 / 数値範囲 / 列挙値など、**フィールド単位の単純な構造的制約**のみ
    - **適用外**: 状態遷移ルール（例: 「完了済みなら再完了不可」）、複数フィールドにまたがる業務不変条件、集約間の整合制約 — これらは集約のメソッド（例: `Todo.Complete()`）に Go コードとして書く
    - **理由**: 業務不変条件をメソッドで表現することで業務知識がコードから読める（Vernon IDDD の Always Valid 原則と整合）。一方で「文字数」「必須」のような構造制約は宣言的に書いた方が読みやすく、Java の Bean Validation（JSR-380）/ C# の DataAnnotations が同等の流儀を採っている。「業務 = コード、構造 = タグ」の境界を引くことで両者の利点を取る
- タイムスタンプは扱いを 2 つに分ける:
  - 業務上の意味があるもの（例: 注文日時 `PlacedAt`、期限 `DueAt`）→ ドメインに持つ。業務上の意味がわかる名前を付ける
  - 行の生成・更新時刻（`created_at` / `updated_at`）→ DB の補助 metadata。ドメインに持たせない

現在の定義では、ドメイン層は `internal/domain/` 以下に **ドメインモデル単位**でディレクトリを切る。各ディレクトリ配下には Aggregate Root と Value Object を、サブディレクトリを切らずに同じ階層に配置する（kgrzybek、IDDD など DDD 主流の流儀）。

Value Object や Domain Service が多数になった場合は、`internal/domain/value/`、`internal/domain/service/` のように **DDD パターン別のサブディレクトリで分ける選択肢もある**。判断は規模・チームの慣習による。

### usecase 層

#### 指針

- ドメインオブジェクトの生成や、状態の変更、リポジトリを使用した永続化を行う
- ドメインオブジェクトをユースケースから外に漏らさない

#### 命名と粒度（業務動詞ベース）

- ユースケース名は **業務動詞**（業務上の出来事 / 状態遷移を表す語）で命名する。「ユースケース = 利用ケース、業務上の使われ方」という語の意味と整合させる
- CRUD（Create / Update / Delete / Get）は **永続化レイヤーの語彙**。ユースケース層の語彙ではない
  - NG: `CreateOrder`, `UpdateOrder`, `DeleteOrder`（データ操作の語彙）
  - OK: `PlaceOrder`（注文を確定する）、`ShipOrder`（発送する）、`CancelOrder`（キャンセルする）、`ApplyCoupon`（クーポンを適用する）（業務上の行為）
- ユースケースの粒度は **業務的意図** で判断する:
  - 業務的意図が違う → 別ユースケースに分ける（例: 注文の `ChangeShippingAddress` と `ChangeQuantity` は配送 / 金額に影響、別の不変条件）
  - 業務的意図が同じ → 1 ユースケースにまとめる（例: GitHub issue の title + body は「内容を改訂する」共通意図）
  - 「フィールドごとに分ける」は過剰。判断基準はフィールドではなく業務的意図
- 業務動詞が浮かびづらいときの探し方:
  - 抽象的なドメインだと CRUD の引力に負ける → 具体化（業務イメージのある近接ドメインを借りる、例: スケジュール / GitHub issue / Slack API）すると見える
  - 「データの操作」ではなく「業務上の出来事 / 状態遷移」を書き出す
- **読み取り系（Query）も書き込み系と同じく業務動詞で命名する**（例: `ViewTodo`）。リポジトリ層の `FindByID` / `List...` の汎用語命名とは層が違うため非対称（リポジトリ = 永続化の関心、ユースケース = 業務操作の関心）

#### Go での体現

- **ユースケースの対象別に名前空間を切り、その下にユースケースごとのパッケージを置く**（`internal/usecase/<対象>/<ユースケース動詞>/usecase.go`）
  - 例: `internal/usecase/todo/define/usecase.go`
  - イメージは **Slack API の名前空間構造**（`chat.postMessage`, `users.info`, `conversations.archive` のような `chat.*` / `users.*` / `conversations.*`）
  - kgrzybek/modular-monolith-with-ddd の `Application/<対象>/<ユースケース>/` 二段階階層を Go に翻訳した形でもある
- パッケージ名はユースケース動詞（例: `define`）、外からは `define.UseCase` / `define.NewUseCase()` で参照する
  - パッケージ名を `<対象>`（例: `todo`）にしない理由: `domain/todo` との衝突回避 + ユースケース単位の境界を明示するため
- 例外として、関連する複数のユースケースで共通型が必要なら、大枠のパッケージ（対象ディレクトリ直下）に置く
- 入力型は `Param`、戻り値型は `Result`（各ユースケースパッケージ内で独立定義）
- 同じドメインから生まれた情報でも、ユースケースごとに意味（セマンティック）が違うため、`Result` は別物として扱う（構造が同じに見えても共通化しない）
  - 原典系（Vernon IDDD、kgrzybek 等）も基本はユースケースごとの DTO。Java / C# は MapStruct / AutoMapper / LINQ projection など型変換を支援する道具が豊富で派生型を作るのが容易だが、Go はこれらの道具に乏しく、struct embedding の「全部入り」になりがち。本プロジェクトでは Go の制約から「常に別物」を強い制限として運用する
- 「Todo を返している」のではなく「そのユースケースの結果を返している」という意識で設計する
- メソッド名は `Execute` で統一
  - パッケージ名（例: `define`）でユースケース名は既に伝わっているため、メソッド名は統一語で十分（`define.UseCase.Execute()` で「define を実行する」と読める）。Effective Go の「単一メソッド interface はメソッド名 + er」は interface 命名の慣習で、struct ベースのユースケースには適用しない
- ドメインオブジェクトをユースケースから外に漏らさない（`Result` に詰め替える）
- ドメインエラー・インフラエラーは `UseCaseError` に変換して返す（エラー設計は `errors.md` を参照）

### infra 層 ここはまだ未確定

- DB、外部 API、メッセージキュー等の**具体的な実装**を持つ
- domain 層が定義した interface（Repository 等）を実装する
- 外部ライブラリへの依存はこの層に閉じ込める

### presentation 層 ここはまだ未確定

- HTTP ハンドラ、gRPC サービス、CLI コマンドなど、アプリを外に公開する層
- 責務:
  - リクエスト DTO のデコード
  - 入力の形式バリデーション
    - スコープ: 必須項目の有無、型、文字数、数値範囲、列挙値など、リクエスト構造上の制約
    - 業務不変条件（業務ルール違反）のチェックはここでは行わない。ドメイン層の責務
  - usecase 呼び出し
  - ドメインオブジェクト → レスポンス DTO 変換
  - エラー → HTTP ステータス変換
- ビジネスロジック・DB アクセスは持たない
