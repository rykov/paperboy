package server

import (
	"github.com/rykov/paperboy/config"
	"github.com/spf13/afero"

	"context"
	"os"
	"path/filepath"
)

// ===== Campaigns listing resolver ======

func (r *Resolver) Campaigns(ctx context.Context) ([]*Campaign, error) {
	campaigns := []*Campaign{}
	walkFn := func(path, key string, fi os.FileInfo, err error) {
		campaigns = append(campaigns, &Campaign{
			subject: key, // FIXME
			param:   key,
		})
	}
	err := walkFilesByExt(config.Config.ContentDir, ".md", walkFn)
	return campaigns, err
}

type Campaign struct {
	param   string
	subject string
}

func (c *Campaign) Param() string {
	return c.param
}

func (c *Campaign) Subject() string {
	return c.subject
}

// ===== Recipient list listing resolver ======

func (r *Resolver) Lists(ctx context.Context) ([]*List, error) {
	lists := []*List{}
	walkFn := func(path, key string, fi os.FileInfo, err error) {
		lists = append(lists, &List{
			name:  key, // FIXME
			param: key,
		})
	}
	err := walkFilesByExt(config.Config.ListDir, ".yaml", walkFn)
	return lists, err
}

type List struct {
	param string
	name  string
}

func (c *List) Param() string {
	return c.param
}

func (c *List) Name() string {
	return c.name
}

// Iteration helper to find all files with a certain extension in a directory
func walkFilesByExt(dir, ext string, walkFn func(path, key string, fi os.FileInfo, err error)) error {
	return afero.Walk(config.Config.AppFs, dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}

		pathExt := filepath.Ext(path)
		if pathExt != ext {
			return nil
		}

		// Remove dir prefix and extension
		key, _ := filepath.Rel(dir, path)
		key = key[:len(key)-len(pathExt)]
		walkFn(path, key, fi, nil)
		return nil
	})
}
