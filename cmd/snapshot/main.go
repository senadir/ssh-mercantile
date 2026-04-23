// snapshot: render each TUI view statically to stdout to sanity-check
// the layout without needing a pty. Not part of the shipping binary.
package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"mercantile/api"
	"mercantile/tui"
)

func main() {
	c := api.New("")
	m := tui.NewModel(c, "https://mercantile.wordpress.org")

	// size the model
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = mm.(*tui.Model)

	// run Init (fires loads) then wait for results. Since we aren't running
	// the bubbletea program loop, we call the commands inline.
	if init := m.Init(); init != nil {
		go func() {
			msgs := []tea.Msg{init()}
			for _, msg := range msgs {
				mm2, _ := m.Update(msg)
				m = mm2.(*tui.Model)
			}
		}()
	}
	time.Sleep(3 * time.Second)

	fmt.Println("=== WELCOME ===")
	fmt.Println(m.View())
	fmt.Println()

	// simulate pressing 'a' -> apparel
	mm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = mm.(*tui.Model)
	if cmd != nil {
		go func(c tea.Cmd) {
			if msg := c(); msg != nil {
				mm2, _ := m.Update(msg)
				m = mm2.(*tui.Model)
			}
		}(cmd)
	}
	time.Sleep(3 * time.Second)

	fmt.Println("=== APPAREL BROWSE ===")
	fmt.Println(m.View())
}
