package pgerrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestClassify(t *testing.T) {
	classifier := NewPostgresErrorClassifier()

	testCases := []struct {
		name string
		err  error
		want PGErrorClassification
	}{
		{name: "nil", err: nil, want: NonRetriable},
		{name: "plain error", err: errors.New("boom"), want: NonRetriable},
		{
			name: "connection exception is retriable",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionException},
			want: Retriable,
		},
		{
			name: "connection failure is retriable",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionFailure},
			want: Retriable,
		},
		{
			name: "wrapped connection failure is retriable",
			err:  fmt.Errorf("query: %w", &pgconn.PgError{Code: pgerrcode.ConnectionFailure}),
			want: Retriable,
		},
		{
			name: "unique violation is not retriable",
			err:  &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			want: NonRetriable,
		},
		{
			name: "syntax error is not retriable",
			err:  &pgconn.PgError{Code: pgerrcode.SyntaxError},
			want: NonRetriable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, classifier.Classify(tc.err))
		})
	}
}
