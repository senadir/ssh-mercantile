package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/styles"
)

func (m *Model) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
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
	case "up", "k":
		if m.browseIdx > 0 {
			m.browseIdx--
		}
	case "down", "j":
		if m.browseIdx < len(m.browseItems)-1 {
			m.browseIdx++
		}
	case "home", "g":
		m.browseIdx = 0
	case "end", "G":
		m.browseIdx = len(m.browseItems) - 1
	case "enter":
		if m.browseIdx < len(m.browseItems) {
			p := m.browseItems[m.browseIdx]
			m.push(viewProduct)
			m.productLoading = true
			m.product = nil
			return m, loadProductCmd(m.client, p.ID)
		}
	case "c":
		m.push(viewCart)
		return m, loadCartCmd(m.client)
	case "/":
		m.push(viewSearch)
		m.searchInput.Focus()
		m.searchInput.SetValue("")
		return m, nil
	}
	return m, nil
}

func (m *Model) viewBrowse() string {
	title := styles.Title.Render(m.browseTitle)
	count := styles.Muted.Render(fmt.Sprintf("(%d)", len(m.browseItems)))
	header := title + " " + count

	if m.browseLoading {
		return "\n" + header + "\n\n" + styles.Muted.Render("  Loading…")
	}
	if len(m.browseItems) == 0 {
		return "\n" + header + "\n\n" + styles.Muted.Render("  No products found.")
	}

	maxVisible := m.height - 8
	if maxVisible < 8 {
		maxVisible = 8
	}
	start := 0
	if m.browseIdx >= maxVisible {
		start = m.browseIdx - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.browseItems) {
		end = len(m.browseItems)
	}

	var rows []string
	for i := start; i < end; i++ {
		p := m.browseItems[i]
		displayName := stripHTML(p.Name)
		name := truncate(displayName, 50)
		price := formatPriceFromPrices(p.Prices.Price, p.Prices)
		typ := ""
		if p.Type == "variable" {
			typ = styles.Muted.Render(" (options)")
		}
		line := fmt.Sprintf("%-50s %s%s", name, styles.Accent.Render(price), typ)
		if i == m.browseIdx {
			line = styles.Selected.Render("▸ " + truncate(displayName, 48) + "  " + formatPriceFromPrices(p.Prices.Price, p.Prices))
		} else {
			line = "  " + line
		}
		rows = append(rows, line)
	}

	if end < len(m.browseItems) {
		rows = append(rows, styles.Muted.Render(fmt.Sprintf("  … %d more (scroll to view)", len(m.browseItems)-end)))
	}

	body := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return "\n" + header + "\n\n" + body
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n < 1 {
		return ""
	}
	return s[:n-1] + "…"
}
