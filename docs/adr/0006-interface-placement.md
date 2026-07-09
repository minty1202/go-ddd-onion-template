# ADR-0006: interface の配置方針

## ステータス

採択（2026-05-01）

## コンテキスト

ADR-0005 で presentation 層の設計方針が確定し、handler が usecase を呼ぶ構造が必要になった。これに伴い、層境界における interface の扱いを横断的に整理する必要が出た。

特に未整理だった点:

- Go の慣習（"Accept interfaces, return structs"、消費側で interface 定義）と DDD 原典（Vernon IDDD、Java スタイル：提供側で interface 定義）の関係
- domain 層の `todo.Repository` が domain 側にある理由と、他の層境界（presentation ↔ usecase など）に同じ思想を持ち込むべきかの判断軸
- Rust の trait（名目的型付け、提供側で定義）と Go の interface（構造的型付け、消費側で定義可能）の性質差

本リポジトリは新規 Go プロジェクトの土台として使う前提のため、interface の配置方針を一度確定し、今後の層境界設計で迷わないようにする。

## 検討と判断

### 1. Go の interface の性質（前提整理）

Go の interface は **構造的型付け（structural typing / duck typing）**。型のメソッドセットが interface の要求と一致すれば、明示的な宣言なしに自動的に satisfy する。提供側の型は interface の存在を知らなくてよい。

これは Rust の trait（名目的型付け、`impl Trait for Type` の宣言が必須）とは性質が異なる。Rust では trait は提供側に置かれる必然性があるが、Go では消費側で interface を定義しても提供側は影響を受けない。

### 2. Go の慣習: "Accept interfaces, return structs"

Go コミュニティの基本指針（Rob Pike, Go Code Review Comments）:

- 関数 / handler は **interface を受け取る**（消費側で interface 定義）
- 関数 / handler は **struct を返す**（提供側は具象を返す）
- interface は「使う側の都合」で定義し、「実装側」は関与しない

標準ライブラリも消費側 interface の例が多い（`io.Reader` は呼び出す側が定義し、`bytes.Buffer` などはそれを「知らずに」実装する形）。

### 3. 慣習からの逸脱が許される場面

Go の慣習は絶対ではない。以下のケースでは提供側で interface を定義することが正当化される:

- **依存性逆転（DIP）が必要なシーム**: Repository / Port-Adapter のように、自然な依存方向を逆転させる必要がある場合。内側（消費される側）で interface を定義する
- **ライブラリ / フレームワークが拡張ポイントを提供する場合**: `net/http.Handler`, `sort.Interface`, `database/sql/driver.Driver` のように、ライブラリがプラグインの形を定義し、ユーザーコードが実装する形
- **複数実装を前提にした標準契約**: `io.Reader`, `error` のように、最初から複数実装が想定される標準的な抽象

これらは「慣習を曲げている」のではなく、「消費側 interface が合理的でない場面として最初から認められている枠」。

### 4. domain 層の `todo.Repository` の位置付け

`todo.Repository` は domain 層に interface 定義がある。これは上記 3 のうち **「依存性逆転（DIP）が必要なシーム」** に該当する:

- 自然な依存方向: domain → infra（domain が DB を使いたい）
- これを **逆転** させる: infra → domain（infra が domain の interface を実装）
- 逆転の目的: domain が infra の具体に依存しないようにする（オニオン / ヘキサゴナルアーキテクチャの中核）

DDD 原典（Vernon IDDD）でも Repository は domain 側に interface を置く。これは Java の慣習と DDD の原則が一致した結果であり、Go の慣習からは「逸脱」だが**「依存性逆転」という正当な理由がある逸脱**として広く受け入れられている。

### 5. handler ↔ usecase は逸脱の理由がない

handler → usecase の依存は:

- **自然な依存方向**（外側 → 内側）。逆転の必要がない
- アーキテクチャ的なシームでもない（普通の呼び出し関係）
- ライブラリ / フレームワーク的拡張ポイントでもない（アプリケーション内の普通の層境界）

3 で挙げた逸脱の理由のいずれにも該当しないため、**Go の慣習をそのまま適用** = **消費側（handler）で interface を定義** が筋となる。

DDD 原典（Java 例）では Application Service も interface として export されているが、これは Java の慣習（Spring Bean、Java EE の慣習）に従った結果であり、DDD の原則そのものが提供側 interface を要求しているわけではない。Go では同じ DDD の原則を**「依存方向は内側に向ける」**部分だけ尊重し、interface の置き場所は Go の慣習に従えばよい。

## 決定

### 原則

**「消費側で interface を定義する」**（Go の慣習に従う）。

### 例外

以下のケースでは提供側 / 内側で interface を定義する:

1. **依存性逆転が必要なシーム**（Repository, Port-Adapter 系）
   - 内側（消費される側）で interface 定義
2. **ライブラリ / フレームワークの拡張ポイント**
   - 提供側で interface 定義
3. **複数実装を前提にした標準契約**
   - 抽象を提供する側で interface 定義

### 本プロジェクトでの具体的な適用

| 関係 | 該当ケース | interface の場所 |
|------|----------|----------------|
| domain ↔ infra (Repository) | 例外 1（DIP） | domain 側 |
| presentation ↔ usecase | 原則 | presentation 側（handler 内で interface 定義） |

usecase 層の struct（`define.UseCase` 等）は interface を export しない。提供側は具象を返し、interface は消費側（handler）が必要な形で定義する。

## 帰結

### 利点

- Go の慣習（"Accept interfaces, return structs"）と一貫する
- 構造的型付けを最大限活用できる（提供側は interface の存在を知らなくてよい）
- usecase 層が interface を export しないため、提供側コードが消費者の都合を背負わない（YAGNI に沿う）
- Rust / Java 経験者にとっては初期違和感があるが、その違和感の正体（構造的型付け vs 名目的型付け）が ADR で言語化されているため、判断軸が伝達できる
- 「Repository が domain にある」と「handler が interface を定義する」は両立し、矛盾しない（前者は依存性逆転の例外、後者は原則）

### 欠点・コスト

- 同じ interface を複数の消費者が定義することがあり得る（DRY 違反に見える）。ただし interface のサイズは消費者ごとの最小限（必要なメソッドだけ）が望ましいため、結果として別物になることが多く、実害は小さい
- 消費者ごとに小さな interface が増えるため、一見「重複」に見える。Go コミュニティはこの方が望ましいと考えている（"The bigger the interface, the weaker the abstraction" - Rob Pike）

### スコープ外（別 ADR で扱う）

- mockery によるモック生成の対象範囲（消費側 interface もモック対象にするか） — 必要時に検討
- 例外 2 / 例外 3 が本プロジェクトで発生したときの個別判断 — その時点で判断

## 参考リンク

- ADR-0005: presentation 層の設計方針（gRPC + Slack API スタイル）
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs (Rob Pike)](https://go-proverbs.github.io/) — "The bigger the interface, the weaker the abstraction." / "Don't design with interfaces, discover them."
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) — "Interfaces"
- Vaughn Vernon, "Implementing Domain-Driven Design" (IDDD), ch.10 / ch.12
