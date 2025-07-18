package config

import (
	"os"
	"testing"
)

func TestGetBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"empty_value", "", false, false},
		{"empty_value_default_true", "", true, true},
		{"true_value", "true", false, true},
		{"false_value", "false", true, false},
		{"one_value", "1", false, true},
		{"zero_value", "0", true, false},
		{"on_value", "on", false, true},
		{"off_value", "off", true, false},
		{"yes_value", "yes", false, true},
		{"no_value", "no", true, false},
		{"invalid_value", "invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}
			defer os.Unsetenv(key)

			result := GetBool(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetBool(%q, %v) = %v, expected %v", key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{"empty_value", "", "default", "default"},
		{"set_value", "custom", "default", "custom"},
		{"empty_string", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_STRING"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}
			defer os.Unsetenv(key)

			result := GetString(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetString(%q, %q) = %q, expected %q", key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"empty_value", "", 100, 100},
		{"valid_int", "42", 100, 42},
		{"invalid_int", "invalid", 100, 100},
		{"zero_value", "0", 100, 0},
		{"negative_value", "-5", 100, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}
			defer os.Unsetenv(key)

			result := GetInt(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetInt(%q, %d) = %d, expected %d", key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue float64
		expected     float64
	}{
		{"empty_value", "", 3.14, 3.14},
		{"valid_float", "2.71", 3.14, 2.71},
		{"invalid_float", "invalid", 3.14, 3.14},
		{"zero_value", "0", 3.14, 0.0},
		{"negative_value", "-1.5", 3.14, -1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_FLOAT"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			} else {
				os.Unsetenv(key)
			}
			defer os.Unsetenv(key)

			result := GetFloat64(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetFloat64(%q, %f) = %f, expected %f", key, tt.defaultValue, result, tt.expected)
			}
		})
	}
}

func TestConfigAccessors(t *testing.T) {
	// Test with clean environment
	os.Unsetenv(EnvDebugMode)
	os.Unsetenv(EnvLogLevel)
	os.Unsetenv(EnvSaveDirectory)
	os.Unsetenv(EnvAutoSaveEnabled)

	// Test defaults
	if GetDebugMode() != DefaultDebugMode {
		t.Errorf("GetDebugMode() = %v, expected %v", GetDebugMode(), DefaultDebugMode)
	}
	if GetLogLevel() != DefaultLogLevel {
		t.Errorf("GetLogLevel() = %q, expected %q", GetLogLevel(), DefaultLogLevel)
	}
	if GetSaveDirectory() != DefaultSaveDirectory {
		t.Errorf("GetSaveDirectory() = %q, expected %q", GetSaveDirectory(), DefaultSaveDirectory)
	}
	if GetAutoSaveEnabled() != DefaultAutoSaveEnabled {
		t.Errorf("GetAutoSaveEnabled() = %v, expected %v", GetAutoSaveEnabled(), DefaultAutoSaveEnabled)
	}

	// Test with environment variables
	os.Setenv(EnvDebugMode, "true")
	os.Setenv(EnvLogLevel, "ERROR")
	os.Setenv(EnvSaveDirectory, "custom_saves")
	os.Setenv(EnvAutoSaveEnabled, "false")

	if GetDebugMode() != true {
		t.Errorf("GetDebugMode() = %v, expected true", GetDebugMode())
	}
	if GetLogLevel() != "ERROR" {
		t.Errorf("GetLogLevel() = %q, expected ERROR", GetLogLevel())
	}
	if GetSaveDirectory() != "custom_saves" {
		t.Errorf("GetSaveDirectory() = %q, expected custom_saves", GetSaveDirectory())
	}
	if GetAutoSaveEnabled() != false {
		t.Errorf("GetAutoSaveEnabled() = %v, expected false", GetAutoSaveEnabled())
	}

	// Cleanup
	os.Unsetenv(EnvDebugMode)
	os.Unsetenv(EnvLogLevel)
	os.Unsetenv(EnvSaveDirectory)
	os.Unsetenv(EnvAutoSaveEnabled)
}

func TestGetConfig(t *testing.T) {
	// Clean environment
	os.Unsetenv(EnvDebugMode)
	os.Unsetenv(EnvLogLevel)
	os.Unsetenv(EnvSaveDirectory)
	os.Unsetenv(EnvAutoSaveEnabled)

	config := GetConfig()
	if config == nil {
		t.Fatal("GetConfig() returned nil")
	}

	// Test defaults
	if config.DebugMode != DefaultDebugMode {
		t.Errorf("config.DebugMode = %v, expected %v", config.DebugMode, DefaultDebugMode)
	}
	if config.LogLevel != DefaultLogLevel {
		t.Errorf("config.LogLevel = %q, expected %q", config.LogLevel, DefaultLogLevel)
	}
	if config.SaveDirectory != DefaultSaveDirectory {
		t.Errorf("config.SaveDirectory = %q, expected %q", config.SaveDirectory, DefaultSaveDirectory)
	}
	if config.AutoSaveEnabled != DefaultAutoSaveEnabled {
		t.Errorf("config.AutoSaveEnabled = %v, expected %v", config.AutoSaveEnabled, DefaultAutoSaveEnabled)
	}

	// Test with environment variables
	os.Setenv(EnvDebugMode, "true")
	os.Setenv(EnvLogLevel, "DEBUG")
	os.Setenv(EnvSaveDirectory, "custom_saves")
	os.Setenv(EnvAutoSaveEnabled, "false")

	config = GetConfig()
	if config.DebugMode != true {
		t.Errorf("config.DebugMode = %v, expected true", config.DebugMode)
	}
	if config.LogLevel != "DEBUG" {
		t.Errorf("config.LogLevel = %q, expected DEBUG", config.LogLevel)
	}
	if config.SaveDirectory != "custom_saves" {
		t.Errorf("config.SaveDirectory = %q, expected custom_saves", config.SaveDirectory)
	}
	if config.AutoSaveEnabled != false {
		t.Errorf("config.AutoSaveEnabled = %v, expected false", config.AutoSaveEnabled)
	}

	// Cleanup
	os.Unsetenv(EnvDebugMode)
	os.Unsetenv(EnvLogLevel)
	os.Unsetenv(EnvSaveDirectory)
	os.Unsetenv(EnvAutoSaveEnabled)
}

func TestSetUnsetEnv(t *testing.T) {
	key := "TEST_SET_UNSET"
	value := "test_value"

	// Test SetEnv
	SetEnv(key, value)
	if os.Getenv(key) != value {
		t.Errorf("SetEnv failed: os.Getenv(%q) = %q, expected %q", key, os.Getenv(key), value)
	}

	// Test UnsetEnv
	UnsetEnv(key)
	if os.Getenv(key) != "" {
		t.Errorf("UnsetEnv failed: os.Getenv(%q) = %q, expected empty string", key, os.Getenv(key))
	}
}

func TestEnvironmentVariableKeys(t *testing.T) {
	// Test that all environment variable keys are properly defined
	expectedKeys := []string{
		EnvDebugMode,
		EnvLogLevel,
		EnvSaveDirectory,
		EnvAutoSaveEnabled,
	}

	for _, key := range expectedKeys {
		if key == "" {
			t.Errorf("Environment variable key is empty")
		}
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that default values are reasonable
	if DefaultLogLevel == "" {
		t.Error("DefaultLogLevel is empty")
	}
	if DefaultSaveDirectory == "" {
		t.Error("DefaultSaveDirectory is empty")
	}
}

// Benchmark tests
func BenchmarkGetBool(b *testing.B) {
	key := "BENCHMARK_BOOL"
	os.Setenv(key, "true")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetBool(key, false)
	}
}

func BenchmarkGetString(b *testing.B) {
	key := "BENCHMARK_STRING"
	os.Setenv(key, "benchmark_value")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetString(key, "default")
	}
}

func BenchmarkGetInt(b *testing.B) {
	key := "BENCHMARK_INT"
	os.Setenv(key, "42")
	defer os.Unsetenv(key)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetInt(key, 0)
	}
}

func BenchmarkGetConfig(b *testing.B) {
	// Set up some environment variables
	os.Setenv(EnvDebugMode, "true")
	os.Setenv(EnvLogLevel, "DEBUG")
	os.Setenv(EnvSaveDirectory, "saves")
	os.Setenv(EnvAutoSaveEnabled, "true")
	defer func() {
		os.Unsetenv(EnvDebugMode)
		os.Unsetenv(EnvLogLevel)
		os.Unsetenv(EnvSaveDirectory)
		os.Unsetenv(EnvAutoSaveEnabled)
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetConfig()
	}
}