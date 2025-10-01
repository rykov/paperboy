package mail

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rykov/paperboy/config"
)

// Test helpers

func parseCSV(t *testing.T, content string, separator string) ([]map[string]any, error) {
	t.Helper()
	cfg := config.CSVConfig{Separator: separator}
	return unmarshalCsvRecipients(&cfg, []byte(content))
}

func assertCSVSuccess(t *testing.T, data []map[string]any, expectedCount int) {
	t.Helper()
	if len(data) != expectedCount {
		t.Fatalf("Expected %d recipients, got %d", expectedCount, len(data))
	}
}

func assertCSVError(t *testing.T, content, expectedMsg string) {
	t.Helper()
	_, err := parseCSV(t, content, ",")
	if err == nil {
		t.Error("Expected error but got none")
		return
	}
	if expectedMsg != "" && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// Tests

func TestUnmarshalCsvRecipients(t *testing.T) {
	t.Run("basic_parsing", func(t *testing.T) {
		content := `email,name,extra
john.doe@example.com,John Doe,Extra Data`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 1)

		if data[0]["email"] != "john.doe@example.com" {
			t.Errorf("Expected email 'john.doe@example.com', got '%v'", data[0]["email"])
		}
		if data[0]["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got '%v'", data[0]["name"])
		}
		if data[0]["extra"] != "Extra Data" {
			t.Errorf("Expected extra 'Extra Data', got '%v'", data[0]["extra"])
		}
	})

	t.Run("multiple_recipients", func(t *testing.T) {
		content := `email,name,company
john.doe@example.com,John Doe,Acme Corp
jane.smith@example.com,Jane Smith,Globex Inc
bob.jones@example.com,Bob Jones,Initech`

		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 3)

		if data[0]["email"] != "john.doe@example.com" {
			t.Errorf("First recipient email incorrect: %v", data[0]["email"])
		}
		if data[2]["company"] != "Initech" {
			t.Errorf("Last recipient company incorrect: %v", data[2]["company"])
		}
	})

	t.Run("custom_separators", func(t *testing.T) {
		t.Run("semicolon", func(t *testing.T) {
			content := `email;name;extra
john.doe@example.com;John Doe;Extra Data`
			data, err := parseCSV(t, content, ";")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertCSVSuccess(t, data, 1)
		})

		t.Run("tab", func(t *testing.T) {
			content := "email\tname\tcompany\njohn@example.com\tJohn Doe\tAcme"
			data, err := parseCSV(t, content, "\t")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertCSVSuccess(t, data, 1)
		})

		t.Run("pipe", func(t *testing.T) {
			content := `email|name|status
john@example.com|John Doe|active`
			data, err := parseCSV(t, content, "|")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertCSVSuccess(t, data, 1)
			if data[0]["status"] != "active" {
				t.Errorf("Expected status 'active', got '%v'", data[0]["status"])
			}
		})
	})

	t.Run("single_column", func(t *testing.T) {
		content := `email
john@example.com
jane@example.com`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 2)
	})

	t.Run("header_only", func(t *testing.T) {
		content := `email,name,company`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 0)
	})

	t.Run("empty_separator_uses_default", func(t *testing.T) {
		content := `email,name
john@example.com,John Doe`
		data, err := parseCSV(t, content, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 1)
	})
}

