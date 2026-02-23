package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_Flags(t *testing.T) {
	cfg, err := Resolve("https://api.example.com", "my-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://api.example.com" {
		t.Errorf("expected URL https://api.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "my-token" {
		t.Errorf("expected token my-token, got %s", cfg.APIToken)
	}
}

func TestResolve_EnvVars(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "https://env.example.com")
	t.Setenv("MVDATA_API_TOKEN", "env-token")

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://env.example.com" {
		t.Errorf("expected URL https://env.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "env-token" {
		t.Errorf("expected token env-token, got %s", cfg.APIToken)
	}
}

func TestResolve_FlagsOverrideEnv(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "https://env.example.com")
	t.Setenv("MVDATA_API_TOKEN", "env-token")

	cfg, err := Resolve("https://flag.example.com", "flag-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://flag.example.com" {
		t.Errorf("expected URL https://flag.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "flag-token" {
		t.Errorf("expected token flag-token, got %s", cfg.APIToken)
	}
}

func TestResolve_ConfigFile(t *testing.T) {
	// Create a temp home directory with config file.
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"api_url":"https://file.example.com","api_token":"file-token"}`), 0o644)

	t.Setenv("HOME", tmpHome)
	// Clear env vars to ensure file is the only source.
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://file.example.com" {
		t.Errorf("expected URL https://file.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "file-token" {
		t.Errorf("expected token file-token, got %s", cfg.APIToken)
	}
}

func TestResolve_EnvOverridesFile(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"api_url":"https://file.example.com","api_token":"file-token"}`), 0o644)

	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "https://env.example.com")
	t.Setenv("MVDATA_API_TOKEN", "env-token")

	cfg, err := Resolve("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://env.example.com" {
		t.Errorf("expected URL https://env.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "env-token" {
		t.Errorf("expected token env-token, got %s", cfg.APIToken)
	}
}

func TestResolve_MissingURL(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")

	// Use a temp home with no config file.
	t.Setenv("HOME", t.TempDir())

	_, err := Resolve("", "some-token")
	if err == nil {
		t.Fatal("expected error for missing URL, got nil")
	}
}

func TestResolve_MissingToken(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("HOME", t.TempDir())

	_, err := Resolve("https://api.example.com", "")
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}
