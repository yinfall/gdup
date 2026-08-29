package gdup

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// GvmConfig represents the .gduprc configuration file.
type GvmConfig struct {
	Version string `json:"version"`
}

// FindGvmConfig searches for .gduprc file from cwd upwards, then in home directory.
func FindGvmConfig() string {
	cwd, err := os.Getwd()
	if err == nil {
		dir := cwd
		for {
			p := filepath.Join(dir, ".gduprc")
			if info, err := os.Stat(p); err == nil && !info.IsDir() {
				return p
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	home, err := os.UserHomeDir()
	if err == nil {
		p := filepath.Join(home, ".gduprc")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}

	return ""
}

// ReadGvmConfig reads and parses a .gduprc config file.
func ReadGvmConfig(path string) (*GvmConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg GvmConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// WriteGvmConfig writes a .gduprc config file at the given path.
func WriteGvmConfig(path string, version string) error {
	cfg := GvmConfig{Version: version}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// GetActiveVersion returns the active version from .gduprc config, or empty string.
func GetActiveVersion() string {
	configPath := FindGvmConfig()
	if configPath == "" {
		return ""
	}
	cfg, err := ReadGvmConfig(configPath)
	if err != nil {
		return ""
	}
	return NormalizeVersionTag(cfg.Version)
}
