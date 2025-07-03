package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Plain PlainConfig `yaml:"plain"`
	UI    UIConfig    `yaml:"ui"`
}

// PlainConfig contains Plain API configuration
type PlainConfig struct {
	APIKey      string `yaml:"api_key" kong:"env:PLAIN_API_KEY"`
	Endpoint    string `yaml:"endpoint" kong:"default:https://core-api.uk.plain.com/graphql/v1"`
	WorkspaceID string `yaml:"workspace_id" kong:"env:PLAIN_WORKSPACE_ID"`
}

// UIConfig contains UI configuration
type UIConfig struct {
	Theme     string `yaml:"theme" kong:"default:default"`
	PageSize  int    `yaml:"page_size" kong:"default:20"`
	ShowDebug bool   `yaml:"show_debug" kong:"default:false"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Plain.APIKey == "" {
		return fmt.Errorf("Plain API key is required (set PLAIN_API_KEY environment variable or configure in config file)")
	}
	if c.Plain.Endpoint == "" {
		return fmt.Errorf("Plain API endpoint is required")
	}
	if c.UI.PageSize <= 0 {
		return fmt.Errorf("UI page size must be positive")
	}
	return nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".simple")
	configPath := filepath.Join(configDir, "config.yaml")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configPath, nil
}

// Load loads configuration from file
func Load(configPath string) (*Config, error) {
	// Default configuration
	cfg := &Config{
		Plain: PlainConfig{
			Endpoint: "https://core-api.uk.plain.com/graphql/v1",
		},
		UI: UIConfig{
			Theme:     "default",
			PageSize:  20,
			ShowDebug: false,
		},
	}

	// Load from environment variables
	if apiKey := os.Getenv("PLAIN_API_KEY"); apiKey != "" {
		cfg.Plain.APIKey = apiKey
	}

	// Load from file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(configPath string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(configPath string) error {
	cfg := &Config{
		Plain: PlainConfig{
			APIKey:      "your-api-key-here",
			Endpoint:    "https://core-api.uk.plain.com/graphql/v1",
			WorkspaceID: "your-workspace-id-here",
		},
		UI: UIConfig{
			Theme:     "default",
			PageSize:  20,
			ShowDebug: false,
		},
	}

	return cfg.Save(configPath)
}
