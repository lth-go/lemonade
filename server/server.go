package server

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/lemonade-command/lemonade/lemon"
)

// Server holds the runtime state for the lemonade HTTP server.
type Server struct {
	logger     *slog.Logger
	lineEnding string
	allow      *ipRange
	port       int
	clipboard  Clipboard
	mu         *sync.Mutex
}

// Options bundles the parameters for New.
type Options struct {
	Allow      string
	Port       int
	LineEnding string
	Logger     *slog.Logger
}

// New constructs a Server from explicit options.
func New(opts Options) (*Server, error) {
	r, err := newIPRange(opts.Allow)
	if err != nil {
		return nil, fmt.Errorf("parse --allow: %w", err)
	}
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Server{
		logger:     opts.Logger,
		lineEnding: opts.LineEnding,
		allow:      r,
		port:       opts.Port,
		clipboard:  osClipboard{},
		mu:         &sync.Mutex{},
	}, nil
}

// withClipboard returns a shallow copy of s using the given Clipboard.
// Intended for tests that want an in-memory clipboard. The mutex is shared
// with the original server, which is fine because tests use a fresh server.
func (s *Server) withClipboard(cb Clipboard) *Server {
	cp := *s
	cp.clipboard = cb
	return &cp
}

// Serve starts the HTTP listener. It blocks until the server stops.
func (s *Server) Serve() error {
	mux := http.NewServeMux()
	mux.Handle("/copy", s.middleware(http.HandlerFunc(s.handleCopy)))
	mux.Handle("/paste", s.middleware(http.HandlerFunc(s.handlePaste)))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}
	return srv.ListenAndServe()
}

func (s *Server) handleCopy(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "Copy only supports POST", http.StatusMethodNotAllowed)
		return
	}

	buf, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("read body", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	text := lemon.ConvertLineEnding(string(buf), s.lineEnding)

	s.mu.Lock()
	defer s.mu.Unlock()

	err = s.clipboard.WriteAll(text)
	if err != nil {
		s.logger.Error("clipboard write", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePaste(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Paste only supports GET", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	text, err := s.clipboard.ReadAll()
	if err != nil {
		s.logger.Error("clipboard read", "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = io.WriteString(w, text)
	if err != nil {
		s.logger.Error("write response", "err", err)
		return
	}
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
			return
		}

		remoteIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "RemoteAddr error.", http.StatusInternalServerError)
			return
		}
		if !s.allow.includeStr(remoteIP) {
			http.Error(w, "Not allowed IP.", http.StatusServiceUnavailable)
			s.logger.Info("blocked by allow list", "ip", remoteIP)
			return
		}

		next.ServeHTTP(w, r)
	})
}
