//go:generate bash internal/cssdata/download.sh

package tailwind

import (
	"io"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"github.com/dhamidi/tailwind-go/internal/cssdata"
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

	// Parsed @keyframes rules, keyed by name.
	keyframes map[string]*KeyframesRule

	// Resolved @apply rules from LoadCSS.
	applyRules []generatedRule

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

// New creates a new Engine pre-loaded with the embedded Tailwind CSS v4
// definitions (theme tokens + utilities). Additional CSS can be layered
// on top with [Engine.LoadCSS].
func New() *Engine {
	e := &Engine{
		theme:      &ThemeConfig{Tokens: make(map[string]string)},
		variants:   make(map[string]*VariantDef),
		keyframes:  make(map[string]*KeyframesRule),
		utilIndex:  newUtilityIndex(),
		candidates: make(map[string]struct{}),
	}
	// Load bundled Tailwind CSS (theme tokens).
	// Errors are ignored — the embedded CSS is known-good.
	_ = e.LoadCSS(cssdata.Theme)
	_ = e.LoadCSS(cssdata.Utilities)

	// Register Go-based utilities (replaces CSS-parsed definitions).
	registerGoUtilities(e.utilIndex)

	// Register Go-based variants (replaces CSS-parsed definitions).
	registerGoVariants(e.variants, 1000)

	return e
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

	// Register font-family utilities for custom --font-* theme tokens.
	// Excludes --font-weight-* (handled by the "font" dynamic utility).
	for k := range stylesheet.Theme.Tokens {
		if !strings.HasPrefix(k, "--font-") {
			continue
		}
		suffix := k[len("--font-"):]
		// Skip sub-namespaced tokens that map to other utilities.
		if strings.HasPrefix(suffix, "weight") || strings.HasPrefix(suffix, "size") {
			continue
		}
		utilName := "font-" + suffix
		// Don't override already-registered utilities.
		if _, exists := e.utilIndex.static[utilName]; exists {
			continue
		}
		reg := staticUtility(utilName, decls("font-family", "var("+k+")"))
		e.utilIndex.add(reg)
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

	// Register keyframes
	for _, kf := range stylesheet.Keyframes {
		e.keyframes[kf.Name] = kf
	}

	// Process @apply rules
	if len(stylesheet.ApplyRules) > 0 {
		e.applyRules = append(e.applyRules, e.processApplyRules(stylesheet.ApplyRules)...)
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

	utilCSS := generate(candidates, e.theme, e.utilIndex, e.variants, e.keyframes)
	if len(e.applyRules) > 0 {
		applyCSS := emitCSS(e.applyRules, nil, nil)
		if utilCSS != "" {
			return applyCSS + "\n" + utilCSS
		}
		return applyCSS
	}
	return utilCSS
}

// processApplyRules resolves @apply directives against the utility registry.
func (e *Engine) processApplyRules(rules []*ApplyRule) []generatedRule {
	return e.processApplyRulesWithDepth(rules, 0, make(map[string]bool))
}

// processApplyRulesWithDepth resolves @apply directives with recursion detection.
// It stops at a maximum depth of 10 per spec §20.7, and tracks visited classes
// to detect circular references.
func (e *Engine) processApplyRulesWithDepth(rules []*ApplyRule, depth int, visited map[string]bool) []generatedRule {
	if depth >= 10 {
		return nil
	}

	var result []generatedRule
	for _, ar := range rules {
		for _, cls := range ar.Classes {
			if visited[cls] {
				continue
			}
			visited[cls] = true

			pc := parseClass(cls)
			entry, valueStr := resolveUtility(pc, e.utilIndex)
			if entry == nil {
				delete(visited, cls)
				continue
			}
			decls := resolveEntryDeclarations(entry, valueStr, pc, e.theme)
			if decls == nil {
				delete(visited, cls)
				continue
			}

			sel := ar.Selector
			sel = resolveVariantSelector(sel, pc.Variants, e.variants)
			mediaQueries := resolveVariants(pc.Variants, e.variants)

			result = append(result, generatedRule{
				selector:     sel,
				declarations: decls,
				important:    pc.Important,
				mediaQueries: mediaQueries,
				order:        ar.Order,
			})
			delete(visited, cls)
		}
	}
	return result
}

// Preflight returns the Tailwind CSS preflight (reset/normalize) stylesheet.
// This is a static stylesheet that does not depend on scanned candidates.
// It should be included before utility CSS in the served output.
// Theme references (--theme()) are resolved against the engine's current
// theme configuration.
func (e *Engine) Preflight() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return resolveThemeRefs(string(cssdata.Preflight), e.theme.Tokens)
}

// PreflightCSS is an alias for [Preflight].
func (e *Engine) PreflightCSS() string {
	return e.Preflight()
}

// ThemeCSS returns the theme layer containing all design tokens as CSS
// custom properties on :root, :host. These are the --color-*, --spacing,
// --text-*, --font-*, etc. variables that utility classes reference.
func (e *Engine) ThemeCSS() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.theme.Tokens) == 0 {
		return ""
	}

	// Collect and sort token names for deterministic output.
	names := make([]string, 0, len(e.theme.Tokens))
	for k := range e.theme.Tokens {
		names = append(names, k)
	}
	sort.Strings(names)

	var sb strings.Builder
	sb.WriteString(":root, :host {\n")
	for _, name := range names {
		sb.WriteString("  ")
		sb.WriteString(name)
		sb.WriteString(": ")
		sb.WriteString(e.theme.Tokens[name])
		sb.WriteString(";\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

// PropertiesCSS returns the @property declarations and legacy fallback
// layer for Tailwind's internal --tw-* CSS custom properties. These
// provide default values (e.g., --tw-border-style: solid) so that
// utility CSS referencing these variables is self-sufficient.
func (e *Engine) PropertiesCSS() string {
	return twPropertyDeclarations + twPropertiesFallbackLayer
}

// FullCSS returns the complete Tailwind stylesheet: theme variables,
// preflight/reset, utility CSS, and @property declarations. This is
// the method the HTTP handler uses to produce the served stylesheet.
func (e *Engine) FullCSS() string {
	theme := e.ThemeCSS()
	preflight := e.Preflight()
	utility := e.CSS()
	properties := e.PropertiesCSS()

	var parts []string
	if theme != "" {
		parts = append(parts, theme)
	}
	if preflight != "" {
		parts = append(parts, preflight)
	}
	if utility != "" {
		parts = append(parts, utility)
	}
	if properties != "" {
		parts = append(parts, properties)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n")
}

// resolveThemeRefs replaces all --theme(--token, fallback) references in
// css with the corresponding token value from tokens, or the fallback if
// the token is not found. Resolution is applied recursively so that
// token values containing --theme() are themselves resolved.
func resolveThemeRefs(css string, tokens map[string]string) string {
	const marker = "--theme("
	// Limit iterations to prevent infinite loops from circular refs.
	for range 10 {
		idx := strings.Index(css, marker)
		if idx < 0 {
			break
		}
		// Find the matching closing paren, accounting for nesting.
		start := idx + len(marker)
		depth := 1
		end := start
		for end < len(css) && depth > 0 {
			switch css[end] {
			case '(':
				depth++
			case ')':
				depth--
			}
			if depth > 0 {
				end++
			}
		}
		if depth != 0 {
			// Malformed — no matching paren; stop processing.
			break
		}
		inner := strings.TrimSpace(css[start:end])
		// Split into token name and fallback on the first comma.
		tokenName, fallback := inner, ""
		if comma := strings.IndexByte(inner, ','); comma >= 0 {
			tokenName = strings.TrimSpace(inner[:comma])
			fallback = strings.TrimSpace(inner[comma+1:])
		}
		replacement := fallback
		if v, ok := tokens[tokenName]; ok {
			replacement = v
		}
		css = css[:idx] + replacement + css[end+1:]
	}
	return css
}

// Scan walks fsys and extracts Tailwind class candidates from all
// text files. Binary files (images, fonts, compiled artifacts) are
// automatically skipped.
//
// Scan may be called multiple times with different filesystems;
// candidates accumulate across calls. Call [Engine.Reset] first
// if you need a fresh scan.
//
// Scan reads each text file fully into memory. For extremely large
// files this is acceptable because template/source files are
// typically small.
func (e *Engine) Scan(fsys fs.FS) error {
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return fs.SkipDir
			}
			if name == "node_modules" {
				return fs.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if !hasTextExtension(path) {
			return nil
		}
		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		if !isTextFile(content) {
			return nil
		}
		e.Write(content)
		return nil
	})
	if err != nil {
		return err
	}
	e.Flush()
	return nil
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
