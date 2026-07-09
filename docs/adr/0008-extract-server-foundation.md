# ADR-0008: Connect RPC ベースの個人共通基盤を別 private リポジトリに切り出す

## ステータス

採択（2026-05-04）

## コンテキスト

本リポジトリ（Todo アプリ）は当初から「DDD + オニオンアーキテクチャ + Connect RPC で書く Go プロジェクトのテンプレート」として構築している。Phase 1 / 2 で以下が確立した:

- DDD + オニオンの層構造（domain / usecase / infra / presentation）
- A 層 Connect Interceptor 群（recovery / request_id / logging / normalize）
- 各層のエラー設計（`errdefs.Kind`、`Kind → connect.Code` マッピング）
- server lifecycle（graceful shutdown、Mounter pattern、composition root）

加えて、以下の実態が判明した:

1. **同一個人 (`@minty1202`) が複数の Go プロジェクトを並行して持つ予定**
   - 本リポジトリ（Todo、テンプレート兼ねる開発中プロジェクト）
   - 新サービス（設計段階、Connect RPC 採用予定、同じ DDD + オニオン構造）
2. **共通化したい対象は「コード」ではなく「決定の集合」**
   - 「recovery をどう書くか」「request_id を UUIDv7 にするか」「logging で body を出すか」「Kind → Code マッピング」等は、**全プロジェクトで同じ実装にするのが自然**
   - 各プロジェクトで再議論したくない
3. **OSS 公開は意図しない**（private repo で十分）
4. **Connect RPC を当面の前提とする**（Gin / Echo は対象外）

これにより「フレームワーク基盤を独立した resource に切り出すか」が論点となった。

## 検討と判断

### 1. 候補となる選択肢

- **A. 別 private repo に切り出す**: framework 側を独立リポジトリ化、Todo / 新サービスは `go get` で参照
- **B. 切り出さず、コピー運用**: 各プロジェクトに必要分をコピー、ADR で決定を文書化
- **C. monorepo + `go.work`**: Todo / 新サービス / framework を 1 リポジトリに同居
- **D. 一部のみ extract**（例: normalize のみ）: 高凝集な独立パーツだけ切り出し
- **E. 既存 OSS（kratos / encore 等）に乗る**: 自前 framework を作らない

### 2. 動機との整合性

「**1 度決めたら全プロジェクトで自動適用したい / 各プロジェクトで再議論したくない**」が本決定の主動機。各案の達成度:

| 手段 | 決定の自動適用 | drift 防止 | 動機との適合 |
|---|---|---|---|
| A. 別 private repo | あり (import) | 強い | 高 |
| B. コピー運用 | なし | 弱い (規律次第) | **低** |
| C. monorepo | あり (workspace) | 強い | 中（生死サイクルが違うアプリを同居させる無理がある） |
| D. 一部 extract | 部分のみ | 部分のみ | 中 |
| E. 既存 OSS | あり | 強い | 低（DDD + オニオン + 自分の決定セットが活きない） |

**A が動機を最も直接的に満たす**。

### 3. 反対論への反論

切り出しに反対する典型論点（Rule of Three / 個人 OSS 維持コスト / premature abstraction）は、本ケースでは適用が弱い:

