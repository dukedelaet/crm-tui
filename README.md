# sophie

A full-screen TUI for [crm-cli](https://github.com/jdanielnd/crm-cli) built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

![sophie screenshot](https://i.imgur.com/placeholder.png)

## Features

- **Full-screen TUI** with pink & lavender color scheme
- **Vim-style navigation** — `j`/`k` to move, `h`/`l` to switch tabs
- **Command palette** — press `Ctrl+P` to see all available commands and shortcuts
- **Modal forms** for adding and editing people, organizations, deals, tasks, and interactions
- **Keyboard shortcuts** for common actions:
  - `n` — new (opens add modal for current tab)
  - `e` — edit selected item
  - `d` — delete selected item
  - `v` — view details of selected item
  - `/` — search
  - `r` — refresh data
  - `q` — quit

## Tabs

| Key | Tab | Description |
|-----|-----|-------------|
| `1` / `h` | Status | Dashboard overview |
| `2` / `l` | People | Contact management |
| `3` | Orgs | Organizations |
| `4` | Deals | Sales pipeline |
| `5` | Tasks | Follow-up tracking |
| `6` | Logs | Interaction history |

## Install

```bash
# From source
go install github.com/dukedelaet/sophie@latest

# Or build locally
go build -o sophie ./cmd/
```

Requires [crm-cli](https://github.com/jdanielnd/crm-cli) to be installed and on your PATH.

## Usage

```bash
sophie
```

Run from a terminal with crm-cli configured. Your CRM data lives in `~/.crm/crm.db`.

## Keybindings Reference

### Global
- `Q` / `Ctrl+C` — Quit
- `R` — Refresh all data
- `Ctrl+P` — Open command palette
- `/` — Focus search bar
- `H` / `L` or `Tab` — Navigate tabs

### Within a tab
- `J` / `K` or `↑` / `↓` — Move selection
- `N` — New item (opens modal)
- `E` — Edit selected item
- `D` — Delete selected item
- `V` — View details
- `Enter` — Select / Confirm in modal
- `Esc` — Cancel / Close modal

### In modals
- `Tab` / `Shift+Tab` — Switch fields
- `Ctrl+S` — Save
- `Esc` — Cancel

### Command Palette (`Ctrl+P`)
- `J` / `K` — Navigate commands
- `Enter` — Open matching command form
- `/` — Filter commands
- `Esc` — Close

## Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — Reusable TUI components (list, table, text input, help)
- [crm-cli](https://github.com/jdanielnd/crm-cli) — Backend CRM CLI

## License

MIT
