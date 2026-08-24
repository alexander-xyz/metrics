package pgerrors

import (
	"context"
	"time"
)

var retryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

func WithRetry(ctx context.Context, operation func() error) error {
	classifier := NewPostgresErrorClassifier()

	err := operation()
	if err == nil {
		return nil
	}

	for _, delay := range retryDelays {
		if classifier.Classify(err) == NonRetriable {
			return err
		}

		select {
		case <-ctx.Done():
			return err
		case <-time.After(delay):
		}

		if err = operation(); err == nil {
			return nil
		}
	}

	return err
}
