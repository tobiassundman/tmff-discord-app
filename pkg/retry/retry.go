package retry

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

// Retry retries the given operation until it succeeds or times out.
func Retry(timeout time.Duration, operation func() error) error {
	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.MaxElapsedTime = timeout
	//nolint:mnd // This is a magic number for the maximum interval of the exponential backoff.
	exponentialBackoff.MaxInterval = time.Second * 5

	return backoff.Retry(operation, exponentialBackoff)
}
