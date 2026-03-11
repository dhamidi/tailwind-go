// Package tailwind implements a pure-Go Tailwind CSS engine.
//
// The engine ingests Tailwind's own CSS source (v4 format) to learn
// the utility definitions, theme tokens, and variant rules. It then
// implements [io.Writer] so that any byte stream — template output,
// HTML files, source code — can be piped through it. The engine
// extracts candidate class names from the stream and, on request,
// generates the minimal CSS needed.
//
//	engine := tailwind.New()
//	engine.LoadCSS(tailwindSource) // raw Tailwind v4 CSS
//
//	// Pipe any bytes through — templates, HTML, source files:
//	tmpl.Execute(engine, data)
//	io.Copy(engine, someReader)
//	engine.Write([]byte(`<div class="flex items-center p-4">`))
//
//	// Retrieve the generated CSS:
//	css := engine.CSS()
package tailwind

import (
	"io"
	"sync"
)

// Engine is a Tailwind CSS generator that implements [io.Writer].
//
// Write any bytes to it — it will extract candidate class names from
// the stream. Call [Engine.CSS] to get the generated stylesheet.
//
// Engine is safe for concurrent use. Multiple goroutines may call
// Write simultaneously, and CSS may be called while writes are
// in-progress (it will snapshot the current candidates).
type Engine struct {
	mu sync.RWMutex

	// Parsed Tailwind definitions
	theme     *ThemeConfig
	utilities []*UtilityDef
	variants  map[string]*VariantDef

	// Index for fast utility lookup: prefix → []*UtilityDef
	// Sorted longest-prefix-first for disambiguation.
	utilIndex *utilityIndex

	// Accumulated candidate class names from Write() calls.
	candidates map[string]struct{}

	// Scanner state for the byte stream (handles token boundaries
	// across Write calls).
	scan scanner

	// Optional: pass-through writer. If set, all bytes written to
	// the engine are also written here, making the engine transparent
	// in a pipeline.
	passthrough io.Writer
}

// New creates a new Engine. Call [Engine.LoadCSS] to feed it
// Tailwind's CSS source before piping bytes through.
func New() *Engine {
	return &Engine{
		theme:      &ThemeConfig{Tokens: make(map[string]string)},
		variants:   make(map[string]*VariantDef),
		utilIndex:  newUtilityIndex(),
		candidates: make(map[string]struct{}),
	}
}

// NewPassthrough creates an Engine that also writes all bytes to w.
// This lets you insert the engine transparently into an existing
// pipeline:
//
//	engine := tailwind.NewPassthrough(responseWriter)
//	tmpl.Execute(engine, data) // bytes go to both engine and responseWriter
func NewPassthrough(w io.Writer) *Engine {
	e := New()
	e.passthrough = w
	return e
}

// LoadCSS parses a Tailwind v4 CSS source and populates the engine's
// theme, utility, and variant registries. It can be called multiple
// times to layer additional CSS (e.g., your own @utility definitions
// on top of Tailwind's base).
func (e *Engine) LoadCSS(css []byte) error {
	stylesheet, err := parseStylesheet(css)
	if err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Merge theme tokens
	for k, v := range stylesheet.Theme.Tokens {
		e.theme.Tokens[k] = v
	}

	// Register utilities
	for _, u := range stylesheet.Utilities {
		e.utilities = append(e.utilities, u)
		e.utilIndex.add(u)
	}

	// Register variants
	for _, v := range stylesheet.Variants {
		e.variants[v.Name] = v
	}

	return nil
}

// Write implements [io.Writer]. It scans p for candidate Tailwind
// class names and accumulates them. It never returns an error (unless
// the passthrough writer does).
func (e *Engine) Write(p []byte) (n int, err error) {
	// Passthrough first so errors don't lose data.
	if e.passthrough != nil {
		n, err = e.passthrough.Write(p)
		if err != nil {
			return n, err
		}
	} else {
		n = len(p)
	}

	// Extract candidates from the byte stream.
	tokens := e.scan.feed(p)

	if len(tokens) > 0 {
		e.mu.Lock()
		for _, t := range tokens {
			e.candidates[t] = struct{}{}
		}
		e.mu.Unlock()
	}

	return n, nil
}

// Flush finalizes scanning of any buffered partial token. Call this
// after the last Write if the stream doesn't end with a delimiter.
func (e *Engine) Flush() {
	if tok := e.scan.flush(); tok != "" {
		e.mu.Lock()
		e.candidates[tok] = struct{}{}
		e.mu.Unlock()
	}
}

// Candidates returns the set of candidate class names extracted so far.
func (e *Engine) Candidates() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.candidates))
	for c := range e.candidates {
		out = append(out, c)
	}
	return out
}

// CSS generates the Tailwind CSS for all candidate classes found so far.
// Only utilities that match known definitions are included; unknown
// candidates are silently ignored.
func (e *Engine) CSS() string {
	e.Flush()

	e.mu.RLock()
	candidates := make([]string, 0, len(e.candidates))
	for c := range e.candidates {
		candidates = append(candidates, c)
	}
	e.mu.RUnlock()

	// Read lock is sufficient for generation since we don't mutate
	// theme/utilities/variants during generation.
	e.mu.RLock()
	defer e.mu.RUnlock()

	return generate(candidates, e.theme, e.utilIndex, e.variants)
}

// Reset clears all accumulated candidates, allowing the engine to be
// reused for a new scan pass. Theme/utility/variant definitions are
// preserved.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.candidates = make(map[string]struct{})
	e.scan = scanner{}
}
