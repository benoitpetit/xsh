package utils

import (
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
