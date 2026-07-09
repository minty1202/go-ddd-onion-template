// Package aggregate は集約共通の楽観ロック仕組みを提供する。
package aggregate

import "errors"

// ErrConflict は楽観ロック競合を表す。取得時の version と現在 DB の version が一致しない場合に発生する。
var ErrConflict = errors.New("aggregate: conflict")

// Aggregate は楽観ロック対応集約が満たすべき契約。
// SaveWithLock 等の共通ヘルパーが引数として要求することで、
// Lock を持たない集約をコンパイル時に弾く。
type Aggregate interface {
	Version() int
	SyncVersion(v int)
}

// Lock は楽観ロックの version 管理を共通化する。
// 各集約は private 名前付きフィールド (例: lock aggregate.Lock) で保持し、
// Version() / SyncVersion(v) を集約側で手動公開する。
type Lock struct {
	version int
}

// NewLock は新規生成時の Lock を返す。
func NewLock() Lock { return Lock{version: 0} }

// ReconstructLock は永続化された値から Lock を組み立てる。Repository 実装専用。
func ReconstructLock(v int) Lock { return Lock{version: v} }

func (l *Lock) Version() int      { return l.version }
func (l *Lock) SyncVersion(v int) { l.version = v }
