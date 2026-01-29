package config

// Default configuration values
const (
	// DefaultModel is the default LLM model to use
	DefaultModel = "claude-sonnet-4-5"

	// DefaultProvider is the default LLM provider
	DefaultProvider = "anthropic"

	// DefaultTemperature is the default temperature for LLM requests
	DefaultTemperature = 0.7

	// DefaultMaxTokens is the default maximum tokens for LLM responses
	DefaultMaxTokens = 4096

	// DefaultConfigDir is the default directory for Anvil configuration
	DefaultConfigDir = "~/.anvil"

	// DefaultConfigFile is the default configuration file name
	DefaultConfigFile = "config.yaml"

	// DefaultLogDir is the default directory for log files
	DefaultLogDir = "~/.anvil/logs"

	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"
)

// Config represents the application configuration
type Config struct {
	// LLM configuration
	Model       string  `mapstructure:"model"`
	Provider    string  `mapstructure:"provider"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`

	// Logging configuration
	LogLevel string `mapstructure:"log_level"`
	LogDir   string `mapstructure:"log_dir"`

	// API Keys (stored in OS keychain, not in file)
	// These are not part of the config file
	APIKeys map[string]string `mapstructure:"-"`
}

// NewDefaultConfig returns a new Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		Model:       DefaultModel,
		Provider:    DefaultProvider,
		Temperature: DefaultTemperature,
		MaxTokens:   DefaultMaxTokens,
		LogLevel:    DefaultLogLevel,
		LogDir:      DefaultLogDir,
		APIKeys:     make(map[string]string),
	}
}
