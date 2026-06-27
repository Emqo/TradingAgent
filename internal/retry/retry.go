package retry

import (
	"context"
	"fmt"
	"time"
)

// Config holds retry configuration.
type Config struct {
	// MaxRetries is the maximum number of retries.
	MaxRetries int
	// InitialBackoff is the initial backoff duration.
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration.
	MaxBackoff time.Duration
	// BackoffMultiplier is the multiplier for backoff.
	BackoffMultiplier float64
}

// DefaultConfig returns the default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc[T any] func() (T, error)

// Do executes a function with retry logic.
func Do[T any](ctx context.Context, config Config, fn RetryableFunc[T]) (T, error) {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Check context
		select {
		case <-ctx.Done():
			var zero T
			return zero, ctx.Err()
		default:
		}

		// Execute function
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't sleep on last attempt
		if attempt < config.MaxRetries {
			// Sleep with backoff
			select {
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			case <-time.After(backoff):
			}

			// Increase backoff
			backoff = time.Duration(float64(backoff) * config.BackoffMultiplier)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}
	}

	var zero T
	return zero, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DoSimple executes a simple function with retry logic (no return value).
func DoSimple(ctx context.Context, config Config, fn func() error) error {
	_, err := Do(ctx, config, func() (any, error) {
		return nil, fn()
	})
	return err
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	// Add custom logic here to determine if an error is retryable
	// For now, all errors are retryable
	return true
}
