package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const EnvRogueOptions = "ROGUEOPTS"
const saveFileOption = "file"

type InventoryStyle string

const (
	InventoryOverwrite InventoryStyle = "overwrite"
	InventorySlow      InventoryStyle = "slow"
	InventoryClear     InventoryStyle = "clear"
)

// Options contains the Rogue 5.4.4 user options shared by startup and the UI.
type Options struct {
	Terse          bool
	Flush          bool
	Jump           bool
	SeeFloor       bool
	PassGo         bool
	Tombstone      bool
	InventoryStyle InventoryStyle
	Name           string
	NameConfigured bool
	Fruit          string
	File           string
}

func DefaultOptions() Options {
	return Options{
		SeeFloor:       true,
		Tombstone:      true,
		InventoryStyle: InventoryOverwrite,
		Name:           "Player",
		Fruit:          "slime-mold",
	}
}

// ParseOptions overlays comma-separated Rogue name/value options on defaults.
func ParseOptions(raw string) (Options, error) {
	options := DefaultOptions()
	for token := range strings.SplitSeq(raw, ",") {
		if err := parseOption(&options, token); err != nil {
			return Options{}, err
		}
	}
	return options, nil
}

func parseOption(options *Options, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	name, value, hasValue := strings.Cut(token, "=")
	name = strings.TrimSpace(name)
	if hasValue {
		value = strings.TrimSpace(value)
	}
	switch name {
	case "terse", "flush", "jump", "seefloor", "passgo", "tombstone":
		if hasValue {
			return fmt.Errorf("option %q does not take a value", name)
		}
		setBooleanOption(options, name, true)
	case "noterse", "noflush", "nojump", "noseefloor", "nopassgo", "notombstone":
		if hasValue {
			return fmt.Errorf("option %q does not take a value", name)
		}
		setBooleanOption(options, strings.TrimPrefix(name, "no"), false)
	case "inven", "name", "fruit", saveFileOption:
		if !hasValue || (value == "" && name != saveFileOption) {
			return fmt.Errorf("option %q requires a non-empty value", name)
		}
		return setStringOption(options, name, value)
	default:
		return fmt.Errorf("unknown option %q", name)
	}
	return nil
}

func setStringOption(options *Options, name, value string) error {
	switch name {
	case "inven":
		style := InventoryStyle(value)
		if !validInventoryStyle(style) {
			return fmt.Errorf("invalid inventory style %q", value)
		}
		options.InventoryStyle = style
	case "name":
		options.Name = value
		options.NameConfigured = true
	case "fruit":
		options.Fruit = value
	case saveFileOption:
		path, err := expandHomePath(value)
		if err != nil {
			return err
		}
		options.File = path
	}
	return nil
}

func setBooleanOption(options *Options, name string, value bool) {
	switch name {
	case "terse":
		options.Terse = value
	case "flush":
		options.Flush = value
	case "jump":
		options.Jump = value
	case "seefloor":
		options.SeeFloor = value
	case "passgo":
		options.PassGo = value
	case "tombstone":
		options.Tombstone = value
	}
}
func expandHomePath(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand save path: %w", err)
	}
	return filepath.Join(home, strings.TrimLeft(strings.TrimPrefix(path, "~"), "/")), nil
}

func validInventoryStyle(style InventoryStyle) bool {
	switch style {
	case InventorySlow, InventoryClear, InventoryOverwrite:
		return true
	default:
		return false
	}
}

// LoadOptions reads ROGUEOPTS after the existing .env loader has populated the environment.
func LoadOptions() (Options, error) {
	return ParseOptions(os.Getenv(EnvRogueOptions))
}
