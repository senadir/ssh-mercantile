// apitest: quick sanity check against the live Store API to prove the full
// cart/shipping flow works end-to-end. Not part of the shipping binary.
package main

import (
	"fmt"
	"log"

	"mercantile/api"
)

func main() {
	c := api.New("")

	// Bootstrap the session: /cart returns both the Cart-Token and the Nonce
	// that subsequent write requests need.
	boot, err := c.GetCart()
	must(err)
	fmt.Printf("bootstrap: items=%d nonce=%s token=%s\n", boot.ItemsCount, truncate(c.Nonce(), 10), truncate(c.CartToken(), 16))

	cats, err := c.ListCategories()
	must(err)
	fmt.Printf("categories: %d\n", len(cats))

	// pick a simple product
	_ = cats
	ps, err := c.ListProducts(api.ProductsParams{PerPage: 5})
	must(err)
	fmt.Printf("products: %d (token=%s)\n", len(ps), truncate(c.CartToken(), 16))
	var simple *api.Product
	for i := range ps {
		if ps[i].Type == "simple" && ps[i].IsPurchasable {
			simple = &ps[i]
			break
		}
	}
	if simple == nil {
		log.Fatal("no simple purchasable product")
	}
	fmt.Printf("simple: %d %s (%s)\n", simple.ID, simple.Name, simple.Prices.Price)

	cart, err := c.AddItem(api.AddItemRequest{ID: simple.ID, Quantity: 1})
	must(err)
	fmt.Printf("cart: %d items, total=%s\n", cart.ItemsCount, cart.Totals.TotalPrice)

	// update address to get shipping rates
	addr := api.Address{
		FirstName: "Ada", LastName: "Lovelace",
		Address1: "10 Downing St", City: "London",
		Postcode: "SW1A 2AA", Country: "GB",
		Email: "ada@example.com",
	}
	ship := addr
	ship.Email = ""
	cart, err = c.UpdateCustomer(api.UpdateCustomerRequest{
		BillingAddress:  &addr,
		ShippingAddress: &ship,
	})
	must(err)
	fmt.Printf("after address: calc_shipping=%v packages=%d\n", cart.HasCalculatedShipping, len(cart.ShippingRates))
	for _, p := range cart.ShippingRates {
		fmt.Printf("  package %v rates=%d\n", p.PackageID, len(p.ShippingRates))
		for _, r := range p.ShippingRates {
			fmt.Printf("    %s  %s  %s\n", r.RateID, r.Name, r.Price)
		}
	}

	// pick first rate
	if len(cart.ShippingRates) > 0 && len(cart.ShippingRates[0].ShippingRates) > 0 {
		pkg := cart.ShippingRates[0]
		rate := pkg.ShippingRates[0]
		cart, err = c.SelectShippingRate(pkg.PackageID, rate.RateID)
		must(err)
		var shipTotal string
		if cart.Totals.TotalShipping != nil {
			shipTotal = *cart.Totals.TotalShipping
		}
		fmt.Printf("selected %q; total=%s shipping=%s\n", rate.Name, cart.Totals.TotalPrice, shipTotal)
	}

	fmt.Printf("cart-token: %s\n", c.CartToken())
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
