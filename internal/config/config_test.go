package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_Flags(t *testing.T) {
	cfg, err := Resolve("https://api.example.com", "my-token", "")
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

	cfg, err := Resolve("", "", "")
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

	cfg, err := Resolve("https://flag.example.com", "flag-token", "")
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

func TestResolve_ConfigFile_Legacy(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"api_url":"https://file.example.com","api_token":"file-token"}`), 0o644)

	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	cfg, err := Resolve("", "", "")
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

func TestResolve_ConfigFile_Profiles(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)

	cf := map[string]any{
		"current_profile": "default",
		"profiles": map[string]any{
			"default": map[string]string{
				"api_url":   "https://default.example.com",
				"api_token": "default-token",
			},
			"staging": map[string]string{
				"api_url":   "https://staging.example.com",
				"api_token": "staging-token",
			},
		},
	}
	data, _ := json.Marshal(cf)
	os.WriteFile(filepath.Join(configDir, "config.json"), data, 0o644)

	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	// Default profile
	cfg, err := Resolve("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://default.example.com" {
		t.Errorf("expected URL https://default.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "default-token" {
		t.Errorf("expected token default-token, got %s", cfg.APIToken)
	}
}

func TestResolve_ProfileFlag(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)

	cf := map[string]any{
		"current_profile": "default",
		"profiles": map[string]any{
			"default": map[string]string{
				"api_url":   "https://default.example.com",
				"api_token": "default-token",
			},
			"staging": map[string]string{
				"api_url":   "https://staging.example.com",
				"api_token": "staging-token",
			},
		},
	}
	data, _ := json.Marshal(cf)
	os.WriteFile(filepath.Join(configDir, "config.json"), data, 0o644)

	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")

	cfg, err := Resolve("", "", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://staging.example.com" {
		t.Errorf("expected URL https://staging.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "staging-token" {
		t.Errorf("expected token staging-token, got %s", cfg.APIToken)
	}
}

func TestResolve_ProfileEnvVar(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)

	cf := map[string]any{
		"current_profile": "default",
		"profiles": map[string]any{
			"default": map[string]string{
				"api_url":   "https://default.example.com",
				"api_token": "default-token",
			},
			"staging": map[string]string{
				"api_url":   "https://staging.example.com",
				"api_token": "staging-token",
			},
		},
	}
	data, _ := json.Marshal(cf)
	os.WriteFile(filepath.Join(configDir, "config.json"), data, 0o644)

	t.Setenv("HOME", tmpHome)
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "staging")

	cfg, err := Resolve("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIURL != "https://staging.example.com" {
		t.Errorf("expected URL https://staging.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "staging-token" {
		t.Errorf("expected token staging-token, got %s", cfg.APIToken)
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
	t.Setenv("MVDATA_PROFILE", "")

	cfg, err := Resolve("", "", "")
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
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	_, err := Resolve("", "some-token", "")
	if err == nil {
		t.Fatal("expected error for missing URL, got nil")
	}
}

func TestResolve_MissingToken(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	_, err := Resolve("https://api.example.com", "", "")
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
}

func TestSave_NewProfile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := Save("default", &Profile{
		APIURL:   "https://api.example.com",
		APIToken: "my-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify.
	data, err := os.ReadFile(filepath.Join(tmpHome, ".mvdata", "config.json"))
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var cf map[string]any
	json.Unmarshal(data, &cf)

	if cf["current_profile"] != "default" {
		t.Errorf("expected current_profile default, got %v", cf["current_profile"])
	}

	profiles := cf["profiles"].(map[string]any)
	defaultProfile := profiles["default"].(map[string]any)
	if defaultProfile["api_url"] != "https://api.example.com" {
		t.Errorf("expected api_url https://api.example.com, got %v", defaultProfile["api_url"])
	}
	if defaultProfile["api_token"] != "my-token" {
		t.Errorf("expected api_token my-token, got %v", defaultProfile["api_token"])
	}
}

func TestSave_MigrateLegacy(t *testing.T) {
	tmpHome := t.TempDir()
	configDir := filepath.Join(tmpHome, ".mvdata")
	os.MkdirAll(configDir, 0o755)
	os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"api_url":"https://old.example.com","api_token":"old-token"}`), 0o644)

	t.Setenv("HOME", tmpHome)

	// Save a new profile — this should migrate the old flat config to profiles.
	err := Save("staging", &Profile{
		APIURL:   "https://staging.example.com",
		APIToken: "staging-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(configDir, "config.json"))
	var cf map[string]any
	json.Unmarshal(data, &cf)

	profiles := cf["profiles"].(map[string]any)

	// Old flat config should be migrated to "default" profile.
	defaultProfile := profiles["default"].(map[string]any)
	if defaultProfile["api_url"] != "https://old.example.com" {
		t.Errorf("expected migrated default api_url https://old.example.com, got %v", defaultProfile["api_url"])
	}

	// New profile should also exist.
	stagingProfile := profiles["staging"].(map[string]any)
	if stagingProfile["api_url"] != "https://staging.example.com" {
		t.Errorf("expected staging api_url https://staging.example.com, got %v", stagingProfile["api_url"])
	}
}
