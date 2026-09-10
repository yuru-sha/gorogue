package screen

import (
	"github.com/yuru-sha/gorogue/internal/core/command"
)

// tryMovePlayer executes a movement command from a direction-based UI action.
func (s *GameScreen) tryMovePlayer(dx, dy int) {
	s.addCommandResult(s.executeCommand(command.NewMoveCommand(command.Direction{X: dx, Y: dy})))
}

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

func (s *GameScreen) handleDisarm() {
	s.AddMessage("Disarm trap which direction? (hjklybnu)")
}

func (s *GameScreen) canGoDownstairs() bool {
	return s.dungeonManager != nil && s.dungeonManager.CanGoDownstairs()
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
