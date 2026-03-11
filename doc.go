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
//	engine := tailwind.New()
//
//	// Write any bytes — HTML, templates, source code.
//	engine.Write([]byte(`<div class="flex items-center p-4 bg-blue-500">`))
//
//	// Generate the CSS.
//	css := engine.CSS()
//
// # Pipeline Integration
//
// [NewPassthrough] creates an engine that also forwards all bytes to an
// underlying [io.Writer], making it transparent in a pipeline:
//
//	engine := tailwind.NewPassthrough(responseWriter)
//	tmpl.Execute(engine, data) // bytes flow to both engine and responseWriter
//	css := engine.CSS()
//
// # Custom Definitions
//
// [Engine.LoadCSS] accepts Tailwind v4 CSS directives to extend or override
// the built-in definitions:
//
//	engine := tailwind.New()
//	engine.LoadCSS([]byte(`
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
//   - @keyframes name { ... } — define animation keyframes
//
// Inside utility declarations, these placeholders are resolved at generation time:
//
//   - --value(--namespace) — replaced with the matching theme token value
//   - --value(type1, type2) — resolved based on the value type (length, color, etc.)
package tailwind
