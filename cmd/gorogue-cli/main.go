// GoRogue CLI - Terminal-based CLI mode like PyRogue
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/save"
	"github.com/yuru-sha/gorogue/internal/game/score"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

var (
	debugMode   = flag.Bool("debug", false, "Enable debug mode")
	helpFlag    = flag.Bool("help", false, "Show help information")
	interactive = flag.Bool("interactive", true, "Run in interactive mode")
	seedFlag    = flag.Int64("seed", 0, "Random seed (0 selects one automatically)")
	showScores  = flag.Bool("s", false, "Show persisted high scores and exit")
)

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}
	if *showScores {
		if err := showHighScores(os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read high scores: %v\n", err)
			os.Exit(1)
		}
		return
	}
	// Initialize logger
	if err := logger.Setup(); err != nil {
		panic(err)
	}
	defer logger.Cleanup()

	options, err := config.LoadOptions()
	if err != nil {
		logger.Fatal("Invalid ROGUEOPTS", "error", err)
	}
	// 環境変数で設定されていればそれを使用、フラグで上書き
	debugEnabled := config.GetDebugMode() || *debugMode

	if debugEnabled {
		logger.Info("Starting GoRogue CLI in debug mode",
			"env_debug", config.GetDebugMode(),
			"flag_debug", *debugMode,
		)
		if config.GetDebugMode() {
			config.PrintConfig()
		}
	} else {
		logger.Info("Starting GoRogue CLI")
	}

	seed := *seedFlag
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	// Initialize the same generated game world used by the GUI.
	player := actor.NewPlayerWithSeed(1, 1, seed)
	dungeonManager := dungeon.NewDungeonManagerWithSeed(player, seed)
	saveIntegration := save.NewSaveGameIntegration()
	saveIntegration.SetOptions(&options)
	if err := saveIntegration.Initialize(); err != nil {
		logger.Warn("Failed to initialize save integration", "error", err)
	}
	saveIntegration.SetGameState(player, dungeonManager)

	// Initialize CLI mode
	cliMode := cli.NewCLIModeWithDungeonManager(dungeonManager, player)
	cliMode.SetSaveIntegration(saveIntegration)
	cliMode.SetOptions(&options)
	cliMode.IsActive = true

	if *interactive {
		runInteractiveMode(cliMode)
	} else {
		runBatchMode(cliMode)
	}
}

func showHelp() {
	fmt.Println("GoRogue CLI - Terminal-based roguelike game")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gorogue-cli [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -debug         Enable debug mode")
	fmt.Println("  -help          Show this help")
	fmt.Println("  -interactive   Run in interactive mode (default: true)")
	fmt.Println("  -seed          Set the random seed (0 selects one automatically)")
	fmt.Println("  ROGUEOPTS      Configure Rogue options as comma-separated name or name=value tokens")
	fmt.Println("                 Boolean options use name or noname. See docs/features.md for valid names.")
	fmt.Println("  -s             Show persisted high scores and exit (read-only)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gorogue-cli                    # Start interactive CLI")
	fmt.Println("  gorogue-cli -debug             # Start with debug mode")
	fmt.Println("  gorogue-cli -s                 # List persisted high scores")
	fmt.Println("  echo 'status' | gorogue-cli -interactive=false  # Batch mode")
	fmt.Println()
	fmt.Println("High scores:")
	fmt.Println("  Displays rank, score, player, outcome/cause, and deepest reached floor.")
	fmt.Println("  Missing score files are not created.")
	fmt.Println()
	fmt.Println("Interactive Commands:")
	fmt.Println("  help           Show all available commands")
	fmt.Println("  status         Show player status")
	fmt.Println("  heal [amount]  Heal player")
	fmt.Println("  gold <amount>  Add gold")
	fmt.Println("  create <type>  Create item")
	fmt.Println("  teleport <x> <y>  Teleport player")
	fmt.Println("  quit, exit     Exit CLI")
	fmt.Println()
	fmt.Println("For full command list, run 'help' in interactive mode.")
}

func showHighScores(output io.Writer) error {
	entries, err := score.NewScoreManager().GetAllScores()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(output, "No high scores found (score file does not exist).")
			return nil
		}
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(output, "No high scores found (score list is empty).")
		return nil
	}

	fmt.Fprintln(output, "Rank  Score  Player  Outcome/Cause  Floor")
	for rank := range entries {
		entry := &entries[rank]
		outcome := entry.DeathReason
		if entry.IsVictory {
			outcome = "Victory"
		} else if outcome == "" {
			outcome = "Death"
		}
		fmt.Fprintf(output, "%d  %d  %s  %s  %d\n", rank+1, entry.Score, entry.PlayerName, outcome, entry.DeepestFloor)
	}
	return nil
}

