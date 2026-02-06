package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/dulait/grit/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store GitHub PAT for the current project",
	RunE:  runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	RunE:  runAuthStatus,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFromWorkingDir()
	if err != nil {
		return err
	}

	projectKey := config.ProjectKey(cfg)

	fmt.Printf("Enter GitHub PAT for %s: ", projectKey)

	tokenBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading token: %w", err)
		}
		tokenBytes = []byte(strings.TrimSpace(token))
	}
	fmt.Println()

	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	store := config.NewCompositeTokenStore()
	if err := store.Set(projectKey, token); err != nil {
		return err
	}

	fmt.Printf("Token stored for %s\n", projectKey)
	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadFromWorkingDir()
	if err != nil {
		return err
	}

	projectKey := config.ProjectKey(cfg)
	store := config.NewCompositeTokenStore()

	token, err := store.Get(projectKey)
	if err != nil {
		fmt.Printf("GitHub: not authenticated (%v)\n", err)
	} else {
		maskedToken := token[:4] + "..." + token[len(token)-4:]
		fmt.Printf("GitHub: authenticated (%s)\n", maskedToken)
	}

	if cfg.LLM.Provider == "none" {
		fmt.Println("LLM: none (AI features disabled)")
	} else {
		fmt.Printf("LLM: %s (%s)\n", cfg.LLM.Provider, cfg.LLM.Model)
		if cfg.LLM.Provider == "ollama" {
			fmt.Println("  No API key required for Ollama")
		}
	}

	return nil
}
