package config

import (
	"io/fs"
	"path/filepath"
	"sort"
	"testing"

	"github.com/spf13/afero"
)

func TestFsPathMethods(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)

	// Override default paths for testing
	cfg.ContentDir = "articles"
	cfg.ListDir = "subscribers"
	cfg.LayoutDir = "templates"
	cfg.ThemeDir = "themes"

	// Test ContentPath
	if path := cfg.AppFs.ContentPath("newsletter"); path != "articles/newsletter" {
		t.Errorf("Expected ContentPath 'articles/newsletter', got '%s'", path)
	}

	// Test ListPath
	if path := cfg.AppFs.ListPath("vips"); path != "subscribers/vips" {
		t.Errorf("Expected ListPath 'subscribers/vips', got '%s'", path)
	}

	// Test LayoutPath without theme - create the file first
	afero.WriteFile(memFs, "templates/default.html", []byte("layout"), 0644)
	if path := cfg.AppFs.LayoutPath("default.html"); path != "templates/default.html" {
		t.Errorf("Expected LayoutPath 'templates/default.html', got '%s'", path)
	}
}

func TestFsLayoutPathWithTheme(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)

	cfg.LayoutDir = "layouts"
	cfg.ThemeDir = "themes"
	cfg.Theme = "modern"

	// Create theme layout file
	themePath := "themes/modern/layouts/custom.html"
	afero.WriteFile(memFs, themePath, []byte("theme layout"), 0644)

	// Test theme layout is found
	if path := cfg.AppFs.LayoutPath("custom.html"); path != themePath {
		t.Errorf("Expected theme layout path '%s', got '%s'", themePath, path)
	}

	// Test fallback to regular layout when theme doesn't have the file
	regularPath := "layouts/fallback.html"
	afero.WriteFile(memFs, regularPath, []byte("regular layout"), 0644)

	if path := cfg.AppFs.LayoutPath("fallback.html"); path != regularPath {
		t.Errorf("Expected fallback to regular layout '%s', got '%s'", regularPath, path)
	}
}

func TestFsFindContentPath(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ContentDir = "content"

	// Create content files with different extensions
	afero.WriteFile(memFs, "content/newsletter.md", []byte("markdown content"), 0644)
	afero.WriteFile(memFs, "content/announcement", []byte("no extension"), 0644)

	// Test finding .md file by name without extension
	if path := cfg.AppFs.FindContentPath("newsletter"); path != "content/newsletter.md" {
		t.Errorf("Expected to find 'content/newsletter.md', got '%s'", path)
	}

	// Test finding exact file with extension
	if path := cfg.AppFs.FindContentPath("newsletter.md"); path != "content/newsletter.md" {
		t.Errorf("Expected to find 'content/newsletter.md', got '%s'", path)
	}

	// Test finding file without extension
	if path := cfg.AppFs.FindContentPath("announcement"); path != "content/announcement" {
		t.Errorf("Expected to find 'content/announcement', got '%s'", path)
	}

	// Test file that doesn't exist
	if path := cfg.AppFs.FindContentPath("missing"); path != "" {
		t.Errorf("Expected empty path for missing file, got '%s'", path)
	}
}

func TestFsFindListPath(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ListDir = "lists"

	// Create list files with different extensions
	afero.WriteFile(memFs, "lists/subscribers.yaml", []byte("yaml list"), 0644)
	afero.WriteFile(memFs, "lists/vips.yml", []byte("yml list"), 0644)
	afero.WriteFile(memFs, "lists/exact-name", []byte("no extension"), 0644)

	// Test finding .yaml file
	if path := cfg.AppFs.FindListPath("subscribers"); path != "lists/subscribers.yaml" {
		t.Errorf("Expected to find 'lists/subscribers.yaml', got '%s'", path)
	}

	// Test finding .yml file
	if path := cfg.AppFs.FindListPath("vips"); path != "lists/vips.yml" {
		t.Errorf("Expected to find 'lists/vips.yml', got '%s'", path)
	}

	// Test finding exact file
	if path := cfg.AppFs.FindListPath("exact-name"); path != "lists/exact-name" {
		t.Errorf("Expected to find 'lists/exact-name', got '%s'", path)
	}

	// Test file that doesn't exist
	if path := cfg.AppFs.FindListPath("missing"); path != "" {
		t.Errorf("Expected empty path for missing file, got '%s'", path)
	}
}

func TestFsFileExtensionPriority(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ListDir = "lists"

	// Create files with different extensions - exact match should have priority
	afero.WriteFile(memFs, "lists/test", []byte("exact file"), 0644)
	afero.WriteFile(memFs, "lists/test.yaml", []byte("yaml file"), 0644)
	afero.WriteFile(memFs, "lists/test.yml", []byte("yml file"), 0644)

	// Exact match should have priority over extensions
	if path := cfg.AppFs.FindListPath("test"); path != "lists/test" {
		t.Errorf("Expected exact match 'lists/test', got '%s'", path)
	}

	// When exact doesn't exist, extension should be found
	memFs.Remove("lists/test")
	if path := cfg.AppFs.FindListPath("test"); path != "lists/test.yaml" && path != "lists/test.yml" {
		t.Errorf("Expected to find extension file, got '%s'", path)
	}
}

