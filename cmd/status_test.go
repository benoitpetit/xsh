package cmd

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestConnectivityStatusLabelDistinguishesSkippedFromFailure(t *testing.T) {
	if got := connectivityStatusLabel(connectivityStatus{}); got != "Not checked" {
		t.Fatalf("unchecked connectivity label = %q, want %q", got, "Not checked")
	}
	if got := connectivityStatusLabel(connectivityStatus{Checked: true}); got != "Issues" {
		t.Fatalf("failed connectivity label = %q, want %q", got, "Issues")
	}
	if got := connectivityStatusLabel(connectivityStatus{Checked: true, CanReachX: true, DiscoveryChecked: true, DiscoveryWorks: true}); got != "OK" {
		t.Fatalf("successful connectivity label = %q, want %q", got, "OK")
	}
}

func TestLocalStatusSkipsConnectivityAndEndpointChecks(t *testing.T) {
	checks := statusChecks{
		connectivity: func(context.Context) connectivityStatus {
			t.Fatal("local status ran connectivity check")
			return connectivityStatus{}
		},
		endpointHealth: func(context.Context) (bool, []string) {
			t.Fatal("local status ran endpoint check")
			return false, nil
		},
	}

	status := collectSystemStatusWithChecks(true, false, checks)
	if status.Connectivity.Checked {
		t.Fatal("local status marked connectivity as checked")
	}
	if !strings.Contains(status.EndpointHealth.Message, "skipped") {
		t.Fatalf("local endpoint message = %q, want it to say the check was skipped", status.EndpointHealth.Message)
	}
}

func TestProbeConnectivityUsesLiveHTTPResponseAndDoesNotClaimDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Errorf("request method = %q, want HEAD", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	status := probeConnectivity(context.Background(), server.Client(), server.URL)
	if !status.Checked || !status.CanReachX {
		t.Fatalf("successful probe status = %#v", status)
	}
	if status.DiscoveryChecked || status.DiscoveryWorks {
		t.Fatalf("reachability probe claimed endpoint discovery was checked: %#v", status)
	}
}

func TestDiscoveryNetworkErrorUsesErrorTypes(t *testing.T) {
	if !isNetworkProbeError(&net.OpError{Op: "dial", Net: "tcp", Err: errors.New("transport failed")}) {
		t.Fatal("typed network error was not recognized")
	}
	if !isNetworkProbeError(&url.Error{Op: "Head", URL: "https://x.com/", Err: &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("transport failed")}}) {
		t.Fatal("wrapped transport error was not recognized")
	}
	if isNetworkProbeError(&url.Error{Op: "Head", URL: "https://x.com/", Err: errors.New("stopped after redirect limit")}) {
		t.Fatal("non-network URL error was classified as a network error")
	}
	if isNetworkProbeError(errors.New("server reported: connection refused")) {
		t.Fatal("plain error text was classified as a network error")
	}
}
