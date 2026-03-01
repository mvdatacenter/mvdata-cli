package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigure(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	root := newRootCmd()
	var outBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&outBuf)

	// Simulate stdin input: blank line for default URL, then a token.
	root.SetIn(strings.NewReader("\nmvd_my_token_12345\n"))
	root.SetArgs([]string{"configure"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, "Profile \"default\" saved") {
		t.Errorf("expected save confirmation, got: %s", out)
	}

	// Verify config was written.
	data, err := os.ReadFile(filepath.Join(tmpHome, ".mvdata", "config.json"))
	if err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	var cf map[string]any
	json.Unmarshal(data, &cf)

	profiles := cf["profiles"].(map[string]any)
	defaultProfile := profiles["default"].(map[string]any)
	if defaultProfile["api_url"] != "https://console-api.mvdatacenter.com" {
		t.Errorf("expected default API URL, got: %v", defaultProfile["api_url"])
	}
	if defaultProfile["api_token"] != "mvd_my_token_12345" {
		t.Errorf("expected token mvd_my_token_12345, got: %v", defaultProfile["api_token"])
	}
}

func TestConfigure_WithProfile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	root := newRootCmd()
	var outBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&outBuf)
	root.SetIn(strings.NewReader("https://staging-api.example.com\nstaging-token\n"))
	root.SetArgs([]string{"configure", "--profile", "staging"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := outBuf.String()
	if !strings.Contains(out, "Profile \"staging\" saved") {
		t.Errorf("expected save confirmation for staging, got: %s", out)
	}

	data, _ := os.ReadFile(filepath.Join(tmpHome, ".mvdata", "config.json"))
	var cf map[string]any
	json.Unmarshal(data, &cf)

	profiles := cf["profiles"].(map[string]any)
	staging := profiles["staging"].(map[string]any)
	if staging["api_url"] != "https://staging-api.example.com" {
		t.Errorf("expected staging URL, got: %v", staging["api_url"])
	}
}

func TestConfigure_MissingToken(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	root := newRootCmd()
	var outBuf bytes.Buffer
	root.SetOut(&outBuf)
	root.SetErr(&outBuf)
	root.SetIn(strings.NewReader("\n\n")) // blank URL (default) and blank token
	root.SetArgs([]string{"configure"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}
