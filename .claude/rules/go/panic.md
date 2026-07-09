---
paths:
  - "**/*.go"
---
# panic の方針

- 外部入力・環境に依存するエラー → error を返す
- 到達したらバグ（プログラムのロジック的にありえない）→ panic
  - switch の網羅後の default
  - Must パターン（入力がハードコードで失敗しないはず）
  - API 契約に対する明確な違反（nil を渡してはいけない関数に nil が来た等）

## 例

```go
// OK: switch で全ケースを網羅した後の default
switch e.Kind {
case errdefs.InvalidArgument:
    return http.StatusBadRequest
case errdefs.NotFound:
    return http.StatusNotFound
case errdefs.FailedPrecondition:
    return http.StatusConflict
case errdefs.AlreadyExists:
    return http.StatusConflict
case errdefs.Internal:
    return http.StatusInternalServerError
default:
    panic("unreachable")
}

// OK: 到達したらバグ（想定される固定エラー値を全て列挙済み）
switch {
case errors.Is(err, todo.ErrInvalid):
    return errdefs.NewInvalidArgument(err)
case errors.Is(err, todo.ErrAlreadyCompleted):
    return errdefs.NewFailedPrecondition(err)
default:
    panic("unreachable")
}

// OK: Must パターン（ハードコードの入力が失敗しないはず）
var nameRe = regexp.MustCompile(`^[a-z]+$`)

// NG: 外部入力に依存するエラーで panic してはいけない
err := validate.Struct(t)
if err != nil {
    // panic("validation failed") ← NG
    return nil, ErrInvalid // ← OK（固定エラー値に畳む）
}
```
