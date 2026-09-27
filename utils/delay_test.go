package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDelaySecondsZeroReturnsImmediately(t *testing.T) {
	started := time.Now()
	DelaySeconds(0)
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("zero delay took %s", elapsed)
	}
}

func TestSleepWithContextReturnsImmediatelyForZero(t *testing.T) {
	start := time.Now()
	if err := SleepWithContext(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("zero delay took too long: %s", time.Since(start))
	}
}

func TestSleepWithContextHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SleepWithContext(ctx, time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatalf("SleepWithContext() error = %v, want context.Canceled", err)
	}
}
