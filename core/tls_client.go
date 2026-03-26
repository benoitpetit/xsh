// Package core provides TLS fingerprinting and HTTP client with browser impersonation.
// This implements real TLS JA3 fingerprint spoofing using uTLS with HTTP/2 support.
package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// TLSClientConfig holds configuration for TLS client
type TLSClientConfig struct {
	FingerprintType TLSFingerprintType
	Proxy           string
	Timeout         time.Duration
}

// DefaultTLSClientConfig returns default TLS configuration
func DefaultTLSClientConfig() *TLSClientConfig {
	return &TLSClientConfig{
		FingerprintType: BestChromeTarget(),
		Proxy:           "",
		Timeout:         30 * time.Second,
	}
}

// chromeClientHelloIDs maps Chrome versions to uTLS ClientHello IDs
// All versions map to HelloChrome_120 for compatibility with utls v1.6.7
var chromeClientHelloIDs = map[TLSFingerprintType]utls.ClientHelloID{
	Chrome120: utls.HelloChrome_120,
	Chrome123: utls.HelloChrome_120,
	Chrome124: utls.HelloChrome_120,
	Chrome126: utls.HelloChrome_120,
	Chrome127: utls.HelloChrome_120,
	Chrome131: utls.HelloChrome_120,
	Chrome133: utls.HelloChrome_120,
}

// uTLSTransport is a custom http.RoundTripper that uses uTLS
type uTLSTransport struct {
	clientHelloID utls.ClientHelloID
	tlsConfig     *utls.Config
	proxy         string
	dialer        *net.Dialer

	// http2Transport is used for HTTP/2 connections
	http2Transport *http2.Transport
}

// RoundTrip implements http.RoundTripper
func (t *uTLSTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Use HTTP/2 transport for all HTTPS requests
	if req.URL.Scheme == "https" {
		return t.http2Transport.RoundTrip(req)
	}
	// Fall back to HTTP/1.1 for HTTP requests
	return t.roundTripHTTP1(req)
}

// roundTripHTTP1 handles HTTP/1.1 requests
func (t *uTLSTransport) roundTripHTTP1(req *http.Request) (*http.Response, error) {
	conn, err := t.dial(req.Context(), req.URL.Host)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := req.Write(conn); err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(conn), req)
}

// dial creates a connection (direct or through proxy)
func (t *uTLSTransport) dial(ctx context.Context, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
		port = "443"
		addr = net.JoinHostPort(host, port)
	}

	if t.proxy != "" {
		return t.dialProxy(ctx, addr, host)
	}
	return t.dialDirect(ctx, addr, host)
}

// dialDirect creates a direct TLS connection
func (t *uTLSTransport) dialDirect(ctx context.Context, addr, host string) (net.Conn, error) {
	tcpConn, err := t.dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	config := t.tlsConfig.Clone()
	config.ServerName = host

	uConn := utls.UClient(tcpConn, config, t.clientHelloID)
	if err := uConn.HandshakeContext(ctx); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return uConn, nil
}

// dialProxy creates a TLS connection through HTTP CONNECT proxy
func (t *uTLSTransport) dialProxy(ctx context.Context, addr, host string) (net.Conn, error) {
	proxyURL, err := url.Parse(t.proxy)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	// Connect to proxy
	proxyConn, err := t.dialer.DialContext(ctx, "tcp", proxyURL.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", addr, addr)
	if _, err := proxyConn.Write([]byte(connectReq)); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to write CONNECT: %w", err)
	}

	// Read response
	buf := make([]byte, 1024)
	n, err := proxyConn.Read(buf)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}

	if !strings.Contains(string(buf[:n]), "200") {
		proxyConn.Close()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", string(buf[:n]))
	}

	// Clone config and set ServerName
	config := t.tlsConfig.Clone()
	config.ServerName = host

	// Wrap with uTLS
	uConn := utls.UClient(proxyConn, config, t.clientHelloID)
	if err := uConn.HandshakeContext(ctx); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	return uConn, nil
}

// newUTLSHTTPClient creates an HTTP client with real TLS fingerprinting and HTTP/2 support
func newUTLSHTTPClient(proxy string) (*http.Client, error) {
	// Select a random Chrome version for rotation
	chromeVersion := selectRandomChromeVersion()

	clientHelloID, ok := chromeClientHelloIDs[chromeVersion]
	if !ok {
		clientHelloID = utls.HelloChrome_120
	}

	tlsConfig := &utls.Config{
		MinVersion:   utls.VersionTLS12,
		MaxVersion:   utls.VersionTLS13,
		CipherSuites: getChromeCipherSuites(),
		CurvePreferences: []utls.CurveID{
			utls.X25519,
			utls.CurveP256,
			utls.CurveP384,
		},
		PreferServerCipherSuites: false,
		InsecureSkipVerify:       false,
		NextProtos:               []string{"h2", "http/1.1"},
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &uTLSTransport{
		clientHelloID: clientHelloID,
		tlsConfig:     tlsConfig,
		proxy:         proxy,
		dialer:        dialer,
	}

	// Configure HTTP/2 transport with uTLS
	// We ignore the passed *tls.Config and use our uTLS config instead
	transport.http2Transport = &http2.Transport{
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return transport.dialTLS(context.Background(), addr)
		},
		TLSClientConfig: &tls.Config{}, // Empty config, we handle TLS ourselves
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		Jar:       nil, // We manage cookies manually
	}, nil
}

