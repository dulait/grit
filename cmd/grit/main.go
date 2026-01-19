// Command grit is a CLI tool for managing GitHub issues with LLM assistance.
package main

import (
	"fmt"
	"os"

	"github.com/dulait/grit/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
