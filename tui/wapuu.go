package tui

import (
	_ "embed"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// wapuuASCII is the WordPress mascot Wapuu, drawn in `█` block characters.
// Embedded from wapuu.txt so the art stays as a viewable text file rather
// than a giant string literal.
//
//go:embed wapuu.txt
var wapuuASCII string

// wapuuRainbow cycles through the spectrum as we walk down Wapuu's rows.
// 7 colors keep adjacent rows distinct without looking computer-random.
var wapuuRainbow = []string{
	"#ef4444", // red
	"#fb923c", // orange
	"#facc15", // yellow
	"#22c55e", // green
	"#3b82f6", // blue
	"#8b5cf6", // indigo
	"#ec4899", // pink
}

// renderWapuu returns the Wapuu art compressed to half height using
// Unicode half-block characters (▀ ▄ █), with each output row tinted to
// the next color in the rainbow palette. Compression is lossless: two
// source rows become one cell row, encoded by which halves are filled.
//
//	source     → cell
//	█ over █   → █  (both halves)
//	█ over ·   → ▀  (top half only)
//	· over █   → ▄  (bottom half only)
//	· over ·   →    (empty)
func renderWapuu() string {
	lines := strings.Split(wapuuASCII, "\n")

	// First pass: compress to half height (half-blocks) and half width
	// (every other column). Builds raw rune rows without styling.
	var rawRows [][]rune
	for i := 0; i < len(lines); i += 2 {
		top := []rune(lines[i])
		var bot []rune
		if i+1 < len(lines) {
			bot = []rune(lines[i+1])
		}
		cols := len(top)
		if len(bot) > cols {
			cols = len(bot)
		}
		row := make([]rune, 0, cols/2+1)
		for c := 0; c < cols; c += 2 {
			topOn := c < len(top) && top[c] == '█'
			botOn := c < len(bot) && bot[c] == '█'
			switch {
			case topOn && botOn:
				row = append(row, '█')
			case topOn:
				row = append(row, '▀')
			case botOn:
				row = append(row, '▄')
			default:
				row = append(row, ' ')
			}
		}
		// Trim trailing spaces — they'd otherwise inflate the row width
		// and skew the eventual centering.
		for len(row) > 0 && row[len(row)-1] == ' ' {
			row = row[:len(row)-1]
		}
		rawRows = append(rawRows, row)
	}

	// Find the shortest leading-whitespace run across all non-empty rows.
	// Stripping that common prefix from every row removes the source's
	// historical left padding without breaking the art's internal shape.
	minLead := -1
	for _, row := range rawRows {
		if len(row) == 0 {
			continue
		}
		lead := 0
		for lead < len(row) && row[lead] == ' ' {
			lead++
		}
		if minLead == -1 || lead < minLead {
			minLead = lead
		}
	}
	if minLead < 0 {
		minLead = 0
	}
	for i := range rawRows {
		if len(rawRows[i]) > minLead {
			rawRows[i] = rawRows[i][minLead:]
		} else {
			rawRows[i] = nil
		}
	}

	// Find max row width so we can pad every row to the same length —
	// uniform width is what lets lipgloss centering treat the art as a
	// single block instead of scattering rows independently.
	maxLen := 0
	for _, row := range rawRows {
		if len(row) > maxLen {
			maxLen = len(row)
		}
	}

	// Time-based rainbow offset — shifts every 500ms so the bands appear
	// to cascade down through Wapuu. UnixMilli/500 gives an integer that
	// advances 2x per second, which we add to each row's index when
	// picking a palette color.
	offset := int(time.Now().UnixMilli() / 500)

	// Second pass: pad to maxLen, colorize, join.
	var out strings.Builder
	for rowIdx, row := range rawRows {
		for len(row) < maxLen {
			row = append(row, ' ')
		}
		color := lipgloss.Color(wapuuRainbow[(rowIdx+offset)%len(wapuuRainbow)])
		out.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(row)))
		out.WriteByte('\n')
	}
	return strings.TrimRight(out.String(), "\n")
}
