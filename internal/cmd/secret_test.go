package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdk "github.com/mvdatacenter/mvdata-sdk-go/mvdata"
)

type secretRequest struct {
	method string
	uri    string
	body   map[string]any
}

type fakeConsole struct {
	t        *testing.T
	routes   map[string]func(w http.ResponseWriter)
	requests []secretRequest
}

func (f *fakeConsole) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := secretRequest{method: r.Method, uri: r.URL.RequestURI()}
	if b, _ := io.ReadAll(r.Body); len(b) > 0 {
		if err := json.Unmarshal(b, &req.body); err != nil {
			f.t.Errorf("%s %s: body is not JSON: %s", r.Method, req.uri, b)
		}
	}
	f.requests = append(f.requests, req)
	answer, ok := f.routes[r.Method+" "+req.uri]
	if !ok {
		f.t.Errorf("unexpected request %s %s", r.Method, req.uri)
		w.WriteHeader(http.StatusTeapot)
		return
	}
	answer(w)
}

func answerJSON(status int, v any) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(v)
	}
}

func answerError(status int, code string) func(http.ResponseWriter) {
	return answerJSON(status, map[string]any{"code": code, "message": "console says " + code})
}

func answerNoContent(w http.ResponseWriter) { w.WriteHeader(http.StatusNoContent) }

func runSecret(t *testing.T, console *fakeConsole, stdin string, args ...string) (string, string, int) {
	t.Helper()
	console.t = t
	server := httptest.NewServer(console)
	defer server.Close()

	root := newRootCmd()
	var stdout, stderr bytes.Buffer
	root.SetArgs(withAuth(server.URL, args...))
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	status := 0
	if err := root.Execute(); err != nil {
		stderr.WriteString("Error: " + errorMessage(err) + "\n")
		status = exitStatus(err)
	}
	return stdout.String(), stderr.String(), status
}

var (
	testAuthor = &sdk.Identity{Email: "ops@example.com", Type: "user"}
	testStore  = sdk.SecretStore{Name: "app", Description: "the app", SecretCount: 2, LastModifiedBy: testAuthor}
	testSecret = sdk.SecretMetadata{StoreName: "app", Name: "db", Version: 3, SyncState: "in-sync", LastModifiedBy: testAuthor}
)

func TestSecretCommandsCallTheirRoute(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		route    string
		answer   func(http.ResponseWriter)
		wantBody map[string]any
		wantOut  []string
	}{
		{"store list", []string{"secret-store", "list"}, "GET /secret-stores",
			answerJSON(200, map[string]any{"secretStores": []sdk.SecretStore{testStore}}), nil,
			[]string{"NAME", "app", "the app", "2", "ops@example.com"}},
		{"store create", []string{"secret-store", "create", "app", "--description", "the app"}, "POST /secret-stores",
			answerJSON(201, testStore), map[string]any{"name": "app", "description": "the app"},
			[]string{"app", "the app"}},
		{"store get", []string{"secret-store", "get", "app"}, "GET /secret-stores/app",
			answerJSON(200, testStore), nil, []string{"app", "the app"}},
		{"store update", []string{"secret-store", "update", "app", "--description", ""}, "PATCH /secret-stores/app",
			answerJSON(200, sdk.SecretStore{Name: "app"}), map[string]any{"description": ""}, []string{"app"}},
		{"store delete", []string{"secret-store", "delete", "app"}, "DELETE /secret-stores/app",
			answerNoContent, nil, nil},
		{"secret list", []string{"secret", "list", "--store", "app"}, "GET /secrets?storeName=app",
			answerJSON(200, map[string]any{"secrets": []sdk.SecretMetadata{testSecret}}), nil,
			[]string{"STORE", "app", "db", "3", "in-sync", "ops@example.com"}},
		{"secret get", []string{"secret", "get", "app/db"}, "GET /secrets/app/db",
			answerJSON(200, testSecret), nil, []string{"app", "db", "3"}},
		{"secret delete", []string{"secret", "delete", "app/db"}, "DELETE /secrets/app/db",
			answerNoContent, nil, nil},
		{"plural secrets alias", []string{"secrets", "get", "app/db"}, "GET /secrets/app/db",
			answerJSON(200, testSecret), nil, []string{"db"}},
		{"plural secret-stores alias", []string{"secret-stores", "get", "app"}, "GET /secret-stores/app",
			answerJSON(200, testStore), nil, []string{"app"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &fakeConsole{routes: map[string]func(http.ResponseWriter){tt.route: tt.answer}}
			out, errOut, status := runSecret(t, console, "", tt.args...)
			if status != 0 {
				t.Fatalf("exit status %d, stderr %q", status, errOut)
			}
			if len(console.requests) != 1 {
				t.Fatalf("expected one request, got %+v", console.requests)
			}
			if got := console.requests[0].body; tt.wantBody != nil && !jsonEqual(got, tt.wantBody) {
				t.Errorf("body = %v, want %v", got, tt.wantBody)
			}
			for _, want := range tt.wantOut {
				if !strings.Contains(out, want) {
					t.Errorf("output %q lacks %q", out, want)
				}
			}
			if tt.wantOut == nil && out != "" {
				t.Errorf("expected no output, got %q", out)
			}
		})
	}
}

