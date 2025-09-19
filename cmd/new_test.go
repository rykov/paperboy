package cmd

import (
	"testing"
	"text/template"

	"github.com/spf13/afero"
)

func TestNewCmd(t *testing.T) {
	cmd := newCmd()

	if cmd == nil {
		t.Fatal("newCmd() returned nil")
	}

	if cmd.Use != "new [path]" {
		t.Errorf("Expected Use to be 'new [path]', got %s", cmd.Use)
	}

	if cmd.Short != "Create new content for a campaign" {
		t.Errorf("Expected specific short description, got %s", cmd.Short)
	}

	if cmd.Example != "paperboy new the-announcement.md" {
		t.Errorf("Expected specific example, got %s", cmd.Example)
	}
}

func TestInitCmd(t *testing.T) {
	cmd := initCmd()

	if cmd == nil {
		t.Fatal("initCmd() returned nil")
	}

	if cmd.Use != "init [path]" {
		t.Errorf("Expected Use to be 'init [path]', got %s", cmd.Use)
	}
}

func TestNewProjectCmd(t *testing.T) {
	cmd := newProjectCmd()

	if cmd == nil {
		t.Fatal("newProjectCmd() returned nil")
	}

	if cmd.Use != "project [path]" {
		t.Errorf("Expected Use to be 'project [path]', got %s", cmd.Use)
	}

	if cmd.Short != "Create new project directory" {
		t.Errorf("Expected specific short description, got %s", cmd.Short)
	}
}

func TestNewCmdSubcommands(t *testing.T) {
	cmd := newCmd()

	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("new command should have subcommands")
	}

	var foundProject, foundList bool
	for _, subCmd := range subcommands {
		switch subCmd.Use {
		case "project [path]":
			foundProject = true
		case "list [path]":
			foundList = true
		}
	}

	if !foundProject {
		t.Error("Expected 'project' subcommand")
	}

	if !foundList {
		t.Error("Expected 'list' subcommand")
	}
}

func TestPathToName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"announcement.md", "Announcement"},
		{"the-big-news.md", "The big news"},
		{"simple", "Simple"},
		{"/path/to/file.txt", "File"},
		{"multi_word_file.md", "Multi word file"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := pathToName(tc.input)
			if result != tc.expected {
				t.Errorf("pathToName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestWriteTemplate(t *testing.T) {
	fs := afero.NewMemMapFs()

	content := "Hello {{.Name}}"
	data := map[string]string{"Name": "World"}
	path := "test.txt"

	err := writeTemplate(fs, path, content, data, true)
	if err != nil {
		t.Fatalf("writeTemplate failed: %v", err)
	}

	exists, err := afero.Exists(fs, path)
	if err != nil {
		t.Fatalf("Error checking file existence: %v", err)
	}
	if !exists {
		t.Error("File should exist after writeTemplate")
	}

	fileContent, err := afero.ReadFile(fs, path)
	if err != nil {
		t.Fatalf("Error reading file: %v", err)
	}

	expected := "Hello World"
	if string(fileContent) != expected {
		t.Errorf("File content = %q, want %q", string(fileContent), expected)
	}
}

func TestWriteTemplateFileExists(t *testing.T) {
	fs := afero.NewMemMapFs()
	path := "existing.txt"

	afero.WriteFile(fs, path, []byte("existing content"), 0644)

	err := writeTemplate(fs, path, "new content", nil, true)
	if err == nil {
		t.Error("Expected error when file exists")
	}

	expectedError := path + " already exists"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestRenderTemplate(t *testing.T) {
	content := "Hello {{.Name}}, welcome to {{.Place}}"
	data := map[string]string{
		"Name":  "Alice",
		"Place": "Wonderland",
	}

	result, err := renderTemplate(content, data)
	if err != nil {
		t.Fatalf("renderTemplate failed: %v", err)
	}

	expected := "Hello Alice, welcome to Wonderland"
	if result.String() != expected {
		t.Errorf("renderTemplate result = %q, want %q", result.String(), expected)
	}
}

func TestRenderTemplateNoData(t *testing.T) {
	content := "Simple content without templates"

	result, err := renderTemplate(content, nil)
	if err != nil {
		t.Fatalf("renderTemplate failed: %v", err)
	}

	if result.String() != content {
		t.Errorf("renderTemplate result = %q, want %q", result.String(), content)
	}
}

func TestNewProjectDirs(t *testing.T) {
	expected := []string{"content", "layouts", "lists", "themes"}

	if len(newProjectDirs) != len(expected) {
		t.Errorf("Expected %d project dirs, got %d", len(expected), len(newProjectDirs))
	}

	for i, dir := range expected {
		if i >= len(newProjectDirs) || newProjectDirs[i] != dir {
			t.Errorf("Expected project dir %q at index %d", dir, i)
		}
	}
}

func TestTemplateConstants(t *testing.T) {
	templates := map[string]string{
		"contentTemplate": contentTemplate,
		"listTemplate":    listTemplate,
		"configTemplate":  configTemplate,
	}

	for name, tmpl := range templates {
		t.Run(name, func(t *testing.T) {
			if tmpl == "" {
				t.Errorf("%s should not be empty", name)
			}

			_, err := template.New("test").Parse(tmpl)
			if err != nil {
				t.Errorf("%s is not a valid template: %v", name, err)
			}
		})
	}
}

func TestNewProjectCmdArgs(t *testing.T) {
	cmd := newProjectCmd()

	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"no args", []string{}, false},
		{"one arg", []string{"path"}, false},
		{"two args", []string{"path1", "path2"}, true},
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

func TestConfigFileConstant(t *testing.T) {
	if configFile != "config.toml" {
		t.Errorf("Expected configFile to be 'config.toml', got %q", configFile)
	}
}
