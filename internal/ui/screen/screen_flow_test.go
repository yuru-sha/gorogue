package screen

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/game/actor"
)

func TestGameScreenRogueStatusAndTopMessage(t *testing.T) {
	player := actor.NewPlayerWithSeed(1, 1, 42)
	player.Gold = 12
	player.HP, player.MaxHP = 7, 10
	player.Strength, player.MaxStrength = 16, 18
	player.Level, player.Exp = 3, 42
	screen := NewGameScreen(80, 24, player)
	screen.messages = []string{"A prompt."}
	grid := gruid.NewGrid(80, 24)
	screen.Draw(&grid)
	var top, bottom strings.Builder
	for x := range 80 {
		top.WriteRune(grid.At(gruid.Point{X: x, Y: 0}).Rune)
		bottom.WriteRune(grid.At(gruid.Point{X: x, Y: 23}).Rune)
	}
	if got := strings.TrimRight(top.String(), " "); got != "A prompt." {
		t.Fatalf("top message = %q", got)
	}
	want := "Level: 1 Gold: 12 Hp: 7(10) Str: 16(18) Arm: 0 Exp: 3/42"
	if got := strings.TrimRight(bottom.String(), " "); got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
}

func TestGameScreenHungerStatusLabels(t *testing.T) {
	tests := []struct {
		name  string
		state int
		want  string
	}{
		{name: "satisfied", state: actor.HungerSatisfied, want: ""},
		{name: "hungry", state: actor.HungerHungry, want: " Hungry"},
		{name: "weak", state: actor.HungerWeak, want: " Weak"},
		{name: "faint", state: actor.HungerFainting, want: " Faint"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := actor.NewPlayerWithSeed(1, 1, 42)
			player.HungerState = tt.state
			screen := NewGameScreen(80, 24, player)
			grid := gruid.NewGrid(80, 24)
			screen.Draw(&grid)
			var status strings.Builder
			for x := range 80 {
				status.WriteRune(grid.At(gruid.Point{X: x, Y: 23}).Rune)
			}
			if got := strings.TrimRight(status.String(), " "); !strings.HasSuffix(got, tt.want) {
				t.Fatalf("status = %q, want hunger suffix %q", got, tt.want)
			}
		})
	}
}

func TestGameScreenHelpAndSequenceInputIsolation(t *testing.T) {
	screen := NewGameScreen(80, 24, actor.NewPlayerWithSeed(1, 1, 42))
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "?"}); got != state.StateHelp {
		t.Fatalf("help state = %v", got)
	}
	screen.SetLevel(newTestFloor(5, 5))
	helpGrid := gruid.NewGrid(80, 24)
	screen.Draw(&helpGrid)
	if helpGrid.At(gruid.Point{X: 1, Y: 2}).Rune != '@' {
		t.Fatal("help prompt hid the gameplay map")
	}
	pos := screen.player.Position
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "l"}); got != state.StateGame {
		t.Fatalf("help response state = %v", got)
	}
	if screen.player.Position != pos {
		t.Fatal("help response also executed a gameplay command")
	}
	if !strings.Contains(strings.Join(screen.messages, " "), "Move East") {
		t.Fatalf("help response = %v", screen.messages)
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "?"}); got != state.StateHelp {
		t.Fatalf("help list state = %v", got)
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "*"}); got != state.StateHelp {
		t.Fatalf("command list state = %v", got)
	}
	if len(screen.messages) <= screen.helpPageSize() {
		t.Fatalf("command list has %d entries, want multiple pages", len(screen.messages))
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: gruid.KeySpace}); got != state.StateHelp || screen.helpPage != 1 {
		t.Fatalf("command list next page = %v at page %d", got, screen.helpPage)
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "l"}); got != state.StateGame {
		t.Fatalf("command list dismissal state = %v", got)
	}
	if screen.player.Position != pos {
		t.Fatal("command list dismissal also executed a gameplay command")
	}
	screen.StartDeathSequence([]string{"You died."}, []string{"Score: 1"})
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "l"}); got != state.StateGameOver || screen.sequenceStage != 1 {
		t.Fatalf("death score stage = %v at stage %d", got, screen.sequenceStage)
	}
	if screen.player.Position != pos {
		t.Fatal("death acknowledgement reached gameplay")
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "l"}); got != state.StateQuit {
		t.Fatalf("death completion state = %v, want StateQuit", got)
	}
	screen.StartVictorySequence([]string{"Won."}, []string{"Score: 1"})
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "h"}); got != state.StateVictory || screen.sequenceStage != 1 {
		t.Fatalf("victory score stage = %v at stage %d", got, screen.sequenceStage)
	}
	if got := screen.HandleInput(gruid.MsgKeyDown{Key: "h"}); got != state.StateQuit {
		t.Fatalf("victory completion state = %v, want StateQuit", got)
	}
}
