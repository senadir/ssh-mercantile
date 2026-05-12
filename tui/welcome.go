package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

const welcomeLogo = `█▀▄▀█ █▀▀ █▀█ █▀▀ ▄▀█ █▄ █ ▀█▀ █ █   █▀▀
█ ▀ █ ██▄ █▀▄ █▄▄ █▀█ █ ▀█  █  █ █▄▄ ██▄`

// partyMinHeight is the minimum terminal height for the full Wapuu
// takeover (8 sparkle rows + 28-row Wapuu + spacing + bar + header +
// footer ≈ 44). Below this we render a compact kaomoji fallback.
const partyMinHeight = 40

// viewPartyCompact is the cramped-terminal fallback for party mode.
// Same color-cycling animation philosophy, fits in ~8 rows.
func (m *Model) viewPartyCompact() string {
	centered := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)

	// Cycle the face's color through the rainbow every 500ms — same
	// cadence as the full Wapuu's band cascade, so the two views feel
	// like the same effect at different scales.
	offset := int(time.Now().UnixMilli() / 500)
	color := lipgloss.Color(wapuuRainbow[offset%len(wapuuRainbow)])
	face := lipgloss.NewStyle().Foreground(color).Bold(true).Render("ʕ•ᴥ•ʔ")
	msg := styles.Subtle.Render("wapuu wants more room, resize taller for rainbows")

	return "\n" +
		centered.Render(face) + "\n\n" +
		centered.Render(msg) + "\n\n" +
		m.renderCountdownBar()
}

func (m *Model) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	keyStr := strings.ToLower(key.String())

	// Easter egg: track arrow-key sequences against the konami code. Arrows
	// have no other meaning on the welcome screen, so this is safe to do
	// before the regular switch.
	if keyStr == "up" || keyStr == "down" || keyStr == "left" || keyStr == "right" {
		m.konamiBuf = append(m.konamiBuf, keyStr)
		if len(m.konamiBuf) > len(konamiSeq) {
			m.konamiBuf = m.konamiBuf[len(m.konamiBuf)-len(konamiSeq):]
		}
		if konamiMatch(m.konamiBuf) {
			m.konamiBuf = nil
			m.partyUntil = time.Now().Add(partyDuration)
			return m, tea.Tick(partyDuration, func(time.Time) tea.Msg { return endPartyMsg{} })
		}
		return m, nil
	}

	switch keyStr {
	case "q":
		return m, tea.Quit
	case "a":
		m.browseSource = "apparel"
		m.push(viewBrowse)
		m.browseLoading = true
		return m, loadProductsCmd(m.client, api.ProductsParams{Category: "apparel"}, "Apparel")
	case "c":
		m.browseSource = "accessories"
		m.push(viewBrowse)
		m.browseLoading = true
		return m, loadProductsCmd(m.client, api.ProductsParams{Category: "accessories"}, "Accessories")
	case "b":
		m.push(viewAbout)
		return m, nil
	case "t":
		m.push(viewCart)
		return m, loadCartCmd(m.client)
	case "/":
		m.push(viewSearch)
		m.searchInput.Focus()
		m.searchInput.SetValue("")
		return m, nil
	case "l":
		// list all
		m.browseSource = "all"
		m.push(viewBrowse)
		m.browseLoading = true
		return m, loadProductsCmd(m.client, api.ProductsParams{}, "All Products")
	}
	return m, nil
}

func (m *Model) viewWelcome() string {
	if m.partyMode() {
		if m.height < partyMinHeight {
			// Terminal isn't tall enough to fit full Wapuu — show a
			// compact kaomoji fallback that still gestures at the
			// easter egg without clipping.
			return m.viewPartyCompact()
		}
		// Easter egg payload: a tall sparkle field at the top (where
		// they have room to drift upward), Wapuu centered below with
		// rainbow bands cascading down through him, and the depleting
		// countdown bar at the bottom just above the footer message.
		centered := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
		return m.renderSparkles() + "\n" + centered.Render(renderWapuu()) + "\n\n" + m.renderCountdownBar()
	}

	logo := styles.Brand.Render(welcomeLogo)

	intro := lipgloss.NewStyle().
		Foreground(styles.ColorFg).
		Render(wrap("The WordPress merch store — now browsable from your terminal. Built on WooCommerce Store API. Cart lives server-side; checkout hands off to the web.", m.contentWidth()))

	shortcuts := []struct{ key, label string }{
		{"a", "Apparel"},
		{"c", "Accessories"},
		{"b", "About"},
		{"/", "Search"},
		{"l", "List all products"},
		{"t", "View cart"},
		{"q", "Quit"},
	}

	var sb strings.Builder
	sb.WriteString(styles.Subtle.Render("Shortcuts") + "\n\n")
	for _, s := range shortcuts {
		sb.WriteString("  ")
		sb.WriteString(styles.Key.Render(padR(s.key, 3)))
		sb.WriteString("  ")
		sb.WriteString(s.label)
		sb.WriteString("\n")
	}

	box := styles.Box.Render(sb.String())

	blocks := lipgloss.JoinVertical(lipgloss.Left,
		"",
		logo,
		"",
		intro,
		"",
		box,
	)

	return lipgloss.NewStyle().PaddingLeft(2).Render(blocks)
}

func padR(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
