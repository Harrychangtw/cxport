package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Brand colors
	colorPrimary   = lipgloss.Color("#B07CFF") // purple, like claude
	colorSecondary = lipgloss.Color("#FF7CBB") // pink accent
	colorSuccess   = lipgloss.Color("#7CFFB0") // green
	colorWarning   = lipgloss.Color("#FFD07C") // amber
	colorDim       = lipgloss.Color("#666666")
	colorBright    = lipgloss.Color("#FFFFFF")
	colorSubtle    = lipgloss.Color("#888888")
	colorBg        = lipgloss.Color("#1A1A2E")
	colorBgLight   = lipgloss.Color("#2A2A3E")

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorBright).
			Background(colorPrimary).
			Padding(0, 1)

	// Section headers
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Selected item
	selectedStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// Dim / secondary text
	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	subtleStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Highlighted text
	accentStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	// Input field
	inputStyle = lipgloss.NewStyle().
			Foreground(colorBright).
			Background(colorBgLight).
			Padding(0, 1)

	// Status bar at bottom
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)

	// File path in list
	filePathStyle = lipgloss.NewStyle().
			Foreground(colorBright)

	// Directory path style
	dirStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Box / panel border
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// Success message
	successStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	// Warning
	warningStyle = lipgloss.NewStyle().
			Foreground(colorWarning)

	// Help key
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Help description
	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorSubtle)

	// Logo / brand
	logoStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	// Cursor for the list
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)
)
