package core

import (
	"strings"
	"testing"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestNewEngineRegistersSaveLoadState(t *testing.T) {
	logger.Setup()
	engine := NewEngineWithSeed(12345)
	engine.stateManager.SetState(state.StateSaveLoad)
	grid := gruid.NewGrid(80, 50)
	engine.stateManager.Draw(&grid)

	var rows []string
	for y := 0; y < 50; y++ {
		var row strings.Builder
		for x := 0; x < 80; x++ {
			row.WriteRune(grid.At(gruid.Point{X: x, Y: y}).Rune)
		}
		rows = append(rows, row.String())
	}
	if !strings.Contains(strings.Join(rows, "\n"), "SAVE/LOAD GAME") {
		t.Fatal("save/load state is not registered")
	}
}
