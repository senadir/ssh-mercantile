package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

// address form field indexes
const (
	fFirstName = iota
	fLastName
	fEmail
	fAddress1
	fAddress2
	fCity
	fState
	fPostcode
	fCountry
	fPhone
	fCount
)

var addrLabels = []string{
	"First name",
	"Last name",
	"Email",
	"Address line 1",
	"Address line 2",
	"City",
	"State / Province",
	"Postcode",
	"Country (ISO-2, e.g. US, GB, DE)",
	"Phone",
}

func (m *Model) setupAddressInputs() {
	m.addrInputs = make([]textinput.Model, fCount)
	for i := range m.addrInputs {
		ti := textinput.New()
		ti.Prompt = ""
		ti.CharLimit = 120
		ti.Width = 40
		if i == fCountry {
			ti.CharLimit = 2
			ti.Width = 6
		}
		m.addrInputs[i] = ti
	}
}

func (m *Model) focusAddressInput(i int) {
	for j := range m.addrInputs {
		if j == i {
			m.addrInputs[j].Focus()
		} else {
			m.addrInputs[j].Blur()
		}
	}
}

func (m *Model) collectAddress() api.Address {
	g := func(i int) string { return strings.TrimSpace(m.addrInputs[i].Value()) }
	return api.Address{
		FirstName: g(fFirstName),
		LastName:  g(fLastName),
		Address1:  g(fAddress1),
		Address2:  g(fAddress2),
		City:      g(fCity),
		State:     g(fState),
		Postcode:  g(fPostcode),
		Country:   strings.ToUpper(g(fCountry)),
		Phone:     g(fPhone),
		Email:     g(fEmail),
	}
}

func (m *Model) validateAddress() string {
	a := m.collectAddress()
	switch {
	case a.FirstName == "":
		return "First name is required"
	case a.LastName == "":
		return "Last name is required"
	case a.Address1 == "":
		return "Street address is required"
	case a.City == "":
		return "City is required"
	case a.Postcode == "":
		return "Postcode is required"
	case a.Country == "":
		return "Country is required"
	case len(a.Country) != 2:
		return "Country must be an ISO-2 code (e.g. US, GB, DE)"
	}
	return ""
}

func (m *Model) updateAddress(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.pop()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			m.addrIdx = (m.addrIdx + 1) % fCount
			m.focusAddressInput(m.addrIdx)
			return m, nil
		case "shift+tab", "up":
			m.addrIdx = (m.addrIdx - 1 + fCount) % fCount
			m.focusAddressInput(m.addrIdx)
			return m, nil
		case "enter":
			if err := m.validateAddress(); err != "" {
				return m, m.setStatus(err, true)
			}
			addr := m.collectAddress()
			billing := addr
			shipping := addr
			shipping.Email = ""
			m.push(viewShipping)
			return m, updateCustomerCmd(m.client, api.UpdateCustomerRequest{
				ShippingAddress: &shipping,
				BillingAddress:  &billing,
			})
		}
	}
	var cmd tea.Cmd
	m.addrInputs[m.addrIdx], cmd = m.addrInputs[m.addrIdx].Update(msg)
	return m, cmd
}

func (m *Model) viewAddress() string {
	title := styles.Title.Render("Shipping address")
	hint := styles.Muted.Render("We'll calculate shipping rates for this address.")

	var rows []string
	for i, ti := range m.addrInputs {
		label := addrLabels[i]
		marker := "  "
		if i == m.addrIdx {
			marker = styles.Key.Render("▸ ")
		}
		rows = append(rows, marker+styles.Subtle.Render(padR(label+":", 32))+ti.View())
	}

	blocks := []string{"", title, hint, ""}
	blocks = append(blocks, rows...)
	blocks = append(blocks, "", "  "+styles.Selected.Render(" Continue to shipping (enter) "))
	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}
