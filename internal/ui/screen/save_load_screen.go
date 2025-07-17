// Package screen セーブ/ロード画面のUI実装
// セーブファイルの管理、選択、削除などの機能を提供
package screen

import (
	"fmt"
	"strconv"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// SaveLoadMode represents the current mode of the save/load screen
type SaveLoadMode int

const (
	ModeSave SaveLoadMode = iota
	ModeLoad
	ModeDelete
)

// SaveLoadScreen handles the save/load interface
type SaveLoadScreen struct {
	width    int
	height   int
	mode     SaveLoadMode
	selected int

	// Save system integration
	saveIntegration *save.SaveGameIntegration

	// UI state
	confirmDelete bool
	selectedSlot  int
	message       string
	messageColor  gruid.Color

	// Save slot information
	saveSlots []string

	// Colors
	colorNormal    gruid.Color
	colorSelected  gruid.Color
	colorHighlight gruid.Color
	colorError     gruid.Color
	colorSuccess   gruid.Color
	colorEmpty     gruid.Color
}

// NewSaveLoadScreen creates a new save/load screen
func NewSaveLoadScreen(width, height int, saveIntegration *save.SaveGameIntegration) *SaveLoadScreen {
	return &SaveLoadScreen{
		width:           width,
		height:          height,
		mode:            ModeSave,
		selected:        0,
		saveIntegration: saveIntegration,
		confirmDelete:   false,
		selectedSlot:    -1,
		message:         "",
		saveSlots:       make([]string, save.MaxSaveSlots),
		colorNormal:     gruid.Color(0xFFFFFF), // White
		colorSelected:   gruid.Color(0xFFFF00), // Yellow
		colorHighlight:  gruid.Color(0x00FF00), // Green
		colorError:      gruid.Color(0xFF0000), // Red
		colorSuccess:    gruid.Color(0x00FF00), // Green
		colorEmpty:      gruid.Color(0x808080), // Gray
	}
}

// SetMode sets the screen mode
func (s *SaveLoadScreen) SetMode(mode SaveLoadMode) {
	s.mode = mode
	s.selected = 0
	s.confirmDelete = false
	s.message = ""
	s.updateSaveSlots()
}

// HandleInput handles input events
func (s *SaveLoadScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		return s.handleKeyDown(string(msg.Key))
	}
	return state.StateGame
}

// handleKeyDown handles key press events
func (s *SaveLoadScreen) handleKeyDown(key string) state.GameState {
	// Handle confirmation dialog
	if s.confirmDelete {
		switch key {
		case "y", "Y":
			s.performDelete()
			s.confirmDelete = false
			return state.StateGame
		case "n", "N", "Escape":
			s.confirmDelete = false
			s.message = ""
			return state.StateGame
		}
		return state.StateGame
	}

	// Handle main navigation
	switch key {
	case "Up", "k":
		if s.selected > 0 {
			s.selected--
		}

	case "Down", "j":
		maxItems := s.getMaxItems()
		if s.selected < maxItems-1 {
			s.selected++
		}

	case "Enter", "Space":
		return s.handleSelection()

	case "Escape", "q":
		return state.StateGame

	case "s", "S":
		s.SetMode(ModeSave)

	case "l", "L":
		s.SetMode(ModeLoad)

	case "d", "D":
		s.SetMode(ModeDelete)

	case "r", "R":
		s.updateSaveSlots()
		s.setMessage("Save slots refreshed", s.colorSuccess)

	case "h", "H", "?":
		s.showHelp()

	// Quick save/load
	case "F5":
		return s.performQuickSave()

	case "F9":
		return s.performQuickLoad()

	// Number keys for direct slot selection
	case "1", "2", "3":
		if num, err := strconv.Atoi(key); err == nil && num >= 1 && num <= save.MaxSaveSlots {
			s.selected = num - 1
			return s.handleSelection()
		}
	}

	return state.StateGame
}

// handleSelection handles selection of a save slot
func (s *SaveLoadScreen) handleSelection() state.GameState {
	if s.selected >= s.getMaxItems() {
		return state.StateGame
	}

	slot := s.selected

	switch s.mode {
	case ModeSave:
		return s.performSave(slot)

	case ModeLoad:
		return s.performLoad(slot)

	case ModeDelete:
		return s.performDeletePrompt(slot)
	}

	return state.StateGame
}

// performSave performs save operation
func (s *SaveLoadScreen) performSave(slot int) state.GameState {
	if err := s.saveIntegration.SaveGame(slot); err != nil {
		s.setMessage(fmt.Sprintf("Save failed: %v", err), s.colorError)
		logger.Error("Save failed", "slot", slot, "error", err)
		return state.StateGame
	}

	s.setMessage(fmt.Sprintf("Game saved to slot %d", slot+1), s.colorSuccess)
	s.updateSaveSlots()
	logger.Info("Game saved via UI", "slot", slot)

	return state.StateGame
}

