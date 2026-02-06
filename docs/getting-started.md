# Getting Started

This guide walks you through installing grit, setting up a project, and creating your first issue.

## Prerequisites

- A GitHub account
- A GitHub repository you want to manage issues for
- A [GitHub Personal Access Token](https://github.com/settings/tokens) with `repo` scope

## Installation

### Option 1: Go install

Requires Go 1.25 or later.

```bash
go install github.com/dulait/grit/cmd/grit@latest
```

### Option 2: Download a release binary

Download the archive for your platform from the [releases page](https://github.com/dulait/grit/releases).

**Linux / macOS:**

```bash
# Download (replace VERSION and OS/ARCH as needed)
curl -LO https://github.com/dulait/grit/releases/latest/download/grit_VERSION_OS_ARCH.tar.gz
tar -xzf grit_VERSION_OS_ARCH.tar.gz
sudo mv grit /usr/local/bin/
```

**Windows:**

1. Download the `.zip` file from the releases page
2. Extract `grit.exe`
3. Move it to a directory in your `PATH`

### Option 3: Self-update an existing installation

```bash
grit update
```

### Verify

```bash
grit version
```

## Initialize a project

Navigate to the root of your Git repository and run:

```bash
grit init
```

The wizard will prompt you for:

1. **GitHub owner** — your GitHub username or organization name
2. **Repository name** — the name of the repo on GitHub
3. **LLM provider** — choose one:
   - **none** — no AI features, manual issue creation only
   - **groq** — free cloud AI (requires a free API key from [groq.com](https://console.groq.com))
   - **ollama** — local AI, runs on your machine (requires [Ollama](https://ollama.com) installed with ~4 GB of disk space)
   - **anthropic** — Claude AI, highest quality (requires a paid API key from [Anthropic](https://console.anthropic.com))
4. **API key** — if the chosen provider requires one
5. **Model** — defaults are provided for each provider; press Enter to accept

This creates a `.grit/` directory in your project with a `config.yaml` file. See the [Configuration](configuration.md) guide for details on the config format.

## Authenticate with GitHub

```bash
grit auth login
```

Paste your GitHub Personal Access Token when prompted. The token is stored securely in your system keyring.

To verify authentication:

```bash
grit auth status
```

## Create your first issue

grit supports three ways to create issues.

### AI-assisted (recommended)

Describe the problem in plain English and let the LLM generate a structured issue:

```bash
grit issue create "the login page returns a 500 error when the email field is empty"
```

grit will show you the generated issue and ask for confirmation before creating it.

### Interactive

Run the command with no arguments to be prompted for each field:

```bash
grit issue create
```

### Explicit

Specify everything with flags:

```bash
grit issue create -t "Fix login 500 error" -d "Submitting the login form with an empty email causes a server error." -l "bug" -a "username"
```

## Launch the TUI

Run `grit` with no subcommand to open the interactive terminal interface:

```bash
grit
```

You'll see a list of open issues for your repository. From here you can browse, create, edit, close, and search issues — all without leaving the terminal. See the [TUI Guide](tui-guide.md) for a full walkthrough.

## Next steps

- [CLI Reference](cli-reference.md) — every command, flag, and argument
- [TUI Guide](tui-guide.md) — interactive mode walkthrough and keybindings
- [Configuration](configuration.md) — config file format and LLM provider setup
- [Updating](updating.md) — keeping grit up to date
