// Package screen セーブ/ロード画面のUI実装
// PyRogue準拠のシンプルなセーブ/ロードシステム
package screen

import (
	"fmt"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/save"
)

// SaveLoadMode represents the current mode of the save/load screen
type SaveLoadMode int

const (
	ModeSave SaveLoadMode = iota
	ModeLoad
)

// SaveLoadScreen handles the save/load interface (PyRogue style)
type SaveLoadScreen struct {
	width    int
	height   int
	mode     SaveLoadMode
	selected int

	// Save system integration
	saveManager     *save.SaveManager
	saveIntegration *save.SaveGameIntegration
	onLoad          func(*actor.Player, *dungeon.DungeonManager)

	// UI state
	message      string
	messageColor gruid.Color

	// Save file information
	saveInfo string

	// Colors
	colorNormal    gruid.Color
	colorSelected  gruid.Color
	colorHighlight gruid.Color
	colorError     gruid.Color
	colorSuccess   gruid.Color
	colorEmpty     gruid.Color

	// Menu options
	menuOptions []string
}

// NewSaveLoadScreen creates a new save/load screen
func NewSaveLoadScreen(width, height int, saveManager *save.SaveManager) *SaveLoadScreen {
	return &SaveLoadScreen{
		width:          width,
		height:         height,
		mode:           ModeSave,
		selected:       0,
		saveManager:    saveManager,
		message:        "",
		menuOptions:    []string{"Save Game", "Load Game", "Back"},
		colorNormal:    gruid.Color(0xFFFFFF), // White
		colorSelected:  gruid.Color(0xFFFF00), // Yellow
		colorHighlight: gruid.Color(0x00FF00), // Green
		colorError:     gruid.Color(0xFF0000), // Red
		colorSuccess:   gruid.Color(0x00FF00), // Green
		colorEmpty:     gruid.Color(0x808080), // Gray
	}
}

// NewSaveLoadScreenWithIntegration creates a save/load screen connected to a game state.
func NewSaveLoadScreenWithIntegration(width, height int, integration *save.SaveGameIntegration) *SaveLoadScreen {
	if integration == nil {
		return NewSaveLoadScreen(width, height, save.NewSaveManager())
	}

	screen := NewSaveLoadScreen(width, height, integration.GetSaveManager())
	screen.saveIntegration = integration
	return screen
}

// SetOnLoad registers the callback used to apply loaded state to the engine.
func (s *SaveLoadScreen) SetOnLoad(callback func(*actor.Player, *dungeon.DungeonManager)) {
	s.onLoad = callback
}

// SetMode sets the screen mode
func (s *SaveLoadScreen) SetMode(mode SaveLoadMode) {
	s.mode = mode
	s.selected = 0
	s.message = ""
	s.updateSaveInfo()
}

// HandleInput handles input events
func (s *SaveLoadScreen) HandleInput(msg gruid.Msg) state.GameState {
	if keyMsg, ok := msg.(gruid.MsgKeyDown); ok {
		return s.handleKeyDown(string(keyMsg.Key))
	}
	return state.StateGame
}

// handleKeyDown handles key press events
func (s *SaveLoadScreen) handleKeyDown(key string) state.GameState {
	// Handle main navigation
	switch key {
	case "Up", "k":
		if s.selected > 0 {
			s.selected--
		} else {
			s.selected = len(s.menuOptions) - 1 // Wrap around
		}

	case "Down", "j":
		if s.selected < len(s.menuOptions)-1 {
			s.selected++
		} else {
			s.selected = 0 // Wrap around
		}

	case "Enter", "Space":
		return s.handleSelection()

	case "Escape", "q":
		return state.StateGame

	case "s", "S":
		return s.performSave()

	case "l", "L":
		return s.performLoad()

	case "h", "H", "?":
		s.showHelp()

	// Quick save/load
	case "F5":
		return s.performSave()

	case "F9":
		return s.performLoad()
	}

	return state.StateGame
}

// handleSelection handles menu selection
func (s *SaveLoadScreen) handleSelection() state.GameState {
	switch s.selected {
	case 0: // Save Game
		return s.performSave()
	case 1: // Load Game
		return s.performLoad()
	case 2: // Back
		return state.StateGame
	default:
		return state.StateGame
	}
}

