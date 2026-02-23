package cmd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestKubernetesCreate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusCreated, sdk.KubernetesCluster{
			Name: "k8s-main", Version: "1.29", NodeInstanceType: "c1", NodeCount: 3,
			Endpoint: "https://k8s-main.k8s.mvdatacenter.com", Status: "provisioning",
		})),
		"kubernetes", "create", "--name", "k8s-main", "--version", "1.29", "--node-type", "c1", "--node-count", "3",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/kubernetes" {
		t.Errorf("expected /kubernetes, got %s", gotPath)
	}
	if !strings.Contains(out, "k8s-main") {
		t.Errorf("expected output to contain 'k8s-main', got: %s", out)
	}
}

func TestKubernetesCreate_JSON(t *testing.T) {
	out, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.KubernetesCluster{
			Name: "k8s-main", Version: "1.29", NodeInstanceType: "c1", NodeCount: 3,
			Endpoint: "https://k8s-main.k8s.mvdatacenter.com", Status: "provisioning",
		}),
		"kubernetes", "create", "--name", "k8s-main", "--version", "1.29", "--node-type", "c1", "--node-count", "3", "-o", "json",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var c sdk.KubernetesCluster
	if err := json.Unmarshal([]byte(out), &c); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if c.NodeCount != 3 {
		t.Errorf("expected nodeCount 3, got %d", c.NodeCount)
	}
}

func TestKubernetesGet(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, sdk.KubernetesCluster{
			Name: "k8s-main", Version: "1.29", NodeInstanceType: "c1", NodeCount: 3,
			Endpoint: "https://k8s-main.k8s.mvdatacenter.com", Status: "running",
		})),
		"kubernetes", "get", "--name", "k8s-main",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("expected GET, got %s", gotMethod)
	}
	if gotPath != "/kubernetes/k8s-main" {
		t.Errorf("expected /kubernetes/k8s-main, got %s", gotPath)
	}
	if !strings.Contains(out, "k8s-main") {
		t.Errorf("expected output to contain 'k8s-main', got: %s", out)
	}
}

func TestKubernetesUpdate(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, jsonHandler(http.StatusOK, sdk.KubernetesCluster{
			Name: "k8s-main", Version: "1.29", NodeInstanceType: "c1", NodeCount: 5,
			Endpoint: "https://k8s-main.k8s.mvdatacenter.com", Status: "running",
		})),
		"kubernetes", "update", "--name", "k8s-main", "--node-count", "5",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("expected PATCH, got %s", gotMethod)
	}
	if gotPath != "/kubernetes/k8s-main" {
		t.Errorf("expected /kubernetes/k8s-main, got %s", gotPath)
	}
	if !strings.Contains(out, "k8s-main") {
		t.Errorf("expected output to contain 'k8s-main', got: %s", out)
	}
}

func TestKubernetesDelete(t *testing.T) {
	var gotMethod, gotPath string
	out, err := executeWithServer(
		captureHandler(&gotMethod, &gotPath, noContentHandler()),
		"kubernetes", "delete", "--name", "k8s-main",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("expected DELETE, got %s", gotMethod)
	}
	if gotPath != "/kubernetes/k8s-main" {
		t.Errorf("expected /kubernetes/k8s-main, got %s", gotPath)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected output to contain 'deleted', got: %s", out)
	}
}

func TestKubernetesCreate_MissingFlags(t *testing.T) {
	_, err := executeWithServer(
		jsonHandler(http.StatusCreated, sdk.KubernetesCluster{}),
		"kubernetes", "create", "--name", "k8s-main",
	)
	if err == nil {
		t.Fatal("expected error for missing flags")
	}
}
