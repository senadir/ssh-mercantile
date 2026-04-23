package api

type Prices struct {
	Price              string `json:"price"`
	RegularPrice       string `json:"regular_price"`
	SalePrice          string `json:"sale_price"`
	CurrencyCode       string `json:"currency_code"`
	CurrencySymbol     string `json:"currency_symbol"`
	CurrencyMinorUnit  int    `json:"currency_minor_unit"`
	CurrencyDecimalSep string `json:"currency_decimal_separator"`
	CurrencyThousandSep string `json:"currency_thousand_separator"`
	CurrencyPrefix     string `json:"currency_prefix"`
	CurrencySuffix     string `json:"currency_suffix"`
}

type Category struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Parent int    `json:"parent"`
	Count  int    `json:"count"`
}

type AttributeTerm struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type Attribute struct {
	ID            int             `json:"id"`
	Name          string          `json:"name"`
	Taxonomy      string          `json:"taxonomy"`
	HasVariations bool            `json:"has_variations"`
	Terms         []AttributeTerm `json:"terms"`
}

type VariationAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type VariationRef struct {
	ID         int                  `json:"id"`
	Attributes []VariationAttribute `json:"attributes"`
}

type Product struct {
	ID               int            `json:"id"`
	Name             string         `json:"name"`
	Slug             string         `json:"slug"`
	Type             string         `json:"type"`
	ShortDescription string         `json:"short_description"`
	Description      string         `json:"description"`
	OnSale           bool           `json:"on_sale"`
	IsPurchasable    bool           `json:"is_purchasable"`
	IsInStock        bool           `json:"is_in_stock"`
	HasOptions       bool           `json:"has_options"`
	Prices           Prices         `json:"prices"`
	PriceHTML        string         `json:"price_html"`
	Categories       []Category     `json:"categories"`
	Attributes       []Attribute    `json:"attributes"`
	Variations       []VariationRef `json:"variations"`
	Permalink        string         `json:"permalink"`
}

type CartItemVariation struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type CartItemTotals struct {
	LineSubtotal string `json:"line_subtotal"`
	LineTotal    string `json:"line_total"`
	CurrencyCode string `json:"currency_code"`
	CurrencyMinorUnit int `json:"currency_minor_unit"`
	CurrencyPrefix    string `json:"currency_prefix"`
	CurrencySuffix    string `json:"currency_suffix"`
}

type CartItem struct {
	Key        string              `json:"key"`
	ID         int                 `json:"id"`
	Quantity   int                 `json:"quantity"`
	Name       string              `json:"name"`
	ShortDescription string        `json:"short_description"`
	Variation  []CartItemVariation `json:"variation"`
	Prices     Prices              `json:"prices"`
	Totals     CartItemTotals      `json:"totals"`
}

type Address struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Company   string `json:"company"`
	Address1  string `json:"address_1"`
	Address2  string `json:"address_2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Postcode  string `json:"postcode"`
	Country   string `json:"country"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone"`
}

type ShippingRate struct {
	RateID       string `json:"rate_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DeliveryTime string `json:"delivery_time"`
	Price        string `json:"price"`
	Taxes        string `json:"taxes"`
	MethodID     string `json:"method_id"`
	InstanceID   int    `json:"instance_id"`
	Selected     bool   `json:"selected"`
	CurrencyCode       string `json:"currency_code"`
	CurrencyMinorUnit  int    `json:"currency_minor_unit"`
	CurrencyPrefix     string `json:"currency_prefix"`
	CurrencySuffix     string `json:"currency_suffix"`
}

type ShippingPackage struct {
	PackageID   any            `json:"package_id"`
	Name        string         `json:"name"`
	Destination Address        `json:"destination"`
	ShippingRates []ShippingRate `json:"shipping_rates"`
}

type CartTotals struct {
	TotalItems    string `json:"total_items"`
	TotalShipping *string `json:"total_shipping"`
	TotalDiscount string `json:"total_discount"`
	TotalTax      string `json:"total_tax"`
	TotalPrice    string `json:"total_price"`
	CurrencyCode  string `json:"currency_code"`
	CurrencyMinorUnit int `json:"currency_minor_unit"`
	CurrencyPrefix    string `json:"currency_prefix"`
	CurrencySuffix    string `json:"currency_suffix"`
}

type Cart struct {
	Items                []CartItem        `json:"items"`
	ItemsCount           int               `json:"items_count"`
	Totals               CartTotals        `json:"totals"`
	NeedsShipping        bool              `json:"needs_shipping"`
	NeedsPayment         bool              `json:"needs_payment"`
	HasCalculatedShipping bool             `json:"has_calculated_shipping"`
	ShippingRates        []ShippingPackage `json:"shipping_rates"`
	ShippingAddress      Address           `json:"shipping_address"`
	BillingAddress       Address           `json:"billing_address"`
	Errors               []any             `json:"errors"`
}
