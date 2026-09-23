package cmd

import (
	"testing"

	"github.com/benoitpetit/xsh/core"
)

func TestEndpointDoctorWarningProvidesRepairCommand(t *testing.T) {
	result := endpointCheckResult(core.EndpointStats{
		TotalCount:  64,
		StaticCount: 64,
	})

	if result.Status != "warn" {
		t.Fatalf("checkEndpoints() status = %q, want warn in the test environment", result.Status)
	}
	if result.Remediation != "xsh endpoints refresh" {
		t.Fatalf("checkEndpoints() remediation = %q, want %q", result.Remediation, "xsh endpoints refresh")
	}
}
