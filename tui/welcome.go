package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

const welcomeLogo = `
  █▀▄▀█ █▀▀ █▀█ █▀▀ ▄▀█ █▄ █ ▀█▀ █ █   █▀▀
  █ ▀ █ ██▄ █▀▄ █▄▄ █▀█ █ ▀█  █  █ █▄▄ ██▄
`

func (m *Model) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch strings.ToLower(key.String()) {
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
	logo := styles.Brand.Render(welcomeLogo)

	intro := lipgloss.NewStyle().
		Foreground(styles.ColorFg).
		Render("The WordPress merch store — now browsable from your terminal.\nBuilt on WooCommerce Store API. Cart lives server-side; checkout hands off to the web.")

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

	// center horizontally on wide terminals
	if m.width > 0 {
		return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(blocks)
	}
	return blocks
}

func padR(s string, n int) string {
	for len(s) < n {
		s += " "
	}
	return s
}
