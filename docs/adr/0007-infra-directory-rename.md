# ADR-0007: infrastructure ディレクトリの infra へのリネーム

## ステータス

採択（2026-05-01）

## コンテキスト

これまで `internal/infrastructure/` 配下に DB アクセス・Repository 実装・外部 API 連携などのインフラ層コードを置いていた。Phase 1.4 で `internal/server/` の組み立てに着手する前段階で、ディレクトリ名と Go の慣習との整合を再評価する必要が出た。

論点:

- DDD 原典（Vernon IDDD、Eric Evans 青本）の表記は **Infrastructure Layer**。一語で書くなら `infrastructure/`
- Go コミュニティは package 名を短く保つ慣習がある（Effective Go: "Good package names are short and clear"）
- 標準ライブラリも `crypto`, `encoding`, `text`, `bytes` のように短い名前を採用
- 既存コードでは `internal/infrastructure/` のまま使っており、import パスや文書の参照が複数箇所に存在
- 後発の Go 新規プロジェクトの土台として使う前提のため、慣習からの逸脱は最小化したい

## 検討と判断

### 1. Go 慣習との比較

Go の慣習に照らすと `infra/` が自然な選択:

- **短さ**: 標準ライブラリのレイヤー的 package 名（`crypto/`, `net/`, `encoding/`）と同等の長さ
- **意味の一意性**: `infra` は IT 業界で「インフラ」として通用する省略形であり、誤解の余地が小さい
- **import 文の見た目**: `internal/infra/repository/...` のほうが `internal/infrastructure/repository/...` より読みやすい

`infrastructure/` は冗長で、毎回タイプする / 読む際のノイズが大きい。

### 2. DDD 原典との整合

Vernon IDDD の表記は "Infrastructure Layer" だが、これは英文での書き方の規定であり、コードのディレクトリ名やパッケージ名を強制するものではない。Java / C# の DDD 実装でも、パッケージ名は言語の慣習に従って付ける（`com.example.infrastructure` でも `com.example.infra` でもよい）。

Go では言語の慣習を優先する判断が妥当。**概念名としては「infra 層 / インフラ層」と呼び、コード上のディレクトリ名 / パッケージ名も `infra` で統一する**。

### 3. リネームのコスト

このタイミングでリネームするコストと、後でやるコストの比較:

- **今やる**: 影響を受ける Go ファイルは 3 ファイル（import path 修正）、`sqlc.yaml` 1 行、ドキュメント数ファイル。1 回の bulk replace で完了
- **後でやる**: 新しい infra 配下のファイル（pool.go、その他）、新規追加される sub-package、新規 ADR / ルールの記述などが増えるため、リネーム対象が増えていく

**今が最もコストの低いタイミング**。

### 4. 他のレイヤー名との整合

`domain/`, `usecase/`, `presentation/` は現状維持:

- `domain` — DDD の中核概念で短く、これ以上省略できない
- `usecase` — DDD 原典の "Application Layer" を本プロジェクトでは `usecase/` と命名済み（既決、ADR-0002 等）
- `presentation` — 11 文字とやや長いが、`pres` 等の省略形は意味が伝わりにくい。慣習的にも `presentation/` で定着

`infrastructure` だけが「DDD 原典名そのまま」で他レイヤーより長く、Go 慣習との乖離が大きい。リネーム対象として妥当。

## 決定

- ディレクトリ: `internal/infrastructure/` → `internal/infra/`
- ドキュメント上の表記: `infra 層 / インフラ層` で統一（DDD 原典の "Infrastructure Layer" は概念説明部分でのみ使用可）
- 他レイヤー（`domain/`, `usecase/`, `presentation/`）は現状維持

## 影響範囲

リネーム時の更新対象:

- Go ファイルの import path（3 ファイル）
- `sqlc.yaml` の `out` パス
- ADR-0004 / ADR-0006 のパス・層名表記
- `.claude/rules/go/architecture.md` / `errors.md` / `naming.md` の表記
- 古いメモ系 md（`architecture-v1-feature-first.md`, `architecture-v2.md`, `2026-04-17-*.md` 等）は履歴のため対象外

## 帰結

### 利点

- Go の慣習（短い package 名）に沿う
- import 文 / ファイルパスが読みやすくなる
- 新規プロジェクトの土台として参照されたとき、Go プロジェクトとして自然な命名になっている
- 「infrastructure 層」と「infra 層」が混在していた表記が `infra` で統一される

### 欠点・コスト

- DDD 原典の表記（"Infrastructure Layer"）と異なる省略形を使うため、DDD 純粋派から見ると違和感があり得る
- 一度リネーム作業のための bulk replace が必要（一過性のコスト）

### スコープ外

- 他レイヤー名の変更（`domain`, `usecase`, `presentation`）は本 ADR の対象外
- レイヤーをまたぐ概念名（例: ADR で「インフラ層」と書くか「infra 層」と書くか）の細かな表記揺れ — 必要に応じて統一する程度で、別 ADR は不要

## 参考リンク

- [Effective Go: Package names](https://go.dev/doc/effective_go#package-names)
- [Go Code Review Comments: Package Names](https://github.com/golang/go/wiki/CodeReviewComments#package-names)
- ADR-0004: Collection-Oriented Repository への切替と楽観ロック採用
- ADR-0006: interface の配置方針
- Vaughn Vernon, "Implementing Domain-Driven Design" (IDDD)