func jsonEqual(a, b map[string]any) bool {
	ja, _ := json.Marshal(a)
	jb, _ := json.Marshal(b)
	return bytes.Equal(ja, jb)
}

func TestSecretListJSONIsAnArrayEvenWithOneRow(t *testing.T) {
	console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
		"GET /secrets?storeName=app": answerJSON(200, map[string]any{"secrets": []sdk.SecretMetadata{testSecret}}),
	}}
	out, _, status := runSecret(t, console, "", "secret", "list", "--store", "app", "-o", "json")
	if status != 0 {
		t.Fatalf("exit status %d", status)
	}
	var rows []sdk.SecretMetadata
	if err := json.Unmarshal([]byte(out), &rows); err != nil || len(rows) != 1 {
		t.Fatalf("expected a one-row array, got %q (%v)", out, err)
	}
}

func TestSecretGetWithoutRevealNeverAsksForTheValue(t *testing.T) {
	for _, format := range []string{"table", "json"} {
		t.Run(format, func(t *testing.T) {
			console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
				"GET /secrets/app/db": answerJSON(200, testSecret),
			}}
			_, _, status := runSecret(t, console, "", "secret", "get", "app/db", "-o", format)
			if status != 0 {
				t.Fatalf("exit status %d", status)
			}
			for _, r := range console.requests {
				if strings.HasSuffix(r.uri, "/reveal") {
					t.Errorf("get without --reveal called %s %s", r.method, r.uri)
				}
			}
		})
	}
}

