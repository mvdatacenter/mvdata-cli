package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

func TestLogin(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	var pollCount atomic.Int32

	server := startMockServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device/authorize":
			json.NewEncoder(w).Encode(sdk.DeviceAuth{
				DeviceCode:      "device-code-123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://console.mvdatacenter.com/device",
				ExpiresIn:       900,
				Interval:        0, // No delay in tests
			})
		case "/auth/device/token":
			if pollCount.Add(1) == 1 {
				// First poll: pending
				json.NewEncoder(w).Encode(sdk.DeviceTokenResponse{Status: "pending"})
			} else {
				// Second poll: complete
				json.NewEncoder(w).Encode(sdk.DeviceTokenResponse{
					Status:      "complete",
					APIToken:    "mvd_test_token_1234567890",
					AccountName: "comsol",
					Email:       "alice@customer.com",
				})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer server.Close()

	out, err := executeCommand("login", "--api-url", server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "ABCD-1234") {
		t.Errorf("expected user code in output, got: %s", out)
	}
	if !strings.Contains(out, "alice@customer.com") {
		t.Errorf("expected email in output, got: %s", out)
	}
	if !strings.Contains(out, "comsol") {
		t.Errorf("expected account name in output, got: %s", out)
	}

	// Verify config was saved.
	data, err := os.ReadFile(filepath.Join(tmpHome, ".mvdata", "config.json"))
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}
	if !strings.Contains(string(data), "mvd_test_token_1234567890") {
		t.Errorf("expected token in config file, got: %s", string(data))
	}
}

func TestLogin_WithProfile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	server := startMockServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/device/authorize":
			json.NewEncoder(w).Encode(sdk.DeviceAuth{
				DeviceCode:      "device-code-123",
				UserCode:        "ABCD-1234",
				VerificationURI: "https://console.mvdatacenter.com/device",
				ExpiresIn:       900,
				Interval:        0,
			})
		case "/auth/device/token":
			json.NewEncoder(w).Encode(sdk.DeviceTokenResponse{
				Status:      "complete",
				APIToken:    "mvd_staging_token",
				AccountName: "staging-acct",
				Email:       "bob@customer.com",
			})
		}
	})
	defer server.Close()

	out, err := executeCommand("login", "--api-url", server.URL, "--profile", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "profile: staging") {
		t.Errorf("expected 'profile: staging' in output, got: %s", out)
	}

	// Verify the staging profile was saved.
	data, _ := os.ReadFile(filepath.Join(tmpHome, ".mvdata", "config.json"))
	var cf map[string]any
	json.Unmarshal(data, &cf)

	profiles := cf["profiles"].(map[string]any)
	staging := profiles["staging"].(map[string]any)
	if staging["api_token"] != "mvd_staging_token" {
		t.Errorf("expected staging token, got: %v", staging["api_token"])
	}
}
