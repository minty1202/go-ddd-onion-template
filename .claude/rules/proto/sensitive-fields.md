---
paths:
  - "**/*.proto"
---
# センシティブフィールドのマーキング

proto に **パスワード / API トークン / セッショントークン / 認証ヘッダ / 基本的な PII (氏名・メアド・電話番号 等)** を追加する際は、フィールドオプション `[debug_redact = true]` を付ける（`google.protobuf.FieldOptions` の標準オプション、protoc v22 / 2023 年導入）。

## 例

```proto
message LoginRequest {
  string username = 1;
  string password = 2 [debug_redact = true];
}
```

## マスクの実体は protobuf 標準に任せる

マスクは **protobuf 公式 formatter** (`prototext.Format` / `protojson.Marshal`) が `debug_redact = true` のフィールドを自動で `[REDACTED]` 化する。

- 自前の redact ロジックは書かない
- 本体デバッグログが必要な箇所では `prototext.Format(msg)` を直接呼ぶだけでよい (= 自動マスクされる)

## レビュー観点

proto に新フィールドを追加する PR では、機密性のあるフィールドに `[debug_redact = true]` が付いているか確認する。

## カバー範囲外

- マスク戦略の細分化 (hash 化 / 末尾 N 文字残す等)
- PII 区分階層 (GDPR 厳格対応 / 金融・医療系)
- 決済情報 — Stripe 等の決済プロバイダに委譲する設計のためアプリでは保持しない

要件が出たら独自オプション併用で対応する。
