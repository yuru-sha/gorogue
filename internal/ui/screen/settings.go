package screen

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/core/state"
)

var optionLabels = [...]string{
	"terse", "flush", "jump", "seefloor", "passgo", "tombstone", "inven", "name", "fruit", "file",
}

func (s *GameScreen) SetOptions(options *config.Options) {
	s.options = *options
	if s.saveIntegration != nil {
		s.saveIntegration.SetOptions(options)
	}
	if s.cliMode != nil {
		s.cliMode.SetOptions(options)
	}
}

func (s *GameScreen) openOptions() state.GameState {
	s.presentation = presentationSettings
	s.optionsCursor = 0
	s.optionsEditing = false
	return state.StateSettings
}

func (s *GameScreen) handleOptionsInput(key gruid.Key) state.GameState {
	if s.optionsEditing {
		s.editOptionValue(key)
		return state.StateSettings
	}
	switch key {
	case gruid.KeyEscape:
		s.presentation = presentationPlay
		return state.StateGame
	case "j", gruid.KeyArrowDown:
		s.optionsCursor = (s.optionsCursor + 1) % len(optionLabels)
	case "k", gruid.KeyArrowUp:
		s.optionsCursor = (s.optionsCursor + len(optionLabels) - 1) % len(optionLabels)
	case gruid.KeySpace, gruid.KeyEnter:
		switch {
		case s.optionsCursor < 6:
			s.toggleOption(s.optionsCursor)
		case s.optionsCursor == 6:
			s.cycleInventoryStyle()
		default:
			s.optionsEditOriginal = s.currentTextOption()
			s.optionsEditOriginalConfig = s.options.NameConfigured
			s.optionsEditing = true
			s.optionsReplaceOnInput = true
		}
	}
	s.SetOptions(&s.options)
	return state.StateSettings
}

func (s *GameScreen) toggleOption(index int) {
	switch index {
	case 0:
		s.options.Terse = !s.options.Terse
	case 1:
		s.options.Flush = !s.options.Flush
	case 2:
		s.options.Jump = !s.options.Jump
	case 3:
		s.options.SeeFloor = !s.options.SeeFloor
	case 4:
		s.options.PassGo = !s.options.PassGo
	case 5:
		s.options.Tombstone = !s.options.Tombstone
	}
}

func (s *GameScreen) cycleInventoryStyle() {
	styles := [...]config.InventoryStyle{config.InventoryOverwrite, config.InventorySlow, config.InventoryClear}
	for i, style := range styles {
		if style == s.options.InventoryStyle {
			s.options.InventoryStyle = styles[(i+1)%len(styles)]
			return
		}
	}
	s.options.InventoryStyle = config.InventoryOverwrite
}

func (s *GameScreen) editOptionValue(key gruid.Key) {
	switch key {
	case gruid.KeyEscape:
		s.setTextOption(s.optionsEditOriginal)
		s.options.NameConfigured = s.optionsEditOriginalConfig
		s.optionsEditing = false
		s.optionsReplaceOnInput = false
	case gruid.KeyEnter:
		s.finishOptionEdit()
	case gruid.KeyBackspace:
		text := []rune(s.currentTextOption())
		if s.optionsReplaceOnInput {
			s.setTextOption("")
			s.optionsReplaceOnInput = false
		} else if len(text) > 0 && (len(text) > 1 || s.optionsCursor == 9) {
			s.setTextOption(string(text[:len(text)-1]))
		}
	default:
		value := string(key)
		if r, size := utf8.DecodeRuneInString(value); size > 0 && size == len(value) && !unicode.IsControl(r) {
			if s.optionsReplaceOnInput {
				s.setTextOption(value)
				s.optionsReplaceOnInput = false
			} else {
				s.setTextOption(s.currentTextOption() + value)
			}
		}
	}
}

func (s *GameScreen) finishOptionEdit() {
	switch s.optionsCursor {
	case 7:
		switch {
		case s.options.Name == "":
			s.options.Name = s.optionsEditOriginal
			s.options.NameConfigured = s.optionsEditOriginalConfig
		case s.options.Name != s.optionsEditOriginal:
			s.options.NameConfigured = true
		default:
			s.options.NameConfigured = s.optionsEditOriginalConfig
		}
	case 8:
		if s.options.Fruit == "" {
			s.options.Fruit = s.optionsEditOriginal
		}
	}
	s.optionsEditing = false
	s.optionsReplaceOnInput = false
	s.SetOptions(&s.options)
}

func (s *GameScreen) currentTextOption() string {
	switch s.optionsCursor {
	case 7:
		return s.options.Name
	case 8:
		return s.options.Fruit
	case 9:
		return s.options.File
	default:
		return ""
	}
}

func (s *GameScreen) setTextOption(value string) {
	switch s.optionsCursor {
	case 7:
		s.options.Name = value
	case 8:
		s.options.Fruit = value
	case 9:
		s.options.File = value
	}
}

func (s *GameScreen) drawOptions(grid *gruid.Grid) {
	rows := []string{
		fmt.Sprintf("terse: %t", s.options.Terse),
		fmt.Sprintf("flush: %t", s.options.Flush),
		fmt.Sprintf("jump: %t", s.options.Jump),
		fmt.Sprintf("seefloor: %t", s.options.SeeFloor),
		fmt.Sprintf("passgo: %t", s.options.PassGo),
		fmt.Sprintf("tombstone: %t", s.options.Tombstone),
		fmt.Sprintf("inven: %s", s.options.InventoryStyle),
		"name: " + s.options.Name,
		"fruit: " + s.options.Fruit,
		"file: " + s.options.File,
	}
	s.drawText(grid, 0, 0, "User options (j/k select, space edit, Esc close)", gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
	for i, row := range rows {
		if i == s.optionsCursor {
			row = "> " + row
		}
		if s.optionsEditing && i == s.optionsCursor {
			row += "_"
		}
		s.drawText(grid, 0, i+2, row, gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
	}
	if s.optionsEditing {
		s.drawText(grid, 0, len(rows)+3, "Type value; Enter applies, Esc cancels.", gruid.Style{Fg: 0xFFFFFF, Bg: 0x000000})
	}
}

func optionsHelpText() []string {
	return []string{
		"O: open user options; Space toggles/cycles, Enter edits text values.",
		"terse, flush, jump, seefloor, passgo and tombstone are Rogue options.",
		"inven accepts overwrite, slow or clear.",
		"name sets saved player identity; fruit stores the preferred fruit name.",
		"file selects the save file path; blank uses the default location.",
		"GUI-native display, movement and death presentation remain in effect.",
	}
}
