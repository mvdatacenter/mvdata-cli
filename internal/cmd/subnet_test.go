package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestSubnetCreate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusCreated, sdk.Subnet{Name: "public", VPCName: "production", CIDRBlock: "10.0.1.0/24"})),
		"subnet", "create", "--name", "public", "--vpc", "production", "--cidr", "10.0.1.0/24",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/subnets" {
		t.Errorf("expected /subnets, got %s", gotPath)
	}
	if !strings.Contains(out, "public") {
		t.Errorf("expected output to contain 'public', got: %s", out)
	}
}

func TestSubnetCreate_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.Subnet{Name: "public", VPCName: "production", CIDRBlock: "10.0.1.0/24"}),
		"subnet", "create", "--name", "public", "--vpc", "production", "--cidr", "10.0.1.0/24", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var s sdk.Subnet
	if err := json.Unmarshal([]byte(out), &s); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if s.VPCName != "production" {
		t.Errorf("expected vpcName production, got %s", s.VPCName)
	}
}

func TestSubnetGet(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, sdk.Subnet{Name: "public", VPCName: "production", CIDRBlock: "10.0.1.0/24"})),
		"subnet", "get", "--name", "public",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/subnets/public" {
		t.Errorf("expected /subnets/public, got %s", gotPath)
	}
	if !strings.Contains(out, "public") {
		t.Errorf("expected output to contain 'public', got: %s", out)
	}
}

func TestSubnetDelete(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, noContentHandler()),
		"subnet", "delete", "--name", "public",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/subnets/public" {
		t.Errorf("expected /subnets/public, got %s", gotPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", out)
	}
}

func TestSubnetCreate_MissingFlags(t *testing.T) {
	_, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.Subnet{}),
		"subnet", "create", "--name", "public",
	)
	if err == nil {
		t.Fatal("expected error for missing --vpc and --cidr flags")
	}
}