func TestSecretGetRevealPrintsTheValue(t *testing.T) {
	revealed := answerJSON(200, map[string]any{"storeName": "app", "name": "db", "version": 3, "value": "s3cret"})

	t.Run("table prints the value and one newline", func(t *testing.T) {
		console := &fakeConsole{routes: map[string]func(http.ResponseWriter){"POST /secrets/app/db/reveal": revealed}}
		out, _, status := runSecret(t, console, "", "secret", "get", "app/db", "--reveal")
		if status != 0 || out != "s3cret\n" {
			t.Errorf("status %d, output %q, want \"s3cret\\n\"", status, out)
		}
		if len(console.requests) != 1 {
			t.Errorf("expected the reveal route alone, got %+v", console.requests)
		}
	})

	t.Run("json prints the SecretValue object", func(t *testing.T) {
		console := &fakeConsole{routes: map[string]func(http.ResponseWriter){"POST /secrets/app/db/reveal": revealed}}
		out, _, status := runSecret(t, console, "", "secret", "get", "app/db", "--reveal", "-o", "json")
		var got map[string]any
		if err := json.Unmarshal([]byte(out), &got); err != nil || status != 0 {
			t.Fatalf("status %d, output %q (%v)", status, out, err)
		}
		want := map[string]any{"storeName": "app", "name": "db", "version": float64(3), "value": "s3cret"}
		if !jsonEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestSecretPutChoosesCreateOrUpdate(t *testing.T) {
	created := answerJSON(201, testSecret)
	updated := answerJSON(200, testSecret)
	tests := []struct {
		name  string
		route map[string]func(http.ResponseWriter)
		want  []string
	}{
		{"missing secret is created", map[string]func(http.ResponseWriter){
			"GET /secrets/app/db": answerError(404, "not_found"),
			"POST /secrets":       created,
		}, []string{"GET /secrets/app/db", "POST /secrets"}},
		{"existing secret is updated", map[string]func(http.ResponseWriter){
			"GET /secrets/app/db": answerJSON(200, testSecret),
			"PUT /secrets/app/db": updated,
		}, []string{"GET /secrets/app/db", "PUT /secrets/app/db"}},
		{"a create that loses a race updates once", map[string]func(http.ResponseWriter){
			"GET /secrets/app/db": answerError(404, "not_found"),
			"POST /secrets":       answerError(409, "conflict"),
			"PUT /secrets/app/db": updated,
		}, []string{"GET /secrets/app/db", "POST /secrets", "PUT /secrets/app/db"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &fakeConsole{routes: tt.route}
			_, errOut, status := runSecret(t, console, "s3cret\n", "secret", "put", "app/db")
			if status != 0 {
				t.Fatalf("exit status %d, stderr %q", status, errOut)
			}
			var got []string
			for _, r := range console.requests {
				got = append(got, r.method+" "+r.uri)
				if r.method != http.MethodGet && r.body["value"] != "s3cret" {
					t.Errorf("%s %s sent value %v, want s3cret", r.method, r.uri, r.body["value"])
				}
			}
			if strings.Join(got, ", ") != strings.Join(tt.want, ", ") {
				t.Errorf("requests = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecretPutCreateSendsTheStoreAndName(t *testing.T) {
	console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
		"GET /secrets/app/db": answerError(404, "not_found"),
		"POST /secrets":       answerJSON(201, testSecret),
	}}
	runSecret(t, console, "v", "secret", "put", "app/db")
	want := map[string]any{"storeName": "app", "name": "db", "value": "v"}
	if got := console.requests[len(console.requests)-1].body; !jsonEqual(got, want) {
		t.Errorf("body = %v, want %v", got, want)
	}
}

func TestSecretPutStopsWhenTheReadFails(t *testing.T) {
	console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
		"GET /secrets/app/db": answerError(403, "forbidden"),
	}}
	_, _, status := runSecret(t, console, "v", "secret", "put", "app/db")
	if status != 4 || len(console.requests) != 1 {
		t.Errorf("status %d after %d requests, want 4 after the read alone", status, len(console.requests))
	}
}

func TestSecretPutReadsTheValue(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "value")
	if err := os.WriteFile(file, []byte("from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"stdin drops one newline", "s3cret\n", nil, "s3cret"},
		{"stdin drops one CRLF", "s3cret\r\n", nil, "s3cret"},
		{"stdin drops only one newline", "s3cret\n\n", nil, "s3cret\n"},
		{"stdin without a newline", "s3cret", nil, "s3cret"},
		{"keep-newline keeps it", "s3cret\n", []string{"--keep-newline"}, "s3cret\n"},
		{"from-file drops one newline", "ignored", []string{"--from-file", file}, "from-file"},
		{"from-file with keep-newline", "ignored", []string{"--from-file", file, "--keep-newline"}, "from-file\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
				"GET /secrets/app/db": answerJSON(200, testSecret),
				"PUT /secrets/app/db": answerJSON(200, testSecret),
			}}
			args := append([]string{"secret", "put", "app/db"}, tt.args...)
			_, errOut, status := runSecret(t, console, tt.stdin, args...)
			if status != 0 {
				t.Fatalf("exit status %d, stderr %q", status, errOut)
			}
			if got := console.requests[1].body["value"]; got != tt.want {
				t.Errorf("sent %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecretPutWithAMissingFileWritesNothing(t *testing.T) {
	console := &fakeConsole{routes: map[string]func(http.ResponseWriter){}}
	_, errOut, status := runSecret(t, console, "", "secret", "put", "app/db", "--from-file", filepath.Join(t.TempDir(), "absent"))
	if status != 1 || len(console.requests) != 0 || !strings.Contains(errOut, "reading value") {
		t.Errorf("status %d, %d requests, stderr %q", status, len(console.requests), errOut)
	}
}

func TestSecretRefMustBeStoreSlashName(t *testing.T) {
	for _, ref := range []string{"db", "/db", "app/", "app/db/x"} {
		t.Run(ref, func(t *testing.T) {
			console := &fakeConsole{routes: map[string]func(http.ResponseWriter){}}
			_, errOut, status := runSecret(t, console, "", "secret", "get", ref)
			if status != 1 || len(console.requests) != 0 || !strings.Contains(errOut, "<store>/<name>") {
				t.Errorf("status %d, %d requests, stderr %q", status, len(console.requests), errOut)
			}
		})
	}
}

func TestSecretFailuresExitWithTheirCodesStatus(t *testing.T) {
	tests := []struct {
		httpStatus int
		code       string
		want       int
	}{
		{404, "not_found", 3},
		{403, "forbidden", 4},
		{403, "limit_exceeded", 5},
		{409, "conflict", 6},
		{400, "invalid", 7},
		{503, "unavailable", 8},
		{500, "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
				"GET /secret-stores/app": answerError(tt.httpStatus, tt.code),
			}}
			_, errOut, status := runSecret(t, console, "", "secret-store", "get", "app")
			if status != tt.want {
				t.Errorf("exit status %d, want %d", status, tt.want)
			}
			if !strings.HasPrefix(errOut, "Error: ") {
				t.Errorf("stderr %q does not start with \"Error: \"", errOut)
			}
		})
	}
}

func TestLimitExceededNamesTheLimit(t *testing.T) {
	console := &fakeConsole{routes: map[string]func(http.ResponseWriter){
		"POST /secret-stores": answerJSON(403, map[string]any{
			"code": "limit_exceeded", "message": "Secret limit reached", "limit": "maxSecrets", "current": 10, "max": 10,
		}),
	}}
	_, errOut, status := runSecret(t, console, "", "secret-store", "create", "app")
	if status != 5 || !strings.Contains(errOut, "maxSecrets limit reached: 10/10") {
		t.Errorf("status %d, stderr %q", status, errOut)
	}
}
