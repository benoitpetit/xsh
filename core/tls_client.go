// Package core provides TLS fingerprinting and HTTP client with browser impersonation.
// This implements real TLS JA3 fingerprint spoofing using uTLS with HTTP/2 support.
package core

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
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
	stopCancel := context.AfterFunc(req.Context(), func() { _ = conn.Close() })

	writeRequest := req.Write
	if t.proxy != "" {
		writeRequest = req.WriteProxy
	}
	if err := writeRequest(conn); err != nil {
		stopCancel()
		_ = conn.Close()
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		stopCancel()
		_ = conn.Close()
		return nil, err
	}
	resp.Body = &connBody{ReadCloser: resp.Body, conn: conn, stopCancel: stopCancel}
	return resp, nil
}

type connBody struct {
	io.ReadCloser
	conn       net.Conn
	stopCancel func() bool
}

func (b *connBody) Close() error {
	if b.stopCancel != nil {
		b.stopCancel()
	}
	err := b.ReadCloser.Close()
	if closeErr := b.conn.Close(); err == nil {
		err = closeErr
	}
	return err
}

// dial creates a plain HTTP/1.1 connection (direct or through an HTTP proxy).
// HTTPS requests use dialTLS and the uTLS path below.
func (t *uTLSTransport) dial(ctx context.Context, addr string) (net.Conn, error) {
	dialer := t.dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 10 * time.Second}
	}

	if t.proxy == "" {
		return dialer.DialContext(ctx, "tcp", addr)
	}

	proxyURL, err := ResolveProxy(t.proxy, func(string) string { return "" })
	if err != nil {
		return nil, err
	}
	if proxyURL == nil {
		return nil, fmt.Errorf("proxy URL is empty")
	}
	proxyHost := proxyURL.Hostname()
	proxyPort := proxyURL.Port()
	if proxyPort == "" {
		proxyPort = "80"
	}
	return dialer.DialContext(ctx, "tcp", net.JoinHostPort(proxyHost, proxyPort))
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
	proxyURL, err := ResolveProxy(t.proxy, func(string) string { return "" })
	if err != nil {
		return nil, err
	}
	if proxyURL == nil {
		return nil, fmt.Errorf("proxy URL is empty")
	}
	proxyHost := proxyURL.Hostname()
	proxyPort := proxyURL.Port()
	if proxyPort == "" {
		proxyPort = "80"
		if proxyURL.Scheme == "https" {
			proxyPort = "443"
		}
	}

	// Connect to proxy
	dialer := t.dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 10 * time.Second}
	}
	proxyConn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(proxyHost, proxyPort))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}
	stopCancel := context.AfterFunc(ctx, func() { _ = proxyConn.Close() })
	closeProxy := func() {
		stopCancel()
		_ = proxyConn.Close()
	}

	// Send CONNECT request
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		credentials := base64.StdEncoding.EncodeToString([]byte(proxyURL.User.Username() + ":" + password))
		connectReq += "Proxy-Authorization: Basic " + credentials + "\r\n"
	}
	connectReq += "\r\n"
	if _, err := proxyConn.Write([]byte(connectReq)); err != nil {
		closeProxy()
		return nil, fmt.Errorf("failed to write CONNECT: %w", err)
	}

	// Read and parse a possibly split CONNECT response without over-reading the
	// first TLS bytes that follow the header terminator.
	response, err := readProxyHeaders(ctx, proxyConn)
	if err != nil {
		closeProxy()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}
	lines := strings.Split(string(response), "\r\n")
	statusFields := strings.Fields(lines[0])
	if len(statusFields) < 2 {
		closeProxy()
		return nil, fmt.Errorf("proxy CONNECT returned malformed status: %q", lines[0])
	}
	statusCode, err := strconv.Atoi(statusFields[1])
	if err != nil || statusCode != http.StatusOK {
		closeProxy()
		return nil, fmt.Errorf("proxy CONNECT failed: %s", lines[0])
	}

	// Clone config and set ServerName
	config := t.tlsConfig.Clone()
	config.ServerName = host

	// Wrap with uTLS
	uConn := utls.UClient(proxyConn, config, t.clientHelloID)
	if err := uConn.HandshakeContext(ctx); err != nil {
		closeProxy()
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}
	stopCancel()

	return uConn, nil
}

func readProxyHeaders(ctx context.Context, conn net.Conn) ([]byte, error) {
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetReadDeadline(deadline)
	} else {
		_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	}
	var response []byte
	var one [1]byte
	for len(response) < 64*1024 {
		n, err := conn.Read(one[:])
		if n > 0 {
			response = append(response, one[:n]...)
			if bytes.HasSuffix(response, []byte("\r\n\r\n")) {
				return response, nil
			}
		}
		if err != nil {
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	return nil, fmt.Errorf("proxy CONNECT response headers exceed 64 KiB")
}

// newUTLSHTTPClient creates an HTTP client with real TLS fingerprinting and HTTP/2 support
func newUTLSHTTPClient(proxy string, chromeVersion TLSFingerprintType) (*http.Client, error) {
	if chromeVersion == "" {
		chromeVersion = BestChromeTarget()
	}

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
		DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			return transport.dialTLS(ctx, addr)
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

// BestChromeTarget returns the best available Chrome target
func BestChromeTarget() TLSFingerprintType {
	return Chrome127
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
	return GetUserAgentForVersion(BestChromeTarget())
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
	return GetSecChUaForVersion(BestChromeTarget())
}

// GetSecChUaForVersion returns the sec-ch-ua header for a specific version
func GetSecChUaForVersion(target TLSFingerprintType) string {
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
	return GetSecChUaFullVersionListForVersion(BestChromeTarget())
}

// GetSecChUaFullVersionListForVersion returns the full version list header for a specific version
func GetSecChUaFullVersionListForVersion(target TLSFingerprintType) string {
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
