package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/skip2/go-qrcode"

	"mercantile/styles"
)

func (m *Model) updateCheckout(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *Model) checkoutURL() string {
	token := m.client.CartToken()
	base := strings.TrimRight(m.siteURL, "/")
	return fmt.Sprintf("%s/checkout?session=%s", base, token)
}

// renderQR returns a QR code for the given URL using Unicode half-block chars.
// Each character row encodes two QR rows; a bit is "on" when it's a dark module.
func renderQR(url string) (string, error) {
	q, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return "", err
	}
	bits := q.Bitmap()
	// Bitmap includes a quiet zone border; trim 2 modules on each side for density.
	// Actually, keep as-is — the quiet zone helps scanners.
	rows := len(bits)
	cols := 0
	if rows > 0 {
		cols = len(bits[0])
	}
	var b strings.Builder
	for y := 0; y < rows; y += 2 {
		for x := 0; x < cols; x++ {
			top := bits[y][x]
			var bot bool
			if y+1 < rows {
				bot = bits[y+1][x]
			}
			switch {
			case top && bot:
				b.WriteString("█")
			case top && !bot:
				b.WriteString("▀")
			case !top && bot:
				b.WriteString("▄")
			default:
				b.WriteString(" ")
			}
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func (m *Model) viewCheckout() string {
	title := styles.Title.Render("Complete your order")
	url := m.checkoutURL()

	qr, err := renderQR(url)
	if err != nil {
		qr = styles.Error.Render("(failed to render QR: " + err.Error() + ")")
	} else {
		qr = lipgloss.NewStyle().Foreground(styles.ColorFg).Render(qr)
	}

	hint := styles.Subtle.Render("Scan this QR code or open the URL below on your phone to finish checkout.")
	urlLine := lipgloss.NewStyle().Foreground(styles.ColorAccent).Render(url)

	summary := ""
	if m.cart != nil {
		total := formatCartTotal(m.cart.Totals.TotalPrice, m.cart.Totals)
		summary = styles.Muted.Render(fmt.Sprintf("Order total: %s · %d item(s)", total, m.cart.ItemsCount))
	}

	blocks := []string{
		"",
		title,
		hint,
		"",
		qr,
		"",
		urlLine,
		"",
		summary,
	}
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}
