// +build windows

package browser

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/billgraziano/dpapi"
)

// chromeCookiePath is the default Chrome cookie path on Windows
const chromeCookiePath = "Google/Chrome/User Data/Default/Network/Cookies"

// chromeKeyPath is the path to Chrome's Local State file
const chromeKeyPath = "Google/Chrome/User Data/Local State"

// dataBlob represents the DPAPI data structure
type dataBlob struct {
	cbData uint32
	pbData *byte
}

// decryptChromeCookieWindows decrypts Chrome cookies on Windows using DPAPI
func decryptChromeCookieWindows(encrypted []byte) (string, error) {
	if len(encrypted) == 0 {
		return "", nil
	}

	// Chrome on Windows v80+ uses a specific format:
	// [version prefix][encrypted key][nonce][ciphertext][tag]
	// Version prefix is usually "v10" or "v11"

	if len(encrypted) < 3 {
		return "", fmt.Errorf("encrypted data too short")
	}

	prefix := string(encrypted[:3])
	if prefix != "v10" && prefix != "v11" {
		// Try direct DPAPI decryption (older format)
		return decryptDPAPIDirect(encrypted)
	}

	// New format: need to decrypt the key first, then decrypt the cookie
	encrypted = encrypted[3:] // Remove prefix

	// Get the encrypted key from Local State
	encryptedKey, err := getEncryptedKeyFromLocalState()
	if err != nil {
		// Fallback: try DPAPI on the whole thing
		return decryptDPAPIDirect(encrypted)
	}

	// Decrypt the key using DPAPI
	masterKeyStr, err := dpapi.Decrypt(string(encryptedKey))
	if err != nil {
		return "", fmt.Errorf("failed to decrypt master key: %w", err)
	}
	masterKey := []byte(masterKeyStr)

	// Parse the encrypted cookie data
	if len(encrypted) < 12+16 { // nonce + minimum ciphertext + tag
		return "", fmt.Errorf("encrypted cookie data too short")
	}

	// Extract nonce, ciphertext, and tag
	// Format: [nonce:12][ciphertext][tag:16]
	nonce := encrypted[:12]
	ciphertextAndTag := encrypted[12:]

	if len(ciphertextAndTag) < 16 {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	ciphertext := ciphertextAndTag[:len(ciphertextAndTag)-16]
	tag := ciphertextAndTag[len(ciphertextAndTag)-16:]

	// Combine ciphertext and tag for GCM
	ciphertextWithTag := append(ciphertext, tag...)

	// Decrypt using AES-256-GCM
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertextWithTag, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// decryptDPAPIDirect uses Windows DPAPI directly for simple decryption
func decryptDPAPIDirect(encrypted []byte) (string, error) {
	decrypted, err := dpapi.Decrypt(string(encrypted))
	if err != nil {
		return "", fmt.Errorf("DPAPI decryption failed: %w", err)
	}
	return decrypted, nil
}

// getEncryptedKeyFromLocalState retrieves the encrypted master key from Chrome's Local State
func getEncryptedKeyFromLocalState() ([]byte, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return nil, fmt.Errorf("LOCALAPPDATA not set")
	}

	localStatePath := filepath.Join(localAppData, chromeKeyPath)
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Local State: %w", err)
	}

	var localState struct {
		OSCrypt struct {
			EncryptedKey string `json:"encrypted_key"`
		} `json:"os_crypt"`
	}

	if err := json.Unmarshal(data, &localState); err != nil {
		return nil, fmt.Errorf("failed to parse Local State: %w", err)
	}

	if localState.OSCrypt.EncryptedKey == "" {
		return nil, fmt.Errorf("no encrypted key found")
	}

	// Decode base64
	encryptedKey, err := base64.StdEncoding.DecodeString(localState.OSCrypt.EncryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key: %w", err)
	}

	// Check for DPAPI prefix
	if len(encryptedKey) < 5 || string(encryptedKey[:5]) != "DPAPI" {
		return nil, fmt.Errorf("invalid key format")
	}

	// Remove DPAPI prefix
	return encryptedKey[5:], nil
}

// GetWindowsChromePaths returns Chrome cookie paths for Windows
func GetWindowsChromePaths() []string {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return nil
	}

	return []string{
		filepath.Join(localAppData, "Google/Chrome/User Data/Default/Network/Cookies"),
		filepath.Join(localAppData, "Google/Chrome/User Data/Default/Cookies"),
		filepath.Join(localAppData, "Google/Chrome/User Data/Profile 1/Network/Cookies"),
		filepath.Join(localAppData, "BraveSoftware/Brave-Browser/User Data/Default/Network/Cookies"),
		filepath.Join(localAppData, "Microsoft/Edge/User Data/Default/Network/Cookies"),
		filepath.Join(localAppData, "Chromium/User Data/Default/Network/Cookies"),
	}
}

// IsChromeAvailableWindows checks if Chrome is available on Windows
func IsChromeAvailableWindows() bool {
	paths := GetWindowsChromePaths()
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// sysDPAPI is a fallback using direct syscalls if the dpapi package fails
func sysDPAPI(encrypted []byte) ([]byte, error) {
	dll := syscall.NewLazyDLL("crypt32.dll")
	proc := dll.NewProc("CryptUnprotectData")

	inBlob := dataBlob{
		cbData: uint32(len(encrypted)),
		pbData: &encrypted[0],
	}

	var outBlob dataBlob

	ret, _, err := proc.Call(
		uintptr(unsafe.Pointer(&inBlob)),
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&outBlob)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("CryptUnprotectData failed: %v", err)
	}

	// Copy the data
	data := make([]byte, outBlob.cbData)
	copy(data, (*[1 << 30]byte)(unsafe.Pointer(outBlob.pbData))[:outBlob.cbData:outBlob.cbData])

	// Free the memory
	localFree := syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")
	localFree.Call(uintptr(unsafe.Pointer(outBlob.pbData)))

	return data, nil
}
