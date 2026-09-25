package screen

import (
	"fmt"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// HandleInput handles input events
func (s *GameScreen) HandleInput(msg gruid.Msg) state.GameState {
	keyMsg, ok := msg.(gruid.MsgKeyDown)
	if !ok {
		return state.StateGame
	}
	nextState := s.handleInputKey(keyMsg.Key)
	if s.player != nil && !s.player.IsAlive() {
		return state.StateGameOver
	}
	return nextState
}

func (s *GameScreen) handleInputKey(key gruid.Key) state.GameState {
	switch s.inputMode {
	case ModeEquip:
		return s.handleEquipInput(key)
	case ModeUnequip:
		return s.handleUnequipInput(key)
	case ModeDrop:
		return s.handleDropInput(key)
	case ModeUse:
		return s.handleUseInput(key)
	case ModeQuaff:
		return s.handleQuaffInput(key)
	case ModeRead:
		return s.handleReadInput(key)
	case ModeReadTarget:
		return s.handleReadTargetInput(key)
	case ModeEat:
		return s.handleEatInput(key)
	case ModeThrow, ModeZap:
		return s.handleDirectionalItemInput(key)
	case ModeCall:
		return s.handleCallInput(key)
	case ModeCLI:
		return s.handleCLIInput(key)
	case ModeDirection:
		return s.handleDirectionInput(key)
	default:
		return s.handleNormalInput(key)
	}
}

// handleNormalInput handles input in normal mode
//
//nolint:gocyclo // Normal-mode input is the exhaustive mapping of gameplay commands.
func (s *GameScreen) handleNormalInput(key gruid.Key) state.GameState {
	// Parse the key into a command
	cmd := s.cmdParser.Parse(key)

	switch cmd.Type {
	// Movement commands
	case command.CmdMoveWest, command.CmdMoveEast, command.CmdMoveNorth, command.CmdMoveSouth,
		command.CmdMoveNorthWest, command.CmdMoveNorthEast, command.CmdMoveSouthWest, command.CmdMoveSouthEast,
		command.CmdRun:
		s.addCommandResult(s.executeCommand(cmd))
	case command.CmdLook:
		s.handleLook()
	case command.CmdInventory:
		s.showInventory()
	case command.CmdPickUp:
		s.handlePickUp()
	case command.CmdDrop:
		s.enterDropMode()
	case command.CmdUse:
		s.enterUseMode()
	case command.CmdQuaff:
		s.enterQuaffMode()
	case command.CmdRead:
		s.enterReadMode()
	case command.CmdWield, command.CmdWear, command.CmdRingOn, command.CmdEquip:
		s.enterEquipModeFor(cmd.Type)
	case command.CmdTakeOff:
		s.addCommandResult(s.executeCommand(cmd, "armor"))
	case command.CmdRingOff, command.CmdUnequip:
		s.enterUnequipModeFor(cmd.Type)
	case command.CmdEat:
		s.enterEatMode()
	case command.CmdWait:
		s.handleWait()
	case command.CmdSearch:
		s.handleSearch()
	case command.CmdFindTrap:
		s.handleFindTrap()
	case command.CmdThrow:
		s.handleThrow()
	case command.CmdZap:
		s.handleZap()
	case command.CmdRepeat:
		s.handleRepeat()
	case command.CmdCharacter, command.CmdDiscover:
		s.addCommandResult(s.executeCommand(cmd))
	case command.CmdCall:
		s.handleCall()
	case command.CmdOpen:
		s.handleOpenDoor()
	case command.CmdClose:
		s.handleCloseDoor()
	case command.CmdFight:
		s.handleFight()
	// Stair commands
	case command.CmdGoUpstairs:
		return s.handleStairs(true)
	case command.CmdGoDownstairs:
		return s.handleStairs(false)

	// System commands
	case command.CmdQuit:
		logger.Info("Quit requested")
		return state.StateMenu
	case command.CmdHelp:
		return state.StateHelp
	case command.CmdSymbol:
		return state.StateSymbol
	case command.CmdEscape:
		s.directionCallback = nil
		s.AddMessage("Canceled.")
	case command.CmdWizard:
		s.wizardMode.Toggle()
		status := "OFF"
		if s.wizardMode.IsActive {
			status = "ON"
		}
		s.AddMessage(fmt.Sprintf("ウィザードモード: %s", status))
	case command.CmdCLI:
		s.enterCLIMode()

	default:
		// Check if it's a wizard command
		if s.wizardMode.IsActive && len(string(key)) == 1 {
			result := s.wizardMode.ExecuteCommand(rune(string(key)[0]))
			if result != "" {
				s.AddMessage(result)
			}
		}
	}
	return state.StateGame
}

func (s *GameScreen) handleCall() {
	if s.player == nil || s.player.Inventory == nil || s.player.IdentifyMgr == nil {
		s.AddMessage("Inventory is unavailable.")
		return
	}
	s.callItemLetter = 0
	s.callNameBuffer = ""
	s.inputMode = ModeCall
	s.AddMessage("Call which unidentified item? (a-z)")
}

func (s *GameScreen) handleCallInput(key gruid.Key) state.GameState {
	if key == gruid.KeyEscape {
		s.callItemLetter = 0
		s.callNameBuffer = ""
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	}
	if s.callItemLetter == 0 {
		return s.selectCallItem(key)
	}
	return s.editCallName(key)
}

func (s *GameScreen) selectCallItem(key gruid.Key) state.GameState {
	letter := []rune(string(key))
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		s.AddMessage("Select an item letter from a-z.")
		return state.StateGame
	}
	_, gameItem := s.player.Inventory.GetItemByLetter(letter[0])
	if gameItem == nil {
		s.AddMessage("No item in that slot.")
		return state.StateGame
	}
	if s.player.IdentifyMgr.IsIdentified(gameItem) {
		s.AddMessage("That item is already identified.")
		return state.StateGame
	}
	s.callItemLetter = letter[0]
	s.AddMessage("Call it what? (Enter to accept, Esc to cancel)")
	return state.StateGame
}

