package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"

	"github.com/Harrychangtw/cxport/internal/config"
	"github.com/Harrychangtw/cxport/internal/export"
	"github.com/Harrychangtw/cxport/internal/fs"
)

type view int

const (
	viewMain view = iota
	viewPicker
	viewPresets
	viewPresetSave
	viewExportDone
	viewHelp
)

type Model struct {
	// State
	currentView view
	root        string
	width       int
	height      int

	// File index
	allFiles []fs.FileEntry
	fileStrs []string // just the paths for fuzzy matching

	// Selection
	selected     []string // relative paths
	selectedSet  map[string]bool
	selCursor    int

	// Picker
	pickerInput   textinput.Model
	pickerMatches []fuzzy.Match
	pickerCursor  int

	// Presets
	cfg          config.Config
	presetNames  []string
	presetCursor int

	// Preset save
	presetNameInput textinput.Model

	// Export
	format     export.Format
	lastResult export.Result

	// Help
	showHelp bool

	// Messages
	statusMsg string
}

func NewModel(root string, format string) Model {
	cfg, _ := config.Load()

	ti := textinput.New()
	ti.Placeholder = "Type to search files..."
	ti.CharLimit = 256
	ti.Width = 60

	pi := textinput.New()
	pi.Placeholder = "Preset name..."
	pi.CharLimit = 64
	pi.Width = 40

	f := export.FormatXML
	if format == "md" {
		f = export.FormatMarkdown
	}

	m := Model{
		currentView:     viewMain,
		root:            root,
		selected:        []string{},
		selectedSet:     make(map[string]bool),
		pickerInput:     ti,
		presetNameInput: pi,
		cfg:             cfg,
		presetNames:     cfg.GetPresetNames(),
		format:          f,
	}

	return m
}

type fileIndexMsg struct {
	files []fs.FileEntry
}

func indexFiles(root string) tea.Cmd {
	return func() tea.Msg {
		files, _ := fs.Walk(root)
		return fileIndexMsg{files: files}
	}
}

func (m Model) Init() tea.Cmd {
	return indexFiles(m.root)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case fileIndexMsg:
		m.allFiles = msg.files
		m.fileStrs = make([]string, len(msg.files))
		for i, f := range msg.files {
			m.fileStrs[i] = f.Path
		}
		return m, nil
	}

	switch m.currentView {
	case viewMain:
		return m.updateMain(msg)
	case viewPicker:
		return m.updatePicker(msg)
	case viewPresets:
		return m.updatePresets(msg)
	case viewPresetSave:
		return m.updatePresetSave(msg)
	case viewExportDone:
		return m.updateExportDone(msg)
	case viewHelp:
		return m.updateHelp(msg)
	}

	return m, nil
}

// --- Main view ---

func (m Model) updateMain(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Add):
			m.currentView = viewPicker
			m.pickerInput.SetValue("")
			m.pickerInput.Focus()
			m.pickerCursor = 0
			m.updatePickerMatches()
			return m, textinput.Blink
		case key.Matches(msg, keys.Presets):
			m.currentView = viewPresets
			m.presetNames = m.cfg.GetPresetNames()
			sort.Strings(m.presetNames)
			m.presetCursor = 0
			return m, nil
		case key.Matches(msg, keys.Export):
			if len(m.selected) == 0 {
				m.statusMsg = "Nothing selected to export"
				return m, nil
			}
			return m.doExport()
		case key.Matches(msg, keys.Format):
			if m.format == export.FormatXML {
				m.format = export.FormatMarkdown
				m.statusMsg = "Format: Markdown"
			} else {
				m.format = export.FormatXML
				m.statusMsg = "Format: XML"
			}
			return m, nil
		case key.Matches(msg, keys.Help):
			m.currentView = viewHelp
			return m, nil
		case key.Matches(msg, keys.Remove):
			if len(m.selected) > 0 && m.selCursor < len(m.selected) {
				path := m.selected[m.selCursor]
				delete(m.selectedSet, path)
				m.selected = append(m.selected[:m.selCursor], m.selected[m.selCursor+1:]...)
				if m.selCursor >= len(m.selected) && m.selCursor > 0 {
					m.selCursor--
				}
			}
			return m, nil
		case key.Matches(msg, keys.ClearAll):
			m.selected = []string{}
			m.selectedSet = make(map[string]bool)
			m.selCursor = 0
			m.statusMsg = "Cleared all selections"
			return m, nil
		case key.Matches(msg, keys.Up):
			if m.selCursor > 0 {
				m.selCursor--
			}
			return m, nil
		case key.Matches(msg, keys.Down):
			if m.selCursor < len(m.selected)-1 {
				m.selCursor++
			}
			return m, nil
		default:
			// Auto-open picker when user types a printable character
			if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
				typed := string(msg.Runes)
				m.currentView = viewPicker
				m.pickerInput.SetValue(typed)
				m.pickerInput.Focus()
				// Move cursor to end of input
				m.pickerInput, _ = m.pickerInput.Update(msg)
				m.pickerCursor = 0
				m.updatePickerMatches()
				return m, textinput.Blink
			}
		}
	}
	return m, nil
}

