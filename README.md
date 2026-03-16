# tailwind-go

A pure-Go implementation of Tailwind CSS v4 — zero dependencies, `io.Writer` interface.

## Quick Start

Install:

```sh
go get github.com/dhamidi/tailwind-go
```

Example:

```go
package main

import (
    "fmt"
    tailwind "github.com/dhamidi/tailwind-go"
)

func main() {
    tw := tailwind.New()
    tw.Write([]byte(`<div class="flex items-center p-4 text-blue-500">`))
    fmt.Println(tw.CSS())
}
```

Output:

```css
.flex {
  display: flex;
}

.items-center {
  align-items: center;
}

.p-4 {
  padding: calc(4 * 0.25rem);
}

.text-blue-500 {
  color: var(--color-blue-500);
}
```

## How-To Guides

### Use with `html/template`

```go
tw := tailwind.New()
tmpl := template.Must(template.ParseFiles("page.html"))
tmpl.Execute(tw, data)
css := tw.CSS()
```

### Transparent pipeline with `NewPassthrough`

```go
tw := tailwind.NewPassthrough(responseWriter)
tmpl.Execute(tw, data) // bytes go to responseWriter AND tw
css := tw.CSS()
// Inject css into a <style> tag or serve separately
```

### Stream from `io.Reader` with `io.Copy`

```go
tw := tailwind.New()
io.Copy(tw, someReader)
css := tw.CSS()
```

### Add custom utilities

```go
tw := tailwind.New()
tw.LoadCSS([]byte(`
@theme {
  --color-brand: #e11d48;
}
@utility brand-bg {
  background-color: var(--color-brand);
}
`))
```

### Reuse across requests

```go
tw := tailwind.New() // create once at startup
// per request:
tw.Reset()
tw.Write([]byte(pageHTML))
css := tw.CSS()
```

### Serve CSS in production

`NewHandlerFromFS` scans an `fs.FS` for Tailwind class candidates and returns a ready-to-use HTTP handler with content-hashed URLs, immutable cache headers, and gzip compression.

```go
h, err := tailwind.NewHandlerFromFS(templateFS)
if err != nil {
    log.Fatal(err)
}
mux.Handle(h.URL(), h)
// In templates: <link rel="stylesheet" href="{{.TailwindURL}}">
```

Use `WithPrefix` to customize the URL path prefix (default is `/tailwind`):

```go
h, err := tailwind.NewHandlerFromFS(templateFS)
if err != nil {
    log.Fatal(err)
}
h.WithPrefix("/css")
h.Build()
mux.Handle(h.URL(), h)
```

### Serve preflight CSS

The engine provides access to individual CSS layers and a complete, self-sufficient stylesheet:

```go
tw := tailwind.New()
tw.Scan(templateFS)
preflight := tw.PreflightCSS()   // static reset, cache aggressively
utilities := tw.CSS()             // regenerated per content scan
theme := tw.ThemeCSS()            // :root token definitions (--spacing, --color-*, etc.)
properties := tw.PropertiesCSS()  // @property defaults for --tw-* variables
full := tw.FullCSS()              // theme + preflight + utilities + properties (self-sufficient)
```

`FullCSS()` produces a complete stylesheet that includes theme token definitions, preflight, utilities, and `@property`/fallback declarations — so the generated CSS is self-sufficient without needing external stylesheets.

`CSS()` also emits `@property` declarations for any `--tw-*` custom properties used by the generated utilities. This ensures that composite patterns (e.g., `translate-x-4` referencing `var(--tw-translate-y)`) work correctly even when only some utilities from a group are used.

## Explanation

The engine works by ingesting Tailwind's own CSS v4 source as its specification. Rather than hardcoding a list of utilities, it parses `@theme`, `@utility`, `@variant`, and `@keyframes` directives from real Tailwind CSS. This means the engine's behavior is defined entirely by the CSS it loads — the same way Tailwind v4 itself works.

The scanner is format-agnostic: it extracts candidate class-name tokens from raw bytes without parsing HTML, JSX, or any other template syntax. You can feed it any text — markup, source code, Markdown — and it will find anything that looks like a Tailwind class. False positives are harmless; unrecognized candidates are silently discarded during CSS generation.

The `Engine` is safe for concurrent use. It uses `sync.RWMutex` internally, so concurrent `Write` and `CSS` calls are safe. Call `Reset()` between requests to clear accumulated candidates while keeping all loaded definitions intact.

## Supported Tailwind Features

