# Updating

## Self-update

If you installed grit from a GitHub release binary, update to the latest version with:

```bash
grit update
```

This will:

1. Check the [latest release](https://github.com/dulait/grit/releases/latest) on GitHub
2. Download the correct binary for your OS and architecture
3. Replace the current executable in place

If you are already on the latest version, grit will tell you and exit.

### Dev builds

If you are running a development build (version shows `dev`), grit cannot determine your current version. It will ask for confirmation before updating:

```
You are running a dev build — the current version cannot be determined.
Update to the latest release anyway? [Y/n]:
```

## Other update methods

### Go install

If you installed with `go install`:

```bash
go install github.com/dulait/grit/cmd/grit@latest
```

### Manual download

Download the latest archive from the [releases page](https://github.com/dulait/grit/releases) and replace the binary in your `PATH`.

## Checking your version

```bash
grit version
```

Displays the current version, commit SHA, and build date.
