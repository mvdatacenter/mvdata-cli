package cmd

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "mvdata version") {
		t.Errorf("expected 'mvdata version' in output, got: %s", out)
	}
}

func TestMissingAuth(t *testing.T) {
	// Commands that require auth should fail without --api-url and --api-token.
	// Clear env vars to avoid inheriting from the environment.
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	_, err := executeCommand("vpc", "get", "--name", "test")
	if err == nil {
		t.Fatal("expected error for missing auth, got nil")
	}
	if !strings.Contains(err.Error(), "API URL is required") {
		t.Errorf("expected 'API URL is required' error, got: %v", err)
	}
}

func TestMissingToken(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	_, err := executeCommand("vpc", "get", "--name", "test", "--api-url", "http://localhost")
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
	if !strings.Contains(err.Error(), "API token is required") {
		t.Errorf("expected 'API token is required' error, got: %v", err)
	}
}

func TestHelp(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, cmd := range []string{"vpc", "subnet", "instance", "key", "kubernetes", "instance-types", "version", "login", "api-key", "configure"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("expected help to contain %q", cmd)
		}
	}
}

func TestOutputFlag(t *testing.T) {
	out, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--output") {
		t.Error("expected --output flag in help")
	}
}
