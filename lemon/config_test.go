package lemon

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// withTempConfig writes a TOML config and points HOME at a temp dir so
// LoadConfig picks it up. Cleanup is automatic via t.Cleanup.
func withTempConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	path := filepath.Join(dir, ".config", "lemonade.toml")
	err := os.MkdirAll(filepath.Dir(path), 0o700)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	err = os.WriteFile(path, []byte(content), 0o600)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	conf, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if conf.Host != "" || conf.Port != 0 || conf.Allow != "" || conf.LineEnding != "" || conf.LogLevel != "" {
		t.Errorf("expected zero Config, got %+v", conf)
	}
}

func TestLoadConfig_DecodesValues(t *testing.T) {
	withTempConfig(t, `port = 1234
host = '192.168.1.1'
allow = '10.0.0.0/8'
line-ending = 'crlf'
log-level = 'warn'
`)
	conf, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if conf.Host != "192.168.1.1" {
		t.Errorf("Host = %q, want 192.168.1.1", conf.Host)
	}
	if conf.Port != 1234 {
		t.Errorf("Port = %d, want 1234", conf.Port)
	}
	if conf.Allow != "10.0.0.0/8" {
		t.Errorf("Allow = %q, want 10.0.0.0/8", conf.Allow)
	}
	if conf.LineEnding != "crlf" {
		t.Errorf("LineEnding = %q, want crlf", conf.LineEnding)
	}
	if conf.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want warn", conf.LogLevel)
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	withTempConfig(t, `this is not = = valid toml [[`)
	_, err := LoadConfig()
	if err == nil {
		t.Error("expected error for invalid TOML")
	}
}

func TestMerge_ConfigFillsUnsetFlags(t *testing.T) {
	t.Setenv("SSH_CLIENT", "")
	t.Setenv("WSL_HOST", "")
	c := Config{Port: 2489, Allow: "0.0.0.0/0,::/0", LogLevel: "info"} // cobra defaults
	src := Config{Host: "192.168.1.1", Port: 1234, LineEnding: "crlf", LogLevel: "warn"}
	c.Merge(src, map[string]bool{})

	if c.Host != "192.168.1.1" {
		t.Errorf("Host = %q, want 192.168.1.1", c.Host)
	}
	if c.Port != 1234 {
		t.Errorf("Port = %d, want 1234 (config overrides default)", c.Port)
	}
	if c.LineEnding != "crlf" {
		t.Errorf("LineEnding = %q, want crlf", c.LineEnding)
	}
	if c.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want warn", c.LogLevel)
	}
}

func TestMerge_ExplicitFlagWins(t *testing.T) {
	t.Setenv("SSH_CLIENT", "")
	t.Setenv("WSL_HOST", "")
	c := Config{Port: 5566, LogLevel: "debug"}
	src := Config{Port: 1234, LogLevel: "error"}
	c.Merge(src, map[string]bool{"port": true, "log-level": true})

	if c.Port != 5566 {
		t.Errorf("Port = %d, want 5566 (flag wins)", c.Port)
	}
	if c.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want debug (flag wins)", c.LogLevel)
	}
}

func TestMerge_EnvHostFallback(t *testing.T) {
	t.Setenv("SSH_CLIENT", "203.0.113.5 22 22")
	t.Setenv("WSL_HOST", "")
	c := Config{}
	c.Merge(Config{}, map[string]bool{})

	if c.Host != "203.0.113.5" {
		t.Errorf("Host = %q, want 203.0.113.5", c.Host)
	}
}

func TestMerge_EnvHostFallback_WSL(t *testing.T) {
	t.Setenv("SSH_CLIENT", "")
	t.Setenv("WSL_HOST", "192.168.2.2")
	c := Config{}
	c.Merge(Config{}, map[string]bool{})

	if c.Host != "192.168.2.2" {
		t.Errorf("Host = %q, want 192.168.2.2", c.Host)
	}
}

func TestMerge_ExplicitHostSuppressesEnv(t *testing.T) {
	t.Setenv("SSH_CLIENT", "203.0.113.5 22 22")
	c := Config{Host: "example.com"}
	c.Merge(Config{}, map[string]bool{"host": true})

	if c.Host != "example.com" {
		t.Errorf("Host = %q, want example.com", c.Host)
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"Debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"critical", slog.LevelError + 4},
		{"unknown", slog.LevelInfo}, // fallback
		{"", slog.LevelInfo},        // fallback
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := ParseLevel(c.in); got != c.want {
				t.Errorf("ParseLevel(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}
