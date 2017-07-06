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
	Config.SetFs(afs)
	AppFs = &fs{afs}
}

type fs struct {
	afero.Fs
}

func (f *fs) ContentPath(name string) string {
	p := filepath.Join(Config.GetString("contentDir"), name)
	return f.findFileWithExtension([]string{p}, contentExts)
}

func (f *fs) ListPath(name string) string {
	p := filepath.Join(Config.GetString("listDir"), name)
	return f.findFileWithExtension([]string{p}, listExts)
}

func (f *fs) layoutPath(name string) string {
	p := []string{filepath.Join(Config.GetString("layoutDir"), name)}
	if t := Config.GetString("theme"); t != "" {
		p = append(p, filepath.Join(Config.GetString("themesDir"), t, p[0]))
	}
	return f.findFileWithExtension(p, []string{})
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
