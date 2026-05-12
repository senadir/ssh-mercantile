package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

// product view layout:
//  [0..n-1]  attribute pickers (one per variation attribute)
//  [n]       Add to cart
//  [n+1]     Back

func (m *Model) productFocusCount() int {
	if m.product == nil {
		return 2
	}
	return len(m.variationAttributes()) + 2
}

// addToCartDisabled reports whether the Add to cart button should be
// shown as disabled — i.e. the product is variable but the user hasn't
// picked every option yet.
func (m *Model) addToCartDisabled() bool {
	if m.product == nil {
		return true
	}
	if m.product.Type != "variable" {
		return false
	}
	for _, a := range m.variationAttributes() {
		if m.productChoice[a.Taxonomy] == "" {
			return true
		}
	}
	return false
}

// triggerAddToCart centralises the logic for pressing Add to cart via
// either the "a" hotkey or Enter on the focused button. It respects the
// disabled and in-flight states so a double-press can't fire two API
// calls.
func (m *Model) triggerAddToCart() (tea.Model, tea.Cmd) {
	if m.addingToCart {
		return m, nil
	}
	if m.addToCartDisabled() {
		return m, m.setStatus("Pick every option first", true)
	}
	m.addingToCart = true
	return m, m.addCurrentToCart()
}

func (m *Model) variationAttributes() []api.Attribute {
	if m.product == nil {
		return nil
	}
	var out []api.Attribute
	for _, a := range m.product.Attributes {
		if a.HasVariations {
			out = append(out, a)
		}
	}
	return out
}

func (m *Model) updateProduct(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.productIdx > 0 {
			m.productIdx--
		}
	case "down", "j":
		if m.productIdx < m.productFocusCount()-1 {
			m.productIdx++
		}
	case "left", "h":
		m.cycleAttr(-1)
	case "right", "l":
		m.cycleAttr(1)
	case "a":
		return m.triggerAddToCart()
	case "c":
		m.push(viewCart)
		return m, loadCartCmd(m.client)
	case "enter":
		attrs := m.variationAttributes()
		switch m.productIdx {
		case len(attrs):
			return m.triggerAddToCart()
		case len(attrs) + 1:
			m.pop()
			return m, nil
		default:
			// on an attribute row: cycle
			m.cycleAttr(1)
		}
	}
	return m, nil
}

func (m *Model) cycleAttr(dir int) {
	attrs := m.variationAttributes()
	if m.productIdx >= len(attrs) {
		return
	}
	a := attrs[m.productIdx]
	if len(a.Terms) == 0 {
		return
	}
	if m.productChoice == nil {
		m.productChoice = map[string]string{}
	}
	current := m.productChoice[a.Taxonomy]
	idx := -1
	for i, t := range a.Terms {
		if t.Slug == current {
			idx = i
			break
		}
	}
	idx += dir
	if idx < 0 {
		idx = len(a.Terms) - 1
	}
	if idx >= len(a.Terms) {
		idx = 0
	}
	m.productChoice[a.Taxonomy] = a.Terms[idx].Slug
}

func (m *Model) addCurrentToCart() tea.Cmd {
	if m.product == nil {
		return nil
	}
	req := api.AddItemRequest{ID: m.product.ID, Quantity: 1}
	if m.product.Type == "variable" {
		attrs := m.variationAttributes()
		var chosen []api.VariationAttribute
		for _, a := range attrs {
			slug := m.productChoice[a.Taxonomy]
			if slug == "" {
				return m.setStatus(fmt.Sprintf("Select %s first", a.Name), true)
			}
			chosen = append(chosen, api.VariationAttribute{
				Name:  a.Taxonomy, // the API accepts the taxonomy name (e.g. pa_size)
				Value: slug,
			})
		}
		// resolve the matching variation id so the API can price correctly
		vID := m.resolveVariationID(chosen)
		if vID == 0 {
			return m.setStatus("No matching variation", true)
		}
		req.ID = vID
		req.Variation = chosen
	}
	return addToCartCmd(m.client, req)
}

func (m *Model) resolveVariationID(chosen []api.VariationAttribute) int {
	if m.product == nil {
		return 0
	}
	want := map[string]string{}
	for _, a := range chosen {
		want[strings.TrimPrefix(a.Name, "pa_")] = a.Value
	}
outer:
	for _, v := range m.product.Variations {
		for _, va := range v.Attributes {
			key := strings.ToLower(va.Name)
			key = strings.TrimPrefix(key, "pa_")
			if w, ok := want[key]; ok && w != va.Value {
				continue outer
			}
		}
		// check all wanted attrs are matched
		match := 0
		for _, va := range v.Attributes {
			key := strings.ToLower(va.Name)
			key = strings.TrimPrefix(key, "pa_")
			if w, ok := want[key]; ok && w == va.Value {
				match++
			}
		}
		if match == len(want) {
			return v.ID
		}
	}
	return 0
}

