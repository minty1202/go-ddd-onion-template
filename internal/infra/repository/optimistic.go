// Package repository は Repository 実装の共通ヘルパーを提供する。
package repository

import (
	"context"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/aggregate"
)

// SaveWithLock は楽観ロック付き Save の共通ヘルパー。
//
// 引数:
//   - agg: aggregate.Aggregate を満たす集約。Lock を持たない型はコンパイル時に弾かれる
//   - persistFn: WHERE id = ? AND version = ? + SET version = version + 1 + RETURNING version 相当の
//     SQL を実行し、新しい version を返す関数。競合時 (影響行数 0 など) は aggregate.ErrConflict を返す
//
// Save 成功時、副作用として agg の version を新しい値に同期する。
// 各 Repository 実装はこのヘルパー経由で Save を実装することで、SyncVersion 呼び忘れを構造的に防ぐ。
func SaveWithLock(
	ctx context.Context,
	agg aggregate.Aggregate,
	persistFn func(ctx context.Context, expectedVersion int) (newVersion int, err error),
) error {
	expected := agg.Version()
	newVersion, err := persistFn(ctx, expected)
	if err != nil {
		return err
	}
	agg.SyncVersion(newVersion)
	return nil
}
