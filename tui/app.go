package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"mercantile/api"
	"mercantile/styles"
)

type view int

const (
	viewWelcome view = iota
	viewBrowse
	viewProduct
	viewCart
	viewAddress
	viewShipping
	viewCheckout
	viewAbout
	viewSearch
)

type Model struct {
	client    *api.Client
	siteURL   string
	width     int
	height    int
	view      view
	viewStack []view

	// shared data
	categories []api.Category
	cart       *api.Cart
	status     string
	statusErr  bool

	// welcome state
	welcomeIdx int

	// browse state
	browseTitle   string
	browseItems   []api.Product
	browseIdx     int
	browseLoading bool
	browseSource  string // "apparel","accessories","all","search:<q>","category:<id>"

	// product state
	product        *api.Product
	productIdx     int // focused attribute index or button
	productChoice  map[string]string // attribute taxonomy -> term slug
	productLoading bool
	addingToCart   bool // Add to cart API call is in flight

	// shared spinner — ticks forever once Init runs, buttons query its View()
	spinner spinner.Model

	// cart state
	cartIdx int

	// address state
	addrInputs []textinput.Model
	addrIdx    int

	// shipping state
	shippingIdx int // flat index across packages*rates
	shippingFlat []shippingOption

	// search state
	searchInput textinput.Model
}

type shippingOption struct {
	pkgIndex  int
	rateIndex int
}

func NewModel(client *api.Client, siteURL string) *Model {
	if siteURL == "" {
		siteURL = "https://mercantile.wordpress.org"
	}
	m := &Model{
		client:  client,
		siteURL: siteURL,
		view:    viewWelcome,
	}
	m.setupAddressInputs()
	m.setupSearchInput()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	// Pale grey-blue — reads as "working" without competing with the
	// surrounding button's bold white text on WP-blue fill.
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#cbd5e1"))
	m.spinner = sp

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		loadCategoriesCmd(m.client),
		loadCartCmd(m.client),
		m.spinner.Tick,
	)
}

func (m *Model) push(v view) {
	m.viewStack = append(m.viewStack, m.view)
	m.view = v
}

func (m *Model) pop() {
	if len(m.viewStack) == 0 {
		m.view = viewWelcome
		return
	}
	last := m.viewStack[len(m.viewStack)-1]
	m.viewStack = m.viewStack[:len(m.viewStack)-1]
	m.view = last
}

func (m *Model) setStatus(text string, isErr bool) tea.Cmd {
	m.status = text
	m.statusErr = isErr
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg { return clearStatusMsg{} })
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case categoriesLoadedMsg:
		m.categories = msg.categories
		return m, nil

	case cartUpdatedMsg:
		m.cart = msg.cart
		// if we are waiting on shipping rates and they just arrived, rebuild flat list
		if m.view == viewShipping && msg.cart != nil {
			m.shippingFlat = buildShippingFlat(msg.cart)
		}
		return m, nil

	case addedToCartMsg:
		m.cart = msg.cart
		m.addingToCart = false
		return m, m.setStatus("Added to cart", false)

	case spinner.TickMsg:
		// Keep the spinner animating. Each tick returns the next tick,
		// so this self-sustains once started from Init.
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case productsLoadedMsg:
		m.browseItems = msg.products
		m.browseTitle = msg.title
		m.browseIdx = 0
		m.browseLoading = false
		return m, nil

	case productLoadedMsg:
		m.product = msg.product
		m.productChoice = map[string]string{}
		// pre-select first term for each attribute as a gentle default
		if m.product != nil {
			for _, a := range m.product.Attributes {
				if a.HasVariations && len(a.Terms) > 0 {
					// leave empty so user is forced to pick
				}
			}
		}
		m.productIdx = 0
		m.productLoading = false
		return m, nil

	case errMsg:
		m.addingToCart = false
		return m, m.setStatus(msg.err.Error(), true)

	case clearStatusMsg:
		m.status = ""
		m.statusErr = false
		return m, nil

	case tea.KeyMsg:
		// global quit
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	switch m.view {
	case viewWelcome:
		return m.updateWelcome(msg)
	case viewBrowse:
		return m.updateBrowse(msg)
	case viewProduct:
		return m.updateProduct(msg)
	case viewCart:
		return m.updateCart(msg)
	case viewAddress:
		return m.updateAddress(msg)
	case viewShipping:
		return m.updateShipping(msg)
	case viewCheckout:
		return m.updateCheckout(msg)
	case viewAbout:
		return m.updateAbout(msg)
	case viewSearch:
		return m.updateSearch(msg)
	}
	return m, nil
}

