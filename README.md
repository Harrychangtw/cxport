# cxport

A fast, interactive TUI for assembling file context for LLM sessions.

Instead of manually editing scripts or running agents that burn tokens reading files, `cxport` gives you a visual, keyboard-driven interface to select exactly the files you need and export them as a single context file.

## Features

- **Fuzzy file search** — VS Code-style fuzzy finder, type to narrow down results instantly
- **Directory selection** — Add entire directories with one keypress
- **Preset system** — Save and load named groups of paths for repeated use
- **Dual format** — Output as XML (Claude-optimized) or Markdown, switchable with `f`
- **Fast** — Single Go binary, indexes your project in milliseconds
- **Respects .gitignore** — Automatically skips noise (node_modules, .git, dist, etc.)

## Install

```bash
go install github.com/Harrychangtw/cxport@latest
```

Or build from source:

```bash
git clone https://github.com/Harrychangtw/cxport.git
cd cxport
go build -o cxport .
```

## Usage

```bash
# Launch in current directory
cxport

# Force markdown output
cxport --format md
```

### Key Bindings

#### Main View
| Key | Action |
|-----|--------|
| `a` | Open fuzzy file picker |
| `p` | Open presets manager |
| `e` | Export selected files |
| `f` | Toggle format (XML/Markdown) |
| `d`/`x` | Remove highlighted file |
| `C` | Clear all selections |
| `↑`/`k` `↓`/`j` | Navigate |
| `?` | Help |
| `q` | Quit |

#### File Picker
| Key | Action |
|-----|--------|
| *type* | Fuzzy search |
| `↑`/`↓` | Navigate results |
| `enter` | Add file |
| `tab` | Add parent directory |
| `esc` | Back |

#### Presets
| Key | Action |
|-----|--------|
| `enter` | Load preset |
| `s` | Save current selection |
| `d` | Delete preset |
| `esc` | Back |

## Output Formats

### XML (default, Claude-optimized)

```xml
<context>
<file path="src/main.go" language="go">
package main
...
</file>
</context>
```

### Markdown

~~~markdown
## File: src/main.go

```go
package main
...
```
~~~

## Config

Presets and settings are stored at `~/.config/cxport/config.json`.

## License

MIT
