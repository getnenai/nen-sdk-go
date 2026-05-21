package nendesktop

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient stands up an httptest server pointed at handler and
// returns a Client wired to it plus a teardown.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := New(Config{APIKey: "sk_nen_test", BaseURL: srv.URL})
	return c, srv.Close
}

func TestListFiles_NoPath_OmitsQuery(t *testing.T) {
	var gotPath, gotRawQuery, gotAuth string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[{"name":"a.txt","size":3,"modified":1700000000.5}]}`))
	})
	defer stop()

	files, err := client.ListFiles(context.Background(), "dsk_1")
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if gotPath != "/desktops/dsk_1/files" {
		t.Errorf("path: got %q", gotPath)
	}
	if gotRawQuery != "" {
		t.Errorf("raw query: got %q, want empty when no WithPath option supplied", gotRawQuery)
	}
	if gotAuth != "Bearer sk_nen_test" {
		t.Errorf("auth: got %q", gotAuth)
	}
	if len(files) != 1 || files[0].Name != "a.txt" || files[0].IsDir {
		t.Fatalf("unexpected files: %+v", files)
	}
}

func TestListFiles_WithPath_AddsQuery(t *testing.T) {
	var gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[]}`))
	})
	defer stop()

	if _, err := client.ListFiles(context.Background(), "dsk_1", WithPath("Documents")); err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if gotRawQuery != "path=Documents" {
		t.Errorf("raw query: got %q want path=Documents", gotRawQuery)
	}
}

func TestListFiles_WithEmptyPath_OmitsQuery(t *testing.T) {
	var gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[]}`))
	})
	defer stop()

	if _, err := client.ListFiles(context.Background(), "dsk_1", WithPath("")); err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if gotRawQuery != "" {
		t.Errorf("raw query: got %q, want empty when WithPath is empty string", gotRawQuery)
	}
}

func TestListFiles_NilOption_Ignored(t *testing.T) {
	var gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[]}`))
	})
	defer stop()

	// A nil option must not panic; it should be skipped, and a real
	// option in the same call should still apply.
	if _, err := client.ListFiles(context.Background(), "dsk_1", nil, WithPath("Documents"), nil); err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if gotRawQuery != "path=Documents" {
		t.Errorf("raw query: got %q want path=Documents", gotRawQuery)
	}
}

func TestListFiles_DecodesIsDir(t *testing.T) {
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"files":[
			{"name":"a.txt","size":5,"modified":1},
			{"name":"Documents","size":0,"modified":2,"is_dir":true}
		]}`))
	})
	defer stop()

	files, err := client.ListFiles(context.Background(), "dsk_1")
	if err != nil {
		t.Fatalf("ListFiles: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 entries, got %d: %+v", len(files), files)
	}

	byName := map[string]File{}
	for _, f := range files {
		byName[f.Name] = f
	}
	if byName["a.txt"].IsDir {
		t.Errorf("a.txt must not be a directory: %+v", byName["a.txt"])
	}
	if !byName["Documents"].IsDir {
		t.Errorf("Documents must be a directory: %+v", byName["Documents"])
	}
}

func TestUploadFile_WithPath_AddsQuery(t *testing.T) {
	var gotPath, gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"success":true,"size":3,"filename":"hi.txt"}`))
	})
	defer stop()

	if _, err := client.UploadFile(context.Background(), "dsk_1", "hi.txt", strings.NewReader("hi!"), "text/plain", WithPath("Documents")); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if gotPath != "/desktops/dsk_1/files/hi.txt" {
		t.Errorf("path: got %q", gotPath)
	}
	if gotRawQuery != "path=Documents" {
		t.Errorf("raw query: got %q want path=Documents", gotRawQuery)
	}
}

func TestUploadFile_NoOption_OmitsQuery(t *testing.T) {
	var gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		_, _ = w.Write([]byte(`{"success":true,"size":3,"filename":"hi.txt"}`))
	})
	defer stop()

	if _, err := client.UploadFile(context.Background(), "dsk_1", "hi.txt", strings.NewReader("hi!"), "text/plain"); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	if gotRawQuery != "" {
		t.Errorf("raw query: got %q want empty", gotRawQuery)
	}
}

func TestDownloadFile_WithPath_AddsQuery(t *testing.T) {
	var gotPath, gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("body"))
	})
	defer stop()

	rc, _, err := client.DownloadFile(context.Background(), "dsk_1", "hi.txt", WithPath("Documents"))
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	_, _ = io.ReadAll(rc)
	_ = rc.Close()

	if gotPath != "/desktops/dsk_1/files/hi.txt" {
		t.Errorf("path: got %q", gotPath)
	}
	if gotRawQuery != "path=Documents" {
		t.Errorf("raw query: got %q want path=Documents", gotRawQuery)
	}
}

func TestDownloadFile_NoOption_OmitsQuery(t *testing.T) {
	var gotRawQuery string
	client, stop := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("body"))
	})
	defer stop()

	rc, _, err := client.DownloadFile(context.Background(), "dsk_1", "hi.txt")
	if err != nil {
		t.Fatalf("DownloadFile: %v", err)
	}
	_, _ = io.ReadAll(rc)
	_ = rc.Close()

	if gotRawQuery != "" {
		t.Errorf("raw query: got %q want empty", gotRawQuery)
	}
}

func TestFile_OmitsIsDirWhenFalse(t *testing.T) {
	// Belt-and-suspenders for the omitempty contract: marshaling a
	// regular file entry must not emit "is_dir" so older clients see
	// byte-identical JSON to today's responses.
	b, err := json.Marshal(File{Name: "a.txt", Size: 3, Modified: 1.5})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	want := `{"name":"a.txt","size":3,"modified":1.5}`
	if got != want {
		t.Errorf("marshaled %s, want %s", got, want)
	}
}