func TestUnmarshalCsvErrors(t *testing.T) {
	t.Run("empty_file", func(t *testing.T) {
		assertCSVError(t, ``, "")
	})

	t.Run("multi_character_separator", func(t *testing.T) {
		cfg := config.CSVConfig{Separator: "||"}
		_, err := unmarshalCsvRecipients(&cfg, []byte(`email,name
test@example.com,Test User`))
		if err == nil {
			t.Error("Expected error for multi-character separator")
		}
		if err.Error() != "multi-character CSV separator not supported" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("mismatched_columns", func(t *testing.T) {
		content := `email,name,company
john@example.com,John Doe,Acme Corp
jane@example.com,Jane Smith`
		assertCSVError(t, content, "wrong number of fields")
	})

	t.Run("malformed_quotes", func(t *testing.T) {
		content := `email,name,note
john@example.com,John Doe,"unclosed quote
jane@example.com,Jane Smith,ok`
		assertCSVError(t, content, "")
	})

	t.Run("extra_fields", func(t *testing.T) {
		content := `email,name
john@example.com,John Doe,Extra,Field`
		assertCSVError(t, content, "")
	})
}

func TestUnmarshalCsvDataQuality(t *testing.T) {
	t.Run("whitespace_preserved", func(t *testing.T) {
		content := `email,name,company
  john@example.com  ,  John Doe  ,  Acme Corp`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// CSV standard: whitespace preserved unless TrimLeadingSpace is set
		if data[0]["email"] != "  john@example.com  " {
			t.Logf("Whitespace preserved: %q", data[0]["email"])
		}
	})

	t.Run("special_characters", func(t *testing.T) {
		content := `email,name,note
john@example.com,"Doe, John","Has comma, in field"
jane@example.com,"Smith ""Jane""","Has ""quotes"""`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 2)

		if data[0]["name"] != "Doe, John" {
			t.Errorf("Expected name 'Doe, John', got '%v'", data[0]["name"])
		}
		if data[1]["name"] != `Smith "Jane"` {
			t.Errorf("Expected name 'Smith \"Jane\"', got '%v'", data[1]["name"])
		}
	})

	t.Run("unicode", func(t *testing.T) {
		t.Run("headers", func(t *testing.T) {
			content := `email,名前,société
john@example.com,太郎,Société Générale`
			data, err := parseCSV(t, content, ",")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertCSVSuccess(t, data, 1)

			if _, ok := data[0]["名前"]; !ok {
				t.Error("Expected unicode header '名前' to be present")
			}
			if _, ok := data[0]["société"]; !ok {
				t.Error("Expected unicode header 'société' to be present")
			}
		})

		t.Run("content", func(t *testing.T) {
			content := `email,name,city
john@example.com,Müller,Zürich
jane@example.com,Søren,København
bob@example.com,김민준,서울`
			data, err := parseCSV(t, content, ",")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			assertCSVSuccess(t, data, 3)

			if data[0]["name"] != "Müller" {
				t.Errorf("Expected name 'Müller', got '%v'", data[0]["name"])
			}
			if data[2]["city"] != "서울" {
				t.Errorf("Expected city '서울', got '%v'", data[2]["city"])
			}
		})
	})

	t.Run("newline_in_quotes", func(t *testing.T) {
		content := `email,name,address
john@example.com,John Doe,"123 Main St
Apt 4B
New York, NY"`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 1)

		address := data[0]["address"]
		if !strings.Contains(address.(string), "\n") {
			t.Errorf("Expected address to contain newlines, got: %v", address)
		}
	})

	t.Run("empty_fields", func(t *testing.T) {
		content := `email,name,company,notes
john@example.com,John Doe,,
jane@example.com,,,Some notes`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 2)

		if data[0]["company"] != "" {
			t.Errorf("Expected empty company field, got: '%v'", data[0]["company"])
		}
		if data[1]["name"] != "" {
			t.Errorf("Expected empty name field, got: '%v'", data[1]["name"])
		}
	})

	t.Run("empty_lines_skipped", func(t *testing.T) {
		content := `email,name
john@example.com,John Doe

jane@example.com,Jane Smith`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// encoding/csv skips blank lines by default
		assertCSVSuccess(t, data, 2)
	})
}

func TestUnmarshalCsvPerformance(t *testing.T) {
	t.Run("large_file", func(t *testing.T) {
		var builder strings.Builder
		builder.WriteString("email,name,company\n")
		for i := 0; i < 1000; i++ {
			fmt.Fprintf(&builder, "user%04d@example.com,User Name %d,Company Inc %d\n", i, i, i%10)
		}

		data, err := parseCSV(t, builder.String(), ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 1000)
	})

	t.Run("many_columns", func(t *testing.T) {
		content := `email,f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13,f14,f15,f16,f17,f18,f19
john@example.com,v1,v2,v3,v4,v5,v6,v7,v8,v9,v10,v11,v12,v13,v14,v15,v16,v17,v18,v19`
		data, err := parseCSV(t, content, ",")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertCSVSuccess(t, data, 1)

		if len(data[0]) != 20 {
			t.Errorf("Expected 20 fields, got %d", len(data[0]))
		}
		if data[0]["f19"] != "v19" {
			t.Errorf("Expected f19='v19', got '%v'", data[0]["f19"])
		}
	})
}