// dialTLS creates a TLS connection for HTTP/2 using uTLS
func (t *uTLSTransport) dialTLS(ctx context.Context, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	if t.proxy != "" {
		return t.dialProxy(ctx, addr, host)
	}
	return t.dialDirect(ctx, addr, host)
}

// getChromeCipherSuites returns Chrome's preferred cipher suites
func getChromeCipherSuites() []uint16 {
	return []uint16{
		utls.TLS_AES_128_GCM_SHA256,
		utls.TLS_AES_256_GCM_SHA384,
		utls.TLS_CHACHA20_POLY1305_SHA256,
		utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		utls.TLS_RSA_WITH_AES_128_CBC_SHA,
		utls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}
}

// selectRandomChromeVersion returns a random Chrome version for rotation
func selectRandomChromeVersion() TLSFingerprintType {
	versions := []TLSFingerprintType{
		Chrome127,
		Chrome126,
		Chrome124,
		Chrome123,
	}
	return versions[rand.Intn(len(versions))]
}

// BestChromeTarget returns the best available Chrome target
func BestChromeTarget() TLSFingerprintType {
	return Chrome127
}

// chromeTargetMutex protects the current target
var chromeTargetMutex sync.RWMutex
var currentChromeTarget TLSFingerprintType = Chrome127

// SyncChromeVersion updates the global Chrome version for headers
func SyncChromeVersion(version TLSFingerprintType) {
	chromeTargetMutex.Lock()
	currentChromeTarget = version
	chromeTargetMutex.Unlock()
}

// GetCurrentChromeTarget returns the current Chrome target for headers
func GetCurrentChromeTarget() TLSFingerprintType {
	chromeTargetMutex.RLock()
	defer chromeTargetMutex.RUnlock()
	return currentChromeTarget
}

// GetUserAgentForVersion returns a Chrome User-Agent string for the given version
func GetUserAgentForVersion(version TLSFingerprintType) string {
	ver := chromeVersionStrings[version]
	if ver == "" {
		ver = "127.0.0.0"
	}

	var platform string
	switch runtime.GOOS {
	case "darwin":
		platform = "Macintosh; Intel Mac OS X 10_15_7"
	case "windows":
		platform = "Windows NT 10.0; Win64; x64"
	default:
		platform = "X11; Linux x86_64"
	}

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", platform, ver)
}

// GetUserAgent returns the current User-Agent string
func GetUserAgent() string {
	target := GetCurrentChromeTarget()
	return GetUserAgentForVersion(target)
}

// GetPlatform returns the platform string for sec-ch-ua-platform
func GetPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	default:
		return "Linux"
	}
}

// GetArchitecture returns the architecture for sec-ch-ua-arch
func GetArchitecture() string {
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		return "arm"
	}
	return "x86"
}

// GetPlatformVersion returns the platform version for sec-ch-ua-platform-version
func GetPlatformVersion() string {
	switch runtime.GOOS {
	case "darwin":
		return "15.0.0"
	case "windows":
		return "10.0.0"
	default:
		return "10.0.0"
	}
}

// GetSecChUa returns the current sec-ch-ua header
func GetSecChUa() string {
	target := GetCurrentChromeTarget()
	ver := chromeVersionStrings[target]
	if ver == "" {
		ver = "127.0.0.0"
	}
	majorVer := ver[:3]
	if majorVer[2] == '.' {
		majorVer = ver[:2]
	}
	return fmt.Sprintf(`"Chromium";v="%s", "Not(A:Brand";v="99", "Google Chrome";v="%s"`, majorVer, majorVer)
}

// GetSecChUaFullVersionList returns the full version list header
func GetSecChUaFullVersionList() string {
	target := GetCurrentChromeTarget()
	ver := chromeVersionStrings[target]
	if ver == "" {
		ver = "127.0.0.0"
	}
	return fmt.Sprintf(`"Google Chrome";v="%s", "Chromium";v="%s", "Not.A/Brand";v="99.0.0.0"`, ver, ver)
}

// GetAcceptLanguage returns the Accept-Language header value
func GetAcceptLanguage() string {
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}

	if lang != "" {
		lang = strings.Split(lang, ".")[0]
		lang = strings.Replace(lang, "_", "-", 1)
		base := strings.Split(lang, "-")[0]
		return fmt.Sprintf("%s,%s;q=0.9,en;q=0.8", lang, base)
	}

	return "en-US,en;q=0.9"
}