func TestFsIsFile(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)

	// Create file and directory
	afero.WriteFile(memFs, "test.txt", []byte("content"), 0644)
	memFs.MkdirAll("testdir", 0755)

	// Test file detection
	if !cfg.AppFs.IsFile("test.txt") {
		t.Error("Should detect file as file")
	}

	// Test directory detection
	if cfg.AppFs.IsFile("testdir") {
		t.Error("Should not detect directory as file")
	}

	// Test non-existent path
	if cfg.AppFs.IsFile("missing") {
		t.Error("Should not detect missing path as file")
	}
}

func TestFsWalkContent(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ContentDir = "content"

	// Create content structure
	afero.WriteFile(memFs, "content/post1.md", []byte("post 1"), 0644)
	afero.WriteFile(memFs, "content/post2.md", []byte("post 2"), 0644)
	afero.WriteFile(memFs, "content/subdir/post3.md", []byte("post 3"), 0644)
	afero.WriteFile(memFs, "content/readme.txt", []byte("readme"), 0644) // Should be ignored
	memFs.MkdirAll("content/emptydir", 0755)                             // Should be ignored

	var found []string
	var keys []string

	err := cfg.AppFs.WalkContent(func(path, key string, fi fs.FileInfo, walkErr error) {
		if walkErr != nil {
			t.Errorf("Walk error: %v", walkErr)
			return
		}
		found = append(found, path)
		keys = append(keys, key)
	})

	if err != nil {
		t.Fatalf("WalkContent failed: %v", err)
	}

	// Should find 3 .md files
	if len(found) != 3 {
		t.Errorf("Expected 3 content files, got %d: %v", len(found), found)
	}

	// Sort for consistent testing
	sort.Strings(found)
	sort.Strings(keys)

	expectedPaths := []string{
		"content/post1.md",
		"content/post2.md",
		"content/subdir/post3.md",
	}
	sort.Strings(expectedPaths)

	for i, expected := range expectedPaths {
		if i >= len(found) || found[i] != expected {
			t.Errorf("Expected path '%s', got '%s'", expected, found[i])
		}
	}

	expectedKeys := []string{"post1", "post2", "subdir/post3"}
	sort.Strings(expectedKeys)

	for i, expected := range expectedKeys {
		if i >= len(keys) || keys[i] != expected {
			t.Errorf("Expected key '%s', got '%s'", expected, keys[i])
		}
	}
}

func TestFsWalkLists(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ListDir = "lists"

	// Create list structure
	afero.WriteFile(memFs, "lists/subscribers.yaml", []byte("subs"), 0644)
	afero.WriteFile(memFs, "lists/vips.yml", []byte("vips"), 0644)
	afero.WriteFile(memFs, "lists/groups/team.yaml", []byte("team"), 0644)
	afero.WriteFile(memFs, "lists/config.toml", []byte("config"), 0644) // Should be ignored

	var found []string
	var keys []string

	err := cfg.AppFs.WalkLists(func(path, key string, fi fs.FileInfo, walkErr error) {
		if walkErr != nil {
			t.Errorf("Walk error: %v", walkErr)
			return
		}
		found = append(found, path)
		keys = append(keys, key)
	})

	if err != nil {
		t.Fatalf("WalkLists failed: %v", err)
	}

	// Should find 3 yaml/yml files
	if len(found) != 3 {
		t.Errorf("Expected 3 list files, got %d: %v", len(found), found)
	}

	// Check that only yaml/yml files are found
	for _, path := range found {
		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			t.Errorf("Should only find yaml/yml files, got: %s", path)
		}
	}
}

func TestFsWalkEmpty(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ContentDir = "content"

	// Create empty content directory
	memFs.MkdirAll("content", 0755)

	var found []string
	err := cfg.AppFs.WalkContent(func(path, key string, fi fs.FileInfo, walkErr error) {
		found = append(found, path)
	})

	if err != nil {
		t.Fatalf("WalkContent failed: %v", err)
	}

	// Should find no files
	if len(found) != 0 {
		t.Errorf("Expected no files in empty directory, got %d: %v", len(found), found)
	}
}

func TestFsWalkNonexistentDirectory(t *testing.T) {
	memFs := afero.NewMemMapFs()
	cfg, _ := LoadConfigFs(t.Context(), memFs)
	cfg.ContentDir = "nonexistent"

	var found []string
	_ = cfg.AppFs.WalkContent(func(path, key string, fi fs.FileInfo, walkErr error) {
		found = append(found, path)
	})

	// Should handle gracefully when directory doesn't exist
	// Note: afero Walk returns no error for missing directories, it just finds nothing
	if len(found) != 0 {
		t.Errorf("Expected no files when directory doesn't exist, got %d: %v", len(found), found)
	}
}
