package config

import (
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	envVarName       = "GRIT_PAT"
	envVarLLMKey     = "GRIT_LLM_KEY"
	keyringService   = "grit"
	keyringLLMPrefix = "grit-llm"
)

type TokenStore interface {
	Get(project string) (string, error)
	Set(project string, token string) error
	Delete(project string) error
}

type EnvTokenStore struct{}

func (s *EnvTokenStore) Get(project string) (string, error) {
	token := os.Getenv(envVarName)
	if token == "" {
		return "", fmt.Errorf("environment variable %s not set", envVarName)
	}
	return token, nil
}

func (s *EnvTokenStore) Set(project string, token string) error {
	return fmt.Errorf("cannot set environment variable from application; set %s manually", envVarName)
}

func (s *EnvTokenStore) Delete(project string) error {
	return fmt.Errorf("cannot unset environment variable from application; unset %s manually", envVarName)
}

type KeyringTokenStore struct{}

func (s *KeyringTokenStore) Get(project string) (string, error) {
	token, err := keyring.Get(keyringService, project)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no token stored for %s", project)
	}
	if err != nil {
		return "", fmt.Errorf("accessing keyring: %w", err)
	}
	return token, nil
}

func (s *KeyringTokenStore) Set(project string, token string) error {
	if err := keyring.Set(keyringService, project, token); err != nil {
		return fmt.Errorf("storing token in keyring: %w", err)
	}
	return nil
}

func (s *KeyringTokenStore) Delete(project string) error {
	if err := keyring.Delete(keyringService, project); err != nil {
		return fmt.Errorf("deleting token from keyring: %w", err)
	}
	return nil
}

type CompositeTokenStore struct {
	stores []TokenStore
}

func NewCompositeTokenStore() *CompositeTokenStore {
	return &CompositeTokenStore{
		stores: []TokenStore{
			&EnvTokenStore{},
			&KeyringTokenStore{},
		},
	}
}

func (s *CompositeTokenStore) Get(project string) (string, error) {
	for _, store := range s.stores {
		token, err := store.Get(project)
		if err == nil {
			return token, nil
		}
	}
	return "", fmt.Errorf("no token found; run 'grit auth login' or set %s", envVarName)
}

func (s *CompositeTokenStore) Set(project string, token string) error {
	return (&KeyringTokenStore{}).Set(project, token)
}

func (s *CompositeTokenStore) Delete(project string) error {
	return (&KeyringTokenStore{}).Delete(project)
}

func ProjectKey(cfg *Config) string {
	return fmt.Sprintf("%s/%s", cfg.Project.Owner, cfg.Project.Repo)
}

func GetLLMKey(provider string) (string, error) {
	key := os.Getenv(envVarLLMKey)
	if key != "" {
		return key, nil
	}

	key, err := keyring.Get(keyringLLMPrefix, provider)
	if err == keyring.ErrNotFound {
		return "", fmt.Errorf("no LLM API key found for %s; set %s or run 'grit auth llm'", provider, envVarLLMKey)
	}
	if err != nil {
		return "", fmt.Errorf("accessing keyring: %w", err)
	}
	return key, nil
}

func SetLLMKey(provider, key string) error {
	if err := keyring.Set(keyringLLMPrefix, provider, key); err != nil {
		return fmt.Errorf("storing LLM key in keyring: %w", err)
	}
	return nil
}
