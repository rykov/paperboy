package config

import (
	"io/fs"
	"path/filepath"
	"slices"

	"github.com/spf13/afero"
)

var (
	contentExts = []string{".md"}
	listExts    = []string{".yaml", ".yml"}
)

type Fs struct {
	Config *AConfig
	afero.Fs
}

func (f *Fs) ContentPath(name string) string {
	return filepath.Join(f.Config.ContentDir, name)
}

func (f *Fs) ListPath(name string) string {
	return filepath.Join(f.Config.ListDir, name)
}

func (f *Fs) LayoutPath(name string) string {
	p := []string{filepath.Join(f.Config.LayoutDir, name)}
	if t := f.Config.Theme; t != "" {
		p = append(p, filepath.Join(f.Config.ThemeDir, t, p[0]))
	}
	return f.findFileWithExtension(p, []string{})
}

func (f *Fs) AssetPath(name string) string {
	return filepath.Join(f.Config.AssetDir, name)
}

func (f *Fs) FindContentPath(name string) string {
	paths := []string{f.ContentPath(name)}
	return f.findFileWithExtension(paths, contentExts)
}

func (f *Fs) FindListPath(name string) string {
	paths := []string{f.ListPath(name)}
	return f.findFileWithExtension(paths, listExts)
}

/*
This will look through all paths, match them with all extensions
and return the first one it finds that exists
*/
func (f *Fs) findFileWithExtension(paths, exts []string) string {
	for _, p := range paths {
		if f.IsFile(p) {
			return p
		}
		for _, e := range exts {
			if pe := p + e; f.IsFile(pe) {
				return pe
			}
		}
	}
	return ""
}

func (fs *Fs) WalkContent(walkFn func(path, key string, fi fs.FileInfo, err error)) error {
	return fs.walkFilesByExts(fs.Config.ContentDir, contentExts, walkFn)
}

func (fs *Fs) WalkLists(walkFn func(path, key string, fi fs.FileInfo, err error)) error {
	return fs.walkFilesByExts(fs.Config.ListDir, listExts, walkFn)
}

// Iteration helper to find all files with multiple possible extensions in a directory
func (pfs *Fs) walkFilesByExts(dir string, exts []string, walkFn func(path, key string, fi fs.FileInfo, err error)) error {
	return afero.Walk(pfs, dir, func(path string, fi fs.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}

		pathExt := filepath.Ext(path)
		if !slices.Contains(exts, pathExt) {
			return nil
		}

		// Remove dir prefix and extension
		key, _ := filepath.Rel(dir, path)
		key = key[:len(key)-len(pathExt)]
		walkFn(path, key, fi, nil)
		return nil
	})
}

func (f *Fs) IsFile(path string) bool {
	s, err := f.Stat(path)
	return err == nil && !s.IsDir()
}

func (f *Fs) isDir(dir string) bool {
	s, err := f.Stat(dir)
	return err == nil && s.IsDir()
}
