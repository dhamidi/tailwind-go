// Package tailwind implements a pure-Go Tailwind CSS v4 engine with zero
// dependencies outside the standard library.
//
// The engine is data-driven: it ingests Tailwind's own CSS source to learn
// utility definitions, theme tokens, and variant rules. It implements
// [io.Writer] so that any byte stream — HTML, Go templates, JSX, or any
// other format — can be piped through it. The engine extracts candidate
// class names from the stream and generates the minimal CSS needed.
//
// # Quick Start
//
// Create an engine pre-loaded with Tailwind v4 definitions, write markup
// to it, and retrieve the generated CSS:
//
//	tw := tailwind.New()
//
//	// Write any bytes — HTML, templates, source code.
//	tw.Write([]byte(`<div class="flex items-center p-4 bg-blue-500">`))
//
//	// Generate the CSS.
//	css := tw.CSS()
//
// # Preflight CSS
//
// The engine provides the Tailwind preflight (CSS reset) stylesheet via
// [Engine.PreflightCSS]. The preflight is independent of the utility CSS
// and should be served separately:
//
//	tw := tailwind.New()
//	preflight := tw.PreflightCSS()  // serve once, cache aggressively
//	utilities := tw.CSS()            // regenerated per content scan
//
// [Engine.FullCSS] combines preflight and utility CSS into a single
// stylesheet, which is what the HTTP handler serves:
//
//	full := tw.FullCSS()  // preflight + utilities
//
// # Pipeline Integration
//
// [NewPassthrough] creates an engine that also forwards all bytes to an
// underlying [io.Writer], making it transparent in a pipeline:
//
//	tw := tailwind.NewPassthrough(responseWriter)
//	tmpl.Execute(tw, data) // bytes flow to both tw and responseWriter
//	css := tw.CSS()
//
// # HTTP Serving
//
// [NewHandlerFromFS] provides a simple, production-ready way to serve
// generated CSS. It scans an [fs.FS] for class candidates, generates
// the CSS, and returns an [http.Handler] with content-hashed URLs,
// immutable cache headers, ETag support, and gzip compression:
//
//	h, err := tailwind.NewHandlerFromFS(templateFS)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	mux.Handle(h.URL(), h)
//
// For more control, use [NewHandler] with an existing [Engine]:
//
//	engine := tailwind.New()
//	engine.Scan(templateFS)
//	h := tailwind.NewHandler(engine)
//	h.Build()
//	mux.Handle(h.URL(), h)
//
// # Scanning Files
//
// [Engine.Scan] walks an [fs.FS] and extracts Tailwind class candidates
// from all text files, automatically skipping binary files:
//
//	engine := tailwind.New()
//	engine.Scan(os.DirFS("./templates"))
//	css := engine.CSS()
//
// # Custom Definitions
//
// [Engine.LoadCSS] accepts Tailwind v4 CSS directives to extend or override
// the built-in definitions:
//
//	tw := tailwind.New()
//	tw.LoadCSS([]byte(`
//		@theme { --color-brand: #e11d48; }
//		@utility brand-bg { background-color: var(--color-brand); }
//		@variant hover (&:hover);
//	`))
//
// # Concurrency
//
// [Engine] is safe for concurrent use. Multiple goroutines may call
// [Engine.Write] simultaneously, and [Engine.CSS] may be called while
// writes are in-progress (it snapshots the current candidates).
// Thread safety is provided by an internal sync.RWMutex.
//
// # Supported Class Syntax
//
// The engine recognizes the full range of Tailwind class syntax:
//
//   - Static utilities: flex, block, hidden
//   - Dynamic utilities: p-4, bg-blue-500, text-lg
//   - Variants: hover:bg-blue-500, dark:md:hover:text-white
//   - Important modifier: !p-4
//   - Negative values: -translate-x-4
//   - Arbitrary values: w-[300px], text-[#ff0000]
//   - Arbitrary properties: [mask-type:alpha]
//   - Fractions: w-1/2, translate-x-2/3
//   - Opacity modifiers: bg-blue-500/75
//   - Type hints: text-[length:1.5em]
//   - Compound variants: group-hover:text-white, peer-focus:ring-2, not-hover:opacity-100, has-*:, in-*:
//   - Named groups/peers: group/sidebar, group-hover/sidebar:flex
//   - Container query variants: @md:flex, @lg:grid
//   - @supports variants: supports-grid:flex, supports-[display:grid]:grid, supports-backdrop-filter:flex, not-supports-[display:grid]:hidden
//   - @starting-style variant: starting:opacity-0
//
// # CSS Directives
//
// [Engine.LoadCSS] accepts CSS containing these directives:
//
//   - @theme { --property: value; } — define design tokens (colors, spacing, breakpoints)
//   - @utility name { ... } — define a static utility (e.g., flex, block)
//   - @utility name-* { ... } — define a dynamic utility (e.g., p-*, bg-*)
//   - @variant name (&:selector); — define a pseudo-class or selector variant
//   - @variant name (@media ...); — define a media query variant
//   - @apply class1 class2; — compose utilities in custom CSS rules
//   - @keyframes name { ... } — define animation keyframes
//
// Inside utility declarations, these placeholders are resolved at generation time:
//
//   - --value(--namespace) — replaced with the matching theme token value
//   - --value(type1, type2) — resolved based on the value type (length, color, etc.)
package tailwind
