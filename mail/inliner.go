package mail

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chris-ramon/douceur/inliner"
	"github.com/spf13/afero"
	"golang.org/x/net/html"
	"path/filepath"
	"strings"
)

func inlineStylesheets(layoutRoot, body string) (string, error) {
	if !AppFs.isDir(layoutRoot) {
		return inliner.Inline(body)
	}

	// Load body into goquery for some inlining fun
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return "", err
	}

	// Let's inline all stylesheet links into "<style/>" tags
	doc.Find("link[rel=stylesheet]").EachWithBreak(func(i int, s *goquery.Selection) bool {
		str, exists := s.Attr("href")
		if !exists {
			badHTML, _ := goquery.OuterHtml(s)
			err = fmt.Errorf("No href attribute for <link>: %s", badHTML)
			return false
		}

		var cssBytes []byte
		path := filepath.Join(layoutRoot, str)
		if cssBytes, err = afero.ReadFile(AppFs, path); err != nil {
			return false
		}

		// Insert nodes manually to avoid injection and escaping
		styleNode := &html.Node{Type: html.ElementNode, Data: "style"}
		textNode := &html.Node{Type: html.TextNode, Data: string(cssBytes)}
		styleNode.FirstChild, styleNode.LastChild = textNode, textNode
		s.ReplaceWithNodes(styleNode)
		return true
	})

	if err != nil {
		return "", err
	}

	if body, err = goquery.OuterHtml(doc.Selection); err != nil {
		return "", err
	}

	return inliner.Inline(body)
}
