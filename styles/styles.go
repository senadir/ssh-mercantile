package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Brand palette. WordPress blue + warm accents. Adaptive: each color has a
	// Light variant (shown on light-background terminals) and a Dark variant.
	ColorBg        = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d0f14"}
	ColorFg        = lipgloss.AdaptiveColor{Light: "#1f2328", Dark: "#e6e6e6"}
	ColorMuted     = lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}
	ColorSubtle    = lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#9ca3af"}
	ColorAccent     = lipgloss.AdaptiveColor{Light: "#2271b1", Dark: "#2271b1"} // WordPress blue
	// Border ring for filled primary buttons. Darker than ColorAccent on
	// light terminals, lighter on dark — so the outline reads as a defined
	// edge rather than vanishing into either the fill or the surrounding bg.
	ColorAccentRing = lipgloss.AdaptiveColor{Light: "#0a4165", Dark: "#a6d1f0"}
	ColorHighlight = lipgloss.AdaptiveColor{Light: "#1e40af", Dark: "#93c5fd"} // selection accent (darker than Accent on light, lighter on dark)
	ColorError     = lipgloss.AdaptiveColor{Light: "#b91c1c", Dark: "#ef4444"}
	ColorSuccess   = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10b981"}
	// Very faint blue wash used as a row highlight. Sits behind text that
	// already carries a fg color; pairs with SelectBar for full-row emphasis.
	ColorRowTint = lipgloss.AdaptiveColor{Light: "#dbeafe", Dark: "#1e3a5f"}

	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg)

	Brand = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorFg)

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
		Foreground(ColorAccent).
		Bold(true)

	Selected = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(ColorAccent).
		Bold(true).
		Padding(0, 1)

	// Left bar marker for the currently-selected row in a list. Paired with
	// SelectedItem for the row's text. Subtler than Selected (which is for
	// buttons/chips): one moving vertical accent instead of a full chip.
	SelectBar = lipgloss.NewStyle().
		Foreground(ColorHighlight).
		Bold(true)

	SelectedItem = lipgloss.NewStyle().
		Foreground(ColorFg).
		Bold(true)

	// Wraps a whole selected row to paint a faint amber wash across its width.
	SelectedRow = lipgloss.NewStyle().
		Background(ColorRowTint)

	Item = lipgloss.NewStyle().
		Foreground(ColorFg).
		Padding(0, 1)

	NavActive = lipgloss.NewStyle().
		Foreground(ColorAccent).
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
