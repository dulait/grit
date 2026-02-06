# TUI Guide

grit includes an interactive terminal interface for browsing and managing issues without leaving the terminal.

## Launching the TUI

Run `grit` with no subcommand:

```bash
grit
```

This requires a configured project (`.grit/config.yaml`) and a stored GitHub token. See [Getting Started](getting-started.md) if you haven't set those up yet.

## Screens

The TUI has four main screens: **List**, **Detail**, **Create**, and **Edit**. You can also open action modals from the Detail screen for quick operations.

---

### List screen

The default screen. Shows a paginated list of issues for your repository.

**Layout:**

- **Header** — project name (`grit · owner/repo`)
- **Search bar** — appears when you press `/`
- **Issue rows** — number, title, state, labels, assignees
- **Status bar** — current filter, page info
- **Help hint** — press `?` for keybindings

**Keybindings:**

| Key | Action |
|-----|--------|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `Enter` / `l` | Open selected issue |
| `c` | Create a new issue |
| `n` | Next page |
| `p` | Previous page |
| `r` | Refresh the list |
| `1` | Filter: open issues |
| `2` | Filter: closed issues |
| `3` | Filter: all issues |
| `/` | Start a search |
| `Esc` | Clear search / exit search mode |
| `?` | Toggle help overlay |
| `q` | Quit |

**Searching:**

Press `/` to activate the search bar. Type your query — results update after a short debounce. Press `Enter` to finalize or `Esc` to cancel and clear the search.

---

### Detail screen

Shows full metadata and body for a single issue.

**Layout:**

- **Header** — `grit · Issue #NUMBER`
- **Title** — bold
- **Metadata** — state, labels, assignees, URL, created/updated timestamps
- **Body** — scrollable viewport

**Keybindings:**

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `Ctrl+u` | Half page up |
| `Ctrl+d` | Half page down |
| `e` | Edit this issue |
| `x` | Close this issue (opens modal) |
| `a` | Assign users (opens modal) |
| `m` | Add a comment (opens modal) |
| `o` | Open in browser |
| `Esc` / `h` / `Backspace` | Back to list |
| `?` | Toggle help overlay |
| `q` | Quit |

---

### Create screen

A multi-step form for creating a new issue. Reached by pressing `c` on the List screen.

**Form fields:**

| Field | Max length | Description |
|-------|-----------|-------------|
| Title | 120 chars | Issue title (required for direct submit) |
| Prompt | 256 chars | Description or natural-language prompt for LLM generation |
| Labels | — | Comma-separated labels (optional) |
| Assignees | — | Comma-separated GitHub usernames (optional) |

**Steps:**

1. **Input** — fill in the form fields
2. **Generating** — LLM processes the prompt (spinner)
3. **Review** — preview the generated issue
4. **Creating** — issue is being posted to GitHub (spinner)
5. **Done** — success message with issue URL

**Keybindings:**

| Key | Action |
|-----|--------|
| `Tab` / `↓` | Next field |
| `Shift+Tab` / `↑` | Previous field |
| `Ctrl+g` | Generate issue with LLM (requires title or prompt) |
| `Ctrl+s` | Submit directly without LLM (requires title) |
| `Esc` | Cancel and return to list |

During review:

| Key | Action |
|-----|--------|
| `Enter` / `Ctrl+s` | Create the issue from the generated content |
| `Esc` | Go back to the input form |

---

### Edit screen

Edit an existing issue's fields. Reached by pressing `e` on the Detail screen.

**Form fields:**

| Field | Max length | Description |
|-------|-----------|-------------|
| Title | 120 chars | Issue title |
| Body | 512 chars | Issue description |
| Labels | — | Comma-separated labels |
| Assignees | — | Comma-separated GitHub usernames |
| State | 10 chars | `open` or `closed` |

**Keybindings:**

| Key | Action |
|-----|--------|
| `Tab` / `↓` | Next field |
| `Shift+Tab` / `↑` | Previous field |
| `Ctrl+s` | Save changes |
| `Esc` | Cancel and return to detail |

---

### Action modals

Quick overlays that appear on top of the Detail screen. Each modal has a text input and submit/cancel controls.

**Close issue** — press `x` on the Detail screen. Optionally type a closing comment, then press `Enter` to close the issue.

**Assign users** — press `a` on the Detail screen. Type comma-separated GitHub usernames, then press `Enter`.

**Add comment** — press `m` on the Detail screen. Type a comment prompt, then press `Enter`. The LLM generates and posts the comment.

**Modal keybindings:**

| Key | Action |
|-----|--------|
| `Enter` | Submit |
| `Esc` | Cancel |

---

## Help overlay

Press `?` on the List or Detail screen to toggle a help overlay showing all available keybindings for the current screen. Press `?` again to dismiss it.

## Navigation summary

```
List ──Enter/l──> Detail ──e──> Edit
  │                 │
  c                 x ──> Close modal
  │                 a ──> Assign modal
  v                 m ──> Comment modal
Create              o ──> Browser
```
