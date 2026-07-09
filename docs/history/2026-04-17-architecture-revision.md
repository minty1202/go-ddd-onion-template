# 議事録: 2026-04-17 — アーキテクチャ方針の再検討と修正（feature-first → B-1 への変更）

## 背景・目的

### 発端

同日に決定した feature-first 方針（`2026-04-17-architecture-decision.md`）について、ユーザーから「一旦冷静になって、決定が最適か見直したい」との指示。特に「feature-first を採用した根拠に偏りがないか」を出発点として再検討することにした。

### 目的

前回の判断根拠の偏り・誤読を洗い直し、Go + DDD + オニオンアーキテクチャを採用する **単一 BC の Web アプリ（1 BC = 1 アプリ前提）** に最も適した構造を、より確かな根拠で決定し直す。

### 判断軸（貫いた基準）

- 前回と同じ基準を維持:
  - 原典（Evans, Vernon, Martin, Fowler, Palermo）と著名 Gopher の発言を強い根拠として扱う
  - 教育用テンプレートは弱い証拠として扱う
  - 推測で「標準」「ベストプラクティス」と断定しない
- **追加基準（今回の議論で明示化）**:
  - 解釈・結論の独立監査を挟む（調査エージェントとは別に監査エージェントを走らせる）
  - 「本番耐用」「教育リファレンス」「テンプレート」の位置付けを都度確認する

---

## 議論の流れ

### 1. 前回判断の偏りの指摘

ユーザーから「前回参照した大規模プロジェクト（Terraform, K8s, Vault 等）は Web アプリとは性質が違うのでは？」との指摘。

さらにより本質的な指摘に発展: 「分岐軸は『Web アプリかどうか』ではなく『**DDD / クリーン / オニオンを明示採用しているか**』では？」

**議論の結果**: 前回の判断根拠（Terraform / K8s 等の大規模 Go プロジェクト）は **DDD / オニオンを採用していない** ため、DDD 採用前提の構造判断の根拠としては不適切だった。

### 2. 調査の再実施（言語横断）

以下 2 本をバックグラウンドで委任：

1. **実運用プロジェクト調査**: DDD / Clean / Onion を明示採用しているプロジェクトの構造傾向（言語横断、14 件）
2. **原典再読**: Evans, Vernon, Palermo, Fowler, Uncle Bob, Bogard の分割指針

#### 調査 1 の結果（暫定）

- 「DDD 採用 = feature-first」は誤り
- 「単一 BC / 複数 BC」で分かれる傾向が見えた
- ただしこの分類は後に「BC と BC 内モジュールの混同」として修正（§5 参照）

#### 調査 2 の結果（要修正箇所を含む）

- **Bogard "Vertical Slice" は DDD/Clean の代替として書かれた記事**。前回 feature-first の原典支持根拠として引用したのは**誤読**
- **Fowler の記述は規模条件つき**（「小規模は layer-first で可、肥大したら feature-first」）。前回条件を省いて無条件 feature-first 支持として引用していた
- Evans の MODULE 原則は二次経由での引用のみ、一次確認が未達

### 3. 検証エージェントによる調査結果の独立検証

調査結果の信頼性を中立に評価させた。

**指摘された主な弱点**:
- Evans Ch.5 MODULE 原則が一次未確認で、原文は**むしろ feature-first 支持方向に読める**
- サンプル 14 件中 C# が 6 件で .NET 文化圏バイアスの可能性
- テンプレート用途での「BC 拡張可能性」の扱いが抜けている
- Fowler を「大きくなったら feature-first」と逆向きにも読める

**総合**: 判断保留（弱く layer-first 寄り）

### 4. BC 概念の誤用修正

ユーザーから「todo, user, schedule は BC ではないよね？」と指摘。

**整理**:
- **BC（Bounded Context）**: モデルの言語・意味が一貫する境界。組織・チーム・言語体系が分岐するレベル
- todo / user / schedule のように同じアプリ内で追加される機能は **1 BC 内のアグリゲート / サブドメイン**
- 前回の調査結果の「複数 BC」という分類で、実は **BC 内の複数モジュール** だったものを混同していた（wild-workouts の 3 service、modular-monolith の 5 モジュールなど）

### 5. 前提確定: 1 BC = 1 アプリ

ユーザーから「BC を扱うような規模にする予定ないです。BC が発生する規模になるならアプリごと分裂する」と前提確定。

