// Package utils provides utility functions for xsh.
package utils

import (
	"math"
	"math/rand"
	"time"
)

// Delay sleeps with randomized jitter
// Actual delay = base * uniform(1 - jitterFactor, 1 + jitterFactor)
func Delay(base float64, jitterFactor float64) {
	if base <= 0 {
		base = 1.5
	}
	if jitterFactor <= 0 {
		jitterFactor = 0.5
	}
	
	min := base * (1 - jitterFactor)
	max := base * (1 + jitterFactor)
	actual := min + rand.Float64()*(max-min)
	
	if actual < 0.1 {
		actual = 0.1
	}
	
	time.Sleep(time.Duration(actual * float64(time.Second)))
}

// WriteDelay sleeps for a random duration appropriate for write operations
func WriteDelay() {
	minDelay := 1.5
	maxDelay := 4.0
	actual := minDelay + rand.Float64()*(maxDelay-minDelay)
	time.Sleep(time.Duration(actual * float64(time.Second)))
}

// BackoffDelay performs exponential backoff with jitter
func BackoffDelay(attempt int, base float64, maxDelay float64) {
	if base <= 0 {
		base = 2.0
	}
	if maxDelay <= 0 {
		maxDelay = 60.0
	}
	
	wait := math.Min(base*math.Pow(2, float64(attempt))+rand.Float64(), maxDelay)
	time.Sleep(time.Duration(wait * float64(time.Second)))
}
