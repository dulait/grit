# CLI Reference

Complete reference for every grit command, subcommand, and flag.

## Global behavior

Running `grit` with no subcommand launches the interactive [TUI](tui-guide.md).

## Commands

- [`grit init`](#grit-init)
- [`grit auth login`](#grit-auth-login)
- [`grit auth status`](#grit-auth-status)
- [`grit config show`](#grit-config-show)
- [`grit issue create`](#grit-issue-create)
- [`grit issue list`](#grit-issue-list)
- [`grit issue view`](#grit-issue-view)
- [`grit issue edit`](#grit-issue-edit)
- [`grit issue close`](#grit-issue-close)
- [`grit issue comment`](#grit-issue-comment)
- [`grit issue assign`](#grit-issue-assign)
- [`grit issue link`](#grit-issue-link)
- [`grit issue search`](#grit-issue-search)
- [`grit issue sub`](#grit-issue-sub)
- [`grit update`](#grit-update)
- [`grit version`](#grit-version)

---

## `grit init`

Initialize grit in the current directory.

```
grit init
```

Starts an interactive wizard that prompts for GitHub owner/repo, LLM provider, API key, and model. Creates a `.grit/config.yaml` file. See [Configuration](configuration.md) for details on each setting.

---

## `grit auth login`

Store a GitHub Personal Access Token for the current project.

```
grit auth login
```

Prompts for a PAT and stores it securely in the system keyring.

---

## `grit auth status`

Check authentication and LLM configuration status.

```
grit auth status
```

Shows whether a GitHub token is stored and which LLM provider and model are configured.

---

## `grit config show`

Display the current project configuration.

```
grit config show
```

Prints the contents of `.grit/config.yaml` in YAML format.

---

## `grit issue create`

Create a GitHub issue.

```
grit issue create [prompt] [flags]
```

Three modes of operation:

| Mode | Usage | Description |
|------|-------|-------------|
| Interactive | `grit issue create` | Prompts for title, description, labels, and assignees |
| AI-assisted | `grit issue create "prompt"` | LLM generates a structured issue from the prompt |
| Explicit | `grit issue create -t "Title" -d "Body"` | Flags set fields directly |

When using AI-assisted mode, flags override the corresponding generated fields.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | | Issue title |
| `--description` | `-d` | | Issue description / body |
| `--labels` | `-l` | | Comma-separated labels |
| `--assignees` | `-a` | | Comma-separated assignees |
| `--yes` | `-y` | `false` | Skip the confirmation prompt |
| `--raw` | | `false` | Use input verbatim — skip LLM enhancement |

**Examples:**

```bash
# AI-assisted
grit issue create "users can't log in when email is empty"

# Explicit with labels and auto-confirm
grit issue create -t "Fix null pointer in auth" -d "Details here" -l "bug" -y

# Raw mode — no AI processing
grit issue create -t "Update README" -d "Add installation section" --raw
```

---

## `grit issue list`

List repository issues.

```
grit issue list [flags]
```

Displays a paginated list of issues. Interactive pagination prompts let you move between pages.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--state` | `-s` | `open` | Filter by state: `open`, `closed`, or `all` |
| `--assignee` | `-a` | | Filter by assignee username, or `"none"` for unassigned |
| `--label` | `-l` | | Filter by label |
| `--limit` | `-n` | `30` | Results per page |
| `--page` | `-p` | `1` | Page number |

**Examples:**

```bash
# List open issues (default)
grit issue list

# List closed issues, 10 per page
grit issue list -s closed -n 10

# Issues assigned to a specific user
grit issue list -a username

# Unassigned bugs
grit issue list -a none -l bug
```

---

## `grit issue view`

View a single issue.

```
grit issue view <number> [flags]
```

Displays the issue title, state, labels, assignees, timestamps, URL, and full body.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--web` | `-w` | `false` | Open the issue in a browser instead of printing it |

**Examples:**

```bash
grit issue view 42
grit issue view 42 --web
```

---

## `grit issue edit`

Edit an existing issue.

```
grit issue edit <number> [flags]
```

Without flags, displays the current issue and lists the available flags. With flags, applies the specified changes directly.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | | New title |
| `--description` | `-d` | | New description / body |
| `--labels` | `-l` | | Comma-separated labels (replaces all existing labels) |
| `--assignees` | `-a` | | Comma-separated assignees (replaces all existing assignees) |
| `--state` | `-s` | | New state: `open` or `closed` |
| `--enhance` | | `false` | Enhance changes with the LLM |
| `--yes` | `-y` | `false` | Skip the confirmation prompt |

**Examples:**

```bash
# Interactive edit
grit issue edit 42

# Change title and add labels
grit issue edit 42 -t "New title" -l "bug,priority"

# Close an issue via edit
grit issue edit 42 -s closed -y
```

---

## `grit issue close`

Close an issue.

```
grit issue close <number> [reason]
```

Closes the issue. If a reason is provided, it is added as a comment before closing.

**Examples:**

```bash
grit issue close 42
grit issue close 42 "fixed in commit abc1234"
```

---

## `grit issue comment`

Add an AI-generated comment to an issue.

```
grit issue comment <number> <prompt>
```

The LLM generates a comment based on the prompt and the issue context, then posts it.

**Examples:**

```bash
grit issue comment 42 "suggest a fix for this bug"
grit issue comment 42 "summarize the current status"
```

---

## `grit issue assign`

Assign users to an issue.

```
grit issue assign <number> <user> [users...]
```

Adds one or more assignees to the issue.

**Examples:**

```bash
grit issue assign 42 alice
grit issue assign 42 alice bob charlie
```

---

## `grit issue link`

Link two issues with a typed relationship.

```
grit issue link <number> <target-number> [flags]
```

Creates a reference comment on the source issue linking to the target.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--type` | | `related` | Relationship type (see below) |

**Link types:**

| Type | Meaning |
|------|---------|
| `related` | General relationship (default) |
| `blocks` | This issue blocks the target |
| `blocked-by` | This issue is blocked by the target |
| `duplicates` | This issue duplicates the target |
| `parent` | This issue is the parent of the target |
| `child` | This issue is a child of the target |

**Examples:**

```bash
grit issue link 42 43
grit issue link 42 43 --type blocks
grit issue link 50 42 --type duplicates
```

---

## `grit issue search`

Search issues using GitHub's search API.

```
grit issue search <query> [flags]
```

Returns a paginated list of matching issues.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--state` | `-s` | | Filter by state: `open` or `closed` |
| `--label` | `-l` | | Filter by label |
| `--limit` | `-n` | `30` | Results per page |
| `--page` | `-p` | `1` | Page number |

**Examples:**

```bash
grit issue search "login error"
grit issue search "timeout" -s open -l bug
```

---

## `grit issue sub`

Create a sub-issue linked as a child of an existing issue.

```
grit issue sub <parent-number> [prompt] [flags]
```

Works like `grit issue create` but automatically links the new issue as a child of the parent.

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | | Issue title |
| `--description` | `-d` | | Issue description |
| `--labels` | `-l` | | Comma-separated labels |
| `--assignees` | `-a` | | Comma-separated assignees |
| `--yes` | `-y` | `false` | Skip confirmation prompt |

**Examples:**

```bash
grit issue sub 42 "implement the auth middleware for this feature"
grit issue sub 42 -t "Add unit tests" -d "Cover edge cases" -y
```

---

## `grit update`

Update grit to the latest release.

```
grit update
```

Checks GitHub for a newer release, downloads the correct binary for your platform, and replaces the running executable. If you are running a dev build, grit will ask for confirmation first. See [Updating](updating.md) for more details.

---

## `grit version`

Print version information.

```
grit version
```

Displays the version number, commit SHA, and build date.
