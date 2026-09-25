package core

import (
	"testing"
)

func TestNewXClientWithRequestConfigPropagatesReadDelayAndTimeout(t *testing.T) {
	client, err := NewXClientWithRequestConfig(nil, "", "", RequestConfig{Delay: 0, Timeout: 12})
	if err != nil {
		t.Fatal(err)
	}
	if !client.readDelayConfigured || client.readDelay != 0 {
		t.Fatalf("read delay configured=%v value=%v, want true/0", client.readDelayConfigured, client.readDelay)
	}
	if client.requestTimeout.Seconds() != 12 {
		t.Fatalf("request timeout = %s, want 12s", client.requestTimeout)
	}
}

func TestApplyReadDelayHonorsExplicitZero(t *testing.T) {
	client, err := NewXClientWithRequestConfig(nil, "", "", RequestConfig{Delay: 0, Timeout: 30})
	if err != nil {
		t.Fatal(err)
	}
	var got float64
	client.readDelayFunc = func(seconds float64) { got = seconds }

	client.applyReadDelay()

	if got != 0 {
		t.Fatalf("read delay = %v, want explicit zero", got)
	}
}
