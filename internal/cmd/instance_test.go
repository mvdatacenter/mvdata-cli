package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestInstanceCreate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusCreated, sdk.Instance{
			Name: "web-01", InstanceType: "c1", AuthorizedKeyName: "deploy-key",
			PrivateIP: "10.0.0.2", Status: "running", HourlyPrice: 0.50,
		})),
		"instance", "create", "--name", "web-01", "--vpc", "production", "--type", "c1", "--key", "deploy-key",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/vpcs/production/instances" {
		t.Errorf("expected /vpcs/production/instances, got %s", gotPath)
	}
	if !strings.Contains(out, "web-01") {
		t.Errorf("expected output to contain 'web-01', got: %s", out)
	}
}

func TestInstanceCreate_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.Instance{
			Name: "web-01", InstanceType: "c1", AuthorizedKeyName: "deploy-key",
			PrivateIP: "10.0.0.2", Status: "running", HourlyPrice: 0.50,
		}),
		"instance", "create", "--name", "web-01", "--vpc", "production", "--type", "c1", "--key", "deploy-key", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var i sdk.Instance
	if err := json.Unmarshal([]byte(out), &i); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if i.PrivateIP != "10.0.0.2" {
		t.Errorf("expected privateIp 10.0.0.2, got %s", i.PrivateIP)
	}
}

func TestInstanceGet(t *testing.T) {
	// GetInstance uses list endpoint and filters client-side.
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, []sdk.Instance{
			{Name: "web-01", InstanceType: "c1", AuthorizedKeyName: "deploy-key", PrivateIP: "10.0.0.2", Status: "running"},
		})),
		"instance", "get", "--name", "web-01", "--vpc", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/vpcs/production/instances" {
		t.Errorf("expected /vpcs/production/instances, got %s", gotPath)
	}
	if !strings.Contains(out, "web-01") {
		t.Errorf("expected output to contain 'web-01', got: %s", out)
	}
}

func TestInstanceDelete(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, noContentHandler()),
		"instance", "delete", "--name", "web-01", "--vpc", "production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/vpcs/production/instances/web-01" {
		t.Errorf("expected /vpcs/production/instances/web-01, got %s", gotPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", out)
	}
}

func TestInstanceCreate_MissingFlags(t *testing.T) {
	_, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.Instance{}),
		"instance", "create", "--name", "web-01",
	)
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
}
