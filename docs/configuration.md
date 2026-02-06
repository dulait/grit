# Configuration

grit stores per-project configuration in a `.grit/config.yaml` file at the root of your repository. This file is created by `grit init`.

## Config file format

```yaml
version: 1

project:
  owner: "your-username"        # GitHub user or organization
  repo: "your-repo"             # Repository name
  issue_prefix: ""              # Optional prefix for issue titles
  labels:                       # Optional list of allowed labels
    - bug
    - feature
    - docs
  assignees: []                 # Optional list of default assignees

llm:
  provider: "groq"              # LLM provider: none, groq, ollama, anthropic
  model: "llama-3.3-70b-versatile"  # Model name
  base_url: ""                  # Only used by ollama
```

### Project settings

| Field | Required | Description |
|-------|----------|-------------|
| `owner` | Yes | GitHub username or organization that owns the repo |
| `repo` | Yes | Repository name |
| `issue_prefix` | No | String prepended to issue titles when creating issues |
| `labels` | No | Allowed labels — used for validation during issue creation |
| `assignees` | No | Default assignees |

### LLM settings

| Field | Required | Description |
|-------|----------|-------------|
| `provider` | Yes | One of `none`, `groq`, `ollama`, `anthropic` |
| `model` | Yes (unless `none`) | Model identifier for the chosen provider |
| `base_url` | Only for `ollama` | Ollama server URL (default: `http://localhost:11434`) |

## LLM providers

### none

No AI features. Issues are created manually — you provide the title, description, labels, and assignees yourself.

No API key or additional configuration needed.

### groq

Free cloud AI powered by Groq's inference engine.

- **Default model:** `llama-3.3-70b-versatile`
- **Requires:** API key (free) from [console.groq.com](https://console.groq.com)
- **Setup:** `grit init` will prompt for the key, or set the `GRIT_LLM_KEY` environment variable

### ollama

Local AI that runs on your machine. No data leaves your computer.

- **Default model:** `llama3.2`
- **Default URL:** `http://localhost:11434`
- **Requires:** [Ollama](https://ollama.com) installed and running (~4 GB disk space)
- **No API key needed**
- **Setup:** install Ollama, pull a model (`ollama pull llama3.2`), then run `grit init`

To use a different Ollama server, set `base_url` in the config or specify it during `grit init`.

### anthropic

Claude AI from Anthropic. Highest quality results.

- **Default model:** `claude-sonnet-4-20250514`
- **Requires:** API key (paid) from [console.anthropic.com](https://console.anthropic.com)
- **Setup:** `grit init` will prompt for the key, or set the `GRIT_LLM_KEY` environment variable

## Authentication

### GitHub token

grit needs a GitHub Personal Access Token (PAT) with `repo` scope to manage issues.

**Option 1: Store via grit (recommended)**

```bash
grit auth login
```

Stores the token in your system keyring, scoped to the current project.

**Option 2: Environment variable**

```bash
export GRIT_PAT="ghp_your_token_here"
```

The environment variable takes priority over the keyring.

### LLM API key

For providers that require an API key (groq, anthropic):

**Option 1: Set during init**

`grit init` prompts for the API key and stores it in the system keyring.

**Option 2: Environment variable**

```bash
export GRIT_LLM_KEY="your_api_key_here"
```

The environment variable takes priority over the keyring.

### Token lookup order

For both GitHub and LLM tokens, grit checks in this order:

1. Environment variable (`GRIT_PAT` or `GRIT_LLM_KEY`)
2. System keyring

## The `.grit/` directory

The `.grit/` directory contains:

- `config.yaml` — project configuration
- `.gitignore` — ensures sensitive local files are not committed

You should commit `.grit/config.yaml` to your repository so teammates can share the same project configuration. API keys are **not** stored in this file — they live in the system keyring or environment variables.