**帰結**:
- このテンプレートは **1 BC = 1 アプリ** 前提
- 議論すべきは「1 BC 内でアグリゲート / 機能が増えていく前提での最適構造」**だけ**
- 検証エージェントが指摘した「BC 拡張可能性」論点は**無効化**
- 代わりに残るのは「アグリゲート追加時のパッケージ構造の手戻りコスト」

### 6. Evans Ch.5 MODULE 原則の一次確認

Evans DDD 原典 (Part I Ch.5) を一次ソース（PDF）で確認。確認できた主要引用：

> "Choose MODULES that tell the story of the system and contain a cohesive set of concepts. Give the MODULES names that become part of the UBIQUITOUS LANGUAGE."

> "Use packaging to separate the domain layer from other code. Otherwise, leave as much freedom as possible to the domain developers to package the domain objects in ways that support their model and design choices."

> "Unless there is a real intention to distribute code on different servers, keep all the code that implements a single conceptual object in the same MODULE, if not the same object."

さらに Ch.5 p.83 の節 **"The Pitfalls of Infrastructure Driven Packaging"** で、J2EE の Entity Bean / Session Bean 分離を「fragment the implementation of the model objects」「robs an object model of cohesion」と**明示的に批判**していることを確認。

**結論**:
- Ch.5 時点で BC は未導入（Ch.14 で初出）→ MODULE 原則は **BC 内の話**
- Evans は **ドメイン層内を概念単位（= aggregate）で切れ**と強く推奨
- **technical layer による再分割（entity/ repository/ service 分離）を明示批判**

### 7. 解釈の独立監査

Evans 引用の読み取りと、そこから導いた解釈の妥当性を独立監査エージェントに評価させた。

**修正された解釈**:
- Evans は「**ドメイン層を他から物理分離**」を明示支持しているが、UI / Application / Infrastructure の水平分割には**沈黙（黙認）**。前段で「最上位 layer-first を Evans が積極支持」と読んだのは**過剰解釈**
- technical layer 再分割の否定は、「直接禁止ではなく帰結」ではなく、**Ch.5 p.83 で明確に批判されており、もっと強く否定してよい**
- 「最上位 layer-first + ドメイン層内部 feature-first」のハイブリッド解釈自体は原典と矛盾しないが、根拠づけの強さは**条件つき妥当（5 段階で 3）**

### 8. dddsample-core の信頼度検証

前段で「Evans 公認の本番レベル事例」扱いしていた `citerus/dddsample-core` の信頼度を調査。

**結果**: **レベル 3（権威ある教育リファレンス）**

- Evans の Domain Language 社と Citerus 社の共同作業（README で明示）→ 権威性あり
- 公式ページで「"a decent way" であり "the way" ではない」と明言、**本番耐用は否定**
- Citerus 自身が業務で使っている証拠なし
- ddd-crew の学習リスト**非掲載**（モダン DDD 学習の標準道標ではない）
- 2025 年もメンテ継続（Star 5k）

**帰結**: 教育リファレンスとして参考にはするが、本番実運用事例としてはカウントしない。

### 9. Bulletproof React 級 DDD リファレンスの不在確認

ユーザーから「自分が知らないだけで、React の `alan2207/bulletproof-react` に相当する DDD 版のデファクトがあるなら知りたい」との要望。言語横断で調査。

**結果**: **DDD 界隈に Bulletproof React 級の単一デファクトは存在しない**

最も近いのは `kgrzybek/modular-monolith-with-ddd`（C#、13.6k ★）だが、言語横断の共通規範にはなっていない。

**存在しない理由**:
1. 言語・プラットフォーム分散（.NET / Java / Go / TS で別々のリファレンスが発達）
2. ドメイン固有性（DDD は「汎用テンプレート」と相性が悪い）
3. 教材と本番耐用の乖離（モデリングが良い教材は CI / デプロイが薄い、運用系が強いテンプレは DDD を前面に出さない）

### 10. 本番耐用レベルの候補絞り込み

以下の基準で再整理：

- レベル 5（本番運用事例）: **なし**
- レベル 4（本番指向の参考実装、README で明示）: `kgrzybek/modular-monolith-with-ddd`, `ardanlabs/service`, `eShopOnWeb`
- レベル 3（権威ある教育リファレンス）: `dddsample-core`, `IDDD_Samples`, `wild-workouts`, `ddd-by-examples/library`

「1 BC = 1 アプリ」前提で絞ると、残るのは：
- **ardanlabs/service**（Go）
- **eShopOnWeb**（C#）
- 参考: **dddsample-core**（Java、レベル 3）

