# ADR-0005: presentation 層の設計方針（gRPC + Slack API スタイル）

## ステータス

採択（2026-04-30）

## コンテキスト

ADR-0002 / ADR-0004 で usecase 命名の業務動詞化と Collection-Oriented Repository + 楽観ロックを採択した。これに対応する presentation 層（proto / Connect RPC / ハンドラ）の設計方針が未確定だった。

メモリ「Pending: proto granularity」で、proto / presentation の粒度を含む論点を再検討する必要があると記録されていた。本 ADR でこれを解消する。

特に未決だった論点:

- proto の service / メソッド設計（Google AIP のリソース指向 vs Slack API スタイル）
- 楽観ロックトークンの API 表現
- protovalidate と domain 定数の drift 管理（ADR-0003 で「presentation 着手時に決める」と保留されていた）
- ハンドラ Go コードの構造

## 検討と判断

### 1. 設計思想: Slack API スタイル vs Google AIP

Google AIP（API Improvement Proposals、aip.dev）は API 設計の指針集。標準メソッド優先（List / Get / Create / Update / Delete を基本に置き、Custom Method は例外として AIP-136 で許容）が思想の柱。

ADR-0002 で usecase 命名を業務動詞化（`ViewTodo` まで含めて非 CRUD 化）した時点で、AIP の標準メソッド優先思想からは離れている。proto 側だけ AIP に揃えると思想の不整合が生じる。

Slack API は標準 / カスタムの区別なく、最初から全メソッドが業務動詞（`chat.postMessage`, `conversations.archive`）で、ADR-0002 の usecase 構造とそのまま整合する。

判断: **Slack API スタイルを採用。AIP は参考リファレンスとして局所論点で借りる。**

AIP から借りる対象: エラーモデル（`google.rpc.Status` / Code、既に errdefs で gRPC Code 準拠）、AIP-154 etag、AIP-158 ページング、AIP-160 フィルタなど、必要に応じて。

### 2. service の粒度

候補:

- per-action service: 業務動詞ごとに別 service（`service DefineTodoService { rpc Execute }`）
- 1 集約 = 1 service: `service TodoService { rpc DefineTodo, rpc ViewTodo, ... }`

per-action service は AIP からも Slack API（namespace 単位 = 1 service）からも外れる。Connect の生成 interface も 1 service = 1 interface 前提。

判断: **1 集約 = 1 service。`TodoService` 1 つに 5 メソッドを集約する。**

### 3. proto ファイルレイアウト

候補:

- 1 ファイル集約（service + 全 message を 1 .proto に）
- アクション別にファイル分割

Buf / gRPC コミュニティの一般慣習は 1 ファイル集約。生成される Go パッケージはどちらの分け方でも同じ（同一 proto package = 同一 Go package）なので、Go から見た構造は変わらない。

判断: **デフォルトは 1 ファイル集約。1 ファイルが肥大化した時点でアクション別に分割する。**

Todo の場合は 5 RPC 規模なので 1 ファイルで十分。

### 4. 楽観ロックトークンの API 表現

楽観ロックは「クライアントが取得時のトークンを更新リクエストで返す」仕組みで、API としてトークンを expose する必要がある（隠せない）。

候補:

- (a) version 数値: 内部 version カラムをそのまま expose
- (b) etag 文字列: 不透明トークンとして expose（HTTP RFC 7232 / AIP-154 流儀）

機能としては等価。差は API の長期的な柔軟性: (a) は内部のバージョニング方式（連番）を API に固定し、後から hash / timestamp に変えると破壊的変更になる。(b) は内部方式が見えないので変更に強い。

業界事例として GitHub / AWS / Google Cloud API は etag を採用。

判断: **etag 文字列を採用する。**

### 5. etag の配置

候補:

- (i) リソース型（`Todo` message）にフィールドとして持つ（AIP-154 流儀）
- (ii) Request / Response の最上位に etag を置く

(ii) はリスト系 API（`ListTodos` 等）で破綻する。最上位の etag は 1 つしか置けず、各リソースに紐付かないため、結局リソース型に持たせる必要が出る。

判断: **(i) リソース型 `Todo` のフィールドとして配置する。**

### 6. Request / Response の中身

候補:

- A: 各 RPC が必要な最小フィールドだけを持つ
- B: 共通基底メッセージ（全フィールド optional）を全 RPC で使い回す

B は RPC コントラクトが proto から読み切れず、AIP / Google Cloud API / gRPC OSS でも採用例がほぼない。

判断: **A 採用。各 Request / Response はその RPC が必要とする最小フィールドのみを持つ。**

usecase 層が per-action で別々の Param / Result を持つ思想と整合する。

### 7. protovalidate と domain 定数の drift 管理

ADR-0003 で proto-validate を境界バリデーションとして採用済み。proto 側に書く数値（`min_len: 3` 等）と domain 定数（`TitleMinLen` 等）の drift をどう担保するかが残論点だった。

候補:

- (a) ハードコード + CI テストで一致を検証
- (b) コード生成（domain 定数から proto 数値を生成）
- (c) 受容（drift 検証なし）

