# Contributing

Thanks for your interest in contributing to grit! This guide covers the development setup, project structure, and contribution workflow.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25 or later
- [Git](https://git-scm.com)
- A GitHub account

## Getting started

```bash
# Clone the repository
git clone https://github.com/dulait/grit.git
cd grit

# Install dependencies
go mod download

# Build
make build

# Run
./bin/grit version
```

## Project structure

```
cmd/grit/              Entry point (main.go)
internal/
  cli/                 Cobra command definitions
  config/              Configuration loading, token storage
  github/              GitHub API client
  llm/                 LLM provider clients (Anthropic, Groq, Ollama)
  service/             Business logic layer
  tui/                 Bubble Tea TUI components
  updater/             Self-update logic
  errors/              Shared error types
docs/                  User-facing documentation
```

### Key conventions

- **CLI commands** live in `internal/cli/`. Each file defines one or more `cobra.Command` variables and registers them in an `init()` function.
- **TUI screens** live in `internal/tui/`. Each screen is a Bubble Tea `Model` with `Init`, `Update`, and `View` methods.
- **Business logic** lives in `internal/service/`. CLI and TUI code call into the service layer rather than hitting the GitHub API directly.
- **Package documentation** is in `doc.go` files within each package.

## Common tasks

### Build with version info

```bash
make build
```

This injects the version, commit SHA, and build date via ldflags.

### Run tests

```bash
make test
```

### Format and lint

```bash
make fmt
make vet
```

### Install locally

```bash
make install
```

Installs to your `$GOPATH/bin`.

## Submitting a pull request

1. Fork the repository and create a feature branch from `main`
2. Make your changes
3. Run `make fmt && make vet && make test` to verify
4. Commit with a clear message following the project's commit style:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `chore:` for maintenance tasks
5. Push to your fork and open a pull request against `main`

## Adding a new CLI command

1. Create a new file in `internal/cli/` (or add to an existing one)
2. Define the `cobra.Command` variable
3. Register it in an `init()` function with `rootCmd.AddCommand()` or as a subcommand
4. If the command needs config/auth, follow the pattern in `issue.go` using `buildGitHubClient()` and `buildLLMClient()`

## Adding a new TUI screen

1. Create a new file in `internal/tui/`
2. Define a model struct implementing the Bubble Tea `Model` interface
3. Add a new `screen` constant in `app.go`
4. Add navigation messages in `messages.go`
5. Wire routing in the main `appModel.Update()` method

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Keep functions focused and short
- Use descriptive names — avoid abbreviations unless they are widely understood
- Error messages should start with a lowercase letter and not end with punctuation
