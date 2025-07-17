package state

import "github.com/anaseto/gruid"

// GameState represents the current state of the game
type GameState int

const (
	StateMenu GameState = iota
	StateGame
	StateInventory
	StateHelp
	StateGameOver
	StateSaveLoad
)

// State represents a game state interface
type State interface {
	HandleInput(msg gruid.Msg) GameState
	Draw(grid *gruid.Grid)
}

// StateManager manages game states
type StateManager struct {
	currentState GameState
	states       map[GameState]State
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		currentState: StateMenu,
		states:       make(map[GameState]State),
	}
}

// RegisterState registers a state handler
func (sm *StateManager) RegisterState(state GameState, handler State) {
	sm.states[state] = handler
}

// GetCurrentState returns the current state
func (sm *StateManager) GetCurrentState() GameState {
	return sm.currentState
}

// SetState sets the current state
func (sm *StateManager) SetState(state GameState) {
	sm.currentState = state
}

// HandleInput handles input for the current state
func (sm *StateManager) HandleInput(msg gruid.Msg) gruid.Effect {
	if handler, exists := sm.states[sm.currentState]; exists {
		sm.currentState = handler.HandleInput(msg)
		if sm.currentState == StateGameOver {
			return gruid.End()
		}
	}
	return nil
}

// Draw draws the current state
func (sm *StateManager) Draw(grid *gruid.Grid) {
	if handler, exists := sm.states[sm.currentState]; exists {
		handler.Draw(grid)
	}
}
