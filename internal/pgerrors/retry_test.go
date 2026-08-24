package pgerrors

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithRetrySucceedsImmediately(t *testing.T) {
	calls := 0

	err := WithRetry(context.Background(), func() error {
		calls++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestWithRetryStopsOnNonRetriable(t *testing.T) {
	calls := 0
	want := errors.New("синтаксическая ошибка")

	err := WithRetry(context.Background(), func() error {
		calls++
		return want
	})

	assert.ErrorIs(t, err, want)
	assert.Equal(t, 1, calls, "неповторяемая ошибка не повторяется")
}

func TestWithRetryStopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	want := &pgconn.PgError{Code: pgerrcode.ConnectionFailure}

	start := time.Now()
	err := WithRetry(ctx, func() error {
		calls++
		return want
	})

	assert.ErrorIs(t, err, want)
	assert.Equal(t, 1, calls)
	assert.Less(t, time.Since(start), time.Second, "отменённый контекст не ждёт паузу")
}
