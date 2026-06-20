// Package config persists mctop's small set of user preferences under the OS
// config dir, next to the cached OAuth credentials.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the persisted preferences. Vim defaults to true: Load seeds it
// before decoding so a missing file or absent field keeps vim mode on, while an
// explicit false in the file turns it off.
type Config struct {
	Vim bool `json:"vim"`
}

func path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mctop", "config.json"), nil
}

// Load returns the saved preferences, falling back to defaults when there is no
// readable config file.
func Load() Config {
	c := Config{Vim: true}
	p, err := path()
	if err != nil {
		return c
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return c
	}
	_ = json.Unmarshal(data, &c)
	return c
}

// Save writes the preferences, creating the config dir if needed.
func Save(c Config) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
