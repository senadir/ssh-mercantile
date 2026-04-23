package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

func (m *Model) setupSearchInput() {
	ti := textinput.New()
	ti.Placeholder = "sweater, mug, sticker…"
	ti.Prompt = "› "
	ti.CharLimit = 80
	ti.Width = 40
	m.searchInput = ti
}

func (m *Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.pop()
			return m, nil
		case "enter":
			q := strings.TrimSpace(m.searchInput.Value())
			if q == "" {
				return m, m.setStatus("Type a query first", true)
			}
			m.browseSource = "search:" + q
			// Replace the search view with the browse view in the stack.
			m.view = viewBrowse
			m.browseLoading = true
			return m, loadProductsCmd(m.client, api.ProductsParams{Search: q}, "Search: "+q)
		}
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m *Model) viewSearch() string {
	title := styles.Title.Render("Search products")
	hint := styles.Muted.Render("Type your query and press enter.")
	return lipgloss.JoinVertical(lipgloss.Left,
		"", title, hint, "", "  "+m.searchInput.View(),
	)
}
