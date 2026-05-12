package tui

import (
	"math/rand"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Sparkle particles drift upward through a band of fixed height above
// Wapuu during party mode. Each particle is one cell — a random character
// from sparkleChars — that ages by 1 row per spinner tick (~100ms). Once
// a particle's age reaches sparkleRows it falls off the top and is removed.

const (
	sparkleRows     = 8
	sparkleSpawnMax = 4 // max sparkles spawned per tick (random 0..max-1)
)

// sparkleChars cycles between weights — heavier glyphs are rarer, lighter
// ones more common — to give the field some visual texture.
var sparkleChars = []rune{'.', '·', '·', '*', '✦', '✧'}

// sparkleColor is a pale silver-blue, picked so the particles read as
// "stars" against any terminal bg without competing with Wapuu.
var sparkleColor = lipgloss.Color("#cbd5e1")

type sparkleParticle struct {
	x    int  // column position
	age  int  // 0 = freshly spawned (bottom row), sparkleRows-1 = top row
	char rune
}

// advanceSparkles ages existing particles, retires the ones that aged
// past the top, and maybe spawns 0–(sparkleSpawnMax-1) new ones at the
// bottom row. Call once per spinner tick while party mode is active.
func (m *Model) advanceSparkles() {
	if m.width == 0 {
		return
	}
	kept := m.sparkles[:0]
	for _, s := range m.sparkles {
		s.age++
		if s.age < sparkleRows {
			kept = append(kept, s)
		}
	}
	m.sparkles = kept

	spawn := rand.Intn(sparkleSpawnMax)
	for i := 0; i < spawn; i++ {
		m.sparkles = append(m.sparkles, sparkleParticle{
			x:    rand.Intn(m.width),
			age:  0,
			char: sparkleChars[rand.Intn(len(sparkleChars))],
		})
	}
}

// renderSparkles returns a sparkleRows-tall block of text painting the
// current particle positions. Empty cells are spaces. Newest particles
// sit at the bottom row, oldest at the top — so the eye reads it as
// "stars drifting up."
func (m *Model) renderSparkles() string {
	if m.width == 0 {
		return strings.Repeat("\n", sparkleRows-1)
	}
	grid := make([][]rune, sparkleRows)
	for i := range grid {
		grid[i] = make([]rune, m.width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	for _, s := range m.sparkles {
		row := sparkleRows - 1 - s.age
		if row < 0 || row >= sparkleRows {
			continue
		}
		if s.x < 0 || s.x >= m.width {
			continue
		}
		grid[row][s.x] = s.char
	}
	var b strings.Builder
	for i, row := range grid {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(string(row))
	}
	return lipgloss.NewStyle().Foreground(sparkleColor).Render(b.String())
}
