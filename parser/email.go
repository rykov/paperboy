// Copyright 2016n The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

const (
	// TODO(bep) Do we really have to export these?

	// HTMLLead identifies the start of HTML documents.
	HTMLLead = "<"
	// YAMLLead identifies the start of YAML frontmatter.
	YAMLLead = "-"
	// YAMLDelimUnix identifies the end of YAML front matter on Unix.
	YAMLDelimUnix = "---\n"
	// YAMLDelimDOS identifies the end of YAML front matter on Windows.
	YAMLDelimDOS = "---\r\n"
	// YAMLDelim identifies the YAML front matter delimiter.
	YAMLDelim = "---"
	// TOMLLead identifies the start of TOML front matter.
	TOMLLead = "+"
	// TOMLDelimUnix identifies the end of TOML front matter on Unix.
	TOMLDelimUnix = "+++\n"
	// TOMLDelimDOS identifies the end of TOML front matter on Windows.
	TOMLDelimDOS = "+++\r\n"
	// TOMLDelim identifies the TOML front matter delimiter.
	TOMLDelim = "+++"
	// JSONLead identifies the start of JSON frontmatter.
	JSONLead = "{"
	// HTMLCommentStart identifies the start of HTML comment.
	HTMLCommentStart = "<!--"
	// HTMLCommentEnd identifies the end of HTML comment.
	HTMLCommentEnd = "-->"
	// BOM Unicode byte order marker
	BOM = '\ufeff'
)

var (
	delims = regexp.MustCompile(
		"^(" + regexp.QuoteMeta(YAMLDelim) + `\s*\n|` + regexp.QuoteMeta(TOMLDelim) + `\s*\n|` + regexp.QuoteMeta(JSONLead) + ")",
	)
)

// Page represents a parsed content page.
type Email interface {
	FrontMatter() []byte
	Content() []byte
	IsRenderable() bool
	Metadata() (interface{}, error)
}

type page struct {
	render      bool
	frontmatter []byte
	content     []byte
}

func (p *page) Content() []byte {
	return p.content
}

func (p *page) FrontMatter() []byte {
	return p.frontmatter
}

func (p *page) IsRenderable() bool {
	return p.render
}

func (p *page) Metadata() (meta interface{}, err error) {
	frontmatter := p.FrontMatter()

	if len(frontmatter) != 0 {
		fm := DetectFrontMatter(rune(frontmatter[0]))
		meta, err = fm.Parse(frontmatter)
		if err != nil {
			return
		}
	}
	return
}

// ReadFrom reads the content from an io.Reader and constructs a page.
func ReadFrom(r io.Reader) (p Email, err error) {
	reader := bufio.NewReader(r)

	// chomp BOM and assume UTF-8
	if err = chompBOM(reader); err != nil && err != io.EOF {
		return
	}
	if err = chompWhitespace(reader); err != nil && err != io.EOF {
		return
	}
	if err = chompFrontmatterStartComment(reader); err != nil && err != io.EOF {
		return
	}

	firstLine, err := peekLine(reader)
	if err != nil && err != io.EOF {
		return
	}

	newp := new(page)
	newp.render = shouldRender(firstLine)

	if newp.render && isFrontMatterDelim(firstLine) {
		left, right, err := determineDelims(firstLine)
		if err != nil {
			return nil, err
		}
		fm, err := extractFrontMatterDelims(reader, left, right)
		if err != nil {
			return nil, err
		}
		newp.frontmatter = fm
	}

	content, err := extractContent(reader)
	if err != nil {
		return nil, err
	}

	newp.content = content

	return newp, nil
}

func chompBOM(r io.RuneScanner) (err error) {
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			return err
		}
		if c != BOM {
			r.UnreadRune()
			return nil
		}
	}
}

func chompWhitespace(r io.RuneScanner) (err error) {
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			return err
		}
		if !unicode.IsSpace(c) {
			r.UnreadRune()
			return nil
		}
	}
}

func chompFrontmatterStartComment(r *bufio.Reader) (err error) {
	candidate, err := r.Peek(32)
	if err != nil {
		return err
	}

	str := string(candidate)
	if strings.HasPrefix(str, HTMLCommentStart) {
		lineEnd := strings.IndexAny(str, "\n")
		if lineEnd == -1 {
			//TODO: if we can't find it, Peek more?
			return nil
		}
		testStr := strings.TrimSuffix(str[0:lineEnd], "\r")
		if strings.Contains(testStr, HTMLCommentEnd) {
			return nil
		}
		buf := make([]byte, lineEnd)
		if _, err = r.Read(buf); err != nil {
			return
		}
		if err = chompWhitespace(r); err != nil {
			return err
		}
	}

	return nil
}

func chompFrontmatterEndComment(r *bufio.Reader) (err error) {
	candidate, err := r.Peek(32)
	if err != nil {
		return err
	}

	str := string(candidate)
	lineEnd := strings.IndexAny(str, "\n")
	if lineEnd == -1 {
		return nil
	}
	testStr := strings.TrimSuffix(str[0:lineEnd], "\r")
	if strings.Contains(testStr, HTMLCommentStart) {
		return nil
	}

	//TODO: if we can't find it, Peek more?
	if strings.HasSuffix(testStr, HTMLCommentEnd) {
		buf := make([]byte, lineEnd)
		if _, err = r.Read(buf); err != nil {
			return
		}
		if err = chompWhitespace(r); err != nil {
			return err
		}
	}

	return nil
}

