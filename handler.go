package tailwind

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"strings"
	"sync"
)

// Handler is an HTTP handler that serves the Tailwind CSS generated
// by the engine. It provides content-hashed URLs for cache-busting,
// immutable cache headers for CDN/browser caching, and transparent
// gzip compression.
//
// Use [Handler.URL] to get the current URL to embed in HTML. The URL
// changes whenever the CSS content changes (after a new Scan), ensuring
// browsers always fetch the latest stylesheet.
type Handler struct {
	mu     sync.RWMutex
	engine *Engine
	prefix string // URL path prefix, default "/tailwind"

	// Cached state (recomputed on Build)
	css     []byte // raw CSS bytes
	gzipped []byte // gzip-compressed CSS
	hash    string // hex-encoded content hash (first 8 bytes of SHA-256)
	url     string // computed URL: prefix + "-" + hash + ".css"
}

// NewHandler creates a Handler that serves CSS from the given engine.
// The engine should already have its definitions loaded (via New())
// and candidates populated (via Scan or Write).
//
// Call Build() after populating the engine to compute the initial
// CSS and content hash.
func NewHandler(engine *Engine) *Handler {
	return &Handler{
		engine: engine,
		prefix: "/tailwind",
	}
}

// WithPrefix sets the URL path prefix for the handler.
// Default is "/tailwind", producing URLs like "/tailwind-a1b2c3d4.css".
func (h *Handler) WithPrefix(prefix string) *Handler {
	h.prefix = prefix
	return h
}

// Build computes the CSS from the engine's current state and caches
// the result along with its content hash and compressed form.
// Call this after Scan() or Write() to update the served CSS.
//
// Build is safe for concurrent use — it acquires a write lock
// while updating the cached state. Concurrent ServeHTTP calls
// will block briefly during the update.
func (h *Handler) Build() {
	cssStr := h.engine.FullCSS()
	cssBytes := []byte(cssStr)

	sum := sha256.Sum256(cssBytes)
	hashStr := hex.EncodeToString(sum[:8])

	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gz.Write(cssBytes)
	gz.Close()

	h.mu.Lock()
	h.css = cssBytes
	h.gzipped = buf.Bytes()
	h.hash = hashStr
	h.url = h.prefix + "-" + hashStr + ".css"
	h.mu.Unlock()
}

// URL returns the current content-hashed URL path for the CSS.
// Embed this in your HTML <link> tags. The URL changes when
// Build() is called with different CSS content.
//
// Example: "/tailwind-a1b2c3d4e5f67890.css"
func (h *Handler) URL() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.url
}

// ServeHTTP implements [http.Handler]. It serves the cached CSS with
// appropriate caching headers, handles conditional requests, and
// provides transparent gzip compression.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	currentURL := h.url
	currentHash := h.hash
	css := h.css
	gzipped := h.gzipped
	prefix := h.prefix
	h.mu.RUnlock()

	path := r.URL.Path

	// Check if path matches the expected pattern: prefix + "-" + hex + ".css"
	if path == currentURL {
		h.serveCSS(w, r, currentHash, css, gzipped)
		return
	}

	if matchesPattern(path, prefix) {
		// Old/different hash — redirect to current URL.
		http.Redirect(w, r, currentURL, http.StatusFound)
		return
	}

	http.NotFound(w, r)
}

func (h *Handler) serveCSS(w http.ResponseWriter, r *http.Request, hash string, css, gzipped []byte) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("ETag", hash)

	if r.Header.Get("If-None-Match") == hash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")
		w.Write(gzipped)
		return
	}

	w.Write(css)
}

// matchesPattern checks if path matches prefix + "-" + hex + ".css".
func matchesPattern(path, prefix string) bool {
	if !strings.HasPrefix(path, prefix+"-") {
		return false
	}
	rest := path[len(prefix)+1:]
	if !strings.HasSuffix(rest, ".css") {
		return false
	}
	hexPart := rest[:len(rest)-4]
	if len(hexPart) == 0 {
		return false
	}
	for _, c := range hexPart {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// NewHandlerFromFS creates a ready-to-use Handler by scanning fsys
// for Tailwind class candidates. This is the recommended way to set
// up CSS serving in a Go application.
//
// Usage:
//
//	h := tailwind.NewHandlerFromFS(templateFS)
//	mux.Handle(h.URL(), h)
//	// In templates: <link rel="stylesheet" href="{{.TailwindURL}}">
func NewHandlerFromFS(fsys fs.FS) (*Handler, error) {
	engine := New()
	if err := engine.Scan(fsys); err != nil {
		return nil, err
	}
	h := NewHandler(engine)
	h.Build()
	return h, nil
}
