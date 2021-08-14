// +build !withUI

package ui

import (
	"io/fs"
	"testing/fstest"
	"time"
)

var FS fs.FS

func init() {
	FS = fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data:    []byte("UI was not built/embedded"),
			Mode:    fs.FileMode(0644),
			ModTime: time.Now(),
		},
	}
}
