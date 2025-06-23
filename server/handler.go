package server

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rykov/paperboy/config"
	"github.com/urfave/negroni/v3"

	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	ctxZipFileKey = iota
)

// MustSchemaHandler wraps the multipart GraphQL handler
func MustSchemaHandler(schema string, resolver any) http.Handler {
	s := graphql.MustParseSchema(schema, resolver)
	return &handler{Handler: &relay.Handler{s}}
}

type handler struct {
	*relay.Handler
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		h.Handler.ServeHTTP(w, r)
	}

	// Parse form to capture request & zip
	params, file, err := parseMultipartGQL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer os.Remove(file.Name()) // clean up

	// Store zip file into context to handle in GraphQL
	ctx := context.WithValue(r.Context(), ctxZipFileKey, file)

	// Execute GraphQL query
	response := h.Schema.Exec(ctx, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with results
	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

// Accessor for zipfile from context
func RequestZipFile(ctx context.Context) (*os.File, bool) {
	f, ok := ctx.Value(ctxZipFileKey).(*os.File)
	return f, ok
}

// Iterate through form and capture JSON and ZIP files for processing
func parseMultipartGQL(r *http.Request) (params *gqlRequestParams, file *os.File, err error) {
	// Get a streaming reader for the parts
	mr, err := r.MultipartReader()
	if err != nil {
		return nil, nil, err
	}

	// Iterate through each part
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			return params, file, nil
		} else if err != nil {
			return nil, nil, err
		}

		// Inspect incoming headers
		ct := part.Header.Get("Content-Type")
		//fieldName := part.FormName()
		//        fileName := part.FileName() // empty if this part isn't a file

		switch {
		case ct == "application/json":
			params = &gqlRequestParams{}
			if err := json.NewDecoder(part).Decode(params); err != nil {
				return nil, nil, err
			}

		case ct == "application/zip":
			file, err = os.CreateTemp("", "paperboy-zip")
			if err != nil {
				return nil, nil, err
			}

			_, err1 := io.Copy(file, part)
			_, err2 := file.Seek(0, io.SeekStart)

			if err := errors.Join(err1, err2); err != nil {
				os.Remove(file.Name()) // clean up
				return nil, nil, err
			}

		default:
			// unknown type—just skip it
			if _, err := io.Copy(io.Discard, part); err != nil {
				return nil, nil, err
			}
		}
	}
}

type gqlRequestParams struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables"`
}

// WithMiddleware wraps the handler with logging, recovery, etc
func WithMiddleware(h http.Handler, cfg *config.AConfig) http.Handler {
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())

	// Add basic authentication
	if cfg != nil && cfg.ServerAuth != "" {
		expU, expP, _ := strings.Cut(cfg.ServerAuth, ":")
		n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			if u, p, ok := r.BasicAuth(); ok {
				okU := subtle.ConstantTimeCompare([]byte(u), []byte(expU)) == 1
				okP := subtle.ConstantTimeCompare([]byte(p), []byte(expP)) == 1
				if okU && okP {
					next(rw, r)
					return
				}
			}
			s := http.StatusUnauthorized
			http.Error(rw, http.StatusText(s), s)
		})
	}

	n.UseHandler(h)
	return n
}