### 11. 3 プロジェクトの構造比較

| プロジェクト | トップレベル | ドメイン層内 | 特徴 |
|---|---|---|---|
| ardanlabs/service | `app`/`business`/`foundation` | `userbus`/`productbus` (aggregate) | サフィックス命名、stores を aggregate 配下に |
| eShopOnWeb | `ApplicationCore`/`Infrastructure`/`Web` | `Entities/*Aggregate/` | 最上位 layer、aggregate は Entities 内のみ |
| dddsample-core | `domain`/`application`/`infrastructure`/`interfaces` | `domain/model/cargo/` など | Evans 忠実 4 層 |

**共通原則**: 3 つとも **aggregate ごとにフォルダを切る** は完全に共通。

**分岐**:
- **B-1**: トップ layer-first、**ドメイン層の中だけ** aggregate 分割（dddsample-core, eShopOnWeb）
- **B-2**: トップ layer-first、**各層の中で** aggregate 分割（ardanlabs/service）

### 12. B-1 と B-2 のメリット・デメリット整理

| 観点 | B-1（純 layer-first） | B-2（各層 aggregate 分割） |
|---|---|---|
| 複数 aggregate をまたぐユースケース | ○ 自然 | △ 置き場に迷う、命名と実態がズレる |
| ドメインサービスの置き場 | ○ `domain/service/` で自然 | △ 置き場に迷う |
| 1 aggregate の凝集性 | △ 複数層に分散 | ○ grep 一発 |
| Go のパッケージ名衝突 | × エイリアス必要 | ○ サフィックスで回避 |
| DDD / オニオンの教科書性 | ○ 明示的 | △ 物理配置と論理層が乖離 |
| 層の独立性（gRPC / CLI 追加） | ○ 独立 | × app に統合されていて追加が面倒 |
| 学習コスト | ○ 低い | △ サフィックス規則を覚える必要 |

### 13. B-1 採用の決定

**決定**: **B-1（純 layer-first + ドメイン層内 aggregate 分割）** を採用。

**根拠**:
1. Evans Ch.5 原典（一次確認済み）: ドメイン層内を概念単位で切る、technical layer 再分割を明示批判 → B-1 と完全整合
2. 本番指向 / 教育権威ある参考実装（eShopOnWeb, dddsample-core）が B-1 寄り
3. Todo アプリの想定拡張（User × Todo × Tag × Schedule のような**複合ロジック**）で、複数 aggregate をまたぐユースケース・ドメインサービスに自然に対応できる
4. DDD / オニオンの教科書的構造で、認知コストが低い（ユーザーの Rails 経験からくる「構造から一目で分かる」志向にも合致）
5. 層の独立性が高く、将来 gRPC / CLI / 別 Presentation を足しやすい

**B-2 を採らなかった理由**:
- 複数 aggregate をまたぐロジックに弱い
- Application と Presentation の層分離ができず、拡張時に窮屈
- サフィックス命名規則の独自性が学習・保守コストになる

### 14. Application 層の名称

DDD / オニオン用語では「Application 層」が一般的だが、**このプロジェクトでは既存コードに合わせて `usecase/` を継続採用**する。意味としては同じ。

---

## 決定事項まとめ

1. **B-1（純 layer-first + ドメイン層内 aggregate 分割）** を採用
2. 4 層構成: `domain/` / `usecase/` / `infrastructure/` / `presentation/`
3. **Application 層の名称は `usecase/`** で統一（既存コード優先）
4. ドメイン層の内部は **aggregate ごとにサブディレクトリ**（`domain/todo/`, `domain/user/`, `domain/schedule/`）
5. **パッケージ名衝突（`package todo` が複数層で発生）は import エイリアスで対処**（`domainTodo`, `repoTodo` など）
6. `internal/` 直下は引き続きフラット（`platform/` 等のラッパーは作らない）

**前回議事録の決定を「維持」から「白紙化」に戻したもの**（構造が変わって前提が崩れたため、未確定事項に移動）:
- ID は `ulid.ULID` 直接使用 → 要再検討
- errdefs は interface のみの横断契約パッケージ → 要再検討

---

## 未確定事項

