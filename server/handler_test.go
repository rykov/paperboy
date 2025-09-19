package server

import (
	"github.com/google/go-cmp/cmp"

	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseMultipartGQL(t *testing.T) {
	// Prepare JSON and ZIP content
	zipData := []byte("PK\x03\x04dummyzipcontent")
	jsonReq := map[string]any{
		"variables":     map[string]any{"foo": 123},
		"operationName": "TestOp",
		"query":         "{ hello }",
	}

	// Build multipart body via helper
	body, contentType, err := buildMultipartBody(jsonReq, zipData)
	if err != nil {
		t.Fatalf("buildMultipartBody error: %v", err)
	}

	// Create HTTP request
	req := httptest.NewRequest("POST", "/graphql", body)
	req.Header.Set("Content-Type", contentType)

	// Call parseMultipartGQL
	params, file, err := parseMultipartGQL(req)
	if err != nil {
		t.Fatalf("parseMultipartGQL error: %v", err)
	}
	defer os.Remove(file.Name())

	// Check params
	expected := &gqlRequestParams{
		Query:         "{ hello }",
		OperationName: "TestOp",
		Variables:     map[string]interface{}{"foo": float64(123)},
	}
	if params.Query != expected.Query {
		t.Errorf("expected Query %q, got %q", expected.Query, params.Query)
	}
	if params.OperationName != expected.OperationName {
		t.Errorf("expected OperationName %q, got %q", expected.OperationName, params.OperationName)
	}
	if !reflect.DeepEqual(params.Variables, expected.Variables) {
		t.Errorf("expected Variables %v, got %v", expected.Variables, params.Variables)
	}

	// Check file contents
	info, err := os.Stat(file.Name())
	if err != nil {
		t.Fatalf("stat temp file: %v", err)
	}
	if info.Size() != int64(len(zipData)) {
		t.Errorf("expected file size %d, got %d", len(zipData), info.Size())
	}
	if !strings.HasPrefix(filepath.Base(file.Name()), "paperboy-zip") {
		t.Errorf("unexpected temp file name: %s", file.Name())
	}
}

func TestServeHTTP_Multipart_IncludesFile(t *testing.T) {
	// Prepare GraphQL JSON and ZIP data
	resolver := testResolver{
		zipData: []byte("PK\x03\x04dummyzipcontent"),
		gqlVars: map[string]any{"testArg": "123"},
	}

	// Simple GraphQL schema that exposes hasFile
	schema := `type Query { checkFile: Boolean! }`
	h := MustSchemaHandler(schema, &resolver)

	// Prepare GraphQL JSON and ZIP data
	jsonReq := map[string]any{
		"variables": resolver.gqlVars,
		"query":     "{ checkFile }",
	}

	// Build multipart body via helper
	body, contentType, err := buildMultipartBody(jsonReq, resolver.zipData)
	if err != nil {
		t.Fatalf("buildMultipartBody error: %v", err)
	}

	// Create HTTP request and recorder
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/graphql", body)
	req.Header.Set("Content-Type", contentType)

	// Invoke handler
	h.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Errorf("GQL error body: %q", rr.Body.String())
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	// Verify response
	var resp struct {
		Data   struct{ CheckFile bool }
		Errors []map[string]any
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if e := resp.Errors; e != nil {
		t.Errorf("expected no checkFile errors: %+v", e)
	}
	if !resp.Data.CheckFile {
		t.Errorf("expected checkFile=true, got false")
	}
}

// testResolver is used for the ServeHTTP integration test.
type testResolver struct {
	gqlVars map[string]any
	zipData []byte
}

func (r *testResolver) CheckFile(ctx context.Context) (bool, error) {
	if f, ok := RequestZipFile(ctx); !ok {
		return false, errors.New("No zip file in context")
	} else if raw, err := io.ReadAll(f); err != nil {
		return false, err
	} else if d := cmp.Diff(r.zipData, raw); d != "" {
		return false, fmt.Errorf("Zip file mismatch: %s", d)
	}
	return true, nil
}

// buildMultipartBody constructs a multipart body with an application/json part and an application/zip part.
func buildMultipartBody(jsonContent map[string]any, zipContent []byte) (body *bytes.Buffer, contentType string, err error) {
	body = &bytes.Buffer{}
	w := multipart.NewWriter(body)

	// JSON part
	hdr := textproto.MIMEHeader{}
	hdr.Set("Content-Type", "application/json")
	jsonBytes, err := json.Marshal(jsonContent)
	if err != nil {
		return nil, "", err
	}
	jsonPart, err := w.CreatePart(hdr)
	if err != nil {
		return nil, "", err
	}
	if _, err := jsonPart.Write(jsonBytes); err != nil {
		return nil, "", err
	}

	// ZIP part
	hdr = textproto.MIMEHeader{}
	hdr.Set("Content-Type", "application/zip")
	hdr.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, "test.zip"))
	zipPart, err := w.CreatePart(hdr)
	if err != nil {
		return nil, "", err
	}
	if _, err := zipPart.Write(zipContent); err != nil {
		return nil, "", err
	}

	// Close writer to finalize the body
	if err := w.Close(); err != nil {
		return nil, "", err
	}

	return body, w.FormDataContentType(), nil
}