func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}
	var body string
	switch m.view {
	case viewWelcome:
		body = m.viewWelcome()
	case viewBrowse:
		body = m.viewBrowse()
	case viewProduct:
		body = m.viewProduct()
	case viewCart:
		body = m.viewCart()
	case viewAddress:
		body = m.viewAddress()
	case viewShipping:
		body = m.viewShipping()
	case viewCheckout:
		body = m.viewCheckout()
	case viewAbout:
		body = m.viewAbout()
	case viewSearch:
		body = m.viewSearch()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		body,
		m.renderFooter(),
	)
}

func (m *Model) renderHeader() string {
	brand := styles.Brand.Render("mercantile")
	tag := styles.Muted.Render("// the WordPress store, over SSH")

	cartBadge := ""
	if m.cart != nil && m.cart.ItemsCount > 0 {
		cartBadge = styles.Selected.Render(fmt.Sprintf(" cart %d ", m.cart.ItemsCount))
	} else {
		cartBadge = styles.Muted.Render(" cart 0 ")
	}

	nav := m.renderNav()
	left := brand + "  " + tag
	// right-align cartBadge
	pad := m.width - lipgloss.Width(left) - lipgloss.Width(cartBadge) - lipgloss.Width(nav) - 4
	if pad < 2 {
		pad = 2
	}
	top := lipgloss.JoinHorizontal(lipgloss.Left,
		left,
		strings.Repeat(" ", pad),
		nav,
		"  ",
		cartBadge,
	)
	rule := styles.Muted.Render(strings.Repeat("─", m.width))
	return top + "\n" + rule
}

func (m *Model) renderNav() string {
	active := ""
	switch {
	case m.view == viewBrowse && m.browseSource == "apparel":
		active = "apparel"
	case m.view == viewBrowse && m.browseSource == "accessories":
		active = "accessories"
	case m.view == viewAbout:
		active = "about"
	}
	items := []struct{ key, label, slug string }{
		{"A", "Apparel", "apparel"},
		{"C", "Accessories", "accessories"},
		{"B", "About", "about"},
	}
	var parts []string
	for _, it := range items {
		key := styles.Key.Render("[" + it.key + "]")
		label := it.label
		if active == it.slug {
			label = styles.NavActive.Render(label)
		} else {
			label = styles.NavInactive.Render(label)
		}
		parts = append(parts, key+" "+label)
	}
	return strings.Join(parts, "   ")
}

func (m *Model) renderFooter() string {
	var help string
	switch m.view {
	case viewWelcome:
		help = keys("a apparel", "c accessories", "/ search", "t cart", "b about", "q quit")
	case viewBrowse:
		help = keys("↑/↓ navigate", "enter open", "c cart", "esc back", "q quit")
	case viewProduct:
		help = keys("↑/↓ select", "→ next attr", "← prev attr", "a add to cart", "c cart", "esc back")
	case viewCart:
		help = keys("↑/↓ select", "+ qty up", "- qty down", "d remove", "enter checkout", "esc back")
	case viewAddress:
		help = keys("tab next field", "shift+tab prev", "enter submit", "esc back")
	case viewShipping:
		help = keys("↑/↓ select", "enter confirm", "esc back")
	case viewCheckout:
		help = keys("esc back", "q quit")
	case viewAbout:
		help = keys("esc back")
	case viewSearch:
		help = keys("enter search", "esc back")
	}
	status := ""
	if m.status != "" {
		if m.statusErr {
			status = styles.Error.Render("✗ "+m.status) + "  "
		} else {
			status = styles.Success.Render("✓ "+m.status) + "  "
		}
	}
	rule := styles.Muted.Render(strings.Repeat("─", m.width))
	return rule + "\n" + status + styles.HelpBar.Render(help)
}

func keys(items ...string) string {
	var parts []string
	for _, it := range items {
		sp := strings.SplitN(it, " ", 2)
		if len(sp) == 2 {
			parts = append(parts, styles.Key.Render(sp[0])+" "+styles.Muted.Render(sp[1]))
		} else {
			parts = append(parts, styles.Muted.Render(it))
		}
	}
	return strings.Join(parts, "   ")
}
