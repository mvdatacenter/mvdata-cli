package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestKeyCreate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusCreated, sdk.Key{Name: "deploy-key", Key: "ssh-ed25519 AAAA..."})),
		"key", "create", "--name", "deploy-key", "--key", "ssh-ed25519 AAAA...",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/keys" {
		t.Errorf("expected /keys, got %s", gotPath)
	}
	if !strings.Contains(out, "deploy-key") {
		t.Errorf("expected output to contain 'deploy-key', got: %s", out)
	}
}

func TestKeyCreate_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.Key{Name: "deploy-key", Key: "ssh-ed25519 AAAA..."}),
		"key", "create", "--name", "deploy-key", "--key", "ssh-ed25519 AAAA...", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var k sdk.Key
	if err := json.Unmarshal([]byte(out), &k); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if k.Name != "deploy-key" {
		t.Errorf("expected name deploy-key, got %s", k.Name)
	}
}

func TestKeyGet(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, sdk.Key{Name: "deploy-key", Key: "ssh-ed25519 AAAA..."})),
		"key", "get", "--name", "deploy-key",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	// GetKey uses query param: GET /keys?name=deploy-key
	if gotPath != "/keys" {
		t.Errorf("expected /keys, got %s", gotPath)
	}
	if !strings.Contains(out, "deploy-key") {
		t.Errorf("expected output to contain 'deploy-key', got: %s", out)
	}
}

func TestKeyDelete(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, noContentHandler()),
		"key", "delete", "--name", "deploy-key",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/keys/deploy-key" {
		t.Errorf("expected /keys/deploy-key, got %s", gotPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", out)
	}
}

func TestKeyCreate_MissingFlags(t *testing.T) {
	_, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.Key{}),
		"key", "create", "--name", "deploy-key",
	)
	if err == nil {
		t.Fatal("expected error for missing --key flag")
	}
}
