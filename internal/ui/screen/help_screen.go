package screen

import (
	"fmt"
	"sort"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/state"
)

// HelpScreen displays game commands and controls
type HelpScreen struct {
	width, height int
	parser        *command.Parser
	grid          gruid.Grid
}

// NewHelpScreen creates a new help screen
func NewHelpScreen(width, height int) *HelpScreen {
	return &HelpScreen{
		width:  width,
		height: height,
		parser: command.NewParser(),
		grid:   gruid.NewGrid(width, height),
	}
}

// Draw implements the screen interface
func (s *HelpScreen) Draw(dst *gruid.Grid) {
	// Clear the grid
	s.grid.Fill(gruid.Cell{Rune: ' '})

	// Draw title
	title := "GoRogue - Command Help"
	titleX := (s.width - len(title)) / 2
	s.drawString(titleX, 2, title, 0xFFFF00, 0x000000) // Yellow on black

	// Draw subtitle
	subtitle := "Press any key to return to game"
	subtitleX := (s.width - len(subtitle)) / 2
	s.drawString(subtitleX, 4, subtitle, 0x00FFFF, 0x000000) // Cyan on black

	// Get key bindings
	bindings := s.parser.GetKeyBindings()

	// Group commands by category
	categories := map[string][]string{
		"Movement":   []string{"h,j,k,l", "y,u,b,n", "Arrow keys"},
		"Actions":    []string{"x", "i", ",", "d", "a", "q", "r", "w", "t", ".", "s", "o", "c"},
		"Navigation": []string{"<", ">"},
		"System":     []string{"Q", "?", "ESC", "Ctrl+W", ":"},
	}

	// Draw commands by category
	y := 6
	for _, category := range []string{"Movement", "Actions", "Navigation", "System"} {
		// Draw category header
		s.drawString(5, y, fmt.Sprintf("=== %s ===", category), 0x00FF00, 0x000000) // Green on black
		y += 2

		// Sort keys in category for consistent display
		keys := categories[category]
		sort.Strings(keys)

		// Draw commands in category
		for _, key := range keys {
			if desc, ok := bindings[key]; ok {
				// Format the key display
				keyDisplay := fmt.Sprintf("%-12s", key)
				s.drawString(7, y, keyDisplay, 0xFFFFFF, 0x000000) // White on black
				s.drawString(20, y, desc, 0x808080, 0x000000)      // Gray on black
				y++
			}
		}
		y++ // Extra space between categories
	}

	// Draw additional info
	y = s.height - 8
	s.drawString(5, y, "Additional Information:", 0xFFFF00, 0x000000) // Yellow on black
	y += 2
	s.drawString(7, y, "• Vi keys (hjkl) and diagonal movement (yubn) follow vi/nethack conventions", 0x808080, 0x000000) // Gray on black
	y++
	s.drawString(7, y, "• Commands are case-sensitive (q=quaff, Q=quit)", 0x808080, 0x000000) // Gray on black
	y++
	s.drawString(7, y, "• Some commands will prompt for additional input (item selection, etc.)", 0x808080, 0x000000) // Gray on black
	y++
	s.drawString(7, y, "• Wizard mode provides debug features for testing", 0x808080, 0x000000) // Gray on black

	// Copy to destination
	dst.Copy(s.grid)
}

// HandleInput handles input events for the help screen
func (s *HelpScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg.(type) {
	case gruid.MsgKeyDown:
		// Any key returns to game
		return state.StateGame
	}
	return state.StateHelp
}

// drawString draws a string at the specified position
func (s *HelpScreen) drawString(x, y int, str string, fg, bg gruid.Color) {
	for i, r := range str {
		if x+i < s.width && y < s.height {
			s.grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{
				Rune: r,
				Style: gruid.Style{
					Fg: fg,
					Bg: bg,
				},
			})
		}
	}
}

// drawBox draws a box border
func (s *HelpScreen) drawBox(x, y, w, h int, fg, bg gruid.Color) {
	// Top and bottom borders
	for i := x; i < x+w; i++ {
		s.grid.Set(gruid.Point{X: i, Y: y}, gruid.Cell{Rune: '─', Style: gruid.Style{Fg: fg, Bg: bg}})
		s.grid.Set(gruid.Point{X: i, Y: y + h - 1}, gruid.Cell{Rune: '─', Style: gruid.Style{Fg: fg, Bg: bg}})
	}

	// Left and right borders
	for i := y; i < y+h; i++ {
		s.grid.Set(gruid.Point{X: x, Y: i}, gruid.Cell{Rune: '│', Style: gruid.Style{Fg: fg, Bg: bg}})
		s.grid.Set(gruid.Point{X: x + w - 1, Y: i}, gruid.Cell{Rune: '│', Style: gruid.Style{Fg: fg, Bg: bg}})
	}

	// Corners
	s.grid.Set(gruid.Point{X: x, Y: y}, gruid.Cell{Rune: '┌', Style: gruid.Style{Fg: fg, Bg: bg}})
	s.grid.Set(gruid.Point{X: x + w - 1, Y: y}, gruid.Cell{Rune: '┐', Style: gruid.Style{Fg: fg, Bg: bg}})
	s.grid.Set(gruid.Point{X: x, Y: y + h - 1}, gruid.Cell{Rune: '└', Style: gruid.Style{Fg: fg, Bg: bg}})
	s.grid.Set(gruid.Point{X: x + w - 1, Y: y + h - 1}, gruid.Cell{Rune: '┘', Style: gruid.Style{Fg: fg, Bg: bg}})
}
