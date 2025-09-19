package cmd

import (
	"testing"

	"github.com/rykov/paperboy/config"
)

func TestNew(t *testing.T) {
	build := config.BuildInfo{
		Version:   "test-version",
		BuildDate: "test-date",
	}

	cmd := New(build)

	if cmd == nil {
		t.Fatal("New() returned nil command")
	}

	if cmd.Use != "paperboy" {
		t.Errorf("Expected Use to be 'paperboy', got %s", cmd.Use)
	}

	if config.Build.Version != "test-version" {
		t.Errorf("Expected config.Build.Version to be 'test-version', got %s", config.Build.Version)
	}

	if config.Build.BuildDate != "test-date" {
		t.Errorf("Expected config.Build.BuildDate to be 'test-date', got %s", config.Build.BuildDate)
	}
}

func TestNewSubcommands(t *testing.T) {
	build := config.BuildInfo{Version: "test", BuildDate: "test"}
	cmd := New(build)

	expectedCommands := []string{"new", "init", "send", "server", "version", "preview"}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expectedCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand %s not found", expectedCmd)
		}
	}
}

func TestNewConfigFlag(t *testing.T) {
	build := config.BuildInfo{Version: "test", BuildDate: "test"}
	cmd := New(build)

	configFlag := cmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Fatal("Expected --config flag to be present")
	}

	if configFlag.Usage != "config file (default: ./config.yaml)" {
		t.Errorf("Unexpected config flag usage: %s", configFlag.Usage)
	}
}

func TestNewUserError(t *testing.T) {
	err := newUserError("test error: %s", "example")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected := "test error: example"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestNewUserErrorNoArgs(t *testing.T) {
	err := newUserError("simple error")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expected := "simple error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestCommandStructure(t *testing.T) {
	build := config.BuildInfo{Version: "test", BuildDate: "test"}
	cmd := New(build)

	if cmd.Use != "paperboy" {
		t.Errorf("Root command should have Use 'paperboy', got %s", cmd.Use)
	}

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Root command should have subcommands")
	}

	commandNames := make(map[string]bool)
	for _, subCmd := range subcommands {
		commandNames[subCmd.Name()] = true
	}

	requiredCommands := []string{"new", "init", "send", "server", "version", "preview"}
	for _, required := range requiredCommands {
		if !commandNames[required] {
			t.Errorf("Missing required command: %s", required)
		}
	}
}

func TestCobraCommandTypes(t *testing.T) {
	build := config.BuildInfo{Version: "test", BuildDate: "test"}
	cmd := New(build)

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("Expected subcommands to exist")
	}
}
