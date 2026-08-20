package server

import "github.com/atotto/clipboard"

// Clipboard abstracts the host clipboard so the server can be tested with
// an in-memory implementation. The production implementation delegates to
// github.com/atotto/clipboard.
type Clipboard interface {
	WriteAll(text string) error
	ReadAll() (string, error)
}

// osClipboard is the default, production Clipboard backed by atotto/clipboard.
type osClipboard struct{}

func (osClipboard) WriteAll(text string) error { return clipboard.WriteAll(text) }
func (osClipboard) ReadAll() (string, error)   { return clipboard.ReadAll() }