func peekLine(r *bufio.Reader) (line []byte, err error) {
	firstFive, err := r.Peek(5)
	if err != nil {
		return
	}
	idx := bytes.IndexByte(firstFive, '\n')
	if idx == -1 {
		return firstFive, nil
	}
	idx++ // include newline.
	return firstFive[:idx], nil
}

func shouldRender(lead []byte) (frontmatter bool) {
	if len(lead) <= 0 {
		return
	}

	if bytes.Equal(lead[:1], []byte(HTMLLead)) {
		return
	}
	return true
}

func isFrontMatterDelim(data []byte) bool {
	return delims.Match(data)
}

func determineDelims(firstLine []byte) (left, right []byte, err error) {
	switch len(firstLine) {
	case 5:
		fallthrough
	case 4:
		if firstLine[0] == YAMLLead[0] {
			return []byte(YAMLDelim), []byte(YAMLDelim), nil
		}
		return []byte(TOMLDelim), []byte(TOMLDelim), nil
	case 3:
		fallthrough
	case 2:
		fallthrough
	case 1:
		return []byte(JSONLead), []byte("}"), nil
	default:
		err := fmt.Errorf("unable to determine delims from %q", firstLine)
		return nil, nil, err
	}
}

// extractFrontMatterDelims takes a frontmatter from the content bufio.Reader.
// Beginning white spaces of the bufio.Reader must be trimmed before call this
// function.
func extractFrontMatterDelims(r *bufio.Reader, left, right []byte) (fm []byte, err error) {
	var (
		c         byte
		buf       bytes.Buffer
		level     int
		sameDelim = bytes.Equal(left, right)
	)

	// Frontmatter must start with a delimiter. To check it first,
	// pre-reads beginning delimiter length - 1 bytes from Reader
	for i := 0; i < len(left)-1; i++ {
		if c, err = r.ReadByte(); err != nil {
			return nil, fmt.Errorf("unable to read frontmatter at filepos %d: %s", buf.Len(), err)
		}
		if err = buf.WriteByte(c); err != nil {
			return nil, err
		}
	}

	// Reads a character from Reader one by one and checks it matches the
	// last character of one of delimiters to find the last character of
	// frontmatter. If it matches, makes sure it contains the delimiter
	// and if so, also checks it is followed by CR+LF or LF when YAML,
	// TOML case. In JSON case, nested delimiters must be parsed and it
	// is expected that the delimiter only contains one character.
	for {
		if c, err = r.ReadByte(); err != nil {
			return nil, fmt.Errorf("unable to read frontmatter at filepos %d: %s", buf.Len(), err)
		}
		if err = buf.WriteByte(c); err != nil {
			return nil, err
		}

		switch c {
		case left[len(left)-1]:
			if sameDelim { // YAML, TOML case
				if bytes.HasSuffix(buf.Bytes(), left) && (buf.Len() == len(left) || buf.Bytes()[buf.Len()-len(left)-1] == '\n') {
				nextByte:
					c, err = r.ReadByte()
					if err != nil {
						// It is ok that the end delimiter ends with EOF
						if err != io.EOF || level != 1 {
							return nil, fmt.Errorf("unable to read frontmatter at filepos %d: %s", buf.Len(), err)
						}
					} else {
						switch c {
						case '\n':
							// ok
						case ' ':
							// Consume this byte and try to match again
							goto nextByte
						case '\r':
							if err = buf.WriteByte(c); err != nil {
								return nil, err
							}
							if c, err = r.ReadByte(); err != nil {
								return nil, fmt.Errorf("unable to read frontmatter at filepos %d: %s", buf.Len(), err)
							}
							if c != '\n' {
								return nil, fmt.Errorf("frontmatter delimiter must be followed by CR+LF or LF but those can't be found at filepos %d", buf.Len())
							}
						default:
							return nil, fmt.Errorf("frontmatter delimiter must be followed by CR+LF or LF but those can't be found at filepos %d", buf.Len())
						}
						if err = buf.WriteByte(c); err != nil {
							return nil, err
						}
					}
					if level == 0 {
						level = 1
					} else {
						level = 0
					}
				}
			} else { // JSON case
				level++
			}
		case right[len(right)-1]: // JSON case only reaches here
			level--
		}

		if level == 0 {
			// Consumes white spaces immediately behind frontmatter
			if err = chompWhitespace(r); err != nil && err != io.EOF {
				return nil, err
			}
			if err = chompFrontmatterEndComment(r); err != nil && err != io.EOF {
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}
}

func extractContent(r io.Reader) (content []byte, err error) {
	wr := new(bytes.Buffer)
	if _, err = wr.ReadFrom(r); err != nil {
		return
	}
	return wr.Bytes(), nil
}
