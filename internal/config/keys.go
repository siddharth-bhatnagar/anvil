package config

import (
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	// KeyringService is the service name used in the OS keychain
	KeyringService = "anvil"
)

// KeyManager handles secure storage of API keys in the OS keychain
type KeyManager struct {
	service string
}

// NewKeyManager creates a new KeyManager
func NewKeyManager() *KeyManager {
	return &KeyManager{
		service: KeyringService,
	}
}

// SetKey stores an API key in the OS keychain
// provider: the LLM provider name (e.g., "anthropic", "openai")
// apiKey: the API key to store
func (km *KeyManager) SetKey(provider, apiKey string) error {
	if provider == "" {
		return fmt.Errorf("provider name cannot be empty")
	}
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	err := keyring.Set(km.service, provider, apiKey)
	if err != nil {
		return fmt.Errorf("failed to store API key for %s: %w", provider, err)
	}

	return nil
}

// GetKey retrieves an API key from the OS keychain
// provider: the LLM provider name (e.g., "anthropic", "openai")
// Returns the API key or an error if not found
func (km *KeyManager) GetKey(provider string) (string, error) {
	if provider == "" {
		return "", fmt.Errorf("provider name cannot be empty")
	}

	apiKey, err := keyring.Get(km.service, provider)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", fmt.Errorf("no API key found for provider %s", provider)
		}
		return "", fmt.Errorf("failed to retrieve API key for %s: %w", provider, err)
	}

	return apiKey, nil
}

// DeleteKey removes an API key from the OS keychain
// provider: the LLM provider name (e.g., "anthropic", "openai")
func (km *KeyManager) DeleteKey(provider string) error {
	if provider == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	err := keyring.Delete(km.service, provider)
	if err != nil {
		if err == keyring.ErrNotFound {
			return fmt.Errorf("no API key found for provider %s", provider)
		}
		return fmt.Errorf("failed to delete API key for %s: %w", provider, err)
	}

	return nil
}

// HasKey checks if an API key exists for the given provider
// provider: the LLM provider name (e.g., "anthropic", "openai")
func (km *KeyManager) HasKey(provider string) bool {
	_, err := km.GetKey(provider)
	return err == nil
}

// ListProviders returns a list of providers that have API keys stored
// Note: This is a best-effort implementation as the keyring API doesn't
// provide a native way to list all keys. We check common providers.
func (km *KeyManager) ListProviders() []string {
	commonProviders := []string{"anthropic", "openai", "google", "ollama"}
	var providers []string

	for _, provider := range commonProviders {
		if km.HasKey(provider) {
			providers = append(providers, provider)
		}
	}

	return providers
}
