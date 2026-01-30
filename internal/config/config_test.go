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