// performLoad performs load operation
func (s *SaveLoadScreen) performLoad(slot int) state.GameState {
	if !s.saveIntegration.HasSave(slot) {
		s.setMessage("No save file in this slot", s.colorError)
		return state.StateGame
	}

	if err := s.saveIntegration.LoadGame(slot); err != nil {
		s.setMessage(fmt.Sprintf("Load failed: %v", err), s.colorError)
		logger.Error("Load failed", "slot", slot, "error", err)
		return state.StateGame
	}

	s.setMessage(fmt.Sprintf("Game loaded from slot %d", slot+1), s.colorSuccess)
	logger.Info("Game loaded via UI", "slot", slot)

	return state.StateGame
}

// performDeletePrompt shows delete confirmation
func (s *SaveLoadScreen) performDeletePrompt(slot int) state.GameState {
	if !s.saveIntegration.HasSave(slot) {
		s.setMessage("No save file in this slot", s.colorError)
		return state.StateGame
	}

	s.confirmDelete = true
	s.selectedSlot = slot
	s.setMessage(fmt.Sprintf("Delete save slot %d? (y/n)", slot+1), s.colorHighlight)

	return state.StateGame
}

// performDelete performs actual deletion
func (s *SaveLoadScreen) performDelete() {
	if err := s.saveIntegration.DeleteSave(s.selectedSlot); err != nil {
		s.setMessage(fmt.Sprintf("Delete failed: %v", err), s.colorError)
		logger.Error("Delete failed", "slot", s.selectedSlot, "error", err)
		return
	}

	s.setMessage(fmt.Sprintf("Save slot %d deleted", s.selectedSlot+1), s.colorSuccess)
	s.updateSaveSlots()
	logger.Info("Save deleted via UI", "slot", s.selectedSlot)
}

// performQuickSave performs quick save
func (s *SaveLoadScreen) performQuickSave() state.GameState {
	if err := s.saveIntegration.QuickSave(); err != nil {
		s.setMessage(fmt.Sprintf("Quick save failed: %v", err), s.colorError)
		logger.Error("Quick save failed", "error", err)
		return state.StateGame
	}

	s.setMessage("Quick save completed", s.colorSuccess)
	s.updateSaveSlots()
	logger.Info("Quick save completed via UI")

	return state.StateGame
}

// performQuickLoad performs quick load
func (s *SaveLoadScreen) performQuickLoad() state.GameState {
	if !s.saveIntegration.HasSave(0) {
		s.setMessage("No quick save file", s.colorError)
		return state.StateGame
	}

	if err := s.saveIntegration.QuickLoad(); err != nil {
		s.setMessage(fmt.Sprintf("Quick load failed: %v", err), s.colorError)
		logger.Error("Quick load failed", "error", err)
		return state.StateGame
	}

	s.setMessage("Quick load completed", s.colorSuccess)
	logger.Info("Quick load completed via UI")

	return state.StateGame
}

// updateSaveSlots updates save slot information
func (s *SaveLoadScreen) updateSaveSlots() {
	allSaveInfo := s.saveIntegration.GetAllSaveInfo()

	for i := 0; i < save.MaxSaveSlots; i++ {
		if info, exists := allSaveInfo[i]; exists {
			s.saveSlots[i] = info
		} else {
			s.saveSlots[i] = "Empty"
		}
	}
}

// getMaxItems returns the maximum number of items based on current mode
func (s *SaveLoadScreen) getMaxItems() int {
	return save.MaxSaveSlots
}

// setMessage sets a message with color
func (s *SaveLoadScreen) setMessage(message string, color gruid.Color) {
	s.message = message
	s.messageColor = color
}

// showHelp shows help message
func (s *SaveLoadScreen) showHelp() {
	help := "Save/Load Help: ↑↓:Navigate Enter:Select s:Save l:Load d:Delete r:Refresh F5:Quick Save F9:Quick Load"
	s.setMessage(help, s.colorHighlight)
}

// Draw draws the save/load screen
func (s *SaveLoadScreen) Draw(grid *gruid.Grid) {
	// Clear screen
	grid.Fill(gruid.Cell{Rune: ' '})

	// Update save slots
	s.updateSaveSlots()

	// Draw title
	title := s.getTitle()
	s.drawCenteredText(grid, 2, title, s.colorHighlight)

	// Draw instructions
	instructions := s.getInstructions()
	s.drawCenteredText(grid, 4, instructions, s.colorNormal)

	// Draw save slots
	s.drawSaveSlots(grid)

	// Draw auto-save info
	s.drawAutoSaveInfo(grid)

	// Draw controls
	s.drawControls(grid)

	// Draw message
	if s.message != "" {
		s.drawCenteredText(grid, s.height-3, s.message, s.messageColor)
	}

	// Draw confirmation dialog
	if s.confirmDelete {
		s.drawConfirmDialog(grid)
	}
}

