package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultBaseURL = "https://mercantile.wordpress.org/wp-json/wc/store/v1"

type Client struct {
	baseURL   string
	http      *http.Client
	mu        sync.Mutex
	cartToken string
	nonce     string
}

func New(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	// Each client gets its own cookie jar so anonymous WooCommerce sessions
	// (PHPSESSID / wp_woocommerce_session_*) persist for the life of this SSH
	// connection — one shopper per SSH session.
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second, Jar: jar},
	}
}

func (c *Client) CartToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cartToken
}

func (c *Client) setCartToken(token string) {
	if token == "" {
		return
	}
	c.mu.Lock()
	c.cartToken = token
	c.mu.Unlock()
}

func (c *Client) Nonce() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.nonce
}

func (c *Client) setNonce(n string) {
	if n == "" {
		return
	}
	c.mu.Lock()
	c.nonce = n
	c.mu.Unlock()
}

func (c *Client) do(method, path string, query url.Values, body any, out any) error {
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mercantile-ssh/1.0")
	// Note: we intentionally do NOT send the Cart-Token header on requests.
	// On mercantile.wordpress.org, the presence of Cart-Token triggers a
	// server-side 500. The cart is associated with this connection via the
	// session cookie jar instead. We still capture Cart-Token from the
	// *response* because the QR checkout handoff URL needs it.
	if n := c.Nonce(); n != "" {
		req.Header.Set("Nonce", n)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if tok := resp.Header.Get("Cart-Token"); tok != "" {
		c.setCartToken(tok)
	}
	if n := resp.Header.Get("Nonce"); n != "" {
		c.setNonce(n)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(raw, &apiErr)
		if apiErr.Message != "" {
			return fmt.Errorf("store api %d: %s", resp.StatusCode, stripTags(apiErr.Message))
		}
		return fmt.Errorf("store api %d", resp.StatusCode)
	}

	if out == nil {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

// ProductsParams covers the filters we use in the TUI. Category accepts
// either the category ID (int, as string) or slug — the Store API treats
// both the same.
type ProductsParams struct {
	Category string
	Search   string
	Page     int
	PerPage  int
	OrderBy  string
	Order    string
}

func (c *Client) ListProducts(p ProductsParams) ([]Product, error) {
	q := url.Values{}
	if p.Category != "" {
		q.Set("category", p.Category)
	}
	if p.Search != "" {
		q.Set("search", p.Search)
	}
	if p.Page > 0 {
		q.Set("page", strconv.Itoa(p.Page))
	}
	if p.PerPage > 0 {
		q.Set("per_page", strconv.Itoa(p.PerPage))
	} else {
		q.Set("per_page", "30")
	}
	if p.OrderBy != "" {
		q.Set("orderby", p.OrderBy)
	}
	if p.Order != "" {
		q.Set("order", p.Order)
	}
	var out []Product
	if err := c.do("GET", "/products", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetProduct(id int) (*Product, error) {
	var out Product
	if err := c.do("GET", "/products/"+strconv.Itoa(id), nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListCategories() ([]Category, error) {
	q := url.Values{}
	q.Set("per_page", "100")
	q.Set("hide_empty", "true")
	var out []Category
	if err := c.do("GET", "/products/categories", q, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetCart() (*Cart, error) {
	var out Cart
	if err := c.do("GET", "/cart", nil, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type AddItemRequest struct {
	ID        int                  `json:"id"`
	Quantity  int                  `json:"quantity"`
	Variation []VariationAttribute `json:"variation,omitempty"`
}

func (c *Client) AddItem(req AddItemRequest) (*Cart, error) {
	var out Cart
	if err := c.do("POST", "/cart/add-item", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateItem(key string, qty int) (*Cart, error) {
	body := map[string]any{"key": key, "quantity": qty}
	var out Cart
	if err := c.do("POST", "/cart/update-item", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) RemoveItem(key string) (*Cart, error) {
	body := map[string]any{"key": key}
	var out Cart
	if err := c.do("POST", "/cart/remove-item", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type UpdateCustomerRequest struct {
	ShippingAddress *Address `json:"shipping_address,omitempty"`
	BillingAddress  *Address `json:"billing_address,omitempty"`
}

func (c *Client) UpdateCustomer(req UpdateCustomerRequest) (*Cart, error) {
	var out Cart
	if err := c.do("POST", "/cart/update-customer", nil, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SelectShippingRate(packageID any, rateID string) (*Cart, error) {
	body := map[string]any{"package_id": packageID, "rate_id": rateID}
	var out Cart
	if err := c.do("POST", "/cart/select-shipping-rate", nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// stripTags does a crude strip of html tags from API error messages.
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
