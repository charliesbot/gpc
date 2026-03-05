package client

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"net/http"
	"time"

	"google.golang.org/api/googleapi"
)

const (
	baseDelay = 1 * time.Second
	maxDelay  = 30 * time.Second
)

// RetryableError is implemented by errors that can indicate whether a retry
// is appropriate. Types satisfying this interface allow the retry logic to
// distinguish transient failures from permanent ones.
type RetryableError interface {
	error
	Retryable() bool
}

// IsRetryable reports whether err should be retried. It returns true when:
//   - err satisfies RetryableError and its Retryable method returns true, or
//   - err is a googleapi.Error with an HTTP 429 or 5xx status code.
func IsRetryable(err error) bool {
	// Check the RetryableError interface first.
	var re RetryableError
	if errors.As(err, &re) {
		return re.Retryable()
	}

	// Fall back to inspecting Google API HTTP status codes.
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		if apiErr.Code == http.StatusTooManyRequests {
			return true
		}
		if apiErr.Code >= http.StatusInternalServerError {
			return true
		}
	}

	return false
}

// Retry calls fn up to maxRetries times when it returns a retryable error.
// It uses exponential backoff starting at 1 s and capped at 30 s, with random
// jitter added to each delay to avoid thundering-herd problems. The context
// can be used to cancel or time-out the retry loop.
func Retry(ctx context.Context, maxRetries int, fn func() error) error {
	var err error
	for attempt := range maxRetries + 1 {
		err = fn()
		if err == nil {
			return nil
		}

		if !IsRetryable(err) {
			return err
		}

		// Don't sleep after the last attempt.
		if attempt == maxRetries {
			break
		}

		delay := backoff(attempt)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

// backoff computes the delay for a given attempt number using exponential
// backoff with jitter. The base delay doubles on each attempt but is capped
// at maxDelay. A random jitter of [0, baseDelay) is then added.
func backoff(attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	delay := time.Duration(exp) * baseDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	jitter := time.Duration(rand.Int64N(int64(baseDelay)))
	delay += jitter

	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}
