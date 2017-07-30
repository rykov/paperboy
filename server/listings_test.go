package server

import (
	//"github.com/google/go-cmp/cmp"
	//"github.com/jordan-wright/email"
	//"github.com/neelance/graphql-go"
	"github.com/rykov/paperboy/mail"
	"github.com/spf13/afero"

	//"context"
	"encoding/json"
	//"os"
	//"strings"
	"testing"
	//"time"
)

type gqlCampaign struct {
	Subject string
	Param   string
}

func TestCampaignsQuery(t *testing.T) {
	fs := mail.AppFs
	afero.WriteFile(fs, fs.ContentPath("c1.md"), []byte("# Hello"), 0644)
	afero.WriteFile(fs, fs.ContentPath("sub/c2.md"), []byte("# World"), 0644)
	afero.WriteFile(fs, fs.ContentPath("skip.txt"), []byte("Not-content"), 0644)

	response := issueGraphQLQuery(`{
		campaigns {
			subject
			param
		}
	}`)

	if errs := response.Errors; len(errs) > 0 {
		t.Fatalf("GraphQL errors %+v", errs)
	}

	resp := struct {
		Campaigns []gqlCampaign
	}{}

	if err := json.Unmarshal(response.Data, &resp); err != nil {
		t.Fatalf("JSON unmarshal error: %s", err)
	}

	// Check to make sure we are listing the right files
	if len(resp.Campaigns) != 2 {
		t.Fatalf("Incorrect number of campaigns returned")
	}

	// Check the first campaign from root directory
	if c1 := resp.Campaigns[0]; c1.Param != "c1" {
		t.Fatalf("Invalid param for \"c1\" campaign: %s", c1.Param)
	} else if c1.Subject != "c1" {
		t.Fatalf("Invalid subject for \"c1\" campaign: %s", c1.Subject)
	}

	// Check the second campaign from root directory
	if c2 := resp.Campaigns[1]; c2.Param != "sub/c2" {
		t.Fatalf("Invalid param for \"c2\" campaign: %s", c2.Param)
	} else if c2.Subject != "sub/c2" {
		t.Fatalf("Invalid subject for \"c2\" campaign: %s", c2.Subject)
	}
}

type gqlList struct {
	Name  string
	Param string
}

func TestListsQuery(t *testing.T) {
	fs := mail.AppFs
	afero.WriteFile(fs, fs.ListPath("l1.yaml"), []byte("---"), 0644)
	afero.WriteFile(fs, fs.ListPath("sub/l2.yaml"), []byte("---"), 0644)
	afero.WriteFile(fs, fs.ListPath("skip.txt"), []byte("Not-content"), 0644)

	response := issueGraphQLQuery(`{
		lists {
			name
			param
		}
	}`)

	if errs := response.Errors; len(errs) > 0 {
		t.Fatalf("GraphQL errors %+v", errs)
	}

	resp := struct {
		Lists []gqlList
	}{}

	if err := json.Unmarshal(response.Data, &resp); err != nil {
		t.Fatalf("JSON unmarshal error: %s", err)
	}

	// Check to make sure we are listing the right files
	if len(resp.Lists) != 2 {
		t.Fatalf("Incorrect number of lists returned")
	}

	// Check the first campaign from root directory
	if l1 := resp.Lists[0]; l1.Param != "l1" {
		t.Fatalf("Invalid param for \"l1\" list: %s", l1.Param)
	} else if l1.Name != "l1" {
		t.Fatalf("Invalid name for \"l1\" list: %s", l1.Name)
	}

	// Check the second campaign from root directory
	if l2 := resp.Lists[1]; l2.Param != "sub/l2" {
		t.Fatalf("Invalid param for \"l2\" list: %s", l2.Param)
	} else if l2.Name != "sub/l2" {
		t.Fatalf("Invalid name for \"l2\" list: %s", l2.Name)
	}
}
