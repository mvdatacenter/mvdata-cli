package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestAPIKeyCreate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusCreated, sdk.APIKey{
			ID: "abc-123", Name: "terraform-prod", Key: "mvd_test1234",
			Prefix: "mvd_test1234", CreatedAt: "2025-01-01T00:00:00Z",
		})),
		"api-key", "create", "--name", "terraform-prod",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/auth/api-keys" {
		t.Errorf("expected /auth/api-keys, got %s", gotPath)
	}
	if !strings.Contains(out, "terraform-prod") {
		t.Errorf("expected output to contain 'terraform-prod', got: %s", out)
	}
	if !strings.Contains(out, "mvd_test1234") {
		t.Errorf("expected output to contain key, got: %s", out)
	}
}

func TestAPIKeyCreate_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.APIKey{
			ID: "abc-123", Name: "terraform-prod", Key: "mvd_test1234",
			Prefix: "mvd_test1234", CreatedAt: "2025-01-01T00:00:00Z",
		}),
		"api-key", "create", "--name", "terraform-prod", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var key sdk.APIKey
	if err := json.Unmarshal([]byte(out), &key); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if key.Name != "terraform-prod" {
		t.Errorf("expected name terraform-prod, got %s", key.Name)
	}
}

func TestAPIKeyCreate_MissingName(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	_, err := executeCommand("api-key", "create", "--api-url", "http://localhost", "--api-token", "test")
	if err == nil {
		t.Fatal("expected error for missing --name, got nil")
	}
}

func TestAPIKeyList(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, []sdk.APIKey{
			{ID: "abc-123", Name: "key-1", Prefix: "mvd_abc12345", CreatedAt: "2025-01-01T00:00:00Z"},
			{ID: "def-456", Name: "key-2", Prefix: "mvd_def67890", CreatedAt: "2025-01-02T00:00:00Z"},
		})),
		"api-key", "list",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/auth/api-keys" {
		t.Errorf("expected /auth/api-keys, got %s", gotPath)
	}
	if !strings.Contains(out, "key-1") {
		t.Errorf("expected output to contain 'key-1', got: %s", out)
	}
	if !strings.Contains(out, "key-2") {
		t.Errorf("expected output to contain 'key-2', got: %s", out)
	}
}

func TestAPIKeyList_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusOK, []sdk.APIKey{
			{ID: "abc-123", Name: "key-1", Prefix: "mvd_abc12345"},
		}),
		"api-key", "list", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var keys []sdk.APIKey
	if err := json.Unmarshal([]byte(out), &keys); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
}

func TestAPIKeyDelete(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, noContentHandler()),
		"api-key", "delete", "--id", "abc-123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/auth/api-keys/abc-123" {
		t.Errorf("expected /auth/api-keys/abc-123, got %s", gotPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected 'deleted' in output, got: %s", out)
	}
}

func TestAPIKeyDelete_MissingID(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	_, err := executeCommand("api-key", "delete", "--api-url", "http://localhost", "--api-token", "test")
	if err == nil {
		t.Fatal("expected error for missing --id, got nil")
	}
}
