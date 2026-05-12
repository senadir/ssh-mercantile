package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

func buildShippingFlat(cart *api.Cart) []shippingOption {
	var out []shippingOption
	if cart == nil {
		return out
	}
	for pi, pkg := range cart.ShippingRates {
		for ri := range pkg.ShippingRates {
			out = append(out, shippingOption{pkgIndex: pi, rateIndex: ri})
		}
	}
	return out
}

func (m *Model) updateShipping(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.shippingIdx > 0 {
			m.shippingIdx--
		}
	case "down", "j":
		if m.shippingIdx < len(m.shippingFlat)-1 {
			m.shippingIdx++
		}
	case "enter":
		if m.cart == nil || !m.cart.HasCalculatedShipping {
			return m, m.setStatus("No shipping rates available for this address", true)
		}
		if len(m.shippingFlat) == 0 {
			return m, m.setStatus("No shipping options", true)
		}
		opt := m.shippingFlat[m.shippingIdx]
		pkg := m.cart.ShippingRates[opt.pkgIndex]
		rate := pkg.ShippingRates[opt.rateIndex]
		m.push(viewCheckout)
		return m, selectShippingRateCmd(m.client, pkg.PackageID, rate.RateID)
	}
	return m, nil
}

func (m *Model) viewShipping() string {
	title := styles.Title.Render("Choose shipping")

	if m.cart == nil {
		return "\n" + title + "\n\n" + styles.Muted.Render("  Calculating…")
	}
	if !m.cart.HasCalculatedShipping {
		return "\n" + title + "\n\n" + styles.Muted.Render("  Calculating shipping for your address…")
	}
	if len(m.cart.ShippingRates) == 0 {
		return "\n" + title + "\n\n" + styles.Error.Render("  No shipping packages returned. Try a different address.")
	}

	var rows []string
	flatIdx := 0
	for pi, pkg := range m.cart.ShippingRates {
		pkgLabel := pkg.Name
		if pkgLabel == "" {
			pkgLabel = fmt.Sprintf("Package %d", pi+1)
		}
		rows = append(rows, "")
		rows = append(rows, styles.Subtle.Render(pkgLabel))
		if len(pkg.ShippingRates) == 0 {
			rows = append(rows, styles.Muted.Render("  (no rates for this address)"))
			continue
		}
		for _, r := range pkg.ShippingRates {
			var line string
			if flatIdx == m.shippingIdx {
				rowBg := styles.SelectedRow
				bar := styles.SelectBar.Inherit(rowBg).Render("▍ ")
				text := styles.SelectedItem.Inherit(rowBg).Render(formatShippingRate(r))
				line = bar + text
			} else {
				line = "  " + formatShippingRate(r)
			}
			rows = append(rows, line)
			flatIdx++
		}
	}

	blocks := []string{"", title}
	blocks = append(blocks, rows...)
	cta := renderButton(buttonOpts{
		Label:   "Confirm",
		Hotkey:  "enter",
		Variant: btnPrimary,
		Focused: true,
	})
	indent := lipgloss.NewStyle().MarginLeft(2)
	blocks = append(blocks, "", indent.Render(cta))
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}
