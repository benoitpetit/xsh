package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseAssetName(t *testing.T) {
	for _, test := range []struct{ os, arch, want string }{
		{"linux", "amd64", "xsh-linux-amd64"},
		{"darwin", "arm64", "xsh-darwin-arm64"},
		{"windows", "amd64", "xsh-windows-amd64.exe"},
	} {
		got, err := releaseAssetName(test.os, test.arch)
		if err != nil || got != test.want {
			t.Fatalf("releaseAssetName(%q, %q) = %q, %v", test.os, test.arch, got, err)
		}
	}
	if _, err := releaseAssetName("freebsd", "amd64"); err == nil {
		t.Fatal("unsupported platform accepted")
	}
}

func TestCompareReleaseVersions(t *testing.T) {
	for _, test := range []struct {
		current, latest string
		want            int
	}{
		{"0.1.1", "v0.1.2", -1},
		{"v0.1.1", "0.1.1", 0},
		{"0.2.0", "v0.1.9", 1},
		{"dev", "v0.1.1", -1},
	} {
		got, err := compareReleaseVersions(test.current, test.latest)
		if err != nil || got != test.want {
			t.Fatalf("compareReleaseVersions(%q, %q) = %d, %v", test.current, test.latest, got, err)
		}
	}
	if _, err := compareReleaseVersions("0.1.1", "latest"); err == nil {
		t.Fatal("invalid release tag accepted")
	}
}

func TestSelectReleaseAsset(t *testing.T) {
	release := githubRelease{TagName: "v0.1.2"}
	release.Assets = append(release.Assets, struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
		Size               int64  `json:"size"`
	}{"xsh-linux-amd64", "https://github.com/benoitpetit/xsh/releases/download/v0.1.2/xsh-linux-amd64", "sha256:" + strings.Repeat("a", 64), 42})
	if asset, err := selectReleaseAsset(release, "linux", "amd64"); err != nil || asset.Size != 42 {
		t.Fatalf("selectReleaseAsset() = %#v, %v", asset, err)
	}
	release.Assets[0].Digest = ""
	if _, err := selectReleaseAsset(release, "linux", "amd64"); err == nil {
		t.Fatal("missing digest accepted")
	}
	release.Assets[0].Digest = "sha256:" + strings.Repeat("a", 64)
	release.Assets[0].BrowserDownloadURL = "https://example.org/xsh-linux-amd64"
	if _, err := selectReleaseAsset(release, "linux", "amd64"); err == nil {
		t.Fatal("untrusted asset URL accepted")
	}
}

func TestFetchLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"tag_name":"v0.1.2","assets":[]}`)
	}))
	defer server.Close()
	release, err := fetchLatestRelease(context.Background(), server.Client(), server.URL)
	if err != nil || release.TagName != "v0.1.2" {
		t.Fatalf("fetchLatestRelease() = %#v, %v", release, err)
	}
}

func TestInstallReleaseAssetVerifiesBeforeReplacement(t *testing.T) {
	content := []byte("new executable")
	sum := sha256.Sum256(content)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer server.Close()
	dir := t.TempDir()
	executable := filepath.Join(dir, "xsh")
	if err := os.WriteFile(executable, []byte("old executable"), 0755); err != nil {
		t.Fatal(err)
	}
	asset := releaseAsset{URL: server.URL, Size: int64(len(content)), Digest: hex.EncodeToString(sum[:])}
	asset.Digest = strings.Repeat("0", 64)
	if err := installReleaseAsset(context.Background(), server.Client(), asset, executable, "linux"); err == nil {
		t.Fatal("tampered download accepted")
	}
	got, err := os.ReadFile(executable)
	if err != nil || string(got) != "old executable" {
		t.Fatalf("executable after failed update = %q, %v", got, err)
	}
	asset.Digest = hex.EncodeToString(sum[:])
	if err := installReleaseAsset(context.Background(), server.Client(), asset, executable, "linux"); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(executable)
	if err != nil || string(got) != string(content) {
		t.Fatalf("updated executable = %q, %v", got, err)
	}
}
