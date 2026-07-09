package aggregate_test

import (
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/aggregate"
	"github.com/stretchr/testify/assert"
)

func TestNewLock(t *testing.T) {
	lock := aggregate.NewLock()
	assert.Equal(t, 0, lock.Version())
}

func TestTeconstructLock(t *testing.T) {
	lock := aggregate.ReconstructLock(5)
	assert.Equal(t, 5, lock.Version())
}

func TestLock_SyncVersion(t *testing.T) {
	lock := aggregate.NewLock()
	assert.Equal(t, 0, lock.Version())

	lock.SyncVersion(7)
	assert.Equal(t, 7, lock.Version())

	lock.SyncVersion(10)
	assert.Equal(t, 10, lock.Version())
}