- **Rule of Three**: 「抽象化対象が未確定なとき」のための経験則。recovery / request_id / logging のような **既に production 標準が確立された横断的関心事**には当てはまらない（[AHA Programming - Kent C. Dodds](https://kentcdodds.com/blog/aha-programming): 「抽象は形が見えてから」）
- **個人 OSS 維持コスト**: 公開 OSS の話。private repo であれば issue / PR / 公開 changelog / semver 厳密遵守は不要。`GOPRIVATE` 環境変数 + `git tag` で済む（[Go Modules - Private Modules 公式](https://go.dev/ref/mod#private-modules)）
- **premature abstraction**: 両側を同一人物が制御するため、不適合があれば破壊的変更で即対応可能。後方互換負債なし、コミュニケーションコストなし

### 4. Go 公式の推奨

[Go Modules Layout (公式)](https://go.dev/doc/modules/layout) より:

> In case the server repository grows packages that become useful for sharing with other projects, it's best to split these off to separate modules.

本ケースは完全にこのパターン。

### 5. 業界事例

- Uber `fx`: 元々社内共通基盤、後に OSS 化（[uber-go/fx](https://github.com/uber-go/fx)）
- Cloudflare: 内部に複数の Go 共通基盤
- 1Password Passage: monorepo + 共有 package 運用（[ブログ記事](https://passage.1password.com/post/shared-go-packages-in-a-monorepo)）
- 個人 / 社内専用 Go framework は規模 2〜3 プロジェクトから始まる事例あり

### 6. 切り出さない選択肢の評価

- **B. コピー運用**: drift が必然。決定の一貫性を構造的に強制できない。ユーザーの動機と矛盾
- **C. monorepo**: Todo（テンプレート開発）と新サービス（本番想定）はライフサイクルが異なる。同居させる根拠が薄い
- **D. 一部のみ extract**: 切り分け基準が曖昧、「決定セット丸ごと固定」という動機にやや合わない
- **E. 既存 OSS**: DDD + オニオン + 自分の決定セットを活かせない

### 7. 名前と構造

リポジトリ名: **`connectkit`**

理由:
- 中の interceptor / dependencies / router は全て `connect.Interceptor` interface に乗っており、Connect 専用
- middleware / ops は HTTP 全般に使えるが、Connect server を組む文脈で利用される
- 別 framework（Gin / Echo）対応は将来別リポジトリ（`ginkit` 等）として切り出す方針で、同居させない

構造（単一モジュール、単一 root パッケージ）:

```
connectkit/
├── go.mod
├── server.go           ← Server + NewServer + Run
├── mount.go            ← Mounter interface
├── router.go           ← newMux (internal helper)
├── dependencies.go     ← Dependencies struct (Mounters のみ)
├── recovery.go         ← interceptor: recovery
├── request_id.go       ← interceptor: request_id
├── logging.go          ← interceptor: logging
├── normalize.go        ← interceptor: normalize
├── default.go          ← Default() (recovery → request_id → logging → otel → normalize → validate)
├── (Phase D) middleware 系: cors.go / security_headers.go / max_body.go / h2c.go
├── (Phase D) ops 系: health.go / version.go / grpchealth.go / grpcreflect.go / pprof.go
└── (Phase C) config.go ← BaseConfig (AppEnv / DatabaseURL / Port / ShutdownTimeout)
```

サブパッケージを切らず全て root の `package connectkit` にする理由: `connectrpc.com/connect` 公式と同じ flat 構造に揃えると、利用側 import が `import "github.com/minty1202/connectkit"` 1 行で済み、`connectkit.NewServer / connectkit.Default / connectkit.Mounter` のように単一名前空間で扱える。サブパッケージ化すると `server.Server` の stutter や `connectkitserver` のような alias が必要になり、薄い convention layer の名前空間としては煩雑。

`internal/` を使わない理由: private repo で import 元は自分のプロジェクトに限定されるため、`internal/` で更に縛る必要がない + `internal/` の中身は外から見ると import パスが汚くなる。

### 8. 切り出さない部分

以下は**アプリ側に残す**:

- ドメイン層 (`internal/domain/<aggregate>/`)
- usecase 層 (`internal/usecase/<aggregate>/<verb>/`)
- Repository 実装 (`internal/infra/repository/<aggregate>repo/`)
- sqlc 生成コード (`internal/infra/db/`)
- presentation 層 (`internal/presentation/<aggregate>rpc/`)
- composition root (`cmd/server/main.go`)
- Mounter 登録（`modules.go` 相当）
- アプリ固有 config（`connectkit.BaseConfig` を embed して拡張）

以下は**切り出さない**（薄すぎる / 既存ライブラリで足りる）:

- `pgxpool.Pool` のラッパー（`pgxpool.New(ctx, dsn)` を直接呼べばよい）

### 9. ローカル開発体験

`go.work` を使うことで、`connectkit` と `todo_app` / 新サービスを同時に編集して 1 commit に含めることが可能（[Go Workspaces 公式チュートリアル](https://go.dev/doc/tutorial/workspaces)）。

CI / 本番 build では `GOWORK=off` 推奨で、各モジュールが独立して build できる必要がある。

## 決定

- **`connectkit` という名前で別 private リポジトリに切り出す**
- **単一モジュール / 単一 root パッケージ構成**で組む
- **Phase 3 / 4 は最初から `connectkit` 内で実装**（Todo プロジェクト内に作ってから移動するロスを避ける）
- **`GOPRIVATE=github.com/minty1202/*` で個人 private 運用**
- **ローカル開発は `go.work`**（todo_app + connectkit + 新サービスを横断）
- **OSS 公開はしない**（将来的に判断する場合は別 ADR）

## 影響範囲

### 新規作成

- `github.com/minty1202/connectkit` 新リポジトリ
- `connectkit/go.mod`、上記構造
- ローカル `go.work`（todo_app + connectkit）

### Todo プロジェクトの変更

- `internal/server/` 配下を connectkit に移動
- `internal/server/interceptor/` 配下を connectkit に移動
- `internal/usecase/errdefs/` は todo_app に残す（usecase 層の語彙、connectkit に吸収しない）
- `internal/presentation/todorpc/error.go` の Kind → connect.Code マッピングは todo_app の presentation 層に残す（connectkit 経由しない）
- `internal/config/` のうち BaseConfig 該当部分を connectkit に移動、アプリ固有部分は embed で残す
- `cmd/server/main.go` の import 更新

### Phase 3 / 4

- middleware 系（CORS / security_headers / max_body / h2c）を connectkit root に追加
- ops 系（healthz / readyz / grpchealth / grpcreflect / version / pprof）を connectkit root に追加

## 帰結

### 利点

- **決定の一貫性が構造的に強制される**: import = 自動適用、各プロジェクトで再議論不要
- **2 プロジェクト目以降の立ち上げコストが下がる**: `go get connectkit` + composition root の記述だけで A 層 / B 層 / C 層が揃う
- **bug fix / 改善が両プロジェクトに即時伝搬**: drift しない
- **境界が物理的に明確**: アプリコードと framework コードを混同する余地がない
- **Go 公式推奨パターンに沿う**

### 欠点・コスト

- **新リポジトリの初期セットアップ作業**: go.mod 作成、move、import 更新、go.work 設定
- **破壊的変更時に両プロジェクト同時更新が必要**: ただし両側を同一人物が制御するため低コスト
- **将来 framework を変えたい場合（Gin / Echo 等）にこの選択がロックインになる**: 別 `<name>kit` リポジトリを作る前提なので致命的ではない

### スコープ外（別途決定）

- **`connectkit` の OSS 化**: 公開する判断は後日。今は private で運用
- **multi-module 化**: 単一モジュールから始め、必要が出たときに分割を検討
- **Gin / Echo 対応**: 別リポジトリ（`ginkit` 等）として切り出す前提で、`connectkit` には含めない
- **Phase 3 / 4 の具体的実装内容**: 別途実装時に決定

## 参考リンク

- [Go Modules Layout (公式)](https://go.dev/doc/modules/layout)
- [Go Modules - Private Modules (公式)](https://go.dev/ref/mod#private-modules)
- [Go Workspaces チュートリアル (公式)](https://go.dev/doc/tutorial/workspaces)
- [AHA Programming - Kent C. Dodds](https://kentcdodds.com/blog/aha-programming)
- [Refactoring and the Rule of Three](https://understandlegacycode.com/blog/refactoring-rule-of-three/)
- [The Wrong Abstraction - Sandi Metz](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction)
- [Uber fx](https://github.com/uber-go/fx)
- [1Password Passage: Shared Go Packages in a Monorepo](https://passage.1password.com/post/shared-go-packages-in-a-monorepo)
