package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
	"github.com/spf13/cobra"
)

// executeCommand runs a root command with the given args and returns stdout output.
func executeCommand(args ...string) (string, error) {
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

// startMockServer creates a test server and returns its URL.
// The handler should encode responses with json.NewEncoder(w).Encode(...).
func startMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// withAuth returns args with --api-url and --api-token flags prepended.
func withAuth(serverURL string, args ...string) []string {
	return append([]string{"--api-url", serverURL, "--api-token", "test-token"}, args...)
}

// newClientOverride is a helper to set up a root command whose newClient
// returns a client pointed at the test server.
func executeWithServer(handler http.HandlerFunc, args ...string) (string, error) {
	server := startMockServer(handler)
	defer server.Close()
	return executeCommand(withAuth(server.URL, args...)...)
}

// jsonHandler returns a handler that writes statusCode and encodes v as JSON.
func jsonHandler(statusCode int, v any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(v)
	}
}

// noContentHandler returns a handler that writes 204 No Content.
func noContentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// captureHandler captures the request method and path, then delegates to next.
func captureHandler(method, path *string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*method = r.Method
		*path = r.URL.Path
		next(w, r)
	}
}

// mustUnmarshal is a test helper to unmarshal JSON from the output.
func mustUnmarshalVPC(output string) *sdk.VPC {
	var v sdk.VPC
	json.Unmarshal([]byte(output), &v)
	return &v
}

// rootWithServerArgs creates a new root command with --api-url and --api-token set.
func rootWithServerArgs(serverURL string, args ...string) (*cobra.Command, *bytes.Buffer) {
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(withAuth(serverURL, args...))
	return root, &buf
}
