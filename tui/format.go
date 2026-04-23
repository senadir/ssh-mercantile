package tui

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"

	"mercantile/api"
)

var tagRe = regexp.MustCompile(`<[^>]+>`)
var wsRe = regexp.MustCompile(`\s+`)

func stripHTML(s string) string {
	s = tagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = wsRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// formatPrice turns "4700" with minor_unit=2 into "€47,00"
func formatPrice(minor string, unit int, prefix, suffix, decimalSep, thousandSep string) string {
	if minor == "" {
		return ""
	}
	n, err := strconv.ParseInt(minor, 10, 64)
	if err != nil {
		return minor
	}
	neg := n < 0
	if neg {
		n = -n
	}
	if decimalSep == "" {
		decimalSep = "."
	}
	if thousandSep == "" {
		thousandSep = ","
	}
	var intPart, fracPart string
	if unit == 0 {
		intPart = strconv.FormatInt(n, 10)
	} else {
		s := strconv.FormatInt(n, 10)
		for len(s) <= unit {
			s = "0" + s
		}
		intPart = s[:len(s)-unit]
		fracPart = s[len(s)-unit:]
	}
	// thousands
	if len(intPart) > 3 {
		var b strings.Builder
		offset := len(intPart) % 3
		if offset > 0 {
			b.WriteString(intPart[:offset])
			if len(intPart) > offset {
				b.WriteString(thousandSep)
			}
		}
		for i := offset; i < len(intPart); i += 3 {
			b.WriteString(intPart[i : i+3])
			if i+3 < len(intPart) {
				b.WriteString(thousandSep)
			}
		}
		intPart = b.String()
	}
	out := intPart
	if fracPart != "" {
		out += decimalSep + fracPart
	}
	if neg {
		out = "-" + out
	}
	return prefix + out + suffix
}

func formatPriceFromPrices(amount string, p api.Prices) string {
	return formatPrice(amount, p.CurrencyMinorUnit, p.CurrencyPrefix, p.CurrencySuffix, p.CurrencyDecimalSep, p.CurrencyThousandSep)
}

func formatCartTotal(amount string, t api.CartTotals) string {
	return formatPrice(amount, t.CurrencyMinorUnit, t.CurrencyPrefix, t.CurrencySuffix, ",", ".")
}

func formatShippingRate(r api.ShippingRate) string {
	price := formatPrice(r.Price, r.CurrencyMinorUnit, r.CurrencyPrefix, r.CurrencySuffix, ",", ".")
	name := stripHTML(r.Name)
	if r.DeliveryTime != "" {
		return fmt.Sprintf("%s — %s (%s)", name, price, r.DeliveryTime)
	}
	return fmt.Sprintf("%s — %s", name, price)
}