// --- Picker view ---

func (m *Model) updatePickerMatches() {
	query := m.pickerInput.Value()
	if query == "" {
		// Show all files (limited)
		limit := 20
		if len(m.fileStrs) < limit {
			limit = len(m.fileStrs)
		}
		m.pickerMatches = make([]fuzzy.Match, limit)
		for i := 0; i < limit; i++ {
			m.pickerMatches[i] = fuzzy.Match{
				Str:   m.fileStrs[i],
				Index: i,
			}
		}
	} else {
		m.pickerMatches = fuzzy.Find(query, m.fileStrs)
		if len(m.pickerMatches) > 30 {
			m.pickerMatches = m.pickerMatches[:30]
		}
	}
	if m.pickerCursor >= len(m.pickerMatches) {
		m.pickerCursor = 0
	}
}

func (m Model) updatePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.currentView = viewMain
			m.pickerInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			if len(m.pickerMatches) > 0 && m.pickerCursor < len(m.pickerMatches) {
				match := m.pickerMatches[m.pickerCursor]
				entry := m.allFiles[match.Index]
				path := entry.Path
				if m.selectedSet[path] {
					// Deselect
					delete(m.selectedSet, path)
					for i, p := range m.selected {
						if p == path {
							m.selected = append(m.selected[:i], m.selected[i+1:]...)
							break
						}
					}
					m.statusMsg = fmt.Sprintf("Removed: %s", path)
				} else {
					// Select
					m.selected = append(m.selected, path)
					m.selectedSet[path] = true
					m.statusMsg = fmt.Sprintf("Added: %s", path)
				}
			}
			return m, nil
		case msg.Type == tea.KeyUp || (msg.Type == tea.KeyCtrlP):
			if m.pickerCursor > 0 {
				m.pickerCursor--
			}
			return m, nil
		case msg.Type == tea.KeyDown || (msg.Type == tea.KeyCtrlN):
			if m.pickerCursor < len(m.pickerMatches)-1 {
				m.pickerCursor++
			}
			return m, nil
		case key.Matches(msg, keys.Tab):
			// Add entire directory of current match
			if len(m.pickerMatches) > 0 && m.pickerCursor < len(m.pickerMatches) {
				match := m.pickerMatches[m.pickerCursor]
				entry := m.allFiles[match.Index]
				dir := filepath.Dir(entry.Path)
				if !m.selectedSet[dir] {
					m.selected = append(m.selected, dir)
					m.selectedSet[dir] = true
					m.statusMsg = fmt.Sprintf("Added directory: %s", dir)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	prevValue := m.pickerInput.Value()
	m.pickerInput, cmd = m.pickerInput.Update(msg)
	if m.pickerInput.Value() != prevValue {
		m.pickerCursor = 0
	}
	m.updatePickerMatches()
	return m, cmd
}

// --- Presets view ---

func (m Model) updatePresets(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.currentView = viewMain
			return m, nil
		case key.Matches(msg, keys.Save):
			if len(m.selected) == 0 {
				m.statusMsg = "Nothing selected to save"
				return m, nil
			}
			m.currentView = viewPresetSave
			m.presetNameInput.SetValue("")
			m.presetNameInput.Focus()
			return m, textinput.Blink
		case key.Matches(msg, keys.Enter):
			if len(m.presetNames) > 0 && m.presetCursor < len(m.presetNames) {
				name := m.presetNames[m.presetCursor]
				preset := m.cfg.Presets[name]
				m.selected = make([]string, len(preset.Paths))
				copy(m.selected, preset.Paths)
				m.selectedSet = make(map[string]bool)
				for _, p := range m.selected {
					m.selectedSet[p] = true
				}
				m.selCursor = 0
				m.statusMsg = fmt.Sprintf("Loaded preset: %s", name)
				m.currentView = viewMain
			}
			return m, nil
		case key.Matches(msg, keys.Delete):
			if len(m.presetNames) > 0 && m.presetCursor < len(m.presetNames) {
				name := m.presetNames[m.presetCursor]
				m.cfg.DeletePreset(name)
				_ = config.Save(m.cfg)
				m.presetNames = m.cfg.GetPresetNames()
				sort.Strings(m.presetNames)
				if m.presetCursor >= len(m.presetNames) && m.presetCursor > 0 {
					m.presetCursor--
				}
				m.statusMsg = fmt.Sprintf("Deleted preset: %s", name)
			}
			return m, nil
		case key.Matches(msg, keys.Up):
			if m.presetCursor > 0 {
				m.presetCursor--
			}
			return m, nil
		case key.Matches(msg, keys.Down):
			if m.presetCursor < len(m.presetNames)-1 {
				m.presetCursor++
			}
			return m, nil
		}
	}
	return m, nil
}

