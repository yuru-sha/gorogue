package config

import (
	"path/filepath"
	"testing"
)

func TestParseOptions(t *testing.T) {
	got, err := ParseOptions("terse,noflush,jump,seefloor,nopassgo,tombstone,inven=clear,name=Ada Lovelace,fruit=pear,file=/tmp/rogue.sav")
	if err != nil {
		t.Fatal(err)
	}
	want := Options{
		Terse:          true,
		Flush:          false,
		Jump:           true,
		SeeFloor:       true,
		PassGo:         false,
		Tombstone:      true,
		InventoryStyle: InventoryClear,
		Name:           "Ada Lovelace",
		NameConfigured: true,
		Fruit:          "pear",
		File:           "/tmp/rogue.sav",
	}
	if got != want {
		t.Fatalf("ParseOptions() = %+v, want %+v", got, want)
	}
}

func TestParseOptionsDefaultsAndNoPrefix(t *testing.T) {
	got, err := ParseOptions("noterse,noflush,nojump,noseefloor,nopassgo,notombstone")
	if err != nil {
		t.Fatal(err)
	}
	want := DefaultOptions()
	want.SeeFloor = false
	want.Tombstone = false
	if got != want {
		t.Fatalf("ParseOptions() = %+v, want %+v", got, want)
	}
}

func TestParseOptionsRejectsInvalidInput(t *testing.T) {
	for _, raw := range []string{"unknown", "terse=true", "name", "name=", "file", "inven=invalid", "=value"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseOptions(raw); err == nil {
				t.Fatalf("ParseOptions(%q) accepted invalid input", raw)
			}
		})
	}
}

func TestLoadOptionsUsesROGUEOPTSEnvironment(t *testing.T) {
	t.Setenv(EnvRogueOptions, "name=Lin,noterse,inven=clear")
	got, err := LoadOptions()
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Lin" || got.Terse || got.InventoryStyle != InventoryClear {
		t.Fatalf("LoadOptions() = %+v", got)
	}
}

func TestParseOptionsExpandsSavePathHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	got, err := ParseOptions("file=~/rogue.sav")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "rogue.sav"); got.File != want {
		t.Fatalf("save path = %q, want %q", got.File, want)
	}
}

func TestParseOptionsEmptySavePathUsesDefault(t *testing.T) {
	got, err := ParseOptions("file=")
	if err != nil {
		t.Fatal(err)
	}
	if want := DefaultOptions(); got != want {
		t.Fatalf("ParseOptions(file=) = %+v, want defaults %+v", got, want)
	}
}
