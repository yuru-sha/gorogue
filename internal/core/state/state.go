package state

import "github.com/hajimehoshi/ebiten/v2"

// GameState represents the current state of the game
type GameState int

const (
	StateMenu GameState = iota
	StateGame
	StateInventory
	StateHelp
	StateGameOver
)

// State represents a game state interface
type State interface {
	Update() GameState
	Draw(screen *ebiten.Image)
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

// Update updates the current state
func (sm *StateManager) Update() {
	if handler, exists := sm.states[sm.currentState]; exists {
		sm.currentState = handler.Update()
	}
}

// Draw draws the current state
func (sm *StateManager) Draw(screen *ebiten.Image) {
	if handler, exists := sm.states[sm.currentState]; exists {
		handler.Draw(screen)
	}
}
