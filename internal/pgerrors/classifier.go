// Package pgerrors классифицирует ошибки PostgreSQL и повторяет операции,
// завершившиеся повторяемой ошибкой.
package pgerrors

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PGErrorClassification — признак того, имеет ли смысл повторять операцию.
type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

// PostgresErrorClassifier относит ошибки PostgreSQL к повторяемым или нет.
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier создаёт классификатор ошибок PostgreSQL.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return ClassifyPgError(pgErr)
	}

	return NonRetriable
}

// ClassifyPgError относит ошибку PostgreSQL к повторяемым, если она вызвана
// проблемами соединения (класс 08).
func ClassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure,
		pgerrcode.SQLClientUnableToEstablishSQLConnection,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection,
		pgerrcode.TransactionResolutionUnknown,
		pgerrcode.ProtocolViolation:
		return Retriable
	}

	return NonRetriable
}
