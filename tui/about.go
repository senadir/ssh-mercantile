package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/styles"
)

func (m *Model) updateAbout(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch strings.ToLower(key.String()) {
	case "esc":
		m.pop()
		return m, nil
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) viewAbout() string {
	title := styles.Title.Render("About")
	body := wrap(
		"Mercantile is the official WordPress merch store — apparel, accessories, stickers, and more. "+
			"This SSH front-end browses the live catalog over the WooCommerce Store API and hands off "+
			"the checkout to the web via a Cart-Token that's handed to you as a QR code. "+
			"No cookies, no accounts, no JavaScript — just the things you need to pick out a sweatshirt "+
			"without leaving your terminal.",
		m.contentWidth(),
	)
	meta := styles.Muted.Render("Store:    https://mercantile.wordpress.org\nStore API: wp-json/wc/store/v1\nClient:   mercantile-ssh (Wish + Bubbletea)")
	return lipgloss.JoinVertical(lipgloss.Left,
		"", title, "", body, "", meta,
	)
}
