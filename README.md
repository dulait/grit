# Grit

A CLI and TUI tool for LLM-assisted GitHub issue management.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/dulait/grit)](https://github.com/dulait/grit/releases/latest)

## Features

- **Dual interface** — full CLI for scripting and an interactive TUI for day-to-day work
- **LLM-assisted issue creation** — describe a problem in plain English; grit generates a structured issue
- **Multiple LLM providers** — Anthropic (Claude), Groq, Ollama (local), or no AI at all
- **Complete issue management** — create, list, view, edit, close, assign, comment, link, and search
- **Sub-issues** — create child issues linked to a parent
- **Issue linking** — relate issues with typed relationships (blocks, duplicates, parent/child, etc.)
- **Search** — find issues with GitHub's search API, filtered by state and label
- **Self-update** — run `grit update` to fetch the latest release from GitHub
- **Cross-platform** — Linux, macOS, and Windows on amd64 and arm64

<!-- TODO: Add a terminal recording / screenshot of the TUI here -->

## Installation

### Using Go

Requires Go 1.25 or later.

```bash
go install github.com/dulait/grit/cmd/grit@latest
```

### From GitHub Releases

Download the binary for your platform from the [releases page](https://github.com/dulait/grit/releases).

**Linux / macOS:**

```bash
curl -LO https://github.com/dulait/grit/releases/latest/download/grit_VERSION_OS_ARCH.tar.gz
tar -xzf grit_VERSION_OS_ARCH.tar.gz
sudo mv grit /usr/local/bin/
```

**Windows:**

1. Download the `.zip` file from the releases page
2. Extract `grit.exe`
3. Move it to a directory in your `PATH`

### Self-update

If you already have grit installed from a release binary:

```bash
grit update
```

### Verify

```bash
grit version
```

## Quick Start

```bash
# 1. Initialize grit in your project
grit init

# 2. Authenticate with GitHub
grit auth login

# 3. Create an issue with AI assistance
grit issue create "describe your issue here"

# 4. Launch the interactive TUI
grit
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Installation, setup, and your first issue |
| [CLI Reference](docs/cli-reference.md) | Every command, flag, and argument |
| [TUI Guide](docs/tui-guide.md) | Interactive mode walkthrough and keybindings |
| [Configuration](docs/configuration.md) | Config file format and LLM provider setup |
| [Updating](docs/updating.md) | Keeping grit up to date |
| [Contributing](CONTRIBUTING.md) | Development setup and contribution guidelines |

## License

MIT License — see [LICENSE](LICENSE) for details.