// performSave performs save operation
func (s *SaveLoadScreen) performSave() state.GameState {
	if s.saveIntegration == nil {
		s.setMessage("Save integration is not configured", s.colorError)
		return state.StateGame
	}
	result := command.Execute(&command.Context{Save: s.saveIntegration}, command.Command{Type: command.CmdSave})
	if result.Error {
		s.setMessage(result.Message, s.colorError)
		return state.StateGame
	}
	s.setMessage(result.Message, s.colorSuccess)
	return state.StateGame
}

// performLoad performs load operation
func (s *SaveLoadScreen) performLoad() state.GameState {
	if s.saveIntegration == nil || s.onLoad == nil {
		s.setMessage("Load integration is not configured", s.colorError)
		return state.StateGame
	}
	result := command.Execute(&command.Context{Save: s.saveIntegration}, command.Command{Type: command.CmdLoad})
	if result.Error {
		s.setMessage(result.Message, s.colorError)
		return state.StateGame
	}

	s.onLoad(result.Player, result.Dungeon)
	s.setMessage(result.Message, s.colorSuccess)
	return state.StateGame
}

// updateSaveInfo updates save file information
func (s *SaveLoadScreen) updateSaveInfo() {
	if s.saveManager.FileExists() {
		info, err := s.saveManager.GetSaveInfo()
		if err != nil {
			s.saveInfo = "Error reading save file"
		} else {
			s.saveInfo = info
		}
	} else {
		s.saveInfo = "No save file"
	}
}

// setMessage sets a message with color
func (s *SaveLoadScreen) setMessage(message string, color gruid.Color) {
	s.message = message
	s.messageColor = color
}

// showHelp shows help message
func (s *SaveLoadScreen) showHelp() {
	help := "Save/Load Help: ↑↓:Navigate Enter:Select s:Save l:Load F5:Save F9:Load Esc:Back"
	s.setMessage(help, s.colorHighlight)
}

// Draw draws the save/load screen
func (s *SaveLoadScreen) Draw(grid *gruid.Grid) {
	// Clear screen
	grid.Fill(gruid.Cell{Rune: ' '})

	// Update save info
	s.updateSaveInfo()

	// Draw title
	title := "=== SAVE/LOAD GAME ==="
	s.drawCenteredText(grid, 2, title, s.colorHighlight)

	// Draw instructions
	instructions := "Select an option"
	s.drawCenteredText(grid, 4, instructions, s.colorNormal)

	// Draw menu options
	s.drawMenuOptions(grid)

	// Draw save file info
	s.drawSaveFileInfo(grid)

	// Draw controls
	s.drawControls(grid)

	// Draw message
	if s.message != "" {
		s.drawCenteredText(grid, s.height-3, s.message, s.messageColor)
	}
}

// drawMenuOptions draws the menu options
func (s *SaveLoadScreen) drawMenuOptions(grid *gruid.Grid) {
	startY := 8

	for i, option := range s.menuOptions {
		y := startY + i*2

		// Determine colors
		var textColor gruid.Color
		if i == s.selected {
			textColor = s.colorSelected
		} else {
			textColor = s.colorNormal
		}

		// Draw selection indicator
		if i == s.selected {
			s.drawText(grid, 20, y, ">", s.colorSelected)
		}

		// Draw option text
		s.drawText(grid, 22, y, option, textColor)
	}
}

// drawSaveFileInfo draws save file information
func (s *SaveLoadScreen) drawSaveFileInfo(grid *gruid.Grid) {
	infoY := 16

	// Draw save file section header
	s.drawText(grid, 10, infoY, "Save File Status:", s.colorHighlight)

	// Draw save file info
	if s.saveManager.FileExists() {
		s.drawText(grid, 10, infoY+2, s.saveInfo, s.colorNormal)
	} else {
		s.drawText(grid, 10, infoY+2, "No save file found", s.colorEmpty)
	}
}

