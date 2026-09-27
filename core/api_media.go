// Package core provides media upload and download operations for Twitter/X.
package core

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	// UploadURL is the Twitter media upload endpoint
	UploadURL = "https://upload.twitter.com/i/media/upload.json"
	// MaxImageSize is the maximum allowed image size (5 MB)
	MaxImageSize = 5 * 1024 * 1024
)

// Allowed image MIME types
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// MediaUploadResult contains the result of a media upload
type MediaUploadResult struct {
	MediaID          string `json:"media_id"`
	MediaIDString    string `json:"media_id_string"`
	Size             int64  `json:"size"`
	ExpiresAfterSecs int    `json:"expires_after_secs,omitempty"`
}

// ValidateMediaFile validates a media file for upload
func ValidateMediaFile(filePath string) (int64, string, error) {
	// Check file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, "", fmt.Errorf("file not found: %s", filePath)
	}

	if info.IsDir() {
		return 0, "", fmt.Errorf("path is a directory: %s", filePath)
	}

	// Check file size
	if info.Size() > MaxImageSize {
		return 0, "", fmt.Errorf("file too large: %.1fMB (max %dMB)",
			float64(info.Size())/(1024*1024), MaxImageSize/(1024*1024))
	}

	if info.Size() == 0 {
		return 0, "", fmt.Errorf("file is empty: %s", filePath)
	}

	// Detect MIME type
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)

	// Fallback: try to detect from content
	if mimeType == "" {
		file, err := os.Open(filePath)
		if err != nil {
			return 0, "", err
		}
		defer file.Close()

		// Read first 512 bytes for detection
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return 0, "", err
		}
		mimeType = http.DetectContentType(buffer[:n])
	}

	// Normalize MIME type
	mimeType = strings.Split(mimeType, ";")[0]

	if !AllowedImageTypes[mimeType] {
		return 0, "", fmt.Errorf("unsupported image format: %s (allowed: jpeg, png, gif, webp)", mimeType)
	}

	return info.Size(), mimeType, nil
}

// UploadMediaFile uploads an image to Twitter/X and returns the media ID
// Uses the chunked upload protocol: INIT -> APPEND -> FINALIZE
func UploadMediaFile(client *XClient, filePath string) (string, error) {
	// Validate file
	fileSize, mimeType, err := ValidateMediaFile(filePath)
	if err != nil {
		return "", err
	}

	// Step 1: INIT
	mediaID, err := uploadInit(client, fileSize, mimeType)
	if err != nil {
		return "", fmt.Errorf("upload INIT failed: %w", err)
	}

	if Verbose {
		logVerbose("Upload INIT success, media_id: %s", mediaID)
	}

	// Step 2: APPEND
	if err := uploadAppend(client, mediaID, filePath); err != nil {
		return "", fmt.Errorf("upload APPEND failed: %w", err)
	}

	if Verbose {
		logVerbose("Upload APPEND success")
	}

	// Step 3: FINALIZE
	if err := uploadFinalize(client, mediaID); err != nil {
		return "", fmt.Errorf("upload FINALIZE failed: %w", err)
	}

	if Verbose {
		logVerbose("Upload FINALIZE success")
	}

	return mediaID, nil
}

// uploadInit initiates the upload and gets a media ID
func uploadInit(client *XClient, totalBytes int64, mediaType string) (string, error) {
	// Determine media category
	mediaCategory := "tweet_image"
	if mediaType == "image/gif" {
		mediaCategory = "tweet_gif"
	}

	// Build request data
	data := url.Values{}
	data.Set("command", "INIT")
	data.Set("total_bytes", fmt.Sprintf("%d", totalBytes))
	data.Set("media_type", mediaType)
	data.Set("media_category", mediaCategory)

	// Make request
	resp, err := makeUploadRequest(client, data)
	if err != nil {
		return "", err
	}

	// Extract media_id
	mediaID, ok := resp["media_id_string"].(string)
	if !ok || mediaID == "" {
		return "", fmt.Errorf("no media_id in INIT response: %v", resp)
	}

	return mediaID, nil
}

// uploadAppend uploads the actual file data
func uploadAppend(client *XClient, mediaID, filePath string) error {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(data)

	// Build request data
	formData := url.Values{}
	formData.Set("command", "APPEND")
	formData.Set("media_id", mediaID)
	formData.Set("segment_index", "0")
	formData.Set("media_data", encoded)

	// Make request (with longer timeout for large files)
	_, err = makeUploadRequestWithTimeout(client, formData, 60)
	return err
}

// uploadFinalize finalizes the upload
func uploadFinalize(client *XClient, mediaID string) error {
	data := url.Values{}
	data.Set("command", "FINALIZE")
	data.Set("media_id", mediaID)

	_, err := makeUploadRequest(client, data)
	return err
}

// makeUploadRequest makes a request to the upload endpoint
func makeUploadRequest(client *XClient, data url.Values) (map[string]interface{}, error) {
	return makeUploadRequestWithTimeout(client, data, 30)
}

// makeUploadRequestWithTimeout makes an upload request with a custom timeout
func makeUploadRequestWithTimeout(client *XClient, data url.Values, timeout int) (map[string]interface{}, error) {
	formData := make(map[string]string, len(data))
	for key, values := range data {
		if len(values) > 0 {
			formData[key] = values[0]
		}
	}
	return client.RestPostWithOptionsContext(context.Background(), UploadURL, formData, nil, timeout)
}

// DownloadMedia downloads a file from a URL to the specified path
func DownloadMedia(mediaURL, outputPath string) error {
	return DownloadMediaContext(context.Background(), http.DefaultClient, mediaURL, outputPath, defaultMaxResponseBytes)
}

// DownloadMediaContext downloads a bounded response through client and
// publishes it atomically at outputPath after the transfer succeeds.
func DownloadMediaContext(ctx context.Context, client *http.Client, mediaURL, outputPath string, maxBytes int64) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		client = http.DefaultClient
	}
	if maxBytes <= 0 {
		maxBytes = defaultMaxResponseBytes
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	dir := filepath.Dir(outputPath)
	base := filepath.Base(outputPath)
	temp, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		_ = temp.Close()
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	written, err := io.Copy(temp, io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return err
	}
	if written > maxBytes {
		return &BodyLimitError{Limit: maxBytes, Size: written}
	}
	if err := temp.Sync(); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return err
	}
	removeTemp = false
	return nil
}
