package tailwind

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandler_URLDeterministic(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()

	url := h.URL()
	if url == "" {
		t.Fatal("URL is empty after Build")
	}
	if !strings.HasPrefix(url, "/tailwind-") {
		t.Fatalf("URL should start with /tailwind-, got %q", url)
	}
	if !strings.HasSuffix(url, ".css") {
		t.Fatalf("URL should end with .css, got %q", url)
	}
}

func TestHandler_URLChangesWithContent(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()
	url1 := h.URL()

	e.Write([]byte(`class="bg-blue-500"`))
	h.Build()
	url2 := h.URL()

	if url1 == url2 {
		t.Fatal("URL should change when CSS content changes")
	}
}

func TestHandler_URLStableForSameContent(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()
	url1 := h.URL()

	h.Build()
	url2 := h.URL()

	if url1 != url2 {
		t.Fatalf("URL should be stable for same content: %q != %q", url1, url2)
	}
}

func TestHandler_ServeHTTP200(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", h.URL(), nil)
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %q", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "text-red") {
		t.Fatalf("response body should contain CSS, got %q", body)
	}
}

func TestHandler_ServeHTTP304(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", h.URL(), nil)

	h.mu.RLock()
	etag := h.hash
	h.mu.RUnlock()

	req.Header.Set("If-None-Match", etag)
	h.ServeHTTP(rec, req)

	if rec.Code != 304 {
		t.Fatalf("expected 304, got %d", rec.Code)
	}
}

func TestHandler_ServeHTTPGzip(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", h.URL(), nil)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ce := rec.Header().Get("Content-Encoding"); ce != "gzip" {
		t.Fatalf("expected Content-Encoding gzip, got %q", ce)
	}
	if v := rec.Header().Get("Vary"); v != "Accept-Encoding" {
		t.Fatalf("expected Vary Accept-Encoding, got %q", v)
	}

	gr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()
	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}
	if !strings.Contains(string(decompressed), "text-red") {
		t.Fatal("decompressed body should contain CSS")
	}
}

func TestHandler_ServeHTTPRedirect(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/tailwind-0000000000000000.css", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != 302 {
		t.Fatalf("expected 302 redirect, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != h.URL() {
		t.Fatalf("expected redirect to %q, got %q", h.URL(), loc)
	}
}

func TestHandler_ServeHTTP404(t *testing.T) {
	e := New()
	h := NewHandler(e)
	h.Build()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/other-path", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CacheHeaders(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e)
	h.Build()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", h.URL(), nil)
	h.ServeHTTP(rec, req)

	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Fatalf("unexpected Cache-Control: %q", cc)
	}
	if etag := rec.Header().Get("ETag"); etag == "" {
		t.Fatal("ETag header should be set")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Fatalf("unexpected Content-Type: %q", ct)
	}
}

func TestHandler_WithPrefix(t *testing.T) {
	e := New()
	e.Write([]byte(`class="text-red-500"`))
	h := NewHandler(e).WithPrefix("/css")
	h.Build()

	url := h.URL()
	if !strings.HasPrefix(url, "/css-") {
		t.Fatalf("URL should start with /css-, got %q", url)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNewHandlerFromFS(t *testing.T) {
	fsys := fstest.MapFS{
		"template.html": &fstest.MapFile{
			Data: []byte(`<div class="flex items-center p-4">hello</div>`),
		},
	}

	h, err := NewHandlerFromFS(fsys)
	if err != nil {
		t.Fatalf("NewHandlerFromFS: %v", err)
	}

	url := h.URL()
	if url == "" {
		t.Fatal("URL should not be empty")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", url, nil)
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "flex") {
		t.Fatalf("CSS should contain flex utility, got %q", body)
	}
}

// Verify Handler implements http.Handler.
var _ http.Handler = (*Handler)(nil)
