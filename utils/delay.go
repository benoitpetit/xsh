// Package utils provides helper utilities.
package utils

import (
	"math"
	"math/rand"
	"time"
)

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

// WriteDelay sleeps for a random duration appropriate for write operations.
func WriteDelay() {
	Delay(minWriteDelaySec, maxWriteDelaySec)
}

// BackoffDelay sleeps for an exponential backoff duration based on the attempt number.
// If minSec and maxSec are both 0, default bounds are used.
func BackoffDelay(attempt int, minSec, maxSec float64) {
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
	time.Sleep(time.Duration(d * float64(time.Second)))
}