// getTitle returns the screen title
func (s *SaveLoadScreen) getTitle() string {
	switch s.mode {
	case ModeSave:
		return "=== SAVE GAME ==="
	case ModeLoad:
		return "=== LOAD GAME ==="
	case ModeDelete:
		return "=== DELETE SAVE ==="
	default:
		return "=== SAVE/LOAD ==="
	}
}

// getInstructions returns mode-specific instructions
func (s *SaveLoadScreen) getInstructions() string {
	switch s.mode {
	case ModeSave:
		return "Select a slot to save your game"
	case ModeLoad:
		return "Select a slot to load your game"
	case ModeDelete:
		return "Select a slot to delete"
	default:
		return "Select an option"
	}
}

// drawSaveSlots draws the save slot list
func (s *SaveLoadScreen) drawSaveSlots(grid *gruid.Grid) {
	startY := 7

	for i := 0; i < save.MaxSaveSlots; i++ {
		y := startY + i*2

		// Determine colors
		var textColor gruid.Color
		var prefixColor gruid.Color

		if i == s.selected {
			textColor = s.colorSelected
			prefixColor = s.colorSelected
		} else {
			textColor = s.colorNormal
			prefixColor = s.colorNormal
		}

		// Check if slot is empty
		isEmpty := s.saveSlots[i] == "Empty"
		if isEmpty {
			textColor = s.colorEmpty
		}

		// Draw slot prefix
		prefix := fmt.Sprintf("%d) ", i+1)
		s.drawText(grid, 10, y, prefix, prefixColor)

		// Draw slot info
		slotInfo := s.saveSlots[i]
		if isEmpty && s.mode == ModeLoad {
			slotInfo = "Empty - Cannot load"
		} else if isEmpty && s.mode == ModeDelete {
			slotInfo = "Empty - Cannot delete"
		}

		s.drawText(grid, 13, y, slotInfo, textColor)

		// Draw selection indicator
		if i == s.selected {
			s.drawText(grid, 8, y, ">", s.colorSelected)
		}
	}
}

// drawAutoSaveInfo draws auto-save information
func (s *SaveLoadScreen) drawAutoSaveInfo(grid *gruid.Grid) {
	autoSaveY := 7 + save.MaxSaveSlots*2 + 2

	// Draw auto-save section header
	s.drawText(grid, 10, autoSaveY, "Auto-Save:", s.colorHighlight)

	// Check if auto-save exists
	if s.saveIntegration.HasAutoSave() {
		info, err := s.saveIntegration.GetSaveInfo(save.AutoSaveSlot)
		if err != nil {
			s.drawText(grid, 10, autoSaveY+1, "Error reading auto-save", s.colorError)
		} else {
			s.drawText(grid, 10, autoSaveY+1, info, s.colorNormal)
		}
	} else {
		s.drawText(grid, 10, autoSaveY+1, "No auto-save available", s.colorEmpty)
	}
}

// drawControls draws control instructions
func (s *SaveLoadScreen) drawControls(grid *gruid.Grid) {
	controlsY := s.height - 8

	controls := []string{
		"↑↓: Navigate",
		"Enter: Select",
		"s: Save Mode",
		"l: Load Mode",
		"d: Delete Mode",
		"r: Refresh",
		"F5: Quick Save",
		"F9: Quick Load",
		"Esc: Back",
	}

	// Draw controls in two columns
	for i, control := range controls {
		x := 5 + (i%2)*30
		y := controlsY + i/2
		s.drawText(grid, x, y, control, s.colorNormal)
	}
}

