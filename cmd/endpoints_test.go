package cmd

import (
	"os"
	"testing"
)

func TestEndpointDiscoveryStateSeparatesDiscoveryFromVerification(t *testing.T) {
	tests := []struct {
		name        string
		dynamic     bool
		static      bool
		quarantined bool
		want        string
	}{
		{"dynamic", true, true, false, "discovered"},
		{"static fallback", false, true, false, "static_fallback"},
		{"quarantined", true, true, true, "quarantined"},
		{"missing", false, false, false, "unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointDiscoveryState(tt.dynamic, tt.static, tt.quarantined); got != tt.want {
				t.Fatalf("endpointDiscoveryState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEndpointRefreshProgressOnlyInHumanMode(t *testing.T) {
	oldJSON, oldYAML, oldCompact := jsonOutput, yamlOutput, compactMode
	defer func() { jsonOutput, yamlOutput, compactMode = oldJSON, oldYAML, oldCompact }()
	compactMode = false

	jsonOutput, yamlOutput = true, false
	if showEndpointRefreshProgress() {
		t.Fatal("JSON refresh would mix progress text with its payload")
	}
	jsonOutput, yamlOutput = false, true
	if showEndpointRefreshProgress() {
		t.Fatal("YAML refresh would mix progress text with its payload")
	}
	jsonOutput, yamlOutput = false, false
	device, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer device.Close()
	withRuntime(&Runtime{Out: device}, func() {
		if !showEndpointRefreshProgress() {
			t.Fatal("human refresh lost its progress text")
		}
	})
}
