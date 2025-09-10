package parser

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test function
func TestReadFrom(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedError     error
		expectedEmail     *page
		expectedMetaError error
		expectedMeta      interface{}
	}{
		{
			name: "Valid input with frontmatter",
			input: `---
title: Test Email
---
This is the content of the email.`,
			expectedError: nil,
			expectedEmail: &page{
				content: []byte("This is the content of the email."),
				frontmatter: []byte(`---
title: Test Email
---
`),
				render: true,
			},
			expectedMetaError: nil,
			expectedMeta:      map[string]interface{}(map[string]interface{}{"title": "Test Email"}),
		},
		{
			name:          "Valid input without frontmatter",
			input:         `This is the content of the email without frontmatter.`,
			expectedError: nil,
			expectedEmail: &page{
				content:     []byte("This is the content of the email without frontmatter."),
				frontmatter: []byte(nil),
				render:      true,
			},
			expectedMetaError: nil,
		},
		{
			name: "Error on reading",
			input: `---
tittle: "not closed
---
This will cause an error`,
			expectedError: nil,
			expectedEmail: &page{
				content: []byte("This will cause an error"),
				frontmatter: []byte(`---
tittle: "not closed
---
`),
				render: true,
			},
			expectedMetaError: errors.New("yaml: line 2: found unexpected document indicator"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))

			// Call the ReadFrom function
			email, err := ReadFrom(reader)

			// Assertions
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEmail.content, email.Content())
				assert.Equal(t, tt.expectedEmail.frontmatter, email.FrontMatter())
				assert.Equal(t, tt.expectedEmail.render, email.IsRenderable())
			}

			// Check metadata
			meta, err := email.Metadata()
			// Assertions
			if tt.expectedMetaError != nil {
				assert.Error(t, err)
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), tt.expectedMetaError.Error())
			} else {
				assert.Equal(t, tt.expectedMeta, meta)
			}
		})
	}
}
