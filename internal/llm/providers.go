package llm

// ProviderInfo describes an LLM provider available during grit init.
type ProviderInfo struct {
	Name         string
	Description  string
	RequiresKey  bool
	DefaultModel string
	DefaultURL   string // only for providers with configurable endpoint
}

// Providers returns the ordered list of available LLM providers.
func Providers() []ProviderInfo {
	return []ProviderInfo{
		{
			Name:        "none",
			Description: "No AI features, manual issue creation only",
		},
		{
			Name:         "groq",
			Description:  "Free cloud AI (requires API key from groq.com)",
			RequiresKey:  true,
			DefaultModel: "llama-3.3-70b-versatile",
		},
		{
			Name:         "ollama",
			Description:  "Local AI (requires Ollama installed, ~4GB)",
			DefaultModel: "llama3.2",
			DefaultURL:   "http://localhost:11434",
		},
		{
			Name:         "anthropic",
			Description:  "Claude AI (paid, highest quality)",
			RequiresKey:  true,
			DefaultModel: "claude-sonnet-4-20250514",
		},
	}
}

// ProviderByName returns the provider info for the given name, or nil if not found.
func ProviderByName(name string) *ProviderInfo {
	for _, p := range Providers() {
		if p.Name == name {
			return &p
		}
	}
	return nil
}
