package client

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github.com/lemonade-command/lemonade/lemon"
)

// Client talks to a remote lemonade server over HTTP. When the server is
// unreachable it falls back to the local clipboard so copy still works
// from the same host.
type Client struct {
	addr       string
	lineEnding string
	logger     *slog.Logger
	httpClient *http.Client
}

// Options bundles the parameters for New.
type Options struct {
	Host       string
	Port       int
	LineEnding string
	Logger     *slog.Logger
}

// New constructs a Client from explicit options. The HTTP transport disables
// proxy usage so remote copy/paste works even when HTTP_PROXY is set.
func New(opts Options) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil

	if opts.Logger == nil {
		opts.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &Client{
		addr:       fmt.Sprintf("http://%s:%d", opts.Host, opts.Port),
		lineEnding: opts.LineEnding,
		logger:     opts.Logger,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   2 * time.Second,
		},
	}
}

// Copy sends text to the remote server's /copy endpoint. On network error
// it writes to the local clipboard and returns the original error so the
// caller can decide how to react.
func (c *Client) Copy(text string) error {
	url := c.addr + "/copy"
	_, err := c.httpClient.Post(url, "text/plain", strings.NewReader(text))
	if err != nil {
		_ = clipboard.WriteAll(text)
		return err
	}
	return nil
}

// Paste fetches text from the remote server's /paste endpoint and applies
// the configured line-ending conversion.
func (c *Client) Paste() (string, error) {
	resp, err := c.httpClient.Get(c.addr + "/paste")
	if err != nil {
		c.logger.Error("http get", "err", err)
		return "", err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("read body", "err", err)
		return "", err
	}

	return lemon.ConvertLineEnding(string(buf), c.lineEnding), nil
}
