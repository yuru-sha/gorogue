package screen

import "github.com/yuru-sha/gorogue/internal/core/command"

func (s *GameScreen) handleLook() {
	s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdLook}))
}

func (s *GameScreen) handlePickUp() {
	s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdPickUp}))
}

func (s *GameScreen) enterUseMode() {
	s.inputMode = ModeUse
	s.AddMessage("Use what? (a-z, ESC to cancel)")
}

func (s *GameScreen) handleWait() {
	s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdWait}))
}

func (s *GameScreen) handleSearch() {
	s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdSearch}))
}

func (s *GameScreen) handleOpenDoor() {
	s.AddMessage("Which direction?")
	s.inputMode = ModeDirection
	s.directionCallback = s.doOpenDoor
}

func (s *GameScreen) handleCloseDoor() {
	s.AddMessage("Which direction?")
	s.inputMode = ModeDirection
	s.directionCallback = s.doCloseDoor
}

func (s *GameScreen) handleFight() {
	s.AddMessage("Attack which direction? (hjklybnu)")
	s.inputMode = ModeDirection
	s.directionCallback = s.doFight
}

func (s *GameScreen) doFight(dx, dy int) {
	s.addCommandResult(s.executeCommand(command.Command{
		Type:      command.CmdFight,
		Direction: command.Direction{X: dx, Y: dy},
	}))
}
func (s *GameScreen) handleFindTrap() {
	s.AddMessage("Search for a trap in which direction? (hjklybnu)")
	s.inputMode = ModeDirection
	s.directionCallback = s.doFindTrap
}

func (s *GameScreen) doFindTrap(dx, dy int) {
	s.addCommandResult(s.executeCommand(command.Command{
		Type:      command.CmdFindTrap,
		Direction: command.Direction{X: dx, Y: dy},
	}))
}

func (s *GameScreen) handleThrow() {
	s.pendingCommand = command.CmdThrow
	s.AddMessage("Throw in which direction? (hjklybnu)")
	s.inputMode = ModeDirection
	s.directionCallback = s.beginDirectionalItemSelection
}

func (s *GameScreen) handleZap() {
	s.pendingCommand = command.CmdZap
	s.AddMessage("Zap in which direction? (hjklybnu)")
	s.inputMode = ModeDirection
	s.directionCallback = s.beginDirectionalItemSelection
}

func (s *GameScreen) beginDirectionalItemSelection(dx, dy int) {
	s.pendingDirection = command.Direction{X: dx, Y: dy}
	s.AddMessage("Select an item to use (a-z, ESC to cancel).")
	if s.pendingCommand == command.CmdThrow {
		s.inputMode = ModeThrow
	} else {
		s.inputMode = ModeZap
	}
}

func (s *GameScreen) handleRepeat() {
	s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdRepeat}))
}

func (s *GameScreen) doOpenDoor(dx, dy int) {
	s.addCommandResult(s.executeCommand(command.Command{
		Type:      command.CmdOpen,
		Direction: command.Direction{X: dx, Y: dy},
	}))
}

func (s *GameScreen) doCloseDoor(dx, dy int) {
	s.addCommandResult(s.executeCommand(command.Command{
		Type:      command.CmdClose,
		Direction: command.Direction{X: dx, Y: dy},
	}))
}
