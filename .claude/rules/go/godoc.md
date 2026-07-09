---
paths:
  - "**/*.go"
---
# godoc コメント

- **公開識別子** (大文字始まりの関数 / 型 / 定数 / 変数) には godoc を書く
- **godoc は日本語で書く**
- godoc の慣習は維持する: 先頭は識別子名で始める / 句点で終わる
- 内部識別子 (小文字始まり) の godoc は任意。振る舞いの細部や非自明な仕様がある場合のみ書く
- **テスト関数 (`Test*` / `Benchmark*` / `Example*`) は godoc 不要**。テスト名 (`TestX_Y_Z` パターン) で十分自己説明的。テスト内の非自明な「なぜ」はインラインコメントで残す

## 例

```go
// NewRecovery は、下流の interceptor および handler で発生した panic から
// 回復する Connect interceptor を返す。unary / streaming サーバ RPC の両方
// をカバーする。
//
// http.ErrAbortHandler は、net/http の abort セマンティクスを保つために
// 再 panic する。
func NewRecovery() connect.Interceptor {
	// ...
}
```