func (m *Model) viewProduct() string {
	if m.productLoading || m.product == nil {
		return "\n" + styles.Muted.Render("  Loading…")
	}
	p := m.product
	title := styles.Title.Render(stripHTML(p.Name))
	price := styles.Accent.Render(formatPriceFromPrices(p.Prices.Price, p.Prices))
	short := stripHTML(p.ShortDescription)
	if short == "" {
		short = stripHTML(p.Description)
	}
	// When the image sits to the left, the text column is narrower —
	// wrap the description to match so it doesn't extend off-screen.
	textWidth := m.contentWidth()
	sideBySide := len(p.Images) > 0 && m.width >= sideBySideMinWidth
	if sideBySide {
		textWidth = m.width - imageCols - 4
		if textWidth < 30 {
			textWidth = 30
		}
	}
	short = wrap(short, textWidth)

	attrs := m.variationAttributes()

	var rows []string
	for i, a := range attrs {
		current := m.productChoice[a.Taxonomy]
		currentName := "—"
		for _, t := range a.Terms {
			if t.Slug == current {
				currentName = t.Name
				break
			}
		}

		var line string
		if i == m.productIdx {
			rowBg := styles.SelectedRow
			bar := styles.SelectBar.Inherit(rowBg).Render("▍ ")
			label := styles.SelectedItem.Inherit(rowBg).Render(fmt.Sprintf("%s: ", a.Name))
			value := styles.SelectedItem.Inherit(rowBg).Render(currentName)
			hint := styles.Muted.Inherit(rowBg).Render("   (← →)")
			line = bar + label + value + hint
		} else {
			line = fmt.Sprintf("  %s: %s   %s", a.Name, styles.Accent.Render(currentName), styles.Muted.Render("(← →)"))
		}
		rows = append(rows, line)
	}

	addBtn := renderButton(buttonOpts{
		Label:    "Add to cart",
		Hotkey:   "a",
		Variant:  btnPrimary,
		Focused:  m.productIdx == len(attrs),
		Disabled: m.addToCartDisabled(),
		Loading:  m.addingToCart,
		Spinner:  m.spinner.View(),
	})
	backBtn := renderButton(buttonOpts{
		Label:   "Back",
		Hotkey:  "esc",
		Variant: btnSecondary,
		Focused: m.productIdx == len(attrs)+1,
	})

	// Product image (or a same-height placeholder so layout doesn't pop
	// when the async fetch lands).
	var imageBlock string
	if rendered, ok := m.imageCache[p.ID]; ok {
		imageBlock = rendered
	} else if len(p.Images) > 0 {
		placeholder := styles.Muted.Render("(loading image…)")
		imageBlock = strings.Repeat("\n", imageRows/2) + placeholder + strings.Repeat("\n", imageRows/2-1)
	}

	// Build the text column (everything that's not the image).
	textParts := []string{title, price, "", styles.Subtle.Render(short)}
	if len(rows) > 0 {
		textParts = append(textParts, "", styles.Muted.Render("Options"))
		textParts = append(textParts, rows...)
	}
	indent := lipgloss.NewStyle().MarginLeft(2)
	textParts = append(textParts, "", indent.Render(addBtn), "", indent.Render(backBtn))
	textBlock := lipgloss.JoinVertical(lipgloss.Left, textParts...)

	if sideBySide && imageBlock != "" {
		// Image left, text right with a 2-col gutter. Top-aligned so the
		// image and the title share a baseline.
		gutter := lipgloss.NewStyle().MarginLeft(2)
		return "\n" + lipgloss.JoinHorizontal(lipgloss.Top, imageBlock, gutter.Render(textBlock))
	}

	// Stacked fallback for narrow terminals (or no image).
	blocks := []string{""}
	if imageBlock != "" {
		blocks = append(blocks, imageBlock, "")
	}
	blocks = append(blocks, textBlock)
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

// sideBySideMinWidth is the terminal width at and above which we put the
// product image next to the text column instead of stacked above. Image
// is imageCols (30) wide, 2-col gutter, and we want at least 30 cols left
// for buttons / description to breathe.
const sideBySideMinWidth = 62

func (m *Model) contentWidth() int {
	if m.width < 20 {
		return 60
	}
	if m.width > 100 {
		return 100
	}
	return m.width - 4
}

func wrap(s string, w int) string {
	if w <= 0 {
		return s
	}
	var out strings.Builder
	line := 0
	for _, word := range strings.Fields(s) {
		if line == 0 {
			out.WriteString(word)
			line = len(word)
			continue
		}
		if line+1+len(word) > w {
			out.WriteByte('\n')
			out.WriteString(word)
			line = len(word)
		} else {
			out.WriteByte(' ')
			out.WriteString(word)
			line += 1 + len(word)
		}
	}
	return out.String()
}
