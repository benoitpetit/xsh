package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const latestReleaseURL = "https://api.github.com/repos/benoitpetit/xsh/releases/latest"
const maxReleaseMetadata = 2 << 20
const maxBinarySize = 200 << 20

var releaseVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

type releaseAsset struct {
	Name   string
	URL    string
	Digest string
	Size   int64
}

func releaseAssetName(goos, goarch string) (string, error) {
	if goos != "linux" && goos != "darwin" && goos != "windows" {
		return "", fmt.Errorf("no xsh release for %s/%s", goos, goarch)
	}
	if goarch != "amd64" && goarch != "arm64" {
		return "", fmt.Errorf("no xsh release for %s/%s", goos, goarch)
	}
	name := "xsh-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return name, nil
}

func compareReleaseVersions(current, latest string) (int, error) {
	currentParts := releaseVersionPattern.FindStringSubmatch(current)
	latestParts := releaseVersionPattern.FindStringSubmatch(latest)
	if len(latestParts) == 0 {
		return 0, fmt.Errorf("invalid release version %q", latest)
	}
	if len(currentParts) == 0 { // Local and development builds can still update.
		return -1, nil
	}
	for i := 1; i <= 3; i++ {
		a, errA := strconv.ParseUint(currentParts[i], 10, 64)
		b, errB := strconv.ParseUint(latestParts[i], 10, 64)
		if errA != nil || errB != nil {
			return 0, errors.New("release version component is too large")
		}
		if a < b {
			return -1, nil
		}
		if a > b {
			return 1, nil
		}
	}
	return 0, nil
}

func selectReleaseAsset(release githubRelease, goos, goarch string) (releaseAsset, error) {
	if release.Draft || release.Prerelease {
		return releaseAsset{}, errors.New("latest release is not stable")
	}
	if _, err := compareReleaseVersions("0.0.0", release.TagName); err != nil {
		return releaseAsset{}, err
	}
	name, err := releaseAssetName(goos, goarch)
	if err != nil {
		return releaseAsset{}, err
	}
	for _, candidate := range release.Assets {
		if candidate.Name != name {
			continue
		}
		address, err := url.Parse(candidate.BrowserDownloadURL)
		if err != nil || address.Scheme != "https" || address.Host != "github.com" || address.RawQuery != "" || address.Fragment != "" || address.Path != "/benoitpetit/xsh/releases/download/"+release.TagName+"/"+name {
			return releaseAsset{}, fmt.Errorf("invalid release URL for %s", name)
		}
		digest := strings.TrimPrefix(candidate.Digest, "sha256:")
		if !strings.HasPrefix(candidate.Digest, "sha256:") || len(digest) != 64 {
			return releaseAsset{}, fmt.Errorf("missing SHA-256 digest for %s", name)
		}
		if _, err := hex.DecodeString(digest); err != nil {
			return releaseAsset{}, fmt.Errorf("invalid SHA-256 digest for %s: %w", name, err)
		}
		if candidate.Size <= 0 || candidate.Size > maxBinarySize {
			return releaseAsset{}, fmt.Errorf("invalid release size for %s", name)
		}
		return releaseAsset{Name: name, URL: candidate.BrowserDownloadURL, Digest: digest, Size: candidate.Size}, nil
	}
	return releaseAsset{}, fmt.Errorf("release %s has no binary for %s/%s", release.TagName, goos, goarch)
}

func releaseHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" || len(via) >= 10 {
				return errors.New("unsafe release redirect")
			}
			return nil
		},
	}
}

func fetchLatestRelease(ctx context.Context, client *http.Client, address string) (githubRelease, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return githubRelease{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "xsh-updater/"+Version)
	response, err := client.Do(request)
	if err != nil {
		return githubRelease{}, fmt.Errorf("fetch latest release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("fetch latest release: HTTP %d", response.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(response.Body, maxReleaseMetadata)).Decode(&release); err != nil {
		return githubRelease{}, fmt.Errorf("decode latest release: %w", err)
	}
	return release, nil
}

func downloadReleaseAsset(ctx context.Context, client *http.Client, asset releaseAsset, target string) (err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "xsh-updater/"+Version)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("download release: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download release: HTTP %d", response.StatusCode)
	}
	if response.ContentLength > asset.Size {
		return errors.New("release download exceeds advertised size")
	}
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(response.Body, asset.Size+1))
	if err != nil {
		return err
	}
	if n != asset.Size {
		return fmt.Errorf("release download size mismatch: got %d, want %d", n, asset.Size)
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), asset.Digest) {
		return errors.New("release SHA-256 digest mismatch")
	}
	if err := file.Chmod(0755); err != nil {
		return err
	}
	return file.Sync()
}

func installReleaseAsset(ctx context.Context, client *http.Client, asset releaseAsset, executable, goos string) error {
	// Staging beside the executable keeps the final rename on the same filesystem.
	staged, err := os.CreateTemp(filepath.Dir(executable), ".xsh-update-*")
	if err != nil {
		return fmt.Errorf("stage update beside %s: %w", executable, err)
	}
	stagedPath := staged.Name()
	staged.Close()
	defer os.Remove(stagedPath)
	if err := downloadReleaseAsset(ctx, client, asset, stagedPath); err != nil {
		return err
	}
	if goos != "windows" {
		return os.Rename(stagedPath, executable)
	}
	// Windows keeps the running executable open. Move it aside, then place the
	// verified binary at its path; restore it if the second rename fails.
	backup := executable + ".old"
	if _, err := os.Stat(backup); err == nil {
		if err := os.Remove(backup); err != nil {
			return fmt.Errorf("remove stale backup %s: %w", backup, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(executable, backup); err != nil {
		return fmt.Errorf("move running executable: %w", err)
	}
	if err := os.Rename(stagedPath, executable); err != nil {
		if restoreErr := os.Rename(backup, executable); restoreErr != nil {
			return fmt.Errorf("install update: %w (restore failed: %v; original at %s)", err, restoreErr, backup)
		}
		return fmt.Errorf("install update: %w", err)
	}
	_ = os.Remove(backup) // May remain locked until this process exits.
	return nil
}

var updateCheckOnly bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update xsh to the latest compatible release",
	Long:  "Download the latest stable xsh release for this operating system and architecture, verify its SHA-256 digest, and replace the current executable.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := releaseAssetName(runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return err
		}
		client := releaseHTTPClient()
		release, err := fetchLatestRelease(cmd.Context(), client, latestReleaseURL)
		if err != nil {
			return err
		}
		asset, err := selectReleaseAsset(release, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return err
		}
		comparison, err := compareReleaseVersions(Version, release.TagName)
		if err != nil {
			return err
		}
		status := "available"
		if comparison >= 0 {
			status = "up-to-date"
		}
		if status == "available" && !updateCheckOnly {
			executable, err := os.Executable()
			if err != nil {
				return err
			}
			executable, err = filepath.EvalSymlinks(executable)
			if err != nil {
				return err
			}
			if err := installReleaseAsset(cmd.Context(), client, asset, executable, runtime.GOOS); err != nil {
				return fmt.Errorf("update %s: %w", name, err)
			}
			status = "updated"
		}
		return output(map[string]string{"status": status, "current_version": Version, "latest_version": release.TagName, "asset": name}, func() {
			_, _ = fmt.Fprintf(runtimeOutput(), "xsh %s: %s → %s (%s)\n", status, Version, release.TagName, name)
		})
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "Check the latest compatible release without installing it")
}
