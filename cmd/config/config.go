// Package config provides functionality for loading and accessing application configuration.
// The configuration is loaded from a JSON file and stored in memory for easy access.
package config

import (
	"encoding/json"
	"os"
)

// UploadDefaults represents default settings for video uploads.
type UploadDefaults struct {
	DefaultPrivacy     string   `json:"default_privacy"`     // Default privacy status (public, unlisted, private)
	DefaultCategory    string   `json:"default_category"`    // Default YouTube category ID
	DefaultDescription string   `json:"default_description"` // Default video description
	DefaultTags        []string `json:"default_tags"`        // Default tags for videos
}

// Config represents the main application configuration structure.
type Config struct {
	UploadDefaults UploadDefaults `json:"upload_defaults"` // Default upload settings
}

// configInstance holds the singleton instance of loaded configuration
var configInstance *Config

// defaultConfig returns default configuration values
func defaultConfig() *Config {
	return &Config{
		UploadDefaults: UploadDefaults{
			DefaultPrivacy:     "public",
			DefaultCategory:    "22", // People & Blogs
			DefaultDescription: "",
			DefaultTags:        []string{},
		},
	}
}

// loadConfig reads and parses the configuration file from the specified path.
// If the file doesn't exist or can't be parsed, returns default configuration.
func loadConfig(filePath string) *Config {
	file, err := os.ReadFile(filePath)
	if err != nil {
		// Config file not found - use defaults
		return defaultConfig()
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		// Config file invalid - use defaults
		return defaultConfig()
	}

	// Apply defaults for any missing values
	defaults := defaultConfig()
	if config.UploadDefaults.DefaultPrivacy == "" {
		config.UploadDefaults.DefaultPrivacy = defaults.UploadDefaults.DefaultPrivacy
	}
	if config.UploadDefaults.DefaultCategory == "" {
		config.UploadDefaults.DefaultCategory = defaults.UploadDefaults.DefaultCategory
	}

	return &config
}

// init is automatically called when the package is imported.
// It loads the configuration from the default location.
func init() {
	configInstance = loadConfig("config.json")
}

// GetUploadDefaults returns the configured upload defaults.
func GetUploadDefaults() UploadDefaults {
	return configInstance.UploadDefaults
}
