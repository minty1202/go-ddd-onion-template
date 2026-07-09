package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/minty1202/go-ddd-onion-template/internal/domain/aggregate"
	"github.com/minty1202/go-ddd-onion-template/internal/infra/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAggregate struct {
	aggregate.Lock
}

func newMockAggregate(version int) *mockAggregate {
	return &mockAggregate{Lock: aggregate.ReconstructLock(version)}
}

func TestSaveWithLock_Success(t *testing.T) {
	agg := newMockAggregate(5)
	var capturedExpected int

	err := repository.SaveWithLock(context.Background(), agg, func(_ context.Context, expected int) (int, error) {
		capturedExpected = expected
		return 6, nil
	})

	require.NoError(t, err)
	assert.Equal(t, 5, capturedExpected)
	assert.Equal(t, 6, agg.Version())
}

func TestSaveWithLock_COnflict(t *testing.T) {
	agg := newMockAggregate(5)

	err := repository.SaveWithLock(context.Background(), agg, func(_ context.Context, expected int) (int, error) {
		return 0, aggregate.ErrConflict
	})

	require.ErrorIs(t, err, aggregate.ErrConflict)
	assert.Equal(t, 5, agg.Version())
}

func TestSaveWithLock_OtherError(t *testing.T) {
	agg := newMockAggregate(5)
	sentinelErr := errors.New("db error")

	err := repository.SaveWithLock(context.Background(), agg, func(_ context.Context, expected int) (int, error) {
		return 0, sentinelErr
	})

	require.ErrorIs(t, err, sentinelErr)
	assert.Equal(t, 5, agg.Version())
}
