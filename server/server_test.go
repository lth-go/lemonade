package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// memClipboard is an in-memory Clipboard for tests.
type memClipboard struct {
	mu   sync.Mutex
	text string
}

func (m *memClipboard) WriteAll(text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.text = text
	return nil
}

func (m *memClipboard) ReadAll() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.text, nil
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestServer builds a Server in-memory with an in-memory clipboard and
// returns an httptest.Server fronting its mux.
func newTestServer(t *testing.T, allow string) (*httptest.Server, *Server, *memClipboard) {
	t.Helper()
	s, err := New(Options{Allow: allow, LineEnding: "", Logger: newTestLogger()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cb := &memClipboard{}
	s = s.withClipboard(cb)
	mux := http.NewServeMux()
	mux.Handle("/copy", s.middleware(http.HandlerFunc(s.handleCopy)))
	mux.Handle("/paste", s.middleware(http.HandlerFunc(s.handlePaste)))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts, s, cb
}

func TestServer_CopyPasteRoundTrip(t *testing.T) {
	ts, _, _ := newTestServer(t, "127.0.0.1/32,::1/32")

	resp, err := http.Post(ts.URL+"/copy", "text/plain", strings.NewReader("hello world"))
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("copy status = %d, want 200", resp.StatusCode)
	}

	resp, err = http.Get(ts.URL + "/paste")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello world" {
		t.Errorf("paste body = %q, want %q", string(body), "hello world")
	}
}

func TestServer_CopyMethodNotAllowed(t *testing.T) {
	ts, _, _ := newTestServer(t, "127.0.0.1/32")
	resp, err := http.Get(ts.URL + "/copy")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestServer_PasteMethodNotAllowed(t *testing.T) {
	ts, _, _ := newTestServer(t, "127.0.0.1/32")
	resp, err := http.Post(ts.URL+"/paste", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", resp.StatusCode)
	}
}

func TestServer_DeniedIP(t *testing.T) {
	// Allow only 10.0.0.0/24, but the test client connects from 127.0.0.1.
	ts, _, _ := newTestServer(t, "10.0.0.0/24")
	resp, err := http.Post(ts.URL+"/copy", "text/plain", strings.NewReader("x"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", resp.StatusCode)
	}
}

func TestServer_LineEndingConversion(t *testing.T) {
	s, err := New(Options{Allow: "127.0.0.1/32", LineEnding: "crlf", Logger: newTestLogger()})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	s = s.withClipboard(&memClipboard{})
	mux := http.NewServeMux()
	mux.Handle("/copy", s.middleware(http.HandlerFunc(s.handleCopy)))
	mux.Handle("/paste", s.middleware(http.HandlerFunc(s.handlePaste)))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	resp, err := http.Post(ts.URL+"/copy", "text/plain", strings.NewReader("a\nb"))
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	resp.Body.Close()

	resp, err = http.Get(ts.URL + "/paste")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "a\r\nb" {
		t.Errorf("paste body = %q, want %q", string(body), "a\r\nb")
	}
}
