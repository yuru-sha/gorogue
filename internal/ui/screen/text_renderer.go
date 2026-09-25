package screen

import (
	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
)

// TextRenderer owns the terminal grid and delegates screen drawing by state.
type TextRenderer struct {
	grid gruid.Grid
}

func NewTextRenderer(width, height int) *TextRenderer {
	return &TextRenderer{grid: gruid.NewGrid(width, height)}
}

func (r *TextRenderer) Draw(states *state.StateManager) gruid.Grid {
	r.grid.Fill(gruid.Cell{Rune: ' '})
	states.Draw(&r.grid)
	return r.grid
}