// --- Preset save ---

func (m Model) updatePresetSave(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.currentView = viewPresets
			m.presetNameInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			name := strings.TrimSpace(m.presetNameInput.Value())
			if name != "" {
				m.cfg.SavePreset(name, m.selected)
				_ = config.Save(m.cfg)
				m.presetNames = m.cfg.GetPresetNames()
				sort.Strings(m.presetNames)
				m.statusMsg = fmt.Sprintf("Saved preset: %s", name)
				m.currentView = viewMain
				m.presetNameInput.Blur()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.presetNameInput, cmd = m.presetNameInput.Update(msg)
	return m, cmd
}

// --- Export ---

func (m Model) doExport() (Model, tea.Cmd) {
	cwd, _ := os.Getwd()
	outputBase := filepath.Join(cwd, m.cfg.OutputFile)
	result, err := export.Export(m.root, m.selected, outputBase, m.format)
	if err != nil {
		m.statusMsg = fmt.Sprintf("Export error: %v", err)
		return m, nil
	}
	m.lastResult = result
	m.currentView = viewExportDone
	return m, nil
}

func (m Model) updateExportDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.currentView = viewMain
		return m, nil
	}
	return m, nil
}

// --- Help ---

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		m.currentView = viewMain
		return m, nil
	}
	return m, nil
}

// ==================== VIEW ====================

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	switch m.currentView {
	case viewPicker:
		return m.viewPicker()
	case viewPresets:
		return m.viewPresets()
	case viewPresetSave:
		return m.viewPresetSave()
	case viewExportDone:
		return m.viewExportDone()
	case viewHelp:
		return m.viewHelp()
	default:
		return m.viewMain()
	}
}

func (m Model) viewMain() string {
	var b strings.Builder

	// Title
	title := titleStyle.Render(" cxport ")
	formatBadge := dimStyle.Render(fmt.Sprintf(" [%s]", m.format))
	b.WriteString(title + formatBadge + "\n\n")

	// Selected files
	if len(m.selected) == 0 {
		b.WriteString(dimStyle.Render("  No files selected. Press ") +
			helpKeyStyle.Render("a") +
			dimStyle.Render(" to add files.\n\n"))
	} else {
		count := accentStyle.Render(fmt.Sprintf("%d", len(m.selected)))
		b.WriteString(headerStyle.Render(fmt.Sprintf("  Selected (%s):", count)))
		b.WriteString("\n")

		// Show files with cursor
		maxShow := m.height - 12
		if maxShow < 5 {
			maxShow = 5
		}
		start := 0
		if m.selCursor >= maxShow {
			start = m.selCursor - maxShow + 1
		}
		end := start + maxShow
		if end > len(m.selected) {
			end = len(m.selected)
		}

		for i := start; i < end; i++ {
			path := m.selected[i]
			cursor := "  "
			style := filePathStyle
			if i == m.selCursor {
				cursor = cursorStyle.Render("▸ ")
				style = selectedStyle
			}

			// Show dir indicator
			absPath := filepath.Join(m.root, path)
			info, err := os.Stat(absPath)
			icon := "  "
			if err == nil && info.IsDir() {
				icon = dirStyle.Render("■ ")
			} else {
				icon = dimStyle.Render("· ")
			}

			b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, icon, style.Render(path)))
		}

		if len(m.selected) > maxShow {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  ... and %d more\n", len(m.selected)-maxShow)))
		}
		b.WriteString("\n")
	}

	// Actions
	b.WriteString(dimStyle.Render("  ─────────────────────────────\n"))
	actions := []struct{ key, desc string }{
		{"a", "add files"},
		{"p", "presets"},
		{"e", "export"},
		{"f", "format (" + string(m.format) + ")"},
		{"d", "remove"},
		{"C", "clear all"},
		{"?", "help"},
		{"q", "quit"},
	}
	var parts []string
	for _, a := range actions {
		parts = append(parts, helpKeyStyle.Render(a.key)+helpDescStyle.Render(" "+a.desc))
	}
	b.WriteString("  " + strings.Join(parts, dimStyle.Render("  ·  ")) + "\n")

	// Status message
	if m.statusMsg != "" {
		b.WriteString("\n  " + subtleStyle.Render(m.statusMsg) + "\n")
	}

	return b.String()
}

