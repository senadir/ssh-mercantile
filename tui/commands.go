package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"mercantile/api"
)

type (
	categoriesLoadedMsg struct{ categories []api.Category }
	productsLoadedMsg   struct {
		products []api.Product
		title    string
	}
	productLoadedMsg struct{ product *api.Product }
	cartUpdatedMsg   struct{ cart *api.Cart }
	errMsg           struct{ err error }
	addedToCartMsg   struct{ cart *api.Cart }
	statusMsg        struct{ text string }
	clearStatusMsg   struct{}
)

func (e errMsg) Error() string { return e.err.Error() }

func loadCategoriesCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		cats, err := c.ListCategories()
		if err != nil {
			return errMsg{err}
		}
		return categoriesLoadedMsg{cats}
	}
}

func loadProductsCmd(c *api.Client, p api.ProductsParams, title string) tea.Cmd {
	return func() tea.Msg {
		ps, err := c.ListProducts(p)
		if err != nil {
			return errMsg{err}
		}
		return productsLoadedMsg{ps, title}
	}
}

func loadProductCmd(c *api.Client, id int) tea.Cmd {
	return func() tea.Msg {
		p, err := c.GetProduct(id)
		if err != nil {
			return errMsg{err}
		}
		return productLoadedMsg{p}
	}
}

func loadCartCmd(c *api.Client) tea.Cmd {
	return func() tea.Msg {
		cart, err := c.GetCart()
		if err != nil {
			return errMsg{err}
		}
		return cartUpdatedMsg{cart}
	}
}

func addToCartCmd(c *api.Client, req api.AddItemRequest) tea.Cmd {
	return func() tea.Msg {
		cart, err := c.AddItem(req)
		if err != nil {
			return errMsg{err}
		}
		return addedToCartMsg{cart}
	}
}

func updateItemCmd(c *api.Client, key string, qty int) tea.Cmd {
	return func() tea.Msg {
		cart, err := c.UpdateItem(key, qty)
		if err != nil {
			return errMsg{err}
		}
		return cartUpdatedMsg{cart}
	}
}

func removeItemCmd(c *api.Client, key string) tea.Cmd {
	return func() tea.Msg {
		cart, err := c.RemoveItem(key)
		if err != nil {
			return errMsg{err}
		}
		return cartUpdatedMsg{cart}
	}
}

func updateCustomerCmd(c *api.Client, req api.UpdateCustomerRequest) tea.Cmd {
	return func() tea.Msg {
		cart, err := c.UpdateCustomer(req)
		if err != nil {
			return errMsg{err}
		}
		return cartUpdatedMsg{cart}
	}
}

func selectShippingRateCmd(c *api.Client, pkgID any, rateID string) tea.Cmd {
	return func() tea.Msg {
		cart, err := c.SelectShippingRate(pkgID, rateID)
		if err != nil {
			return errMsg{err}
		}
		return cartUpdatedMsg{cart}
	}
}
