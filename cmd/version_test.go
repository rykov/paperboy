package cmd

import (
	"testing"
)

func TestVersionCmd(t *testing.T) {
	cmd := versionCmd()

	if cmd == nil {
		t.Fatal("versionCmd() returned nil")
	}

	if cmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got %s", cmd.Use)
	}

	if cmd.Short != "Print the version number of Paperboy" {
		t.Errorf("Expected specific short description, got %s", cmd.Short)
	}

	if cmd.Run == nil {
		t.Error("Run function should not be nil")
	}

	if cmd.RunE != nil {
		t.Error("RunE function should be nil when Run is set")
	}
}

func TestVersionCmdArgs(t *testing.T) {
	cmd := versionCmd()

	// Version command should accept any number of args (Args should be nil)
	if cmd.Args != nil {
		t.Error("Version command should accept any number of args (Args should be nil)")
	}

	// Test that version command doesn't fail with extra args
	// Since Args is nil, Cobra will accept any number of arguments
}

func TestVersionCmdPanic(t *testing.T) {
	cmd := versionCmd()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Version command should not panic: %v", r)
		}
	}()

	// This will print to stdout and potentially stderr, but shouldn't panic
	cmd.Run(cmd, []string{})
}

func TestVersionCmdStructure(t *testing.T) {
	cmd := versionCmd()

	if cmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got %s", cmd.Use)
	}

	if cmd.Args != nil {
		t.Error("Expected Args to be nil (no argument validation)")
	}

	if cmd.ValidArgs != nil {
		t.Error("Expected ValidArgs to be nil")
	}

	if cmd.Flags().NFlag() != 0 {
		t.Error("Expected no flags for version command")
	}
}
