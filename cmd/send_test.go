package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSendCmd(t *testing.T) {
	cmd := sendCmd()

	if cmd == nil {
		t.Fatal("sendCmd() returned nil")
	}

	if cmd.Use != "send [content] [list]" {
		t.Errorf("Expected Use to be 'send [content] [list]', got %s", cmd.Use)
	}

	if cmd.Short != "Send campaign to recipients" {
		t.Errorf("Expected Short to be 'Send campaign to recipients', got %s", cmd.Short)
	}

	if cmd.Example != "paperboy send the-announcement in-the-know" {
		t.Errorf("Expected specific example, got %s", cmd.Example)
	}
}

func TestSendCmdFlags(t *testing.T) {
	cmd := sendCmd()

	serverFlag := cmd.Flags().Lookup("server")
	if serverFlag == nil {
		t.Fatal("Expected --server flag to be present")
	}

	if serverFlag.Usage != "URL of server" {
		t.Errorf("Unexpected server flag usage: %s", serverFlag.Usage)
	}

	if serverFlag.DefValue != "" {
		t.Errorf("Expected server flag default value to be empty, got %s", serverFlag.DefValue)
	}
}

func TestSendCmdArgs(t *testing.T) {
	cmd := sendCmd()

	if cmd.Args == nil {
		t.Fatal("Args function should not be nil")
	}

	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"valid args", []string{"content", "list"}, false},
		{"too few args", []string{"content"}, true},
		{"too many args", []string{"content", "list", "extra"}, true},
		{"no args", []string{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Args(cmd, tc.args)
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSendCmdStructure(t *testing.T) {
	cmd := sendCmd()

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}

	if cmd.Run != nil {
		t.Error("Run function should be nil when RunE is set")
	}
}

func TestSendCmdValidation(t *testing.T) {
	cmd := sendCmd()

	testCases := []struct {
		name string
		args []string
		want error
	}{
		{
			name: "exact two args",
			args: []string{"campaign", "list"},
			want: nil,
		},
		{
			name: "one arg",
			args: []string{"campaign"},
			want: cobra.ExactArgs(2)(cmd, []string{"campaign"}),
		},
		{
			name: "three args",
			args: []string{"campaign", "list", "extra"},
			want: cobra.ExactArgs(2)(cmd, []string{"campaign", "list", "extra"}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := cmd.Args(cmd, tc.args)
			if (got == nil) != (tc.want == nil) {
				t.Errorf("Args validation failed for %v: got error=%v, want error=%v", tc.args, got, tc.want)
			}
		})
	}
}
