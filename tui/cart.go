package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/styles"
)

func (m *Model) updateCart(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.cartIdx > 0 {
			m.cartIdx--
		}
	case "down", "j":
		if m.cart != nil && m.cartIdx < len(m.cart.Items)-1 {
			m.cartIdx++
		}
	case "+", "=":
		if m.cart != nil && m.cartIdx < len(m.cart.Items) {
			it := m.cart.Items[m.cartIdx]
			return m, updateItemCmd(m.client, it.Key, it.Quantity+1)
		}
	case "-", "_":
		if m.cart != nil && m.cartIdx < len(m.cart.Items) {
			it := m.cart.Items[m.cartIdx]
			if it.Quantity > 1 {
				return m, updateItemCmd(m.client, it.Key, it.Quantity-1)
			}
		}
	case "d", "x", "delete":
		if m.cart != nil && m.cartIdx < len(m.cart.Items) {
			it := m.cart.Items[m.cartIdx]
			if m.cartIdx > 0 {
				m.cartIdx--
			}
			return m, removeItemCmd(m.client, it.Key)
		}
	case "enter":
		if m.cart != nil && m.cart.ItemsCount > 0 {
			m.push(viewAddress)
			m.addrIdx = 0
			m.focusAddressInput(0)
			return m, nil
		}
		return m, m.setStatus("Cart is empty", true)
	}
	return m, nil
}

func (m *Model) viewCart() string {
	title := styles.Title.Render("Your cart")
	if m.cart == nil {
		return "\n" + title + "\n\n" + styles.Muted.Render("  Loading…")
	}
	if len(m.cart.Items) == 0 {
		return "\n" + title + "\n\n" + styles.Muted.Render("  Your cart is empty. Press 'esc' to browse.")
	}

	var rows []string
	for i, it := range m.cart.Items {
		var variation string
		if len(it.Variation) > 0 {
			var parts []string
			for _, v := range it.Variation {
				parts = append(parts, fmt.Sprintf("%s: %s", stripHTML(v.Attribute), stripHTML(v.Value)))
			}
			variation = " " + styles.Muted.Render("("+strings.Join(parts, ", ")+")")
		}
		name := truncate(stripHTML(it.Name), 46)
		priceText := formatPriceFromPrices(it.Totals.LineTotal, it.Prices)

		var line string
		if i == m.cartIdx {
			rowBg := styles.SelectedRow
			bar := styles.SelectBar.Inherit(rowBg).Render("▍ ")
			nameCell := styles.SelectedItem.Inherit(rowBg).Render(fmt.Sprintf("%-46s", name))
			qty := rowBg.Render(fmt.Sprintf("  x%-3d  ", it.Quantity))
			price := styles.SelectedItem.Inherit(rowBg).Render(priceText)
			line = bar + nameCell + qty + price
		} else {
			bar := "  "
			nameCell := fmt.Sprintf("%-46s", name)
			price := styles.Accent.Render(priceText)
			line = fmt.Sprintf("%s%s  x%-3d  %s", bar, nameCell, it.Quantity, price)
		}
		rows = append(rows, line+variation)
	}

	subtotal := formatCartTotal(m.cart.Totals.TotalItems, m.cart.Totals)
	discount := formatCartTotal(m.cart.Totals.TotalDiscount, m.cart.Totals)
	tax := formatCartTotal(m.cart.Totals.TotalTax, m.cart.Totals)
	total := formatCartTotal(m.cart.Totals.TotalPrice, m.cart.Totals)

	var totals []string
	totals = append(totals, fmt.Sprintf("  %-20s %s", "Items:", subtotal))
	if m.cart.Totals.TotalDiscount != "0" && m.cart.Totals.TotalDiscount != "" {
		totals = append(totals, fmt.Sprintf("  %-20s %s", "Discount:", discount))
	}
	if m.cart.Totals.TotalShipping != nil && *m.cart.Totals.TotalShipping != "" {
		totals = append(totals, fmt.Sprintf("  %-20s %s", "Shipping:", formatCartTotal(*m.cart.Totals.TotalShipping, m.cart.Totals)))
	}
	if m.cart.Totals.TotalTax != "0" && m.cart.Totals.TotalTax != "" {
		totals = append(totals, fmt.Sprintf("  %-20s %s", "Tax:", tax))
	}
	totals = append(totals, "  "+styles.Muted.Render(strings.Repeat("─", 30)))
	totals = append(totals, fmt.Sprintf("  %-20s %s", styles.Title.Render("Total:"), styles.Accent.Render(total)))

	blocks := []string{"", title, ""}
	blocks = append(blocks, rows...)
	blocks = append(blocks, "")
	blocks = append(blocks, totals...)

	cta := renderButton(buttonOpts{
		Label:   "Proceed to checkout",
		Hotkey:  "enter",
		Variant: btnPrimary,
		Focused: true,
	})
	indent := lipgloss.NewStyle().MarginLeft(2)
	blocks = append(blocks, "", indent.Render(cta))

	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}
