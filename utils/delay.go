// Package utils provides helper utilities.
package utils

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// SleepWithContext waits for d or returns when ctx is canceled.
func SleepWithContext(ctx context.Context, d time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if d <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

const (
	defaultDelaySec  = 1.5
	minWriteDelaySec = 1.5
	maxWriteDelaySec = 4.0
	baseBackoffSec   = 1.0
	maxBackoffSec    = 30.0
)

// Delay sleeps for a random duration between minSec and maxSec seconds.
// If both are 0, the default read delay is used.
func Delay(minSec, maxSec float64) {
	if minSec == 0 && maxSec == 0 {
		minSec = defaultDelaySec
		maxSec = defaultDelaySec
	}
	if minSec >= maxSec {
		time.Sleep(time.Duration(minSec * float64(time.Second)))
		return
	}
	d := minSec + rand.Float64()*(maxSec-minSec)
	time.Sleep(time.Duration(d * float64(time.Second)))
}

// DelaySeconds sleeps for an explicitly configured delay. Unlike Delay(0, 0),
// a zero value here means no delay rather than the default read delay.
func DelaySeconds(seconds float64) {
	if seconds <= 0 {
		return
	}
	_ = SleepWithContext(context.Background(), time.Duration(seconds*float64(time.Second)))
}

// WriteDelay sleeps for a random duration appropriate for write operations.
func WriteDelay() {
	Delay(minWriteDelaySec, maxWriteDelaySec)
}

// BackoffDelay sleeps for an exponential backoff duration based on the attempt number.
// If minSec and maxSec are both 0, default bounds are used.
func BackoffDelay(attempt int, minSec, maxSec float64) {
	_ = BackoffDelayWithContext(context.Background(), attempt, minSec, maxSec)
}

// BackoffDelayWithContext is the cancellable form used by request paths.
func BackoffDelayWithContext(ctx context.Context, attempt int, minSec, maxSec float64) error {
	if minSec == 0 && maxSec == 0 {
		minSec = baseBackoffSec
		maxSec = maxBackoffSec
	}
	backoff := minSec * math.Pow(2, float64(attempt))
	if backoff > maxSec {
		backoff = maxSec
	}
	// Add jitter: ±20%
	jitter := backoff * 0.2 * (rand.Float64()*2 - 1)
	d := backoff + jitter
	if d < minSec {
		d = minSec
	}
	return SleepWithContext(ctx, time.Duration(d*float64(time.Second)))
}
