package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestVPCCreate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusCreated, sdk.VPC{Name: "production", CreatedAt: "2025-01-01T00:00:00Z"})),
		"vpc", "create", "--name", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/vpcs" {
		t.Errorf("expected /vpcs, got %s", gotPath)
	}
	if !strings.Contains(out, "production") {
		t.Errorf("expected output to contain 'production', got: %s", out)
	}
}

func TestVPCCreate_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.VPC{Name: "production", CreatedAt: "2025-01-01T00:00:00Z"}),
		"vpc", "create", "--name", "production", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var vpc sdk.VPC
	if err := json.Unmarshal([]byte(out), &vpc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if vpc.Name != "production" {
		t.Errorf("expected name production, got %s", vpc.Name)
	}
}

func TestVPCGet(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, sdk.VPC{Name: "production", CreatedAt: "2025-01-01T00:00:00Z"})),
		"vpc", "get", "--name", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/vpcs/production" {
		t.Errorf("expected /vpcs/production, got %s", gotPath)
	}
	if !strings.Contains(out, "production") {
		t.Errorf("expected output to contain 'production', got: %s", out)
	}
}

func TestVPCDelete(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, noContentHandler()),
		"vpc", "delete", "--name", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/vpcs/production" {
		t.Errorf("expected /vpcs/production, got %s", gotPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", out)
	}
}

func TestVPCCreate_MissingName(t *testing.T) {
	_, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.VPC{}),
		"vpc", "create",
	)
	if err == nil {
		t.Fatal("expected error for missing --name flag")
	}
}
