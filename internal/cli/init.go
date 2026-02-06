package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/dulait/grit/internal/config"
	"github.com/dulait/grit/internal/llm"
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

	provider, err := promptProviderMenu(reader)
	if err != nil {
		return err
	}

	llmCfg := config.LLMConfig{
		Provider: provider.Name,
	}

	if provider.Name == "none" {
		fmt.Println("No LLM configured. Use --raw flag for issue creation.")
	} else {
		if provider.RequiresKey {
			key, err := promptAPIKey(provider.Name)
			if err != nil {
				return err
			}
			if err := config.SetLLMKey(provider.Name, key); err != nil {
				return fmt.Errorf("storing API key: %w", err)
			}
		}

		if provider.DefaultURL != "" {
			baseURL, err := promptWithDefault(reader, "Ollama base URL", provider.DefaultURL)
			if err != nil {
				return err
			}
			llmCfg.BaseURL = baseURL

			if err := llm.CheckOllamaConnection(baseURL); err != nil {
				fmt.Printf("Warning: %v\n", err)
			} else {
				fmt.Println("Ollama connection OK")
			}
		}

		model, err := promptWithDefault(reader, "Model", provider.DefaultModel)
		if err != nil {
			return err
		}
		llmCfg.Model = model
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

	if provider.Name == "none" {
		fmt.Printf("Initialized grit for %s/%s (no LLM)\n", owner, repo)
	} else {
		fmt.Printf("Initialized grit for %s/%s using %s (%s)\n", owner, repo, llmCfg.Provider, llmCfg.Model)
	}
	return nil
}

func promptProviderMenu(reader *bufio.Reader) (*llm.ProviderInfo, error) {
	providers := llm.Providers()

	fmt.Println("Select LLM provider:")
	for i, p := range providers {
		fmt.Printf("  %d. %-10s - %s\n", i+1, p.Name, p.Description)
	}

	for {
		fmt.Printf("Choice [1-%d]: ", len(providers))
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("reading input: %w", err)
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || choice < 1 || choice > len(providers) {
			fmt.Printf("Please enter a number between 1 and %d.\n", len(providers))
			continue
		}

		selected := providers[choice-1]
		return &selected, nil
	}
}

func promptAPIKey(provider string) (string, error) {
	fmt.Printf("Enter API key for %s: ", provider)

	keyBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		key, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("reading API key: %w", err)
		}
		keyBytes = []byte(strings.TrimSpace(key))
	}
	fmt.Println()

	key := strings.TrimSpace(string(keyBytes))
	if key == "" {
		return "", fmt.Errorf("API key cannot be empty")
	}
	return key, nil
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