- 26 color families (red, orange, amber, yellow, lime, green, emerald, teal, cyan, sky, blue, indigo, violet, purple, fuchsia, pink, rose, slate, gray, zinc, neutral, stone, mauve, olive, mist, taupe) with 11 shades each in OKLCH
- Static utilities: `flex`, `block`, `hidden`
- Dynamic utilities with values: `p-4`, `bg-blue-500`, `w-1/2`
- Sizing: `w-*`, `h-*`, `size-*` (sets both width and height), `min-w-*`, `max-w-*`, `min-h-*`, `max-h-*`
- Logical properties: `inline-*`, `block-*`, `min-inline-*`, `max-inline-*`, `min-block-*`, `max-block-*`
- Variants: `hover:bg-blue-500`, `md:flex`, `dark:text-white`
- Stacked variants: `dark:md:hover:bg-blue-500`
- Pseudo-element variants: `before:block`, `after:content-['hello']` — automatically injects `content: var(--tw-content)` when the pseudo-element is the innermost variant and no opacity modifier is present
- Important modifier: `!p-4`
- Negative values: `-translate-x-4`, `-m-4` (only for negatable utilities; non-negatable utilities like padding, width, opacity, and colors silently discard the negative prefix)
- Arbitrary values: `w-[300px]`, `text-[#ff0000]`
- Arbitrary properties: `[mask-type:alpha]`
- Placeholder color: `placeholder-gray-400`, `placeholder-amber-900/60` (sets `::placeholder` text color)
- Opacity modifiers: `bg-blue-500/75`, `text-white/[.5]`
- Fractions: `w-1/2` → `50%`, `aspect-16/9` → `aspect-ratio: 16 / 9` (integer-only utilities like `z-*` and `order-*` do not accept fractions)
- Type hints: `text-[length:1.5em]`
- Custom properties: `w-[--sidebar-width]`
- Arbitrary variants: `[@media(min-width:900px)]:bg-red-500`
- `@theme`, `@utility`, `@variant`, `@keyframes` directives
- Compound variants: `group-hover:text-white`, `peer-focus:ring-2`, `not-hover:opacity-100`, `has-checked:bg-gray-50`, `in-data-current:font-bold`
- `not-*` media negation: `not-dark:bg-white` → `@media not (prefers-color-scheme: dark) { ... }`
- Named groups/peers: `group/sidebar`, `group-hover/sidebar:flex`, `peer-focus/email:text-white`
- Dark mode class strategy: both `@media (prefers-color-scheme: dark)` and class-based (`.dark` selector)
- Container queries: `@md:flex`, `@lg:grid`, built-in `@3xs` through `@7xl`
- `@supports` variants: `supports-grid:flex`, `supports-[display:grid]:grid`, `supports-backdrop-filter:flex`, `not-supports-[display:grid]:hidden`, `[@supports(display:grid)]:flex`
- `@starting-style` variant: `starting:opacity-0`
- `:merge()` function in variant selectors
- `@apply` directive for composing utilities in custom CSS rules
- Typography: `indent-*`, `hyphens-*`, `break-normal`/`break-all`/`break-keep`, `align-*` (vertical-align)
- Font variant numeric: `tabular-nums`, `lining-nums`, `ordinal`, `slashed-zero`, `diagonal-fractions`, etc. — composable via custom properties (e.g., `ordinal tabular-nums` combines both)
- Transitions: `transition`, `transition-colors`, `transition-opacity`, `transition-shadow`, `transition-transform`, `transition-all`, `transition-none`
- Touch action: `touch-auto`, `touch-none`, `touch-pan-x`, `touch-pan-y`, `touch-pinch-zoom`, `touch-manipulation` — composable via custom properties (e.g., `touch-pan-x touch-pan-y` combines both)
- Scroll snap: `snap-x`, `snap-y`, `snap-both`, `snap-mandatory`, `snap-proximity`, `snap-start`, `snap-center`, `snap-end`
- Object fit/position: `object-cover`, `object-contain`, `object-center`, `object-left-top`, etc.
- Containment: `contain-none`, `contain-content`, `contain-strict`, `contain-size`, `contain-inline-size`, `contain-layout`, `contain-paint`, `contain-style` — composable via custom properties (e.g., `contain-size contain-layout` combines both)
- Performance hints: `will-change-auto`, `will-change-scroll`, `will-change-contents`, `will-change-transform`
- Accessibility: `forced-color-adjust-auto`, `forced-color-adjust-none`
- Text size / line height associations: `text-sm` sets both `font-size` and `line-height`
- Gradients: `bg-gradient-to-r` (legacy), `bg-linear-to-r` (v4), `bg-linear-45` (angle), `bg-radial`, `bg-conic-45` — direction utilities set `--tw-gradient-position` and use `linear-gradient(var(--tw-gradient-stops))`, matching upstream TailwindCSS v4. Color stops via `from-*`, `via-*`, `to-*` with position support (`from-5%`, `via-50%`, `to-100%`) and color interpolation modifiers (`bg-linear-to-r/oklch`).
- Default configuration tokens: `--default-transition-duration`, `--default-transition-timing-function`, etc.

For a comprehensive list of all default theme tokens, see [spec.md Appendix A](spec.md).

## Testing

### Differential Fuzzer

