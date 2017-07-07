package mail

import (
	"github.com/spf13/afero"
	"path/filepath"
)

var (
	contentExts = []string{"md"}
	listExts    = []string{"yaml"}
)

var AppFs *fs

func SetFs(afs afero.Fs) {
	AppFs = &fs{afs}
}

type fs struct {
	afero.Fs
}

func (f *fs) ContentPath(name string) string {
	return filepath.Join(Config.ContentDir, name)
}

func (f *fs) ListPath(name string) string {
	return filepath.Join(Config.ListDir, name)
}

func (f *fs) layoutPath(name string) string {
	p := []string{filepath.Join(Config.LayoutDir, name)}
	if t := Config.Theme; t != "" {
		p = append(p, filepath.Join(Config.ThemeDir, t, p[0]))
	}
	return f.findFileWithExtension(p, []string{})
}

func (f *fs) findContentPath(name string) string {
	paths := []string{f.ContentPath(name)}
	return f.findFileWithExtension(paths, contentExts)
}

func (f *fs) findListPath(name string) string {
	paths := []string{f.ListPath(name)}
	return f.findFileWithExtension(paths, listExts)
}

/* This will look through all paths, match them with all extensions
   and return the first one it finds that exists */
func (f *fs) findFileWithExtension(paths, exts []string) string {
	for _, p := range paths {
		if f.isFile(p) {
			return p
		}
		for _, e := range exts {
			if pe := p + "." + e; f.isFile(pe) {
				return pe
			}
		}
	}
	return ""
}

func (f *fs) isFile(path string) bool {
	s, err := f.Stat(path)
	return err == nil && !s.IsDir()
}

func (f *fs) isDir(dir string) bool {
	s, err := f.Stat(dir)
	return err == nil && s.IsDir()
}
