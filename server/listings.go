package server

import (
	"context"
	"os"
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
	err := r.cfg.AppFs.WalkContent(walkFn)
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
	err := r.cfg.AppFs.WalkLists(walkFn)
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