A differential fuzzer generates random Tailwind CSS classes, processes them through both the official TailwindCSS CLI and the Go implementation, and asserts that outputs are semantically equivalent. This catches regressions and missing utility implementations.

Run the fuzzer (requires Node.js/npm):

```bash
go test -tags reference -run TestDifferentialFuzz -v -timeout 30m
```

Configure the number of classes and random seed:

```bash
go test -tags reference -run TestDifferentialFuzz -v -timeout 30m \
  -fuzz-count 1000 -fuzz-seed 123
```

The generator produces classes at multiple complexity levels — static utilities, spacing, sizing, colors, typography, variants, opacity modifiers, negative values, arbitrary values, arbitrary properties, border width per-side (`border-t-2`, `border-x-4`), border color per-side (`border-t-red-500`), rounded corner variants (`rounded-tl-lg`), border spacing, outline width/offset, ring/inset-ring width, ring offset, divide width/style, filters (`blur-sm`, `brightness-125`, `contrast-75`), backdrop filters (`backdrop-blur-lg`, `backdrop-saturate-200`), transforms (`translate-x-4`, `rotate-45`, `scale-110`, `skew-x-6`), shadows (`shadow-lg`, `inset-shadow-sm`, `text-shadow`, `text-shadow-red-500`, `text-shadow-initial`), opacity (`opacity-50`), duration/delay (`duration-300`, `delay-150`), mask utilities (`mask-clip-border`, `mask-linear-from-red-500`, `mask-radial-at-center`), gradient directions (`bg-linear-to-r`), gradient percentage positions (`from-5%`, `via-50%`, `to-100%`), gradient via reset (`via-none`), background variants (`bg-clip-text`, `bg-blend-multiply`), perspective (`perspective-near`, `perspective-origin-center`), transform origin (`origin-top`), advanced gradients (`bg-radial`, `bg-radial-at-t`, `bg-conic`), stroke width (`stroke-0`, `stroke-2`), inset positioning (`inset-4`, `top-8`, `inset-x-auto`), indent (`indent-4`, `-indent-2`), font-stretch (`font-stretch-condensed`), content (`content-none`), animations (`animate-spin`, `animate-bounce`), safe alignment (`items-center-safe`, `justify-end-safe`), touch actions (`touch-pan-x`), cursor variants, box decoration break, break before/after/inside, color scheme, field sizing, contain, caption side, and combinations — using a deterministic pseudo-random seed for reproducibility. The default fuzz count is 1000 classes.

The class generator itself can be tested without Node.js:

```bash
go test -run TestClassGenerator -v
```

Cache the npm install across runs by setting `TAILWIND_FUZZ_CACHE`:

```bash
export TAILWIND_FUZZ_CACHE=/tmp/tw-fuzz
go test -tags reference -run TestDifferentialFuzz -v -timeout 30m
```

## Current Limitations

The following are **out of scope by design** (see [spec.md §1.2](spec.md)):

- **`@import` / `@source` / `@config` directives**: build-tool concerns; the engine uses `io.Writer` for input instead
- **CSS minification**: the engine emits readable CSS; pipe through a minifier if needed
- **File watching / CLI**: the engine is a library, not a build tool

## API Reference

| Symbol | Description |
|--------|-------------|
| `New()` | Create engine pre-loaded with Tailwind v4 CSS |
| `NewPassthrough(w)` | Create engine that also forwards bytes to `w` |
| `NewHandlerFromFS(fsys)` | Scan an `fs.FS` and return a ready-to-use HTTP handler |
| `NewHandler(engine)` | Create an HTTP handler from an existing engine |
| `Engine.Write(p)` | Scan bytes for candidate classes (`io.Writer`) |
| `Engine.CSS()` | Generate CSS for accumulated candidates |
| `Engine.LoadCSS(css)` | Parse and load additional Tailwind v4 CSS |
| `Engine.Candidates()` | Return extracted candidate class names |
| `Engine.Flush()` | Finalize any buffered partial token |
| `Engine.Reset()` | Clear candidates, keep definitions |
| `Engine.Scan(fsys)` | Walk an `fs.FS` and extract class candidates |
| `Engine.Preflight()` | Get the Tailwind preflight/reset stylesheet |
| `Engine.PreflightCSS()` | Alias for `Preflight()` |
| `Engine.ThemeCSS()` | Get `:root` theme token definitions |
| `Engine.PropertiesCSS()` | Get `@property` declarations and fallback layer |
| `Engine.FullCSS()` | Get complete self-sufficient stylesheet (theme + preflight + utilities + properties) |
| `Handler.Build()` | Recompute cached CSS and content hash |
| `Handler.URL()` | Get the current content-hashed URL path |
| `Handler.WithPrefix(p)` | Set URL path prefix (default `/tailwind`) |
| `Handler.ServeHTTP(w, r)` | Serve CSS with caching, ETag, gzip |