func (m Model) viewPicker() string {
	var b strings.Builder

	title := titleStyle.Render(" Add Files ")
	b.WriteString(title + "\n\n")

	// Search input
	b.WriteString("  " + m.pickerInput.View() + "\n\n")

	// Results
	if len(m.pickerMatches) == 0 {
		b.WriteString(dimStyle.Render("  No matches found.\n"))
	} else {
		maxShow := m.height - 8
		if maxShow < 5 {
			maxShow = 5
		}
		start := 0
		if m.pickerCursor >= maxShow {
			start = m.pickerCursor - maxShow + 1
		}
		end := start + maxShow
		if end > len(m.pickerMatches) {
			end = len(m.pickerMatches)
		}

		for i := start; i < end; i++ {
			match := m.pickerMatches[i]
			path := match.Str
			cursor := "  "
			if i == m.pickerCursor {
				cursor = cursorStyle.Render("▸ ")
			}

			// Highlight matched characters
			rendered := highlightMatch(path, match.MatchedIndexes)

			// Show if already selected
			marker := "  "
			if m.selectedSet[path] {
				marker = selectedStyle.Render("✓ ")
			}

			// Show dir/file icon
			icon := dimStyle.Render("· ")
			if m.allFiles[match.Index].IsDir {
				icon = dirStyle.Render("■ ")
			}

			b.WriteString(fmt.Sprintf("  %s%s%s%s\n", cursor, marker, icon, rendered))
		}

		if len(m.pickerMatches) > maxShow {
			b.WriteString(dimStyle.Render(fmt.Sprintf("\n  %d/%d results", maxShow, len(m.pickerMatches))))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  ─────────────────────────────\n"))
	hints := helpKeyStyle.Render("enter") + helpDescStyle.Render(" toggle file") +
		dimStyle.Render("  ·  ") +
		helpKeyStyle.Render("tab") + helpDescStyle.Render(" add dir") +
		dimStyle.Render("  ·  ") +
		helpKeyStyle.Render("esc") + helpDescStyle.Render(" back")
	b.WriteString("  " + hints + "\n")

	return b.String()
}

func highlightMatch(s string, matchedIndexes []int) string {
	if len(matchedIndexes) == 0 {
		return filePathStyle.Render(s)
	}

	matchSet := make(map[int]bool)
	for _, idx := range matchedIndexes {
		matchSet[idx] = true
	}

	var result strings.Builder
	for i, ch := range s {
		if matchSet[i] {
			result.WriteString(accentStyle.Render(string(ch)))
		} else {
			result.WriteString(filePathStyle.Render(string(ch)))
		}
	}
	return result.String()
}

func (m Model) viewPresets() string {
	var b strings.Builder

	title := titleStyle.Render(" Presets ")
	b.WriteString(title + "\n\n")

	if len(m.presetNames) == 0 {
		b.WriteString(dimStyle.Render("  No presets saved yet.\n"))
		b.WriteString(dimStyle.Render("  Select files first, then press ") +
			helpKeyStyle.Render("s") +
			dimStyle.Render(" to save a preset.\n"))
	} else {
		for i, name := range m.presetNames {
			cursor := "  "
			style := filePathStyle
			if i == m.presetCursor {
				cursor = cursorStyle.Render("▸ ")
				style = selectedStyle
			}

			preset := m.cfg.Presets[name]
			count := dimStyle.Render(fmt.Sprintf(" (%d paths)", len(preset.Paths)))
			b.WriteString(fmt.Sprintf("  %s%s%s\n", cursor, style.Render(name), count))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  ─────────────────────────────\n"))
	hints := helpKeyStyle.Render("enter") + helpDescStyle.Render(" load") +
		dimStyle.Render("  ·  ") +
		helpKeyStyle.Render("s") + helpDescStyle.Render(" save current") +
		dimStyle.Render("  ·  ") +
		helpKeyStyle.Render("d") + helpDescStyle.Render(" delete") +
		dimStyle.Render("  ·  ") +
		helpKeyStyle.Render("esc") + helpDescStyle.Render(" back")
	b.WriteString("  " + hints + "\n")

	if m.statusMsg != "" {
		b.WriteString("\n  " + subtleStyle.Render(m.statusMsg) + "\n")
	}

	return b.String()
}

func (m Model) viewPresetSave() string {
	var b strings.Builder

	title := titleStyle.Render(" Save Preset ")
	b.WriteString(title + "\n\n")

	b.WriteString(dimStyle.Render(fmt.Sprintf("  Saving %d selected paths as a preset.\n\n", len(m.selected))))
	b.WriteString("  Name: " + m.presetNameInput.View() + "\n")

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  ─────────────────────────────\n"))
	hints := helpKeyStyle.Render("enter") + helpDescStyle.Render(" save") +
		dimStyle.Render("  ·  ") +
		helpKeyStyle.Render("esc") + helpDescStyle.Render(" cancel")
	b.WriteString("  " + hints + "\n")

	return b.String()
}

func (m Model) viewExportDone() string {
	var b strings.Builder

	title := titleStyle.Render(" Export Complete ")
	b.WriteString(title + "\n\n")

	b.WriteString(successStyle.Render("  ✓ Export successful!") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("Output:"), filePathStyle.Render(m.lastResult.OutputPath)))
	b.WriteString(fmt.Sprintf("  %s %d\n", dimStyle.Render("Files: "), m.lastResult.FileCount))
	b.WriteString(fmt.Sprintf("  %s %s\n", dimStyle.Render("Size:  "), formatBytes(m.lastResult.TotalBytes)))

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Press any key to continue.\n"))

	return b.String()
}

func (m Model) viewHelp() string {
	var b strings.Builder

	title := titleStyle.Render(" Help ")
	b.WriteString(title + "\n\n")

	sections := []struct {
		name  string
		binds []struct{ key, desc string }
	}{
		{
			"Main",
			[]struct{ key, desc string }{
				{"a", "Open file picker to add files"},
				{"p", "Open presets manager"},
				{"e", "Export selected files"},
				{"f", "Toggle output format (XML/Markdown)"},
				{"d/x", "Remove selected file from list"},
				{"C", "Clear all selections"},
				{"↑/k ↓/j", "Navigate selection list"},
				{"?", "Show this help"},
				{"q", "Quit"},
			},
		},
		{
			"File Picker",
			[]struct{ key, desc string }{
				{"type", "Fuzzy search files"},
				{"↑/↓", "Navigate results"},
				{"enter", "Toggle file selection (add/remove)"},
				{"tab", "Add entire parent directory"},
				{"esc", "Return to main view"},
			},
		},
		{
			"Presets",
			[]struct{ key, desc string }{
				{"enter", "Load preset (replaces selection)"},
				{"s", "Save current selection as preset"},
				{"d", "Delete preset"},
				{"esc", "Return to main view"},
			},
		},
	}

	for _, sec := range sections {
		b.WriteString(headerStyle.Render("  "+sec.name) + "\n")
		for _, bind := range sec.binds {
			k := helpKeyStyle.Render(fmt.Sprintf("  %-12s", bind.key))
			d := helpDescStyle.Render(bind.desc)
			b.WriteString(fmt.Sprintf("  %s %s\n", k, d))
		}
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render("  Press any key to close.\n"))

	return b.String()
}

func formatBytes(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func Run(format string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	w, h := lipgloss.Size("")
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	m := NewModel(cwd, format)
	p := tea.NewProgram(m, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
