package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	LogLevel    string
	LogFormat   string
	SessionPath string
	ReposPath   string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		LogFormat:   getEnv("LOG_FORMAT", "text"),
		SessionPath: getEnv("SESSION_PATH", "~/.chatgit/sessions"),
		ReposPath:   getEnv("REPOS_PATH", "~/.chatgit/repos"),
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets integer environment variable with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetLogLevel returns the log level
func (c *Config) GetLogLevel() string {
	return c.LogLevel
}

// GetLogFormat returns the log format
func (c *Config) GetLogFormat() string {
	return c.LogFormat
}

// GetSessionPath returns the session storage path
func (c *Config) GetSessionPath() string {
	return expandPath(c.SessionPath)
}

// GetReposPath returns the repositories metadata path
func (c *Config) GetReposPath() string {
	return expandPath(c.ReposPath)
}

// expandPath expands ~ to user home directory
func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:]
	}
	return path
}

// GetRepoDataPath returns the data directory for a specific repo
func (c *Config) GetRepoDataPath(repoPath string) string {
	return fmt.Sprintf("%s/%s.json", c.GetSessionPath(), hashPath(repoPath))
}

// hashPath creates a simple hash of a path for use as filename
func hashPath(path string) string {
	// Simple hash for now - just replace invalid characters
	// In production, use a proper hash function
	result := ""
	for _, c := range path {
		switch {
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'):
			result += string(c)
		case c == '/' || c == '\\':
			result += "_"
		default:
			result += fmt.Sprintf("%x", c)
		}
	}
	return result
}
