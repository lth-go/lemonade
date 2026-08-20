package lemon

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config carries all shared configuration. It is populated by the command
// layer (cobra) and may also be decoded directly from ~/.config/lemonade.toml
// via the TOML struct tags.
type Config struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	Allow      string `toml:"allow"`
	LineEnding string `toml:"line-ending"`
	LogLevel   string `toml:"log-level"`
}

// LoadConfig reads ~/.config/lemonade.toml into a Config. A missing file
// yields a zero Config and no error.
func LoadConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, nil
	}
	return decodeConfigFile(filepath.Join(home, ".config", "lemonade.toml"))
}

// decodeConfigFile decodes path as TOML. Missing file -> zero value, no error.
func decodeConfigFile(path string) (Config, error) {
	var conf Config
	_, err := toml.DecodeFile(path, &conf)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	return conf, nil
}

// Merge applies non-zero fields from src onto c, but only for fields whose
// name is NOT in changed. Precedence (highest first): command-line flags
// (already set on c and recorded in changed) > config file > environment.
//
// For Host specifically, when neither the flag nor the config file provide a
// value, the SSH_CLIENT / WSL_HOST environment variables are consulted.
func (c *Config) Merge(src Config, changed map[string]bool) {
	if !changed["host"] {
		if src.Host != "" {
			c.Host = src.Host
		} else {
			c.resolveHost()
		}
	}
	if !changed["port"] && src.Port != 0 {
		c.Port = src.Port
	}
	if !changed["allow"] && src.Allow != "" {
		c.Allow = src.Allow
	}
	if !changed["line-ending"] && src.LineEnding != "" {
		c.LineEnding = src.LineEnding
	}
	if !changed["log-level"] && src.LogLevel != "" {
		c.LogLevel = src.LogLevel
	}
}

// resolveHost fills in c.Host from the SSH_CLIENT / WSL_HOST environment
// variables when no explicit --host was given (client mode).
func (c *Config) resolveHost() {
	if c.Host != "" {
		return
	}
	if sshClient := os.Getenv("SSH_CLIENT"); sshClient != "" {
		c.Host = strings.Split(sshClient, " ")[0]
		return
	}
	if wslHost := os.Getenv("WSL_HOST"); wslHost != "" {
		c.Host = wslHost
	}
}
