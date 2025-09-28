package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestVerifyCmd(t *testing.T) {
	cmd := verifyCmd()

	if cmd == nil {
		t.Fatal("verifyCmd() returned nil")
	}

	if cmd.Use != "verify [content] [list]" {
		t.Errorf("Expected Use to be 'verify [content] [list]', got %s", cmd.Use)
	}

	if cmd.Short != "Verify DKIM signatures in rendered emails" {
		t.Errorf("Expected Short to be 'Verify DKIM signatures in rendered emails', got %s", cmd.Short)
	}

	if cmd.Example != "paperboy verify the-announcement customers" {
		t.Errorf("Expected specific example, got %s", cmd.Example)
	}
}

func TestVerifyCmdArgs(t *testing.T) {
	cmd := verifyCmd()

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

func TestVerifyCmdStructure(t *testing.T) {
	cmd := verifyCmd()

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}

	if cmd.Run != nil {
		t.Error("Run function should be nil when RunE is set")
	}
}

func TestVerifyCmdValidation(t *testing.T) {
	cmd := verifyCmd()

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
