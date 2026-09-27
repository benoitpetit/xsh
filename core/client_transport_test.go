package core

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"testing"
)

func TestHTTP1ResponseBodyRemainsReadableUntilClose(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) || strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("sandbox does not allow local listeners: %v", err)
		}
		t.Fatal(err)
	}
	defer listener.Close()
	ready := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				done <- err
				return
			}
			if line == "\r\n" {
				break
			}
		}
		_, _ = fmt.Fprint(conn, "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n")
		close(ready)
		<-release
		_, _ = fmt.Fprint(conn, "5\r\nhello\r\n0\r\n\r\n")
		done <- nil
	}()

	tpt := &uTLSTransport{dialer: &net.Dialer{}}
	target, _ := url.Parse("http://" + listener.Addr().String() + "/")
	resp, err := tpt.RoundTrip(&http.Request{Method: http.MethodGet, URL: target, Header: make(http.Header), Host: target.Host, Body: http.NoBody, GetBody: func() (io.ReadCloser, error) { return http.NoBody, nil }})
	if err != nil {
		t.Fatal(err)
	}
	<-ready
	close(release)
	body, err := io.ReadAll(resp.Body)
	closeErr := resp.Body.Close()
	if err != nil || closeErr != nil || string(body) != "hello" {
		t.Fatalf("response body = %q, read error=%v close error=%v", body, err, closeErr)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestDialProxyHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tpt := &uTLSTransport{proxy: "http://127.0.0.1:1", dialer: &net.Dialer{}}
	_, err := tpt.dialProxy(ctx, "example.com:443", "example.com")
	if err == nil || ctx.Err() == nil {
		t.Fatalf("dialProxy() error = %v, context = %v", err, ctx.Err())
	}
}