(b) はツールチェーンが複雑化し、見返りが小さい。(c) は ADR-0003 の精神（drift リスクを担保する余地）と相性が悪い。

判断: **(a) ハードコード + CI テストで domain 定数との一致を検証する。**

CI テストの具体（テストファイルの場所、検証方法）は実装時に決める。

### 8. ハンドラ Go struct の単位

Connect の生成 interface は 1 service = 1 interface（`TodoServiceHandler`）を出す。実装は 1 struct で受ける形が Connect 公式パターン。

判断: **1 service = 1 struct。`todorpc.Handler` のような形でパッケージ修飾されて参照される（型名にドメイン名を含めず stutter を避ける）。**

### 9. presentation 層のディレクトリ構造

判断: **`internal/presentation/<対象><プロトコル>/` に handler / mapper / error を配置する。**

```
internal/presentation/
└── todorpc/                       ← 対象 (todo) × プロトコル (rpc) の合成命名
    ├── handler.go   ← Connect の TodoServiceHandler interface 実装
    ├── mapper.go    ← proto Request/Response ↔ usecase Param/Result の変換
    └── error.go     ← usecase errdefs → connect.Error の変換
```

ディレクトリ名 + パッケージ名を `todorpc` のように **対象 × プロトコル**で合成する理由:

- ドメインパッケージ名（`todo`）は domain 層だけが名乗る方針。他層で同名パッケージを作ると import alias 必須 + shadowing リスクが増える（`.claude/rules/go/naming.md`）
- 将来 REST など別プロトコルが必要になったとき、`todorest` を**並列に追加するだけ**で対応できる。プロトコル別に依存（Connect 系 / `net/http` 系）が分離する
- 役割軸（`todohandler` / `todoapi`）にすると複数プロトコル対応時に命名が破綻するため不採用

1 ファイルが肥大化した時点で `handler.go` をアクション別に分割する余地は残す（proto ファイル分割の方針と同じ）。

## 決定（まとめ）

### proto 層

- 設計思想: Slack API スタイル（業務動詞 + 1 service 集約）。AIP は参考リファレンスとして局所論点で借りる
- service: 1 集約 = 1 service（`TodoService`）。per-action service は採用しない
- ファイル: デフォルト 1 ファイル集約、肥大化時にアクション別分割
- 楽観ロックトークン: etag 文字列、リソース型 `Todo` のフィールドとして配置
- Request / Response: 各 RPC が最小フィールドのみ持つ（共通基底メッセージは使わない）
- protovalidate: 境界バリデーションとして採用、proto 側に数値ハードコード、domain 定数との drift は CI テストで担保

### Go presentation 層

- ハンドラ struct: 1 service = 1 struct（パッケージ内では `Handler` で命名、外部から見ると `todorpc.Handler` のように呼ばれる。stutter を避けるため型名にドメイン名を含めない）
- ディレクトリ: `internal/presentation/<対象><プロトコル>/{handler,mapper,error}.go`（例: `internal/presentation/todorpc/`）
- 責務分担:
  - handler — Connect が生成する interface を実装
  - mapper — proto ↔ usecase Param/Result の変換
  - error — usecase errdefs → connect.Error の変換

## 帰結

### 利点

- ADR-0002 / ADR-0004 の usecase 設計と思想が一貫する（業務動詞ベース、Aggregate 単位）
- 業界標準の流儀（Slack API、AIP-154 etag、Connect 公式パターン）と整合
- protovalidate でフォーマット validation を境界に閉じ、ドメインを Always Valid の最後の砦にできる
- リソース型に etag を持たせることでリスト系 API にも自然に拡張可能

### 欠点・コスト

- AIP 全面準拠ではないため、Google Cloud API などの完全な慣習互換性はない
- proto 側の数値はハードコードのため、domain 定数との drift を CI テストで検出する追加実装が必要
- usecase Result が現状 version を含まない設計のため、etag マッピングのために Result に `Version` を追加する修正が必要（既存 define / view package を含む）

### スコープ外（今回扱わない）

- ページング（AIP-158）、フィルタ（AIP-160） — 必要になった時点で AIP を局所参照
- ストリーミング RPC — 採用予定なし
- 認証 / 認可 — 別 ADR
- ハンドラ層の interceptor / middleware の構成（protovalidate interceptor 以外） — 必要時に検討

## 参考リンク

- [Slack API: chat methods](https://api.slack.com/methods?filter=chat)
- [Google AIP-121: Resource-oriented design](https://google.aip.dev/121)
- [Google AIP-136: Custom methods](https://google.aip.dev/136)
- [Google AIP-154: Resource freshness validation](https://google.aip.dev/154)
- [HTTP RFC 7232: Conditional Requests (etag)](https://datatracker.ietf.org/doc/html/rfc7232)
- [protovalidate (Buf 公式)](https://buf.build/bufbuild/protovalidate)
- [Connect RPC](https://connectrpc.com/)
- ADR-0002: ユースケース命名の業務動詞化とリポジトリ設計の再評価
- ADR-0003: validation の責務分担とドメインルール定数の export
- ADR-0004: Collection-Oriented Repository への切替と楽観ロック採用