func (s *GameScreen) editCallName(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEnter:
		name, letter := s.callNameBuffer, s.callItemLetter
		s.callItemLetter = 0
		s.callNameBuffer = ""
		s.inputMode = ModeNormal
		if name == "" {
			s.AddMessage("No call name entered.")
			return state.StateGame
		}
		s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdCall}, string(letter), name))
	case gruid.KeyBackspace:
		if s.callNameBuffer != "" {
			s.callNameBuffer = s.callNameBuffer[:len(s.callNameBuffer)-1]
		}
	default:
		input := string(key)
		if len(input) == 1 && input[0] >= 32 && input[0] <= 126 && len(s.callNameBuffer) < 50 {
			s.callNameBuffer += input
			s.AddMessage("Call name: " + s.callNameBuffer)
		}
	}
	return state.StateGame
}

// handleStairs handles stair movement
func (s *GameScreen) handleStairs(goUp bool) state.GameState {
	commandType := command.CmdGoDownstairs
	if goUp {
		commandType = command.CmdGoUpstairs
	}
	result := s.executeCommand(command.Command{Type: commandType})
	s.addCommandResult(result)
	if result.Victory {
		return state.StateVictory
	}
	return state.StateGame
}
func (s *GameScreen) handleEquipInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			selection := int(string(key)[0] - 'a')
			if selection >= len(s.equippableItems) {
				s.AddMessage("Invalid selection.")
			} else {
				selected := s.equippableItems[selection]
				for index, item := range s.player.Inventory.Items {
					if item == selected {
						s.addCommandResult(s.executeCommand(s.equipActionCommand(), string(rune('a'+index))))
						break
					}
				}
			}
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

func (s *GameScreen) equipActionCommand() command.Command {
	action := s.equipAction
	if action == 0 {
		action = command.CmdEquip
	}
	return command.Command{Type: action}
}

func (s *GameScreen) handleUnequipInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	case "w":
		s.unequipSlot("weapon", "weapon")
	case "a":
		s.unequipSlot("armor", "armor")
	case "l":
		s.unequipSlot("ring-left", "left ring")
	case "r":
		s.unequipSlot("ring-right", "right ring")
	default:
		s.AddMessage("Invalid equipment selection.")
	}
	s.inputMode = ModeNormal
	return state.StateGame
}

func (s *GameScreen) unequipSlot(slot, displaySlot string) {
	action := s.unequipAction
	switch action {
	case command.CmdTakeOff:
		if slot != "armor" {
			s.AddMessage("You can only take off armor with T.")
			return
		}
	case command.CmdRingOff:
		if slot != "ring-left" && slot != "ring-right" {
			s.AddMessage("Select a ring hand.")
			return
		}
	default:
		action = command.CmdUnequip
	}
	result := s.executeCommand(command.Command{Type: action}, slot)
	if result.Message == "No item equipped in "+slot+" slot." && displaySlot != slot {
		result.Message = "You have no " + displaySlot + " equipped."
	}
	s.addCommandResult(result)
}

