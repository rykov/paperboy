package mail

import (
	"github.com/spf13/afero"
	"path/filepath"
)

var AppFs *fs

func SetFs(afs afero.Fs) {
	AppFs = &fs{afs}
}

type fs struct {
	afero.Fs
}

func (f *fs) contentPath(name string) string {
	return filepath.Join(Config.GetString("contentDir"), name)
}

func (f *fs) listPath(name string) string {
	return filepath.Join(Config.GetString("listDir"), name)
}

func (f *fs) layoutPath(name string) string {
	if p := filepath.Join(Config.GetString("layoutDir"), name); f.isFile(p) {
		return p
	}

	t := Config.GetString("theme")
	if p := filepath.Join("themes", t, "layouts", name); t != "" && f.isFile(p) {
		return p
	}

	return ""
}

func (f *fs) isFile(path string) bool {
	s, err := f.Stat(path)
	return err == nil && !s.IsDir()
}
