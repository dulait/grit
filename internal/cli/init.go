package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dulait/grit/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize grit in the current directory",
	Long:  "Creates a .grit directory with configuration for connecting to a GitHub repository.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	if config.Exists(cwd) {
		return fmt.Errorf("grit already initialized in this directory")
	}

	reader := bufio.NewReader(os.Stdin)

	owner, err := prompt(reader, "GitHub owner (user or org)")
	if err != nil {
		return err
	}

	repo, err := prompt(reader, "Repository name")
	if err != nil {
		return err
	}

	provider, err := promptWithDefault(reader, "LLM provider (anthropic/openai/ollama)", "anthropic")
	if err != nil {
		return err
	}

	model, err := promptModel(reader, provider)
	if err != nil {
		return err
	}

	llmCfg := config.LLMConfig{
		Provider: provider,
		Model:    model,
	}

	if provider == "ollama" {
		baseURL, err := promptWithDefault(reader, "Ollama base URL", "http://localhost:11434")
		if err != nil {
			return err
		}
		llmCfg.BaseURL = baseURL
	}

	cfg := &config.Config{
		Version: 1,
		Project: config.ProjectConfig{
			Owner: owner,
			Repo:  repo,
		},
		LLM: llmCfg,
	}

	if err := config.Save(cwd, cfg); err != nil {
		return err
	}

	if err := config.WriteGitignore(cwd); err != nil {
		return fmt.Errorf("creating .gitignore: %w", err)
	}

	fmt.Printf("Initialized grit for %s/%s using %s (%s)\n", owner, repo, provider, model)
	return nil
}

func prompt(reader *bufio.Reader, label string) (string, error) {
	fmt.Printf("%s: ", label)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	return strings.TrimSpace(input), nil
}

func promptWithDefault(reader *bufio.Reader, label, defaultVal string) (string, error) {
	fmt.Printf("%s [%s]: ", label, defaultVal)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("reading input: %w", err)
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

func promptModel(reader *bufio.Reader, provider string) (string, error) {
	defaults := map[string]string{
		"anthropic": "claude-sonnet-4-20250514",
		"openai":    "gpt-4o",
		"ollama":    "llama3.2",
	}
	defaultModel := defaults[provider]
	if defaultModel == "" {
		defaultModel = "claude-sonnet-4-20250514"
	}
	return promptWithDefault(reader, "Model", defaultModel)
}
