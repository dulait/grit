package llm

import (
	"fmt"

	"github.com/dulait/grit/internal/config"
)

func NewClient(cfg config.LLMConfig, apiKey string) (Client, error) {
	switch cfg.Provider {
	case "anthropic":
		if apiKey == "" {
			return nil, fmt.Errorf("anthropic requires an API key; set GRIT_LLM_KEY or run 'grit auth llm'")
		}
		return NewAnthropicClient(apiKey, cfg.Model), nil
	case "ollama":
		return NewOllamaClient(cfg.BaseURL, cfg.Model), nil
	case "openai":
		if apiKey == "" {
			return nil, fmt.Errorf("openai requires an API key; set GRIT_LLM_KEY or run 'grit auth llm'")
		}
		return nil, fmt.Errorf("openai provider not yet implemented")
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", cfg.Provider)
	}
}
