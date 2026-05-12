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
		// Easter egg payload: rainbow half-block Wapuu takes over the
		// welcome body for the duration of party mode, with a depleting
		// countdown bar above so the dismiss isn't a surprise.
		centered := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center)
		return m.renderCountdownBar() + "\n\n" + centered.Render(renderWapuu())
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