// drawControls draws control instructions
func (s *SaveLoadScreen) drawControls(grid *gruid.Grid) {
	controlsY := s.height - 8

	controls := []string{
		"↑↓: Navigate",
		"Enter: Select",
		"s: Save",
		"l: Load",
		"F5: Save",
		"F9: Load",
		"Esc: Back",
	}

	// Draw controls in two columns
	for i, control := range controls {
		x := 5 + (i%2)*30
		y := controlsY + i/2
		s.drawText(grid, x, y, control, s.colorNormal)
	}
}

// drawText draws text at the specified position
func (s *SaveLoadScreen) drawText(grid *gruid.Grid, x, y int, text string, color gruid.Color) {
	for i, r := range text {
		if x+i < s.width && y < s.height {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: gruid.Style{Fg: color}})
		}
	}
}

// drawCenteredText draws centered text
func (s *SaveLoadScreen) drawCenteredText(grid *gruid.Grid, y int, text string, color gruid.Color) {
	x := (s.width - len(text)) / 2
	if x < 0 {
		x = 0
	}
	s.drawText(grid, x, y, text, color)
}

// GetSaveManager returns the save manager
func (s *SaveLoadScreen) GetSaveManager() *save.SaveManager {
	return s.saveManager
}

// SetSaveManager sets the save manager
func (s *SaveLoadScreen) SetSaveManager(saveManager *save.SaveManager) {
	s.saveManager = saveManager
}

// GetMode returns the current mode
func (s *SaveLoadScreen) GetMode() SaveLoadMode {
	return s.mode
}

// GetSelectedOption returns the currently selected option
func (s *SaveLoadScreen) GetSelectedOption() int {
	return s.selected
}

// SetSelectedOption sets the selected option
func (s *SaveLoadScreen) SetSelectedOption(option int) {
	if option >= 0 && option < len(s.menuOptions) {
		s.selected = option
	}
}

// GetMessage returns the current message
func (s *SaveLoadScreen) GetMessage() string {
	return s.message
}

// ClearMessage clears the current message
func (s *SaveLoadScreen) ClearMessage() {
	s.message = ""
}

// IsConfirmingDelete returns whether we're confirming a delete (deprecated in PyRogue style)
func (s *SaveLoadScreen) IsConfirmingDelete() bool {
	return false
}

// GetStatus returns the current screen status
func (s *SaveLoadScreen) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"mode":            s.mode,
		"selected_option": s.selected,
		"message":         s.message,
		"has_save_file":   s.saveManager.FileExists(),
		"save_info":       s.saveInfo,
		"menu_options":    s.menuOptions,
	}
}

// RefreshSaveInfo refreshes the save file information
func (s *SaveLoadScreen) RefreshSaveInfo() {
	s.updateSaveInfo()
}

// Validate validates the screen state
func (s *SaveLoadScreen) Validate() error {
	if s.saveManager == nil {
		return fmt.Errorf("save manager not set")
	}

	if s.selected < 0 || s.selected >= len(s.menuOptions) {
		return fmt.Errorf("invalid selected option: %d", s.selected)
	}

	return nil
}

// GetModeString returns the mode as a string
func (s *SaveLoadScreen) GetModeString() string {
	switch s.mode {
	case ModeSave:
		return "Save"
	case ModeLoad:
		return "Load"
	default:
		return "Unknown"
	}
}

// CanPerformAction checks if the selected action can be performed
func (s *SaveLoadScreen) CanPerformAction() bool {
	switch s.selected {
	case 0: // Save Game
		return true // Can always save
	case 1: // Load Game
		return s.saveManager.FileExists()
	case 2: // Back
		return true
	default:
		return false
	}
}

// GetActionDescription returns a description of the action that will be performed
func (s *SaveLoadScreen) GetActionDescription() string {
	switch s.selected {
	case 0: // Save Game
		if s.saveManager.FileExists() {
			return "Overwrite save file"
		}
		return "Save game"
	case 1: // Load Game
		if s.saveManager.FileExists() {
			return "Load game"
		}
		return "Cannot load - no save file"
	case 2: // Back
		return "Back to game"
	default:
		return "Unknown action"
	}
}

// GetAvailableActions returns available actions for the current state
func (s *SaveLoadScreen) GetAvailableActions() []string {
	actions := []string{"Navigate", "Select", "Back"}
	actions = append(actions, "Save", "Load", "Help")
	return actions
}
