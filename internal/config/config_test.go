package config

import (
	"testing"
)

func TestNewDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig()

	if cfg == nil {
		t.Fatal("NewDefaultConfig returned nil")
	}

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %s, want %s", cfg.Model, DefaultModel)
	}

	if cfg.Provider != DefaultProvider {
		t.Errorf("Provider = %s, want %s", cfg.Provider, DefaultProvider)
	}

	if cfg.Temperature != DefaultTemperature {
		t.Errorf("Temperature = %f, want %f", cfg.Temperature, DefaultTemperature)
	}

	if cfg.APIKeys == nil {
		t.Error("APIKeys map should be initialized")
	}
}

func TestNewManager(t *testing.T) {
	mgr := NewManager()

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	if mgr.config == nil {
		t.Error("Manager config should be initialized")
	}

	if mgr.keyManager == nil {
		t.Error("Manager keyManager should be initialized")
	}
}

func TestNewKeyManager(t *testing.T) {
	km := NewKeyManager()

	if km == nil {
		t.Fatal("NewKeyManager returned nil")
	}

	if km.service != KeyringService {
		t.Errorf("service = %s, want %s", km.service, KeyringService)
	}
}

// TestGetConfig tests retrieving configuration
func TestGetConfig(t *testing.T) {
	mgr := NewManager()
	cfg := mgr.GetConfig()

	if cfg == nil {
		t.Fatal("GetConfig returned nil")
	}

	if cfg.Model != DefaultModel {
		t.Errorf("Model = %s, want %s", cfg.Model, DefaultModel)
	}
}

// TestSetConfig tests updating configuration
func TestSetConfig(t *testing.T) {
	mgr := NewManager()
	newCfg := &Config{
		Model:       "gpt-4",
		Provider:    "openai",
		Temperature: 0.5,
		MaxTokens:   2000,
		LogLevel:    "debug",
		LogDir:      "/tmp/logs",
		APIKeys:     make(map[string]string),
	}

	mgr.SetConfig(newCfg)
	cfg := mgr.GetConfig()

	if cfg.Model != "gpt-4" {
		t.Errorf("Model = %s, want gpt-4", cfg.Model)
	}

	if cfg.Provider != "openai" {
		t.Errorf("Provider = %s, want openai", cfg.Provider)
	}

	if cfg.Temperature != 0.5 {
		t.Errorf("Temperature = %f, want 0.5", cfg.Temperature)
	}
}

// TestGetKeyManager tests retrieving key manager
func TestGetKeyManager(t *testing.T) {
	mgr := NewManager()
	km := mgr.GetKeyManager()

	if km == nil {
		t.Fatal("GetKeyManager returned nil")
	}

	if km.service != KeyringService {
		t.Errorf("service = %s, want %s", km.service, KeyringService)
	}
}

// TestKeyManagerValidation tests KeyManager input validation
func TestKeyManagerValidation(t *testing.T) {
	km := NewKeyManager()

	// Test SetKey with empty provider
	err := km.SetKey("", "test-key")
	if err == nil {
		t.Error("expected error for empty provider")
	}

	// Test SetKey with empty API key
	err = km.SetKey("test-provider", "")
	if err == nil {
		t.Error("expected error for empty API key")
	}

	// Test GetKey with empty provider
	_, err = km.GetKey("")
	if err == nil {
		t.Error("expected error for empty provider")
	}

	// Test DeleteKey with empty provider
	err = km.DeleteKey("")
	if err == nil {
		t.Error("expected error for empty provider")
	}
}

// TestKeyManagerGetNonExistent tests getting non-existent key
func TestKeyManagerGetNonExistent(t *testing.T) {
	km := NewKeyManager()

	// Try to get a key that doesn't exist
	_, err := km.GetKey("nonexistent-provider-xyz123")
	if err == nil {
		t.Error("expected error when getting non-existent key")
	}
}

// TestKeyManagerHasKey tests checking for key existence
func TestKeyManagerHasKey(t *testing.T) {
	km := NewKeyManager()

	// Check for a provider that definitely doesn't exist
	exists := km.HasKey("nonexistent-provider-xyz123")
	if exists {
		t.Error("HasKey returned true for non-existent provider")
	}
}

// TestListProviders tests listing providers
func TestListProviders(t *testing.T) {
	km := NewKeyManager()

	// List providers - should return empty or only existing providers
	providers := km.ListProviders()

	// Can be nil or empty slice - both are valid
	// Just check it doesn't panic
	_ = providers
}

// TestConfigDefaults tests that all default values are set correctly
func TestConfigDefaults(t *testing.T) {
	cfg := NewDefaultConfig()

	tests := []struct {
		name  string
		got   interface{}
		want  interface{}
	}{
		{"Model", cfg.Model, DefaultModel},
		{"Provider", cfg.Provider, DefaultProvider},
		{"Temperature", cfg.Temperature, DefaultTemperature},
		{"MaxTokens", cfg.MaxTokens, DefaultMaxTokens},
		{"LogLevel", cfg.LogLevel, DefaultLogLevel},
		{"LogDir", cfg.LogDir, DefaultLogDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}

	if cfg.APIKeys == nil {
		t.Error("APIKeys map should be initialized")
	}
}

// TestManagerAPIKeyCaching tests API key caching in memory
func TestManagerAPIKeyCaching(t *testing.T) {
	mgr := NewManager()

	// Manually add key to cache
	mgr.config.APIKeys["test-provider"] = "test-key"

	// Should return cached key without hitting keychain
	key, err := mgr.GetAPIKey("test-provider")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if key != "test-key" {
		t.Errorf("got key %s, want test-key", key)
	}
}

// TestManagerHasAPIKey tests checking for API key existence
func TestManagerHasAPIKey(t *testing.T) {
	mgr := NewManager()

	// Add key to memory cache
	mgr.config.APIKeys["cached-provider"] = "test-key"

	// Should find cached key
	if !mgr.HasAPIKey("cached-provider") {
		t.Error("expected HasAPIKey to return true for cached key")
	}

	// Should not find non-existent key
	if mgr.HasAPIKey("nonexistent-provider-xyz123") {
		t.Error("expected HasAPIKey to return false for non-existent key")
	}
}
