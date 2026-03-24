# CLAUDE.md

## Project Overview

**cxport** is an interactive TUI tool for selecting and exporting file content into a single context file for LLM sessions. Built with Go and the Charmbracelet Bubble Tea ecosystem.

## Tech Stack

- **Language:** Go 1.22+
- **TUI Framework:** Bubble Tea (`charmbracelet/bubbletea`)
- **Components:** Bubbles (`charmbracelet/bubbles`) — textinput, key bindings
- **Styling:** Lip Gloss (`charmbracelet/lipgloss`)
- **Fuzzy Search:** `sahilm/fuzzy`

## Project Structure

```
main.go                         # Entry point, flag parsing
internal/
  tui/
    app.go                      # Main Bubble Tea model, state machine, all views
    styles.go                   # Lip Gloss color palette and style definitions
    keys.go                     # Key binding definitions
  config/
    config.go                   # Config + preset persistence (~/.config/cxport/)
  export/
    export.go                   # File concatenation (XML + Markdown formatters)
  fs/
    walker.go                   # File system walking, .gitignore, language detection
```

## Build & Run

```bash
go build -o cxport .    # Build
./cxport                # Run in current directory
./cxport --format md    # Force markdown output
./cxport --version      # Print version
```

## Architecture

### State Machine

The TUI uses a single `Model` struct with a `view` enum:
- `viewMain` — Selected files list + action bar
- `viewPicker` — Fuzzy file search (textinput + filtered list)
- `viewPresets` — Load/save/delete preset groups
- `viewPresetSave` — Name input for saving a preset
- `viewExportDone` — Export result summary
- `viewHelp` — Key binding reference

### File Indexing

On launch, `fs.Walk()` indexes all files under cwd, skipping:
- Hidden files/dirs (`.` prefix)
- Common noise dirs: `node_modules`, `vendor`, `dist`, `build`, `.git`, `__pycache__`, `.next`, `.turbo`, `.cache`
- `.gitignore` patterns (simple glob matching)

### Export Formats

- **XML (default):** `<context><file path="..." language="...">content</file></context>`
- **Markdown:** `## File: path` with fenced code blocks

### Config

Stored at `~/.config/cxport/config.json`. Contains:
- `default_format`: "xml" or "md"
- `output_file`: base name for output (extension added by format)
- `presets`: map of named path groups

## Conventions

- All TUI logic lives in `internal/tui/app.go` — single file for the Elm Architecture model
- Styles are centralized in `styles.go`, never inline
- Key bindings defined in `keys.go` using `bubbles/key`
- File paths in selections/presets are relative to the cwd where cxport was launched