// handleDropInput handles input in drop mode
func (s *GameScreen) handleDropInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdDrop}, string(key)))
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleQuaffInput handles input in quaff mode
func (s *GameScreen) handleQuaffInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdQuaff)
}

// handleReadInput handles scroll and, when required, identify-target selection.
func (s *GameScreen) handleReadInput(key gruid.Key) state.GameState {
	if key == gruid.KeyEscape {
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	}
	if len(string(key)) != 1 || string(key)[0] < 'a' || string(key)[0] > 'z' {
		return state.StateGame
	}

	scrollLetter := string(key)
	result := s.executeCommand(command.Command{Type: command.CmdRead}, scrollLetter)
	if !result.NeedsItemSelection {
		s.addCommandResult(result)
		s.inputMode = ModeNormal
		return state.StateGame
	}
	s.AddMessage(result.Message)
	s.pendingReadScroll = rune(string(key)[0])
	s.inputMode = ModeReadTarget
	s.showIdentifyTargets(result.TargetTypes)
	return state.StateGame
}

func (s *GameScreen) handleReadTargetInput(key gruid.Key) state.GameState {
	if key == gruid.KeyEscape {
		s.pendingReadScroll = 0
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	}
	if len(string(key)) != 1 || string(key)[0] < 'a' || string(key)[0] > 'z' {
		return state.StateGame
	}
	result := s.executeCommand(
		command.Command{Type: command.CmdRead},
		string(s.pendingReadScroll), string(key))
	s.pendingReadScroll = 0
	s.inputMode = ModeNormal
	s.addCommandResult(result)
	return state.StateGame
}

// handleEatInput handles food selection.
func (s *GameScreen) handleEatInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdEat)
}

func (s *GameScreen) handleUseInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdUse)
}

func (s *GameScreen) handleItemSelection(key gruid.Key, commandType command.Type) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			s.addCommandResult(s.executeCommand(command.Command{Type: commandType}, string(key)))
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

func (s *GameScreen) handleDirectionalItemInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			cmd := command.Command{Type: s.pendingCommand, Direction: s.pendingDirection}
			s.addCommandResult(s.executeCommand(cmd, string(key)))
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleCLIInput handles input in CLI mode
func (s *GameScreen) handleCLIInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.cliBuffer = ""
		s.AddMessage("CLI mode exited.")
		return state.StateGame
	case gruid.KeyEnter:
		if s.cliBuffer != "" {
			// Execute command
			result := s.cliMode.ExecuteCommand(s.cliBuffer)
			s.AddMessage(fmt.Sprintf("> %s", s.cliBuffer))
			s.AddMessage(result)
			s.player = s.cliMode.Player
			s.dungeonManager = s.cliMode.Dungeon
			s.level = s.cliMode.Level
			if s.wizardMode != nil && s.level != nil {
				s.wizardMode.Player = s.player
				s.wizardMode.SetLevel(s.level)
			}

			// Add to history
			s.cliHistory = append(s.cliHistory, s.cliBuffer)
			if len(s.cliHistory) > 20 {
				s.cliHistory = s.cliHistory[1:]
			}

			s.cliBuffer = ""
		}
		s.inputMode = ModeNormal
		return state.StateGame
	case gruid.KeyBackspace:
		if s.cliBuffer != "" {
			s.cliBuffer = s.cliBuffer[:len(s.cliBuffer)-1]
		}
		return state.StateGame
	default:
		// Add character to buffer
		if len(string(key)) == 1 {
			char := string(key)[0]
			if char >= 32 && char <= 126 { // Printable ASCII
				s.cliBuffer += string(char)
			}
		}
		return state.StateGame
	}
}

// handleDirectionInput handles input in direction mode
func (s *GameScreen) handleDirectionInput(key gruid.Key) state.GameState {
	// Parse the key into a command
	cmd := s.cmdParser.Parse(key)

	switch cmd.Type {
	case command.CmdMoveWest, command.CmdMoveEast, command.CmdMoveNorth, command.CmdMoveSouth,
		command.CmdMoveNorthWest, command.CmdMoveNorthEast, command.CmdMoveSouthWest, command.CmdMoveSouthEast:
		s.inputMode = ModeNormal
		callback := s.directionCallback
		s.directionCallback = nil
		if callback != nil {
			callback(cmd.Direction.X, cmd.Direction.Y)
		}
	case command.CmdEscape:
		s.AddMessage("Canceled.")
		s.inputMode = ModeNormal
		s.directionCallback = nil
	default:
		s.AddMessage("Invalid direction. Use hjklybnu or arrow keys.")
	}

	return state.StateGame
}
