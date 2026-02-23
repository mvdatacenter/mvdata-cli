package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	if err := Print(&buf, "json", data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("expected name test, got %s", result["name"])
	}
}

func TestPrintJSON_Indented(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"name": "test"}

	if err := Print(&buf, "json", data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "  ") {
		t.Error("expected indented JSON output")
	}
}

func TestPrintTable(t *testing.T) {
	var buf bytes.Buffer
	table := &Table{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"web-01", "running"},
			{"web-02", "stopped"},
		},
	}

	if err := Print(&buf, "table", table); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "NAME") {
		t.Error("expected NAME header in output")
	}
	if !strings.Contains(out, "web-01") {
		t.Error("expected web-01 in output")
	}
	if !strings.Contains(out, "web-02") {
		t.Error("expected web-02 in output")
	}
}

func TestPrintTable_WrongType(t *testing.T) {
	var buf bytes.Buffer
	err := Print(&buf, "table", "not a table")
	if err == nil {
		t.Fatal("expected error for wrong type, got nil")
	}
}

func TestPrint_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	err := Print(&buf, "xml", nil)
	if err == nil {
		t.Fatal("expected error for unknown format, got nil")
	}
}
