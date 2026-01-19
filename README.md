# Grit

A CLI tool for LLM-assisted GitHub issue management.

## Installation

### Using Go (requires Go 1.22+)

```bash
go install github.com/dulait/grit/cmd/grit@latest
```

### From GitHub Releases

Download the appropriate binary for your OS from the [releases page](https://github.com/dulait/grit/releases).

#### Linux/macOS

```bash
# Download (replace VERSION and OS/ARCH as needed)
curl -LO https://github.com/dulait/grit/releases/latest/download/grit_VERSION_OS_ARCH.tar.gz
tar -xzf grit_VERSION_OS_ARCH.tar.gz
sudo mv grit /usr/local/bin/
```

#### Windows

1. Download the `.zip` file from releases
2. Extract `grit.exe`
3. Add to your PATH or move to a directory in your PATH

### Verify Installation

```bash
grit version
```

## Quick Start

1. Navigate to your project directory
2. Initialize grit:
   ```bash
   grit init
   ```
3. Authenticate with GitHub:
   ```bash
   grit auth login
   ```
4. Create your first issue:
   ```bash
   grit issue create "describe your issue here"
   ```

## License

MIT License - see [LICENSE](LICENSE) for details.
