package config

import (
	"github.com/spf13/afero"
	"path/filepath"
)

var (
	contentExts = []string{"md"}
	listExts    = []string{"yaml"}
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
			if pe := p + "." + e; f.IsFile(pe) {
				return pe
			}
		}
	}
	return ""
}

func (f *Fs) IsFile(path string) bool {
	s, err := f.Stat(path)
	return err == nil && !s.IsDir()
}

func (f *Fs) isDir(dir string) bool {
	s, err := f.Stat(dir)
	return err == nil && s.IsDir()
}