1. **usecase 層の内部構造**: ファイル単位かフォルダ単位か、ユースケースの命名規則（`CreateTodoUseCase` か `CreateTodo` か）
2. **ユースケースオブジェクトとドメインサービスの使い分け**: 複数 aggregate をまたぐロジックを usecase / domain service のどちらに置くか
3. **apperr の扱い**（前回議事録から継続）
4. **aggregate 間の依存ルール**（前回議事録から継続、feature → aggregate に語彙修正）
5. **横断処理（orchestration）の置き場**（前回議事録から継続、発生時に判断）
6. **presentation 層の内部構造**: HTTP ハンドラのディレクトリ分割方針（aggregate 別 or フラット）
7. **ID の扱い**: 前回「ulid 直接使用」と決めたが、B-1 構造への変更で前提が崩れたため再検討（typed ラッパー、参照 ID の観点含む）
8. **errdefs の設計**: 前回「interface のみの横断契約」と決めたが、B-1 構造の前提で再検討（置き場、interface と具体型、分類軸）

---

## 前回議事録との差分

| 項目 | 前回（`2026-04-17-architecture-decision.md`） | 今回 |
|---|---|---|
| 全体構造 | feature-first（`internal/<feature>/` 配下にオニオン層） | **B-1 = layer-first + ドメイン層内 aggregate 分割** |
| トップレベル | feature と横断的関心事をフラットに並列 | layer（domain/usecase/infrastructure/presentation）+ 横断（errdefs/http/db/config/logger）をフラット |
| aggregate の位置 | feature ルート直下 = 1 aggregate 1 ディレクトリ | `domain/<aggregate>/` にまとまる |
| 層の分離 | 各 feature 内で 4 層分離 | グローバルに 4 層分離 |
| 組み立て関数 | 各 feature の `feature.go` に `New()` | 未定（`cmd/main.go` or 別設計） |

維持された決定:
- `internal/` 直下はフラット（`platform/` / `foundation/` のラッパー不採用）
- ID は `ulid.ULID` 直接使用
- errdefs は interface のみの横断契約

---

## このセッションで追加・更新したルール / メモリ

### ルール

- `.claude/rules/audit-agent.md` を新規作成（監査エージェントの体系化 TODO メモ）
- `.claude/rules/go/architecture.md` を **B-1 構造で全面書き換え**
- 前 feature-first 版を `architecture-v1-feature-first.md` としてルート直下にバックアップ

### メモリ

- `project_stack.md` を B-1 採用に合わせて更新

---

## 参照した情報源（今回新規・再確認）

### 一次確認済み

- [Evans DDD Ch.5 (2003 final manuscript, fabiofumarola mirror)](https://fabiofumarola.github.io/nosql/readingMaterial/Evans03.pdf) — MODULE 節、"The Pitfalls of Infrastructure Driven Packaging" 節
- [Evans DDD Reference (2015, domainlanguage.com)](https://www.domainlanguage.com/wp-content/uploads/2016/05/DDD_Reference_2015-03.pdf) — MODULES 項

### 実運用プロジェクト（構造確認）

- [ardanlabs/service](https://github.com/ardanlabs/service)
- [dotnet-architecture/eShopOnWeb](https://github.com/dotnet-architecture/eShopOnWeb)
- [citerus/dddsample-core](https://github.com/citerus/dddsample-core)
- [kgrzybek/modular-monolith-with-ddd](https://github.com/kgrzybek/modular-monolith-with-ddd)
- [ThreeDotsLabs/wild-workouts-go-ddd-example](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example)
- [VaughnVernon/IDDD_Samples](https://github.com/VaughnVernon/IDDD_Samples)
- [ddd-by-examples/library](https://github.com/ddd-by-examples/library)

### dddsample-core の位置付け確認

- [citerus dddsample 旧公式サイト](https://dddsample.sourceforge.net/) — 自己定義（"a decent way" ではあるが "the way" ではない）
- [ddd-crew/free-ddd-learning-resources](https://github.com/ddd-crew/free-ddd-learning-resources) — dddsample 非掲載
- [heynickc/awesome-ddd](https://github.com/heynickc/awesome-ddd)

### Bulletproof 相当調査

- [alan2207/bulletproof-react](https://github.com/alan2207/bulletproof-react) — React 界のデファクト参考

### 原典（再確認）

- [Martin Fowler - PresentationDomainDataLayering](https://martinfowler.com/bliki/PresentationDomainDataLayering.html)
- [Jimmy Bogard - Vertical Slice Architecture](https://www.jimmybogard.com/vertical-slice-architecture/)
- [Uncle Bob - Screaming Architecture](https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html)
- [Jeffrey Palermo - The Onion Architecture Part 1](https://jeffreypalermo.com/2008/07/the-onion-architecture-part-1/)
