package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Brand palette. WordPress blue + warm accents.
	ColorBg       = lipgloss.Color("#0d0f14")
	ColorFg       = lipgloss.Color("#e6e6e6")
	ColorMuted    = lipgloss.Color("#6b7280")
	ColorSubtle   = lipgloss.Color("#9ca3af")
	ColorAccent   = lipgloss.Color("#2271b1") // WordPress blue
	ColorHighlight = lipgloss.Color("#ffb800")
	ColorError    = lipgloss.Color("#ef4444")
	ColorSuccess  = lipgloss.Color("#10b981")

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg)

	Brand = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight)

	Muted = lipgloss.NewStyle().
		Foreground(ColorMuted)

	Subtle = lipgloss.NewStyle().
		Foreground(ColorSubtle)

	Accent = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true)

	Error = lipgloss.NewStyle().
		Foreground(ColorError).
		Bold(true)

	Success = lipgloss.NewStyle().
		Foreground(ColorSuccess)

	Key = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	Selected = lipgloss.NewStyle().
		Foreground(ColorBg).
		Background(ColorHighlight).
		Bold(true).
		Padding(0, 1)

	Item = lipgloss.NewStyle().
		Foreground(ColorFg).
		Padding(0, 1)

	NavActive = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true).
		Underline(true)

	NavInactive = lipgloss.NewStyle().
		Foreground(ColorSubtle)

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted).
		Padding(1, 2)

	HelpBar = lipgloss.NewStyle().
		Foreground(ColorMuted).
		MarginTop(1)
)
