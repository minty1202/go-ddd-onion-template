# ADR-0003: validation の責務分担とドメインルール定数の export

## ステータス

採択（2026-04-29）

## コンテキスト

domain 層の `TitleMinLen` / `TitleMaxLen` のような業務ルール定数を export して、presentation 層から参照できるようにしている設計を再評価した。

### 元々の意図

`internal/domain/todo/rules.go` で:

- 文字数制約の定数を export
- カスタムバリデータ `todo_title` を `init()` で `validation.Validate` に登録

これは「presentation 層が同じ業務ルールを再利用する」前提で、ハンドラの struct タグで `validate:"required,todo_title"` のように参照できる構造になっていた。

### 問題意識

presentation 層を proto/gRPC で実装する場合、proto から Go の export 定数を import できない。proto に validation を書くなら、数値（`min_len: 3` など）を proto 側にハードコードする必要があり、domain export と二重定義になる。「export して単一情報源（single source of truth）にする」という当初の動機が、proto 採用前提だと成立しない可能性が出てきた。

このため、validation の責務分担と export 定数の意義を整理する必要が生じた。

## 検討と判断

### 1. validation を何層に置くか

考え得る配置は 3 層:

1. **proto-validate（protovalidate / 旧 protoc-gen-validate, PGV）**: proto ファイル内に validation rule をアノテーションで書き、gRPC interceptor（リクエストを横取りする middleware）で handler 到達前に reject する仕組み。Buf 公式が推す現行形
2. **ハンドラ層**: go-playground/validator + struct タグ等で形式バリデーション（formal validation）。domain の export 定数を import して単一情報源と連動させる
3. **domain**: コンストラクタで Always Valid（不正な状態でインスタンスを存在させない、Vernon IDDD 由来）

### 2. 業界調査の結果

独立調査エージェントによる原典・公式・実運用 OSS 横断調査の結果:

- **Buf 公式（protovalidate）/ grpc-ecosystem / Connect RPC**: interceptor で handler 到達前に reject する前提で設計されており、handler 側で再 validate する推奨は出ていない
- **Vernon IDDD（赤本）**: application service は assertion を持たず、引数をそのまま domain に渡し、Always Valid な値オブジェクト / 集約のコンストラクタが invariant を守る
- **Khorikov（Always-Valid Domain Model）/ kgrzybek**: 「application 層の validation」と「domain 層の invariant」を分離する二層論を提唱しているが、これは proto-validate のような **境界 validation がない前提** での議論
- proto-validate を採用した時点で、それが Khorikov / kgrzybek の言う「application 層の validation」を兼ねていると読むのが筋

「proto-validate + handler 独自 validation」を明示的に推奨する一次ソースは見つからなかった。

### 3. ハンドラ層 validation を入れる正当化条件

業界標準を外れて 2 を入れる場合の正当化条件は限定的:

- proto-validate で表現できない相関ルール（cross-field、フィールド間の関係。tenant 単位の制約、外部状態依存など）がある
- domain → proto に値を一方向に流す仕組み（コード生成で定数を埋め込む等）がある

drift（値のずれ）検知 / 運用安全網 / テストの自己完結性は副次的な利益で、これらだけのために重ね掛けする業界慣習はない。

### 4. domain export 定数は presentation 実装に依らない

「presentation が proto/gRPC を採用するなら export の意義が薄れる、REST を採用するなら活きる」という議論は、domain が下位層（presentation）の都合で構造を変える形になり、オニオンアーキテクチャの依存方向（presentation → domain）に反する。

domain は「これがドメインルールである」という事実を提供する責務を持つ。それを presentation 層がどう使うか（proto-validate に書く / ハンドラから参照する / 使わない）は presentation 側の判断。

### 5. 参考: presentation 実装による意義の非対称

domain の構造判断には影響しないが、presentation 実装によって「export 定数がどう使われるか」は変わる:

- **REST（echo / gin など）+ Go**: ハンドラ層で go-playground/validator が標準。Go の struct タグはコンパイル時の文字列リテラルなので `validate:"min=todo.TitleMinLen"` のように定数を直接書けない。カスタムバリデータ経由で間接的に共有する形（現状の `todo_title` 方式）が単一情報源として機能する
- **gRPC + proto-validate**: 境界が proto 側に立つ。proto 内の数値はハードコードになる（domain と drift リスク）。export 定数の役割はテスト用途や proto 生成時の埋め込み用途に変わる

これは presentation 側の都合。domain の責務として「ルール定数を export する」こと自体は presentation 実装に依らず維持する。

## 決定

### validation の責務分担

- **proto-validate（採用時）**: 境界 validation として interceptor で handler 到達前に reject
- **ハンドラ層独自 validation**: **入れない**（業界標準ではないため、また domain の Always Valid と責務が重複するため）
- **domain（Always Valid）**: コンストラクタで業務不変条件チェック（最後の砦）

二層構成（proto-validate + domain Always Valid）を標準とする。

### domain export 定数の扱い

- `TitleMinLen` / `TitleMaxLen` のような domain ルール定数は **export を維持** する
- カスタムバリデータ `todo_title` の登録（`init()` での `validation.Validate.RegisterValidation`）も維持
- 理由: domain は「ドメインルールである」事実を提供する責務を持ち、presentation 実装に依らず独立して存在する
- presentation がこれをどう使うかは presentation 側の判断

## 帰結

### 利点

- Go gRPC + DDD の業界標準形（proto-validate + Always Valid）に整合
- 責務がきれいに分かれる（proto = 仕様書、domain = 業務不変条件）
- domain が presentation 実装に依存しない（オニオンの依存方向と整合）
- 余計な validation コードを書かない

### 欠点・コスト

- gRPC + proto-validate を採用した場合、proto 側に数値がハードコードされ、domain 定数と drift する可能性が残る
- この drift リスクは別途 CI テスト等で担保する余地があるが、本 ADR では具体的な仕組みには立ち入らない

### スコープ外

- presentation 層の選択（proto/gRPC か REST か）は未確定。別 ADR で扱う
- proto と domain の値一致を担保する具体的な仕組み（CI テスト、コード生成での埋め込み）は presentation 着手時に決める
- 相関ルール（cross-field）が必要になった場合の扱いも別途検討

## 参考リンク

- [Protovalidate gRPC and Go quickstart (Buf 公式)](https://protovalidate.com/quickstart/grpc-go/)
- [grpc-ecosystem/go-grpc-middleware protovalidate interceptor](https://pkg.go.dev/github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate)
- [connectrpc.com/validate](https://pkg.go.dev/connectrpc.com/validate)
- [Google AIP-193: Errors](https://google.aip.dev/193)
- [VaughnVernon/IDDD_Samples](https://github.com/VaughnVernon/IDDD_Samples)
- [Khorikov, "Always-Valid Domain Model"](https://enterprisecraftsmanship.com/posts/always-valid-domain-model/)
- [Kamil Grzybek, "Domain Model Validation"](https://www.kamilgrzybek.com/blog/posts/domain-model-validation)
- Vaughn Vernon, "Implementing Domain-Driven Design" (赤本, 2013)
