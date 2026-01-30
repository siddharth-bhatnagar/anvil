package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Manager handles configuration management
type Manager struct {
	config     *Config
	keyManager *KeyManager
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config:     NewDefaultConfig(),
		keyManager: NewKeyManager(),
	}
}

// Load loads configuration from file and environment variables
func (m *Manager) Load() error {
	// Expand home directory in config dir
	configDir := os.ExpandEnv(DefaultConfigDir)
	if configDir[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(home, configDir[2:])
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set config file path
	m.configPath = filepath.Join(configDir, DefaultConfigFile)

	// Configure Viper
	viper.SetConfigFile(m.configPath)
	viper.SetConfigType("yaml")

	// Set environment variable prefix
	viper.SetEnvPrefix("ANVIL")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("model", DefaultModel)
	viper.SetDefault("provider", DefaultProvider)
	viper.SetDefault("temperature", DefaultTemperature)
	viper.SetDefault("max_tokens", DefaultMaxTokens)
	viper.SetDefault("log_level", DefaultLogLevel)
	viper.SetDefault("log_dir", DefaultLogDir)

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		// Check for various "file not found" error types
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Also check for os.PathError (file doesn't exist)
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to read config file: %w", err)
			}
		}
		// Config file doesn't exist yet, will be created on Save()
	}

	// Unmarshal into config struct
	if err := viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load API keys from keychain
	providers := m.keyManager.ListProviders()
	for _, provider := range providers {
		if apiKey, err := m.keyManager.GetKey(provider); err == nil {
			m.config.APIKeys[provider] = apiKey
		}
	}

	return nil
}

// Save saves the configuration to file
// Note: API keys are NOT saved to the file, they remain in the OS keychain
func (m *Manager) Save() error {
	// Update Viper with current config values
	viper.Set("model", m.config.Model)
	viper.Set("provider", m.config.Provider)
	viper.Set("temperature", m.config.Temperature)
	viper.Set("max_tokens", m.config.MaxTokens)
	viper.Set("log_level", m.config.LogLevel)
	viper.Set("log_dir", m.config.LogDir)

	// Write config file
	if err := viper.WriteConfig(); err != nil {
		// If config file doesn't exist, create it
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := viper.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	return nil
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *Config {
	return m.config
}

// SetConfig updates the configuration
func (m *Manager) SetConfig(config *Config) {
	m.config = config
}

// GetKeyManager returns the key manager for API key operations
func (m *Manager) GetKeyManager() *KeyManager {
	return m.keyManager
}

// GetAPIKey retrieves an API key for the given provider
func (m *Manager) GetAPIKey(provider string) (string, error) {
	// Check if key is already loaded in memory
	if apiKey, ok := m.config.APIKeys[provider]; ok && apiKey != "" {
		return apiKey, nil
	}

	// Try to load from keychain
	apiKey, err := m.keyManager.GetKey(provider)
	if err != nil {
		return "", err
	}

	// Cache in memory
	m.config.APIKeys[provider] = apiKey
	return apiKey, nil
}

// SetAPIKey stores an API key for the given provider
func (m *Manager) SetAPIKey(provider, apiKey string) error {
	// Store in keychain
	if err := m.keyManager.SetKey(provider, apiKey); err != nil {
		return err
	}

	// Cache in memory
	m.config.APIKeys[provider] = apiKey
	return nil
}

// DeleteAPIKey removes an API key for the given provider
func (m *Manager) DeleteAPIKey(provider string) error {
	// Remove from keychain
	if err := m.keyManager.DeleteKey(provider); err != nil {
		return err
	}

	// Remove from memory
	delete(m.config.APIKeys, provider)
	return nil
}

// HasAPIKey checks if an API key exists for the given provider
func (m *Manager) HasAPIKey(provider string) bool {
	// Check memory first
	if apiKey, ok := m.config.APIKeys[provider]; ok && apiKey != "" {
		return true
	}

	// Check keychain
	return m.keyManager.HasKey(provider)
}
