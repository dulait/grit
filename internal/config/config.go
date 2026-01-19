package config

// Config represents the complete grit configuration for a project.
type Config struct {
	Version int           `yaml:"version"`
	Project ProjectConfig `yaml:"project"`
	LLM     LLMConfig     `yaml:"llm"`
}

// ProjectConfig defines the GitHub project settings.
type ProjectConfig struct {
	Owner       string   `yaml:"owner"`
	Repo        string   `yaml:"repo"`
	IssuePrefix string   `yaml:"issue_prefix,omitempty"`
	Labels      []string `yaml:"labels,omitempty"`
	Assignees   []string `yaml:"assignees,omitempty"`
}

// LLMConfig defines the LLM provider settings.
type LLMConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	BaseURL  string `yaml:"base_url,omitempty"`
}

// DefaultLLMConfig returns the default LLM configuration using Anthropic.
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider: "anthropic",
		Model:    "claude-sonnet-4-20250514",
	}
}
