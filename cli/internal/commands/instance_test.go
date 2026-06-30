package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestInstanceHealthChecksLivenessAndReadiness(t *testing.T) {
	t.Setenv("OPENPOST_CONFIG_DIR", t.TempDir())

	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/health":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/v1/ready":
			_, _ = w.Write([]byte(`{"status":"ready","database":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	out, err := executeRootCaptureStdout(t, "--instance", srv.URL, "--json", "instance", "health")
	if err != nil {
		t.Fatalf("instance health returned error: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode json output: %v\noutput:\n%s", err, out)
	}
	if got["instance"] != srv.URL {
		t.Fatalf("instance = %q, want %q", got["instance"], srv.URL)
	}
	if got["health"] != "ok" || got["ready"] != "ready" || got["database"] != "ok" {
		t.Fatalf("health output = %#v, want ok/ready/ok", got)
	}
	wantPaths := []string{"/api/v1/health", "/api/v1/ready"}
	if !reflect.DeepEqual(paths, wantPaths) {
		t.Fatalf("paths = %#v, want %#v", paths, wantPaths)
	}
}

func TestInstanceHealthFailsWhenReadinessIsNotReady(t *testing.T) {
	t.Setenv("OPENPOST_CONFIG_DIR", t.TempDir())

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/health":
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		case "/api/v1/ready":
			_, _ = w.Write([]byte(`{"status":"starting","database":"ok"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	_, err := executeRootCaptureStdout(t, "--instance", srv.URL, "instance", "health")
	if err == nil {
		t.Fatal("instance health returned nil error, want readiness failure")
	}
	if !strings.Contains(err.Error(), `readiness: unexpected status "starting"`) {
		t.Fatalf("error = %v, want readiness failure", err)
	}
}

func executeRootCaptureStdout(t *testing.T, args ...string) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		_ = r.Close()
	}()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	root := NewRoot("test")
	root.SetArgs(args)
	execErr := root.Execute()

	_ = w.Close()
	out := <-done
	return out, execErr
}