type scannedInput struct {
	line string
	err  error
	done bool
}

func scanInput(input io.Reader) <-chan scannedInput {
	results := make(chan scannedInput)
	go func() {
		defer close(results)
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			results <- scannedInput{line: scanner.Text()}
		}
		results <- scannedInput{err: scanner.Err(), done: true}
	}()
	return results
}

func saveRecoveryState(cliMode *cli.CLIMode) error {
	if cliMode == nil || cliMode.Save == nil || cliMode.Player == nil || cliMode.Dungeon == nil {
		return fmt.Errorf("save integration or active game state is unavailable")
	}
	cliMode.Save.SetGameState(cliMode.Player, cliMode.Dungeon)
	return cliMode.Save.SaveGame()
}

func receiveHangup() (signals chan os.Signal, stop func()) {
	signals = make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGHUP)
	stop = func() { signal.Stop(signals) }
	return
}

func saveAfterHangup(cliMode *cli.CLIMode) {
	if err := saveRecoveryState(cliMode); err != nil {
		fmt.Fprintf(os.Stderr, "SIGHUP recovery save failed: %v\n", err)
	}
}

func recoverPendingHangup(cliMode *cli.CLIMode, signals <-chan os.Signal) bool {
	select {
	case <-signals:
		saveAfterHangup(cliMode)
		return true
	default:
		return false
	}
}

func runInteractiveMode(cliMode *cli.CLIMode) {
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                            GoRogue CLI Mode                                 ║")
	fmt.Println("║                         GoRogue CLI Interface                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Welcome to GoRogue CLI! Type 'help' for commands, 'quit' to exit.")
	fmt.Println()

	signals, stopSignals := receiveHangup()
	defer stopSignals()
	input := scanInput(os.Stdin)
	for {
		fmt.Print("gorogue> ")
		select {
		case <-signals:
			saveAfterHangup(cliMode)
			return
		case scanned, ok := <-input:
			if recoverPendingHangup(cliMode, signals) {
				return
			}
			if !ok || scanned.done {
				if scanned.err != nil {
					fmt.Printf("Error reading input: %v\n", scanned.err)
				}
				return
			}
			command := strings.TrimSpace(scanned.line)
			if command == "" {
				continue
			}
			switch strings.ToLower(command) {
			case "quit", "exit", "q":
				fmt.Println("Goodbye!")
				return
			case "clear", "cls":
				fmt.Print("\033[2J\033[1;1H")
				continue
			}
			fmt.Println(cliMode.ExecuteCommand(command))
			if cliMode.RunEnded {
				return
			}
			fmt.Println()
		}
	}
}

func runBatchMode(cliMode *cli.CLIMode) {
	fmt.Println("GoRogue CLI - Batch Mode")
	fmt.Println("Reading commands from stdin...")
	fmt.Println()

	signals, stopSignals := receiveHangup()
	input := scanInput(os.Stdin)
	commandCount := 0
	for {
		select {
		case <-signals:
			saveAfterHangup(cliMode)
			stopSignals()
			return
		case scanned, ok := <-input:
			if recoverPendingHangup(cliMode, signals) {
				stopSignals()
				return
			}
			if !ok {
				stopSignals()
				return
			}
			if scanned.done {
				if scanned.err != nil {
					fmt.Printf("Error reading input: %v\n", scanned.err)
					stopSignals()
					os.Exit(1)
				}
				fmt.Printf("Batch mode completed. Executed %d commands.\n", commandCount)
				stopSignals()
				return
			}
			command := strings.TrimSpace(scanned.line)
			if command == "" {
				continue
			}
			commandCount++
			fmt.Printf("[%d] Executing: %s\n", commandCount, command)
			fmt.Println(cliMode.ExecuteCommand(command))
			if cliMode.RunEnded {
				stopSignals()
				return
			}
			fmt.Println("---")
		}
	}
}
