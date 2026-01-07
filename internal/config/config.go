package config

import (
	"os"
	"path/filepath"
)

const (
	AppName      = "todo-tui"
	APIURL       = "https://todo.blackraven.org/api"
	FPS          = 30
)

// Config holds application configuration
type Config struct {
	APIURL     string
	TokenPath  string
	CredsPath  string
	DataDir    string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	dataDir := GetDataDir()
	return &Config{
		APIURL:    APIURL,
		TokenPath: filepath.Join(dataDir, "token"),
		CredsPath: filepath.Join(dataDir, "credentials"),
		DataDir:   dataDir,
	}
}

// Load loads the configuration (currently just returns defaults)
func Load() *Config {
	return DefaultConfig()
}

// GetDataDir returns the data directory path
func GetDataDir() string {
	// Try XDG config directory first
	if configDir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(configDir, AppName)
	}

	// Fallback to home directory
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, "."+AppName)
	}

	// Last resort - current directory
	return "." + AppName
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() error {
	dataDir := GetDataDir()
	return os.MkdirAll(dataDir, 0700)
}
