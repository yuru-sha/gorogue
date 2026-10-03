package main

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestSaveRecoveryStatePersistsCurrentGameWithoutAdvancingTurn(t *testing.T) {
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", t.TempDir())
	player := actor.NewPlayerWithSeed(1, 1, 42)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, 42)
	player.Position.X, player.Position.Y = 7, 8
	integration := save.NewSaveGameIntegration()
	if err := integration.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	settings := integration.GetSettings()
	settings.AutoSave = false
	integration.SetSettings(settings)
	integration.SetGameState(player, dungeonManager)
	cliMode := &cli.CLIMode{Player: player, Dungeon: dungeonManager, Save: integration}
	beforeTurn := integration.GetGameStats().GetTurnCount()
	signals := make(chan os.Signal, 1)
	signals <- syscall.SIGHUP
	if !recoverPendingHangup(cliMode, signals) {
		t.Fatal("recoverPendingHangup() = false, want true")
	}
	loaded, err := integration.GetSaveManager().LoadGame()
	if err != nil {
		t.Fatalf("LoadGame() error = %v", err)
	}
	if loaded.PlayerData.X != 7 || loaded.PlayerData.Y != 8 {
		t.Fatalf("recovery save position = (%d, %d), want (7, 8)", loaded.PlayerData.X, loaded.PlayerData.Y)
	}
	if got := integration.GetGameStats().GetTurnCount(); got != beforeTurn {
		t.Fatalf("recovery save advanced turn count from %d to %d", beforeTurn, got)
	}
}

func TestCLIHandlesHangupWhileWaitingForInput(t *testing.T) {
	if os.Getenv("GOROGUE_HANGUP_CHILD") == "1" {
		if err := logger.Setup(); err != nil {
			t.Fatal(err)
		}
		defer logger.Cleanup()
		player := actor.NewPlayerWithSeed(1, 1, 42)
		dungeonManager := dungeon.NewDungeonManagerWithSeed(player, 42)
		integration := save.NewSaveGameIntegration()
		if err := integration.Initialize(); err != nil {
			t.Fatal(err)
		}
		integration.SetGameState(player, dungeonManager)
		cliMode := cli.NewCLIModeWithDungeonManager(dungeonManager, player)
		cliMode.SetSaveIntegration(integration)
		cliMode.IsActive = true
		runInteractiveMode(cliMode)
		return
	}

	home := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIHandlesHangupWhileWaitingForInput$")
	cmd.Env = append(os.Environ(), "GOROGUE_HANGUP_CHILD=1", "HOME="+home)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	ready, err := bufio.NewReader(stdout).ReadString('>')
	if err != nil {
		t.Fatalf("CLI did not reach input prompt: %v, output %q, stderr %q", err, ready, stderr.String())
	}
	if !strings.Contains(ready, "gorogue>") {
		t.Fatalf("CLI prompt = %q", ready)
	}
	if err := cmd.Process.Signal(syscall.SIGHUP); err != nil {
		t.Fatalf("sending SIGHUP failed: %v", err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("CLI exited after SIGHUP: %v, stderr %q", err, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".gorogue", save.SaveDirectory, save.SaveFileName)); err != nil {
		t.Fatalf("SIGHUP did not persist a recovery save: %v", err)
	}
}
