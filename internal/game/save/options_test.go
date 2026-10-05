package save

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func TestRogueOptionsApplySavePathAndIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("LOG_DIRECTORY", t.TempDir())
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(logger.Cleanup)
	path := filepath.Join(t.TempDir(), "custom", "run.sav")
	integration := NewSaveGameIntegration()
	integration.SetOptions(&config.Options{Name: "Ada", File: path})
	if err := integration.Initialize(); err != nil {
		t.Fatal(err)
	}
	player := actor.NewPlayerWithSeed(1, 1, 42)
	manager := dungeon.NewDungeonManagerWithSeed(player, 42)
	integration.SetGameState(player, manager)
	if err := integration.SaveGame(); err != nil {
		t.Fatal(err)
	}
	if integration.GetSaveManager().GetSaveDirectory() != filepath.Dir(path) {
		t.Fatalf("save directory = %q, want %q", integration.GetSaveManager().GetSaveDirectory(), filepath.Dir(path))
	}
	loaded, err := integration.GetSaveManager().LoadGame()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.GameInfo.CharName != "Ada" {
		t.Fatalf("saved player identity = %q, want Ada", loaded.GameInfo.CharName)
	}
}

func TestUnconfiguredNamePreservesLoadedIdentity(t *testing.T) {
	integration := NewSaveGameIntegration()
	integration.SetGameInfo(GameInfo{CharName: "Saved adventurer"})
	defaults := config.DefaultOptions()
	integration.SetOptions(&defaults)
	if got := integration.GetGameInfo().CharName; got != "Saved adventurer" {
		t.Fatalf("saved identity after unrelated option update = %q, want %q", got, "Saved adventurer")
	}

	options := config.DefaultOptions()
	options.NameConfigured = true
	integration.SetOptions(&options)
	if got := integration.GetGameInfo().CharName; got != options.Name {
		t.Fatalf("explicit default identity = %q, want %q", got, options.Name)
	}
}

func TestRogueOptionsExpandHomeForUIFilePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	integration := NewSaveGameIntegration()
	options := config.DefaultOptions()
	options.File = "~/rogue.sav"
	integration.SetOptions(&options)
	if got, want := integration.GetSaveManager().getSaveFilePath(), filepath.Join(home, "rogue.sav"); got != want {
		t.Fatalf("save file path = %q, want %q", got, want)
	}
}

func TestCustomSavePathsKeepBackupsIsolated(t *testing.T) {
	t.Setenv("LOG_DIRECTORY", t.TempDir())
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(logger.Cleanup)
	dir := t.TempDir()
	alice := NewSaveManager()
	alicePath := filepath.Join(dir, "alice.sav")
	alice.SetSaveFilePath(alicePath)
	if err := alice.Initialize(); err != nil {
		t.Fatal(err)
	}
	bob := NewSaveManager()
	bobPath := filepath.Join(dir, "alice.sav_backup")
	bob.SetSaveFilePath(bobPath)
	if err := bob.Initialize(); err != nil {
		t.Fatal(err)
	}

	aliceSave := createTestSaveData(t)
	aliceSave.PlayerData.X = 1
	if err := alice.SaveGame(aliceSave); err != nil {
		t.Fatal(err)
	}
	aliceSave.PlayerData.X = 2
	if err := alice.SaveGame(aliceSave); err != nil {
		t.Fatal(err)
	}
	bobSave := createTestSaveData(t)
	bobSave.PlayerData.X = 10
	if err := bob.SaveGame(bobSave); err != nil {
		t.Fatal(err)
	}
	bobSave.PlayerData.X = 11
	if err := bob.SaveGame(bobSave); err != nil {
		t.Fatal(err)
	}

	if err := alice.consumeSave(func(*SaveData) error { return nil }); err != nil {
		t.Fatal(err)
	}
	bobBackups, err := bob.listBackupFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(bobBackups) != 1 {
		t.Fatalf("Bob backups after consuming Alice = %v, want one Bob backup", bobBackups)
	}
	if err := bob.RepairSave(); err != nil {
		t.Fatalf("repair Bob save: %v", err)
	}
	restored, err := bob.LoadGame()
	if err != nil {
		t.Fatal(err)
	}
	if restored.PlayerData.X != 10 {
		t.Fatalf("Bob restored x = %d, want 10", restored.PlayerData.X)
	}
}

func TestRelativeAndAbsoluteSavePathsShareBackupNamespace(t *testing.T) {
	t.Setenv("LOG_DIRECTORY", t.TempDir())
	if err := logger.Setup(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(logger.Cleanup)
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	savePath := filepath.Join(t.TempDir(), "shared.sav")
	relativePath, err := filepath.Rel(workingDir, savePath)
	if err != nil {
		t.Fatal(err)
	}

	relativeManager := NewSaveManager()
	relativeManager.SetSaveFilePath(relativePath)
	absoluteManager := NewSaveManager()
	absoluteManager.SetSaveFilePath(savePath)
	firstSave := createTestSaveData(t)
	firstSave.PlayerData.X = 10
	if err := relativeManager.SaveGame(firstSave); err != nil {
		t.Fatal(err)
	}
	firstSave.PlayerData.X = 11
	if err := absoluteManager.SaveGame(firstSave); err != nil {
		t.Fatal(err)
	}

	backups, err := relativeManager.listBackupFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(backups) != 1 {
		t.Fatalf("relative-path backups = %v, want shared backup", backups)
	}
	if err := relativeManager.RepairSave(); err != nil {
		t.Fatal(err)
	}
	restored, err := relativeManager.LoadGame()
	if err != nil {
		t.Fatal(err)
	}
	if restored.PlayerData.X != 10 {
		t.Fatalf("restored x = %d, want 10", restored.PlayerData.X)
	}
}

func TestDefaultBackupPrefixWithMissingOrRelativeHome(t *testing.T) {
	for _, home := range []string{"", "relative-home"} {
		t.Run(home, func(t *testing.T) {
			t.Setenv("HOME", home)
			manager := NewSaveManager()
			if prefix := manager.backupPrefix(); prefix != "rogue_" {
				t.Fatalf("backup prefix = %q, want legacy default prefix rogue_", prefix)
			}
		})
	}
}
