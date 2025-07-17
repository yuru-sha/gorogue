// GoRogue CLI - Terminal-based CLI mode like PyRogue
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yuru-sha/gorogue/internal/config"
	"github.com/yuru-sha/gorogue/internal/core"
	"github.com/yuru-sha/gorogue/internal/core/cli"
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

var (
	debugMode   = flag.Bool("debug", false, "Enable debug mode")
	helpFlag    = flag.Bool("help", false, "Show help information")
	interactive = flag.Bool("interactive", true, "Run in interactive mode")
)

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	// Initialize logger
	if err := logger.Setup(); err != nil {
		panic(err)
	}
	defer logger.Cleanup()

	// 環境変数で設定されていればそれを使用、フラグで上書き
	debugEnabled := config.GetDebugMode() || *debugMode
	
	if debugEnabled {
		logger.Info("Starting GoRogue CLI in debug mode",
			"env_debug", config.GetDebugMode(),
			"flag_debug", *debugMode,
			"save_directory", config.GetSaveDirectory(),
			"auto_save", config.GetAutoSaveEnabled(),
		)
		if config.GetDebugMode() {
			config.PrintConfig()
		}
	} else {
		logger.Info("Starting GoRogue CLI")
	}

	// Initialize game engine
	engine := core.NewEngine()
	if engine == nil {
		fmt.Println("Failed to initialize game engine")
		os.Exit(1)
	}

	// Initialize game world for CLI commands
	player := actor.NewPlayer(1, 1)
	level := &dungeon.Level{
		Width:    80,
		Height:   24,
		Tiles:    make([][]*dungeon.Tile, 80),
		Monsters: make([]*actor.Monster, 0),
		Items:    make([]*item.Item, 0),
		Rooms:    make([]*dungeon.Room, 0),
	}

	// Initialize CLI mode
	cliMode := cli.NewCLIMode(level, player)
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
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gorogue-cli                    # Start interactive CLI")
	fmt.Println("  gorogue-cli -debug             # Start with debug mode")
	fmt.Println("  echo 'status' | gorogue-cli -interactive=false  # Batch mode")
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

func runInteractiveMode(cliMode *cli.CLIMode) {
	fmt.Println("╔══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                            GoRogue CLI Mode                                 ║")
	fmt.Println("║                      PyRogue-compatible CLI Interface                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Welcome to GoRogue CLI! Type 'help' for commands, 'quit' to exit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("gorogue> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return
		case "clear", "cls":
			fmt.Print("\033[2J\033[1;1H") // Clear screen
			continue
		}

		// Execute CLI command
		result := cliMode.ExecuteCommand(input)
		fmt.Println(result)
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
	}
}

func runBatchMode(cliMode *cli.CLIMode) {
	fmt.Println("GoRogue CLI - Batch Mode")
	fmt.Println("Reading commands from stdin...")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	commandCount := 0

	for scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		commandCount++
		fmt.Printf("[%d] Executing: %s\n", commandCount, input)

		result := cliMode.ExecuteCommand(input)
		fmt.Println(result)
		fmt.Println("---")
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Batch mode completed. Executed %d commands.\n", commandCount)
}
