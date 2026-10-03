package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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

func TestRunPrintsErrorToStderrAndExitsOne(t *testing.T) {
	t.Setenv("MVDATA_API_URL", "")
	t.Setenv("MVDATA_API_TOKEN", "")
	t.Setenv("MVDATA_PROFILE", "")
	t.Setenv("HOME", t.TempDir())

	closed := httptest.NewServer(http.NotFoundHandler())
	closedURL := closed.URL
	closed.Close()
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer failing.Close()

	cases := []struct {
		name    string
		args    []string
		message string
	}{
		{"no config", []string{"vpc", "get", "--name", "test"}, "API URL is required"},
		{"refused connection", []string{"vpc", "get", "--name", "test", "--api-url", closedURL, "--api-token", "x"}, "connection refused"},
		{"refused connection json", []string{"vpc", "get", "--name", "test", "--api-url", closedURL, "--api-token", "x", "-o", "json"}, "connection refused"},
		{"server error", []string{"vpc", "get", "--name", "test", "--api-url", failing.URL, "--api-token", "x"}, "500"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tc.args, &stdout, &stderr)
			if code != 1 {
				t.Errorf("exit status = %d, want 1", code)
			}
			got := stderr.String()
			if !strings.HasPrefix(got, "Error: ") || !strings.HasSuffix(got, "\n") || strings.Count(got, "\n") != 1 {
				t.Errorf("stderr = %q, want one \"Error: <message>\" line", got)
			}
			if !strings.Contains(got, tc.message) {
				t.Errorf("stderr = %q, want it to contain %q", got, tc.message)
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestRunSucceedsWithZeroAndNoStderr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"version"}, &stdout, &stderr); code != 0 {
		t.Errorf("exit status = %d, want 0", code)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "mvdata version") {
		t.Errorf("stdout = %q, want the version line", stdout.String())
	}
}
