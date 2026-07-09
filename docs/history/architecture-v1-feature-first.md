> ⚠️ このファイルは 2026-04-17 時点の古いアーキテクチャ定義（feature-first）のバックアップです。
> 同日に再議論して **B-1（純 layer-first + ドメイン層内 aggregate 分割）** に変更しました。
> 経緯は `2026-04-17-architecture-revision.md`、最新定義は `.claude/rules/go/architecture.md` を参照してください。

---

# アーキテクチャルール（DDD + オニオンアーキテクチャ）

## 全体方針

このプロジェクトは **feature-first（機能優先）構造** を採用する。各 feature の内部にオニオン層を持つ。

### 根拠

- Go 公式の立場（Russ Cox が `golang-standards/project-layout` を Go 標準ではないと明言）と、大規模実運用プロジェクト（Terraform, Kubernetes, Vault, CockroachDB, MinIO, Grafana, Consul, Caddy）がほぼ全て feature-first を採用している事実に基づく
- Martin Fowler "PresentationDomainDataLayering" が、レイヤーが大きくなったらトップレベルをドメイン指向モジュールで分割し、その内部を層化せよと推奨している
- Dave Cheney が util / common / helpers のような層・汎用名を避け、behavior（振る舞い）で切れと主張している
- Uncle Bob の Screaming Architecture（トップレベルは業務ドメインを叫ぶべし）とも整合する

## ディレクトリ構造

```
cmd/
  main.go                         ← 薄いエントリポイント
internal/
  todo/                           ← feature（BC ではなく機能単位）
    domain/                       ← package domain
    usecase/                      ← package usecase
    infrastructure/               ← package infrastructure
    presentation/                 ← package presentation（フラット、YAGNI）
    todo.go                       ← package todo、組み立て関数 New() がここ
  user/                           ← 別 feature（同じ内部構造）
    ...
  http/                           ← HTTP サーバー、ルーティング、ミドルウェア
  db/                             ← DB 接続プール
  config/                         ← 設定読み込み
  logger/                         ← ロガーのセットアップ
  errdefs/                        ← 横断的なエラー interface のみ定義
```

### 構造の原則

- **`internal/` 直下はフラット**。business feature と横断的関心事（http, db, config, logger, errdefs 等）を並列配置する。`platform/`, `foundation/` のようなラッパーは作らない（Go 慣習に従う）
- 各 feature は内部に **domain / usecase / infrastructure / presentation の 4 層**を持つ
- `presentation/` は**フラット**（YAGNI）。HTTP 以外（gRPC, CLI 等）が必要になったらそのとき分割する
- feature の起動は **feature ルート直下のファイル**（例: `internal/todo/todo.go`）の組み立て関数 `New()` が担う

## 各層の責務

### domain 層

#### 指針

- ドメインモデルの知識を対応するオブジェクトに書く
- 常に正しいインスタンスしか存在させない

#### Go での体現

- 非公開フィールド + getter で外から壊せないようにする
- コンストラクタでバリデーションし、不正なオブジェクトを作れないようにする
- 状態変更メソッドにドメインルールを持たせる（例: 完了にできる、既に完了済みならエラー）
- **ID は `ulid.ULID` を直接使用**する。feature ごとに typed ID（`type TodoID ulid.ULID` 等）のラッパーは作らない
- **feature 固有の具体エラー型**は domain 層に置き、`internal/errdefs/` の interface を実装する

### usecase 層

#### 指針

- ドメインオブジェクトの生成や、状態の変更、リポジトリを使用した永続化を行う
- ドメインオブジェクトをユースケースから外に漏らさない

#### Go での体現

- 1 ユースケース 1 構造体
- 戻り値は専用の戻り値型に詰めて返す
- メソッド名は `Execute` で統一
- ドメインエラーを UseCaseError に変換して返す
- インフラエラーは ServerError に変換して返す

### infrastructure 層

- DB、外部 API、メッセージキュー等の**具体的な実装**を持つ
- domain 層が定義した interface（リポジトリ等）を実装する
- 外部ライブラリへの依存はこの層に閉じ込める

### presentation 層

- HTTP ハンドラ、gRPC サービス、CLI コマンドなど、feature を外に公開する層
- 責務：
  - リクエスト DTO のデコード
  - 入力バリデーション
  - usecase 呼び出し
  - ドメインオブジェクト → レスポンス DTO 変換
  - エラー → HTTP ステータス変換
- ビジネスロジック・DB アクセスは持たない

## 横断的関心事（cross-cutting concerns）

`internal/` 直下に置かれる feature 以外のパッケージ。層の内側・外側という概念を超えて、全 feature から参照される。

### errdefs（エラー interface のみの契約パッケージ）

- **interface だけを定義**する（例: `ErrNotFound`, `ErrConflict`, `ErrInvalidParameter`）
- Moby (Docker) の `errdefs` パッケージがこのパターンの代表例
- **具体エラー型は各 feature の domain で定義**し、このパッケージの interface を実装する
- 上位層は interface 型アサーション（`errors.As`）で分類する
- 層（domain / usecase）で errdefs を分割しない。カテゴリ（NotFound, Conflict 等）で分類する

### http, db, config, logger

- アプリ起動時に初期化される、外部リソース（ネットワーク / DB / 環境変数 / 出力先）を包むシングルトン的サービス
- business feature からは具体実装を意識せず、必要な interface だけを知る
- `cmd/main.go` がこれらを組み立てて feature に渡す

## まだ決まっていないこと

### 1. `apperr` の扱い

現プロジェクトには `internal/apperr/violation.go` が存在する。errdefs と役割が被る可能性があるため、以下を要判断：

- apperr が何を担っているか（実装を確認）
- errdefs と統合するか、別の役割として残すか
- 残す場合、横断的関心事として `internal/apperr/` のままで位置づけが適切か

### 2. feature 間の依存ルール

feature A が feature B を直接 import していいかの方針が未確定。以下のような選択肢がある：

- 直接 import を許容する（簡素だが結合が強まる、循環のリスク）
- 禁止する（疎結合だが横断処理のときの取り回しを別途設計する必要あり）
- 許容するが方向性を固定する（例: 上位レイヤー的な feature から下位参照のみ許す）

依存方向を守らせるツール（`depguard` 等）で静的検証する運用も合わせて検討。

### 3. 横断処理（orchestration）の置き場

独立した複数 feature が連携する処理（例: User 削除時に関連 Todo を連動削除）の置き場が未確定。
選択肢：

- 専用層を作る（例: `internal/app/` に横断ユースケースを集約）
- イベント駆動にする（各 feature がイベント発行・購読）
- 発生源 feature の usecase に乗せる（結合強め）

**現時点では発生していない問題**のため、将来必要になったタイミングで判断する。
