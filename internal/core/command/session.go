package command

// Session owns the last turn-consuming command shared by all input interfaces.
type Session struct {
	lastCommand Command
	lastArgs    []string
	hasLast     bool
}

func NewSession() *Session {
	return &Session{}
}

func (s *Session) Execute(ctx *Context, cmd Command, args ...string) Result {
	repeating := cmd.Type == CmdRepeat
	if repeating {
		if !s.hasLast {
			return result(ctx, "You haven't performed a command yet.")
		}
		cmd = s.lastCommand
		args = s.lastArgs
	}
	outcome := Execute(ctx, cmd, args...)

	if !repeating && outcome.TurnConsumed {
		s.lastCommand = cmd
		s.lastArgs = append(s.lastArgs[:0], args...)
		s.hasLast = true
	}
	return outcome
}
