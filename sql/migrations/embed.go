package migrations

import "embed"

// FS は //go:embed ディレクティブによって sql/migrations/*.sql の内容を
// ビルド時に Go バイナリへ埋め込んだもの。
// 実行時はディスク上のファイルを開くのではなく、バイナリに含まれた
// データをそのまま参照する。
//
// 用途: dbtest がテスト用 PG コンテナにマイグレーションを当てる際、
// goose の Go API にこの FS を渡すことで SQL を読み込ませる。
//
//go:embed *.sql
var FS embed.FS
