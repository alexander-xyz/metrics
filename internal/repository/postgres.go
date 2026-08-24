package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	models "github.com/alexander-xyz/metrics/internal/model"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (store *PostgresStorage) UpdateGauge(ctx context.Context, name string, value Gauge) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET mtype = EXCLUDED.mtype, value = EXCLUDED.value`,
		name, models.Gauge, float64(value))
	if err != nil {
		return fmt.Errorf("update gauge %s: %w", name, err)
	}

	return nil
}

func (store *PostgresStorage) UpdateCounter(ctx context.Context, name string, value Counter) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET mtype = EXCLUDED.mtype, delta = metrics.delta + EXCLUDED.delta`,
		name, models.Counter, int64(value))
	if err != nil {
		return fmt.Errorf("update counter %s: %w", name, err)
	}

	return nil
}

func (store *PostgresStorage) SetCounter(ctx context.Context, name string, value Counter) error {
	_, err := store.db.ExecContext(ctx, `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET mtype = EXCLUDED.mtype, delta = EXCLUDED.delta`,
		name, models.Counter, int64(value))
	if err != nil {
		return fmt.Errorf("set counter %s: %w", name, err)
	}

	return nil
}

func (store *PostgresStorage) GetGauge(ctx context.Context, name string) (Gauge, error) {
	var value sql.NullFloat64

	row := store.db.QueryRowContext(ctx,
		`SELECT value FROM metrics WHERE id = $1 AND mtype = $2`, name, models.Gauge)
	if err := row.Scan(&value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}

		return 0, fmt.Errorf("get gauge %s: %w", name, err)
	}

	if !value.Valid {
		return 0, ErrNotFound
	}

	return Gauge(value.Float64), nil
}

func (store *PostgresStorage) GetCounter(ctx context.Context, name string) (Counter, error) {
	var delta sql.NullInt64

	row := store.db.QueryRowContext(ctx,
		`SELECT delta FROM metrics WHERE id = $1 AND mtype = $2`, name, models.Counter)
	if err := row.Scan(&delta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}

		return 0, fmt.Errorf("get counter %s: %w", name, err)
	}

	if !delta.Valid {
		return 0, ErrNotFound
	}

	return Counter(delta.Int64), nil
}

func (store *PostgresStorage) GetGauges(ctx context.Context) (map[string]Gauge, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT id, value FROM metrics WHERE mtype = $1 AND value IS NOT NULL`, models.Gauge)
	if err != nil {
		return nil, fmt.Errorf("get gauges: %w", err)
	}
	defer rows.Close()

	gauges := map[string]Gauge{}

	for rows.Next() {
		var (
			name  string
			value float64
		)

		if err := rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("scan gauge: %w", err)
		}

		gauges[name] = Gauge(value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gauges: %w", err)
	}

	return gauges, nil
}

func (store *PostgresStorage) GetCounters(ctx context.Context) (map[string]Counter, error) {
	rows, err := store.db.QueryContext(ctx,
		`SELECT id, delta FROM metrics WHERE mtype = $1 AND delta IS NOT NULL`, models.Counter)
	if err != nil {
		return nil, fmt.Errorf("get counters: %w", err)
	}
	defer rows.Close()

	counters := map[string]Counter{}

	for rows.Next() {
		var (
			name  string
			delta int64
		)

		if err := rows.Scan(&name, &delta); err != nil {
			return nil, fmt.Errorf("scan counter: %w", err)
		}

		counters[name] = Counter(delta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate counters: %w", err)
	}

	return counters, nil
}

func (store *PostgresStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	gauge, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (id, mtype, value) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET mtype = EXCLUDED.mtype, value = EXCLUDED.value`)
	if err != nil {
		return fmt.Errorf("prepare gauge statement: %w", err)
	}
	defer gauge.Close()

	counter, err := tx.PrepareContext(ctx, `
		INSERT INTO metrics (id, mtype, delta) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET mtype = EXCLUDED.mtype, delta = metrics.delta + EXCLUDED.delta`)
	if err != nil {
		return fmt.Errorf("prepare counter statement: %w", err)
	}
	defer counter.Close()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				continue
			}

			if _, err := gauge.ExecContext(ctx, metric.ID, models.Gauge, *metric.Value); err != nil {
				return fmt.Errorf("batch update gauge %s: %w", metric.ID, err)
			}
		case models.Counter:
			if metric.Delta == nil {
				continue
			}

			if _, err := counter.ExecContext(ctx, metric.ID, models.Counter, *metric.Delta); err != nil {
				return fmt.Errorf("batch update counter %s: %w", metric.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