// drawConfirmDialog draws the confirmation dialog
func (s *SaveLoadScreen) drawConfirmDialog(grid *gruid.Grid) {
	// Draw dialog box
	dialogWidth := 40
	dialogHeight := 6
	dialogX := (s.width - dialogWidth) / 2
	dialogY := (s.height - dialogHeight) / 2

	// Draw border
	for y := dialogY; y < dialogY+dialogHeight; y++ {
		for x := dialogX; x < dialogX+dialogWidth; x++ {
			if y == dialogY || y == dialogY+dialogHeight-1 || x == dialogX || x == dialogX+dialogWidth-1 {
				grid.Set(gruid.Point{X: x, Y: y}, gruid.Cell{Rune: '#', Style: gruid.Style{Fg: s.colorHighlight}})
			} else {
				grid.Set(gruid.Point{X: x, Y: y}, gruid.Cell{Rune: ' '})
			}
		}
	}

	// Draw dialog content
	s.drawCenteredText(grid, dialogY+1, "CONFIRM DELETE", s.colorError)
	s.drawCenteredText(grid, dialogY+2, fmt.Sprintf("Delete save slot %d?", s.selectedSlot+1), s.colorNormal)
	s.drawCenteredText(grid, dialogY+3, "This action cannot be undone!", s.colorError)
	s.drawCenteredText(grid, dialogY+4, "Press 'y' to confirm, 'n' to cancel", s.colorNormal)
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

// GetSaveIntegration returns the save integration
func (s *SaveLoadScreen) GetSaveIntegration() *save.SaveGameIntegration {
	return s.saveIntegration
}

// SetSaveIntegration sets the save integration
func (s *SaveLoadScreen) SetSaveIntegration(saveIntegration *save.SaveGameIntegration) {
	s.saveIntegration = saveIntegration
}

// GetMode returns the current mode
func (s *SaveLoadScreen) GetMode() SaveLoadMode {
	return s.mode
}

// GetSelectedSlot returns the currently selected slot
func (s *SaveLoadScreen) GetSelectedSlot() int {
	return s.selected
}

// SetSelectedSlot sets the selected slot
func (s *SaveLoadScreen) SetSelectedSlot(slot int) {
	if slot >= 0 && slot < save.MaxSaveSlots {
		s.selected = slot
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

// IsConfirmingDelete returns whether we're confirming a delete
func (s *SaveLoadScreen) IsConfirmingDelete() bool {
	return s.confirmDelete
}

// GetStatus returns the current screen status
func (s *SaveLoadScreen) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"mode":              s.mode,
		"selected_slot":     s.selected,
		"confirming_delete": s.confirmDelete,
		"message":           s.message,
		"has_auto_save":     s.saveIntegration.HasAutoSave(),
		"save_slots":        s.saveSlots,
	}
}

// GetSaveSlotInfo returns information about a specific save slot
func (s *SaveLoadScreen) GetSaveSlotInfo(slot int) string {
	if slot >= 0 && slot < len(s.saveSlots) {
		return s.saveSlots[slot]
	}
	return "Invalid slot"
}

// RefreshSaveSlots refreshes the save slot information
func (s *SaveLoadScreen) RefreshSaveSlots() {
	s.updateSaveSlots()
}

// Validate validates the screen state
func (s *SaveLoadScreen) Validate() error {
	if s.saveIntegration == nil {
		return fmt.Errorf("save integration not set")
	}

	if s.selected < 0 || s.selected >= save.MaxSaveSlots {
		return fmt.Errorf("invalid selected slot: %d", s.selected)
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
	case ModeDelete:
		return "Delete"
	default:
		return "Unknown"
	}
}

// CanPerformAction checks if the selected action can be performed
func (s *SaveLoadScreen) CanPerformAction() bool {
	switch s.mode {
	case ModeSave:
		return true // Can always save
	case ModeLoad:
		return s.saveIntegration.HasSave(s.selected)
	case ModeDelete:
		return s.saveIntegration.HasSave(s.selected)
	default:
		return false
	}
}

// GetActionDescription returns a description of the action that will be performed
func (s *SaveLoadScreen) GetActionDescription() string {
	slot := s.selected + 1

	switch s.mode {
	case ModeSave:
		if s.saveIntegration.HasSave(s.selected) {
			return fmt.Sprintf("Overwrite save slot %d", slot)
		}
		return fmt.Sprintf("Save to slot %d", slot)
	case ModeLoad:
		if s.saveIntegration.HasSave(s.selected) {
			return fmt.Sprintf("Load from slot %d", slot)
		}
		return "Cannot load from empty slot"
	case ModeDelete:
		if s.saveIntegration.HasSave(s.selected) {
			return fmt.Sprintf("Delete save slot %d", slot)
		}
		return "Cannot delete empty slot"
	default:
		return "Unknown action"
	}
}

// GetAvailableActions returns available actions for the current state
func (s *SaveLoadScreen) GetAvailableActions() []string {
	actions := []string{"Navigate", "Select", "Back"}

	if s.mode != ModeSave {
		actions = append(actions, "Save Mode")
	}
	if s.mode != ModeLoad {
		actions = append(actions, "Load Mode")
	}
	if s.mode != ModeDelete {
		actions = append(actions, "Delete Mode")
	}

	actions = append(actions, "Refresh", "Quick Save", "Quick Load")

	return actions
}
