package client

import (
	"github.com/moby/patternmatcher"
	"resty.dev/v3"

	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
)

const sendGQL = `
  mutation sendCampaign($campaign: String!, $list: String!) {
    sendCampaign(campaign: $campaign, list: $list)
  }
`

type client struct {
	context   context.Context
	serverURL string
}

func New(ctx context.Context, url string) *client {
	return &client{ctx, url}
}

func (c *client) Send(args SendArgs) error {
	pr, ct := streamZipToMultipart(args, func(mw *multipart.Writer) error {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Type", "application/json")
		part, err := mw.CreatePart(header)
		if err != nil {
			return err
		}

		return json.NewEncoder(part).Encode(map[string]any{
			"operationName": "sendCampaign",
			"query":         sendGQL,
			"variables": map[string]any{
				"campaign": args.Campaign,
				"list":     args.List,
			},
		})
	})

	// Capture GraphQL errors
	var output gqlErrorResponse

	// Prepare Resty client and request
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", ct).
		SetResult(&output).
		SetBody(pr).
		Post(c.serverURL)

	// non‐2xx → treat as error
	if err != nil {
		return err
	} else if e := output.Errors; len(e) > 0 {
		return fmt.Errorf("server error: %s", e[0].Message)
	} else if resp.IsError() {
		return fmt.Errorf("server returned %s: %s",
			resp.Status(),
			resp.String(),
		)
	}

	return nil
}

// Common GQL error response
type SendArgs struct {
	ProjectIgnores []string
	ProjectPath    string
	Campaign       string
	List           string
}

// Common GQL error response
type gqlErrorResponse struct {
	Errors []struct {
		Path    []string
		Message string
	}
}

// Create a pipe: the ZIP writer writes to pw, and resty reads from pr.
func streamZipToMultipart(args SendArgs, callback func(*multipart.Writer) error) (io.Reader, string) {
	// Create a pipe: multipart.Writer writes to pw, Resty reads from pr.
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	dirPath := args.ProjectPath
	ignores := args.ProjectIgnores

	go func() {
		defer pw.Close()
		defer mw.Close()

		// Create a single zip-file part
		header := make(textproto.MIMEHeader)
		header.Set("Content-Type", "application/zip") // <- here you set it!
		part, err := mw.CreatePart(header)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		// Stream the ZIP into that part:
		zw := zip.NewWriter(part)

		// Walk through the directory respecting "ClientIgnores" filtering configuration
		zipErr := walkWithIgnores(dirPath, ignores, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			// Create a zip header based on file info
			info, err := d.Info()
			if err != nil {
				return err
			}
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// Compute the relative path in the zip archive
			relPath, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}

			// Normalize to "/" separators
			header.Name = filepath.ToSlash(relPath)
			header.Method = zip.Deflate

			// Create the entry and copy file data
			w, err := zw.CreateHeader(header)
			if err != nil {
				return err
			}
			if _, err := io.Copy(w, f); err != nil {
				return err
			}

			return nil
		})

		// Close the zip writer (flushes headers)
		mwErr := errors.Join(zipErr, zw.Close())

		// Callback to add more parts
		if callback != nil {
			mwErr = errors.Join(mwErr, callback(mw))
		}

		// Propagate any error to the reader side
		if mwErr != nil {
			pw.CloseWithError(mwErr)
		} else {
			pw.Close()
		}
	}()

	return pr, mw.FormDataContentType()
}

// walkWithIgnores does a ".dockerignore"-like filtering when walking a directory
func walkWithIgnores(root string, ignores []string, fn fs.WalkDirFunc) error {
	matcher, err := patternmatcher.New(ignores)
	if err != nil {
		return fmt.Errorf("invalid ignore pattern: %w", err)
	}

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Propagate internal errors
		if err != nil {
			return fn(path, d, err)
		}

		// Compute a path relative to root for pattern matching
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)

		// Skip the root itself from matching
		if rel != "." {
			// patternmatcher expects slash-separated paths
			matched, err := matcher.Matches(rel)
			if err != nil {
				return fmt.Errorf("matching %q: %w", rel, err)
			}
			if matched {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
		}

		// Not ignored
		return fn(path, d, err)
	})
}
