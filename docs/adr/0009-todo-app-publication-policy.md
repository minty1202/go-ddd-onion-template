# ADR-0009: todo_app は将来公開する、connectkit は private を維持する

## ステータス

採択（2026-05-04）

## コンテキスト

ADR-0008 で connectkit を別 private リポジトリに切り出すことが決定された。todo_app 自体の公開方針は ADR-0008 のスコープ外だったため、別途決定が必要となった。

todo_app は Go 新規開発のテンプレートとして構築されており、将来的にテンプレート / 読み物として公開する動機がある。一方で connectkit は ADR-0008 で「OSS 公開はしない」と明示されている。

## 検討と判断

### 1. 制約: public リポジトリは private 依存を持てない

todo_app を公開した瞬間、`go.mod` には `github.com/minty1202/connectkit` への require が残る。これにより外部読者の `git clone && go build` は以下の理由で失敗する:

- `proxy.golang.org` は private repo を fetch できない（404）
- 直接 fetch にフォールバックしても、外部読者は connectkit へのアクセス権がない（403）
- `sum.golang.org` の checksum 検証も成立しない

これは [Go 公式: Private Modules](https://go.dev/ref/mod#private-modules) の前提を裏返した状況で、`GOPRIVATE` は「アクセス権を持つ人が proxy / sum を bypass する」ための環境変数であり、外部公開には効かない。

### 2. 取りうる選択肢

| 案 | todo_app | connectkit | 外部 build |
|---|---|---|---|
| A | public | public | OK |
| B | private | private | （外部公開しない） |
| C | public | private | 失敗するが許容 |
| D | public（vendor commit） | private | OK だが connectkit ソースが事実上公開 |

### 3. 「公開 = 運用負荷」は誤解

「公開すると issues 対応 / PR review / changelog / 厳密 semver が必須」は誤解。GitHub では:

- issues は repo 設定で disable できる（[GitHub Docs - Disabling issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/configuring-issues/disabling-issues)）
- PR は無視 / close / 拒否しても良い
- `v0.x.x` のままなら破壊的変更は慣例的に自由（[SemVer §4](https://semver.org/#spec-item-4): "Major version zero (0.y.z) is for initial development. Anything MAY change at any time."）
- README で "personal project, not maintained for external use" と明示できる

実際のコストは以下の one-time 作業のみ:

- LICENSE 追加（無し = "all rights reserved" 扱いで誰も使えない）
- commit 履歴の secret 混入チェック
- 私的記述の精査
- README / 起動手順の整備

### 4. 案 D（vendor commit）が成立しない理由

`go mod vendor` で `vendor/` ディレクトリを commit すれば外部 build は成立するが、connectkit のソース自体が todo_app の public repo に丸ごと載るため、connectkit を private にする意味が無くなる。privacy としては実質公開と同等なので採用しない。

## 決定

- **todo_app は将来公開する**（テンプレート / 読み物としての価値があるため）
- **外部読者の `go build` 失敗は許容する**（誰かに動かしてもらう想定なし）
- **connectkit は private を維持**（ADR-0008 通り）
- **公開時点で必要な作業**:
  - LICENSE 追加
  - README に「private な `minty1202/connectkit` への access が無いと build できない」と明示
  - commit 履歴の secret 混入スキャン
  - 私的記述の精査

## 影響範囲

- 新規作成: なし（本 ADR のみ）
- todo_app の公開タイミング: 未定。本 ADR は「公開する場合の方針」を確定するもの

## 帰結

### 利点

- todo_app をテンプレート / 読み物として後日公開できる
- 公開前 / 公開後で connectkit 側の設計判断は変わらない（薄い convention layer 原則を保てる）
- ADR-0008 の OSS 化判断を独立に保てる

### 欠点

- 外部読者は `go build` を実行できない（コードを読むことはできる）

### スコープ外

- connectkit の OSS 化判断（ADR-0008 / 別 ADR の議題、本 ADR では現状の private 維持を確認するのみ）
- todo_app の公開タイミングの決定

## 参考リンク

- [Go Modules - Private Modules（公式）](https://go.dev/ref/mod#private-modules)
- [SemVer §4](https://semver.org/#spec-item-4)
- [GitHub Docs - Disabling issues](https://docs.github.com/en/issues/tracking-your-work-with-issues/configuring-issues/disabling-issues)
- ADR-0008: connectkit 切り出し
