package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"mercantile/styles"
)

// buttonVariant chooses the visual weight.
//
//	Primary   — the "continue" action on a screen (Add to cart,
//	            Proceed to checkout, Confirm). Filled when focused.
//	Secondary — alternates (Back, Cancel). Quieter; ghost when focused.
type buttonVariant int

const (
	btnPrimary buttonVariant = iota
	btnSecondary
)

// buttonOpts is the input for renderButton. A single button can be in one
// of several states at once — e.g. focused + loading, or idle + disabled —
// so each state is its own bool rather than an enum.
type buttonOpts struct {
	Label    string
	Hotkey   string // parenthetical hint; "" to skip
	Variant  buttonVariant
	Focused  bool
	Disabled bool   // greyed out, ignores keypresses
	Loading  bool   // shows a spinner; implies focused
	Spinner  string // current spinner frame (e.g. "⣾"); only used when Loading
}

// Shared layout: 3-line rounded box with horizontal padding.
var buttonBase = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(0, 2).
	Bold(true)

// renderButton returns a multi-line string representing a button. The output
// is 3 terminal rows tall regardless of state, so layouts don't jump as
// focus or loading changes.
func renderButton(opts buttonOpts) string {
	inner := opts.Label
	if opts.Hotkey != "" {
		inner = fmt.Sprintf("%s (%s)", opts.Label, opts.Hotkey)
	}
	if opts.Loading {
		// Replace the label with "spinner  Verb…". Feels alive; the user
		// can see the button has received their press and is working.
		inner = fmt.Sprintf("%s  %s…", opts.Spinner, opts.Label)
	}

	// Disabled wins over everything else — no matter the variant or focus,
	// a disabled button reads as "not available right now."
	if opts.Disabled {
		return buttonBase.
			BorderForeground(styles.ColorMuted).
			Foreground(styles.ColorMuted).
			Bold(false).
			Render(inner)
	}

	switch opts.Variant {
	case btnPrimary:
		if opts.Focused || opts.Loading {
			// Filled chip with a traced outline. The trick: each border
			// cell's background matches the fill, so the button reads as
			// one continuous blue shape — but the border glyph itself is
			// drawn in a contrasting ring color. The rounded corner arcs
			// show as a colored stroke over the fill, no corner gap.
			return buttonBase.
				BorderForeground(styles.ColorAccentRing).
				BorderBackground(styles.ColorAccent).
				Foreground(lipgloss.Color("#ffffff")).
				Background(styles.ColorAccent).
				Render(inner)
		}
		// Ghost — blue border and text, transparent inside.
		return buttonBase.
			BorderForeground(styles.ColorAccent).
			Foreground(styles.ColorAccent).
			Render(inner)

	case btnSecondary:
		if opts.Focused {
			// Amber border + tint bg — matches list-row focus language.
			return buttonBase.
				BorderForeground(styles.ColorHighlight).
				Foreground(styles.ColorFg).
				Background(styles.ColorRowTint).
				Render(inner)
		}
		// Quiet: muted border and text.
		return buttonBase.
			BorderForeground(styles.ColorMuted).
			Foreground(styles.ColorSubtle).
			Bold(false).
			Render(inner)
	}
	return inner
}
