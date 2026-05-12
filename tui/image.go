package tui

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/image/draw"
)

// Image render dimensions. Each terminal row encodes 2 vertical pixels via
// the half-block trick, so a 30×15 cell render is effectively 30×30 pixels.
// Small enough to fit any reasonable terminal, big enough to be recognizable.
const (
	imageCols = 30
	imageRows = 15
)

// imageLoadedMsg arrives when an async image fetch completes. productID
// tags it so the View knows whether the result is still relevant (a fast
// click-through could leave a stale image inflight).
type imageLoadedMsg struct {
	productID int
	rendered  string // pre-rendered ANSI string, ready to drop into a view
	err       error
}

// fetchImageCmd downloads, decodes, resizes, and pre-renders a product
// image. Done off the main tea event loop. Cmd returns an imageLoadedMsg
// which the Update handler caches by productID.
func fetchImageCmd(productID int, url string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return imageLoadedMsg{productID: productID, err: err}
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return imageLoadedMsg{productID: productID, err: err}
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return imageLoadedMsg{productID: productID, err: fmt.Errorf("http %d", resp.StatusCode)}
		}
		// Cap body size at 4MB — product images shouldn't approach that.
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if err != nil {
			return imageLoadedMsg{productID: productID, err: err}
		}
		img, _, err := image.Decode(bytes.NewReader(body))
		if err != nil {
			return imageLoadedMsg{productID: productID, err: err}
		}
		return imageLoadedMsg{
			productID: productID,
			rendered:  renderImageHalfBlock(img, imageCols, imageRows),
		}
	}
}

// renderImageHalfBlock resizes the image to cols × (2*rows) pixels, then
// emits one row per two pixel rows. Each cell is `▀` (top half block)
// styled with foreground = top pixel color and background = bottom pixel
// color, so a single cell shows two stacked colored pixels.
func renderImageHalfBlock(img image.Image, cols, rows int) string {
	target := image.NewRGBA(image.Rect(0, 0, cols, 2*rows))
	// NearestNeighbor keeps each output pixel as a single sampled source
	// pixel — no blending, no soft edges. Gives the image a deliberate
	// pixel-art look that leans into the medium's limits.
	draw.NearestNeighbor.Scale(target, target.Bounds(), img, img.Bounds(), draw.Src, nil)

	var b strings.Builder
	for y := 0; y < 2*rows; y += 2 {
		for x := 0; x < cols; x++ {
			tr, tg, tb, _ := target.At(x, y).RGBA()
			br, bg, bb, _ := target.At(x, y+1).RGBA()
			cell := lipgloss.NewStyle().
				Foreground(lipgloss.Color(rgbHex(tr, tg, tb))).
				Background(lipgloss.Color(rgbHex(br, bg, bb))).
				Render("▀")
			b.WriteString(cell)
		}
		if y < 2*rows-2 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// rgbHex converts 16-bit RGB channel values (as returned by image.Color's
// RGBA method) to a CSS-style #rrggbb string for lipgloss.
func rgbHex(r, g, b uint32) string {
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}
