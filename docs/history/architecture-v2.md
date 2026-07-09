---
paths:
  - "internal/**/*.go"
---
# アーキテクチャルール（DDD + オニオンアーキテクチャ）

## 今回の議論で決まったこと

このプロジェクトは **B-1 構造 = layer-first + ドメイン層内 aggregate 分割** を採用する。

具体的に確定したのは以下 4 点:

1. **4 層構成**: `domain` / `usecase` / `infrastructure` / `presentation`
2. **Application 層の名称は `usecase/`**（既存コードに合わせる）
3. **domain 層の内部だけ、aggregate ごとにサブディレクトリを切る**
4. **domain 層以外の層の内部構造、横断的関心事の位置付け、組み立て責務などは全て未確定**（後述）

### 前提

- **1 BC = 1 アプリ**。BC が発生する規模になったらリポジトリごと分ける

### B-1 採用の根拠

- **Evans DDD Ch.5 (2003, 一次確認済み)**:
  - 「Choose MODULES that tell the story of the system and contain a cohesive set of concepts」
  - 「keep all the code that implements a single conceptual object in the same MODULE」
  - p.83 "The Pitfalls of Infrastructure Driven Packaging" で technical layer 再分割（`entity/`, `repository/`, `service/` 分離）を明示批判
  - これらから「ドメイン層内は aggregate 単位で切る」が直接支持される
- **本番指向の参考実装**（`dotnet-architecture/eShopOnWeb`、`citerus/dddsample-core`）が B-1 寄り
- 複数 aggregate をまたぐユースケース・ドメインサービスに自然対応できる
- DDD / オニオンの教科書的構造で認知コストが低い
- 層の独立性が高い

### 不採用となった選択肢

- **feature-first（`internal/<feature>/` 配下にオニオン）**: 前回（`2026-04-17-architecture-decision.md`）で採用していたが、根拠が DDD 非採用プロジェクトに偏り、Bogard "Vertical Slice" を誤読していたため撤回
- **B-2（ardanlabs/service 型、各層で aggregate 分割）**: 複数 aggregate をまたぐロジックに弱い、Application と Presentation の層分離ができない

詳細は `../../2026-04-17-architecture-revision.md` を参照。

## ディレクトリ構造（確定した骨格のみ）

```
internal/
  domain/                         ← ドメイン層（核）
    todo/                         ← aggregate
    user/                         ← aggregate
    ...
  usecase/                        ← ユースケース層（Application 相当）
  infrastructure/                 ← インフラ層
  presentation/                   ← プレゼンテーション層
```

- `internal/` 直下に上記 4 層を配置する
- domain 層の中は **aggregate 単位でサブディレクトリ**
- 上記以外（横断的関心事の配置、各層の内部構造）は**未確定**

## 各層の責務（骨格のみ）

### domain 層

- ドメインモデルの知識を対応するオブジェクトに書く
- 常に正しいインスタンスしか存在させない
- **aggregate ごとにサブディレクトリ**（`domain/todo/`, `domain/user/`）
- Repository は interface だけこの層で定義し、実装は infrastructure 層

### usecase 層

- ドメインオブジェクトを使ってユースケースを実現する層
- 複数 aggregate をまたぐユースケースも自然に書ける（B-1 の強み）

### infrastructure 層

- DB、外部 API 等の**具体的な実装**を持つ
- domain 層が定義した interface（Repository 等）を実装する

### presentation 層

- HTTP / gRPC / CLI 等、アプリを外に公開する層

（Go での具体的な書き方、内部構造、細部は全て下記「まだ決まっていないこと」を参照）

## まだ決まっていないこと

今回の議論では B-1 構造の採用だけを確定させた。以下は全て未確定で、以降の議論で個別に決める。

1. **usecase 層の内部構造**: ファイル単位 / フォルダ単位、命名規則、戻り値型
2. **infrastructure 層の内部構造**: Repository 実装の置き方（aggregate 別 or フラット）
3. **presentation 層の内部構造**: aggregate 別 / エンドポイント別 / フラット
4. **横断的関心事の位置付け**: errdefs, http, config, logger, DB 接続プール等を `internal/` 直下にどう配置するか。前回議事録の「フラット配置」「`platform/` 不採用」は B-1 への変更で前提が崩れたため再検討
5. **パッケージ名衝突の対処**: `internal/domain/todo/` と `internal/infrastructure/repository/todo/` が `package todo` で衝突する。import エイリアス / サフィックス命名（`tododomain`, `todorepo`）/ その他、手段は未決定
6. **ID の扱い**: typed ID ラッパー / primitive、参照 ID の実装反映
7. **errdefs の設計**: 置き場（横断 or layer 別）、interface のみ or 具体型含む、分類軸
8. **apperr の扱い**: `internal/apperr/violation.go` の位置付けを errdefs と合わせて再判断
9. **aggregate 間の依存ルール**: 直接 import 可 / `domain/service/` 経由、`depguard` 等の静的検証
10. **ユースケースとドメインサービスの使い分け**: 複数 aggregate をまたぐロジックの置き場
11. **横断処理（orchestration）の置き場**: 発生時に判断
12. **`cmd/main.go` の組み立て責務**: 前回議事録の前提が崩れたため再検討
