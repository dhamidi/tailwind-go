# tailwind-go: Behavioral Specification

Version: 0.1.0-draft
Target compatibility: Tailwind CSS v4.x

## 1. Design Philosophy

tailwind-go is a pure-Go Tailwind CSS engine with zero dependencies beyond the standard library. Rather than reimplementing Tailwind's utility definitions in Go, the engine **ingests Tailwind's own CSS source file** as its specification. All knowledge of utilities, theme tokens, and variants comes from parsing Tailwind's v4 CSS dialect at runtime. The Go code is a generic CSS-generation engine; the CSS file is the data.

The engine implements `io.Writer`. Any byte stream — template output, HTML files, source code, HTTP response bodies — can be piped through it. The engine extracts candidate class names from the raw bytes without any knowledge of the source format. This makes it usable with Go's `html/template`, `templ`, JSX transpiler output, static HTML, or any other system that produces bytes containing Tailwind classes.

### 1.1 Core Invariants

1. **The CSS source is the single source of truth.** The engine has no hardcoded knowledge of any specific Tailwind utility, theme token, or variant. All behavior derives from the parsed CSS.
2. **The byte scanner is format-agnostic.** It does not parse HTML, Go templates, JSX, or any other syntax. It extracts candidate tokens from raw bytes using delimiter-based tokenization.
3. **False positives are harmless.** The scanner deliberately over-extracts. Tokens that don't match any utility definition are silently discarded during generation.
4. **The engine is transparent in pipelines.** In passthrough mode, all bytes written to the engine are forwarded unchanged to a downstream `io.Writer`. The engine observes but does not modify the byte stream.
5. **Zero non-stdlib dependencies.** The entire implementation uses only packages from Go's standard library.

### 1.2 Non-Goals

- **Build tool / file watcher.** The engine is a library. CLI tooling, file watching, and build integration are out of scope for the core package (though a `cmd/` subpackage may be provided).
- **Browser runtime.** This is a build-time / server-side tool. It does not run in WASM or target the browser.
- **CSS minification / optimization.** The engine generates readable CSS. Minification is a separate concern.
- **Preflight / reset styles generation.** The engine does not generate preflight styles — they come from Tailwind's embedded `preflight.css`. The preflight is available via `PreflightCSS()` with `--theme()` references resolved, but is served independently from the utility CSS.


## 2. Public API

### 2.1 Engine

```go
type Engine struct { /* unexported fields */ }
```

The Engine is the central type. It accumulates candidate class names from written bytes and generates CSS on demand.

#### 2.1.1 Construction

```go
func New() *Engine
```

Creates a new Engine with empty registries. `LoadCSS` must be called before generation will produce any output.

```go
func NewPassthrough(w io.Writer) *Engine
```

Creates an Engine that forwards all bytes written to it to `w` before scanning them. This makes the engine transparent in a pipeline:

```go
engine := tailwind.NewPassthrough(httpResponseWriter)
tmpl.Execute(engine, data)  // bytes go to both engine and response
css := engine.CSS()
```

The passthrough writer receives bytes **before** scanning. If the passthrough write returns an error, that error is returned from `Write` and the bytes are **not** scanned (fail-fast: don't silently lose data in the downstream).

#### 2.1.2 Loading CSS Source

```go
func (e *Engine) LoadCSS(css []byte) error
```

Parses a Tailwind v4 CSS source and populates the engine's internal registries for theme tokens, utility definitions, and variant definitions. May be called multiple times; subsequent calls **merge** into the existing registries:

- Theme tokens: later values overwrite earlier ones for the same key.
- Utilities: later definitions with the same pattern overwrite earlier ones.
- Variants: later definitions with the same name overwrite earlier ones.

This merging behavior supports layering custom definitions on top of Tailwind's base:

```go
engine.LoadCSS(tailwindBaseCSS)  // Tailwind's own source
engine.LoadCSS(myCustomCSS)      // project-specific @utility / @theme overrides
```

Returns a non-nil error if the CSS contains syntax that prevents parsing. Recoverable issues (unknown at-rules, malformed declarations) are silently skipped — the parser is lenient.

#### 2.1.3 Writing Bytes

```go
func (e *Engine) Write(p []byte) (n int, err error)
```

Implements `io.Writer`. Scans `p` for candidate class names and accumulates them. In passthrough mode, writes `p` to the downstream writer first.

Return values:
- **Without passthrough:** Always returns `(len(p), nil)`. The scanner cannot fail.
- **With passthrough:** Returns whatever the passthrough writer returns. If the passthrough write fails, scanning is skipped for that chunk.

The scanner maintains state across calls. A token that spans two `Write` calls (e.g., `bg-bl` | `ue-500`) is correctly reconstructed.

Thread safety: multiple goroutines may call `Write` concurrently. The candidate set is protected by a mutex. However, token reconstruction across calls assumes sequential delivery — concurrent writes may split tokens unpredictably. In practice this is fine because over-extraction is harmless.

#### 2.1.4 Flushing

```go
func (e *Engine) Flush()
```

Finalizes any partial token buffered in the scanner. Called automatically by `CSS()`. Should be called explicitly if you need an accurate `Candidates()` list without generating CSS.

#### 2.1.5 Retrieving Candidates

```go
func (e *Engine) Candidates() []string
```

Returns the current set of extracted candidate class names. Order is non-deterministic (map iteration). Includes all tokens extracted so far, including those that may not match any utility. Calls `Flush()` implicitly.

#### 2.1.6 Generating CSS

```go
func (e *Engine) CSS() string
```

Generates the complete Tailwind CSS stylesheet for all candidates extracted so far. Only candidates that match a known utility definition (or are valid arbitrary properties) produce output. Unknown candidates are silently ignored.

The output is a single string containing valid CSS. Rules are ordered according to the Tailwind layer system (see §10).

Calls `Flush()` implicitly before generation.

Thread safety: may be called concurrently with `Write()`. Takes a snapshot of the current candidate set and reads (but does not modify) the utility/variant/theme registries.

#### 2.1.7 Resetting

```go
func (e *Engine) Reset()
```

Clears all accumulated candidates and resets the scanner state. Theme, utility, and variant definitions are preserved. Use this to re-scan a different set of source files without reloading the CSS.

#### 2.1.8 Preflight CSS

```go
func (e *Engine) PreflightCSS() string
```

Returns the Tailwind preflight (CSS reset) stylesheet with `--theme()` references resolved against the engine's current theme. The preflight is independent of the utility CSS returned by `CSS()` — it must be served separately.

The `--theme()` function syntax is `--theme(--token-name, fallback-value)`. Resolution looks up the token name in the engine's theme tokens; if found, the token value is substituted; if not, the fallback value is used. Resolution is applied recursively so that token values containing `--theme()` are themselves resolved.

Typical usage:

```go
e := tailwind.New()
preflightCSS := e.PreflightCSS()  // serve once, cache aggressively
utilityCSS := e.CSS()              // regenerated per content scan
```


## 3. Bundled CSS Source

### 3.1 Code Generation

The repository includes a `go generate` directive that downloads the current release of the Tailwind CSS v4 source file. The exact mechanism (e.g., a shell script or Go helper in `internal/generate`) is an implementation detail, but the result is a Go source file containing the CSS embedded as a constant or `//go:embed` directive.

Running `go generate ./...` fetches the upstream Tailwind CSS source and writes it into the repository so that it is available at compile time.

### 3.2 Embedded Default CSS

The downloaded CSS file is embedded into the library using Go's `//go:embed` mechanism (or an equivalent compile-time embedding). This embedded CSS serves as the **default CSS source** for the engine.

When `New()` or `NewPassthrough()` is called, the engine automatically loads the embedded Tailwind CSS. No explicit `LoadCSS` call is required for standard Tailwind utilities to work:

```go
engine := tailwind.New()
// engine already knows all standard Tailwind v4 utilities
engine.Write([]byte(`<div class="flex items-center p-4">`))
css := engine.CSS()  // produces valid CSS immediately
```

### 3.3 LoadCSS Is Additive

`LoadCSS` **appends** to the engine's internal CSS buffer. It does not replace the previously loaded CSS (whether from the embedded default or from prior `LoadCSS` calls). Each call parses the provided CSS and merges the resulting definitions into the existing registries:

- Theme tokens: later values overwrite earlier ones for the same key.
- Utilities: later definitions with the same pattern overwrite earlier ones.
- Variants: later definitions with the same name overwrite earlier ones.

This means a typical usage pattern layers custom definitions on top of the bundled default:

```go
engine := tailwind.New()                  // loads embedded Tailwind CSS
engine.LoadCSS(myProjectOverrides)        // adds/overrides on top
engine.LoadCSS(myComponentLibraryCSS)     // further additions
```

All three CSS sources contribute to the final registries. Definitions from later `LoadCSS` calls take precedence over earlier ones when keys collide, but non-colliding definitions from all sources coexist.


## 4. Byte Stream Scanner

### 4.1 Overview

The scanner converts a raw byte stream into a set of candidate class name strings. It has **no knowledge** of any markup language, template syntax, or file format. It operates purely on bytes.

### 4.2 Tokenization Rules

The scanner splits the byte stream into tokens using a delimiter-based approach:

**Token bytes** — bytes that can appear within a Tailwind class name:
- ASCII letters: `a-z`, `A-Z`
- ASCII digits: `0-9`
- Hyphen: `-`
- Underscore: `_`
- Colon: `:` (variant separator)
- Square brackets: `[`, `]` (arbitrary values)
- Forward slash: `/` (fractions and opacity modifiers)
- Dot: `.` (decimal values in arbitrary expressions)
- Hash: `#` (hex colors in arbitrary values)
- Exclamation mark: `!` (important modifier)
- Percent: `%` (percentage values)
- Parentheses: `(`, `)` (inside arbitrary values, e.g., `calc()`)
- Comma: `,` (inside arbitrary values)
- Plus: `+` (inside arbitrary values)
- Asterisk: `*` (inside arbitrary values)
- At-sign: `@` (for `@`-prefixed breakpoint variants)

**Delimiter bytes** — everything else (whitespace, `<`, `>`, `"`, `'`, `=`, `{`, `}`, `;`, etc.) terminates the current token.

### 4.3 Bracket Depth Tracking

Square brackets `[` and `]` receive special handling. When the scanner encounters `[`, it increments a depth counter. While depth > 0, **all bytes** (including what would normally be delimiters) are accumulated into the current token. This is necessary because arbitrary values can contain characters like spaces (encoded as `_`) and commas:

```
grid-cols-[1fr_auto_2fr]    ← single token
bg-[rgb(255,0,0)]           ← single token
[mask-type:alpha]            ← single token
```

The depth counter decrements on `]`. Normal tokenization resumes when depth returns to 0.

### 4.4 Cross-Chunk Token Reconstruction

The scanner maintains a byte buffer (`[]byte`) across `Write` calls. When a non-delimiter byte arrives, it's appended to the buffer. When a delimiter byte arrives, the buffer is flushed as a completed token. This correctly handles tokens split across chunk boundaries.

### 4.5 Candidate Filtering

Before accepting a completed token as a candidate, the scanner applies lightweight rejection filters:

1. **Must start with a letter, `!`, `-`, or `[`.** Rejects tokens starting with digits, punctuation, etc.
2. **Single non-letter characters are rejected.** A lone `-` or `!` is not a class.
3. **URLs are rejected.** Any token containing `://` is discarded.
4. **Unbalanced brackets are rejected.** If `[` and `]` counts don't match, discard.

These filters are deliberately **conservative** (reject obvious non-classes) rather than aggressive. The scanner should never reject a valid Tailwind class. Accepting non-classes is fine — they'll be discarded during generation.

### 4.6 What the Scanner Does NOT Do

- It does not parse HTML attributes to find `class="..."`.
- It does not understand Go template syntax (`{{.Field}}`).
- It does not look for `className=` (JSX).
- It does not handle string interpolation or concatenation.
- It does not detect or extract classes from JavaScript/TypeScript source.

All of these would couple the scanner to specific formats. Instead, the scanner extracts **every plausible token** from the byte stream. The tradeoff is a slightly larger candidate set (more misses during generation), which has negligible performance impact.


## 5. CSS Source Parsing

### 5.1 Supported Tailwind v4 Dialect

The parser understands the following CSS constructs:

#### 5.1.1 `@theme` Blocks

```css
@theme {
  --color-red-500: #ef4444;
  --spacing: 0.25rem;
  --breakpoint-sm: 40rem;
  --font-family-sans: ui-sans-serif, system-ui, sans-serif;
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
  --animate-spin: spin 1s linear infinite;
}
```

Parsed into a flat `map[string]string` of custom property names to values. The `--` prefix is preserved in the map keys.

`@theme` blocks may include a modifier keyword:

```css
@theme inline {
  --color-primary: #3b82f6;
}
```

The modifier keyword is noted but the parsing behavior is the same — all declarations become theme tokens.

Supported modifiers:

- **`inline`** — Theme values are inlined directly into utility declarations rather than emitted as CSS custom properties on `:root`. Useful for tokens that should not be exposed as variables.
- **`static`** — Forces all CSS custom properties in the block to be generated in the output regardless of whether any utility references them. Normally the engine tree-shakes unused theme variables; `static` bypasses this optimization.

```css
@theme static {
  --color-primary: var(--color-red-500);
}
```

With `static`, `--color-primary` will appear as a CSS custom property on `:root` even if no utility like `bg-primary` is used in the scanned content.

##### Theme Namespaces

Theme tokens are organized into namespaces by their property name prefix. The engine recognizes these namespaces for value resolution:

| Prefix | Namespace | Example | Utilities |
|--------|-----------|---------|-----------|
| `--color-` | `color` | `--color-blue-500` | `bg-blue-500`, `text-red-600` |
| `--spacing` | `spacing` | `--spacing` (base), `--spacing-*` (overrides) | `p-4`, `m-2`, `gap-6` |
| `--breakpoint-` | `breakpoint` | `--breakpoint-md` | `md:` variant |
| `--font-family-` | `font-family` | `--font-family-sans` | `font-sans` |
| `--text-` | `text` | `--text-lg` | `text-xl`, `text-base` |
| `--font-weight-` | `font-weight` | `--font-weight-bold` | `font-bold` |
| `--leading-` | `leading` | `--leading-tight` | `leading-tight` |
| `--tracking-` | `tracking` | `--tracking-wide` | `tracking-wide` |
| `--radius-` | `radius` | `--radius-lg` | `rounded-lg` |
| `--shadow-` | `shadow` | `--shadow-md` | `shadow-md` |
| `--inset-shadow-` | `inset-shadow` | `--inset-shadow-sm` | `inset-shadow-sm` |
| `--drop-shadow-` | `drop-shadow` | `--drop-shadow-lg` | `drop-shadow-lg` |
| `--text-shadow-` | `text-shadow` | `--text-shadow-sm` | `text-shadow-sm` |
| `--blur-` | `blur` | `--blur-md` | `blur-md` |
| `--opacity-` | `opacity` | `--opacity-50` | `opacity-50` |
| `--transition-property-` | `transition-property` | `--transition-property-all` | `transition-all` |
| `--ease-` | `ease` | `--ease-in-out` | `ease-in-out` |
| `--animate-` | `animate` | `--animate-spin` | `animate-spin` |
| `--perspective-` | `perspective` | `--perspective-dramatic` | `perspective-dramatic` |
| `--aspect-` | `aspect` | `--aspect-video` | `aspect-video` |
| `--container-` | `container` | `--container-3xs` | `@3xs:` variant |
| `--fill-` | `fill` | `--fill-red-500` | `fill-red-500` |
| `--stroke-` | `stroke` | `--stroke-blue-500` | `stroke-blue-500` |
| `--stroke-width-` | `stroke-width` | `--stroke-width-2` | `stroke-2` |
| `--scale-` | `scale` | `--scale-75` | `scale-75` |
| `--rotate-` | `rotate` | `--rotate-45` | `rotate-45` |
| `--skew-` | `skew` | `--skew-12` | `skew-12` |
| `--translate-` | `translate` | `--translate-full` | `translate-full` |
| `--width-` | `width` | `--width-prose` | `w-prose` |
| `--height-` | `height` | `--height-screen` | `h-screen` |
| `--min-width-` | `min-width` | `--min-width-0` | `min-w-0` |
| `--max-width-` | `max-width` | `--max-width-prose` | `max-w-prose` |
| `--min-height-` | `min-height` | `--min-height-0` | `min-h-0` |
| `--max-height-` | `max-height` | `--max-height-screen` | `max-h-screen` |
| `--cursor-` | `cursor` | `--cursor-pointer` | `cursor-pointer` |
| `--columns-` | `columns` | `--columns-3` | `columns-3` |
| `--grid-template-columns-` | `grid-template-columns` | `--grid-template-columns-header` | `grid-cols-header` |
| `--grid-template-rows-` | `grid-template-rows` | `--grid-template-rows-header` | `grid-rows-header` |
| `--grid-auto-columns-` | `grid-auto-columns` | `--grid-auto-columns-min` | `auto-cols-min` |
| `--grid-auto-rows-` | `grid-auto-rows` | `--grid-auto-rows-min` | `auto-rows-min` |
| `--grid-column-` | `grid-column` | `--grid-column-span-3` | `col-span-3` |
| `--grid-column-start-` | `grid-column-start` | `--grid-column-start-1` | `col-start-1` |
| `--grid-column-end-` | `grid-column-end` | `--grid-column-end-3` | `col-end-3` |
| `--grid-row-` | `grid-row` | `--grid-row-span-3` | `row-span-3` |
| `--grid-row-start-` | `grid-row-start` | `--grid-row-start-1` | `row-start-1` |
| `--grid-row-end-` | `grid-row-end` | `--grid-row-end-3` | `row-end-3` |
| `--order-` | `order` | `--order-1` | `order-1` |
| `--z-index-` | `z-index` | `--z-index-50` | `z-50` |
| `--inset-` | `inset` | `--inset-auto` | `inset-auto` |
| `--margin-` | `margin` | `--margin-auto` | `m-auto` |
| `--padding-` | `padding` | `--padding-0` | `p-0` |
| `--gap-` | `gap` | `--gap-0` | `gap-0` |
| `--scroll-margin-` | `scroll-margin` | `--scroll-margin-0` | `scroll-m-0` |
| `--scroll-padding-` | `scroll-padding` | `--scroll-padding-0` | `scroll-p-0` |
| `--background-image-` | `background-image` | `--background-image-none` | `bg-none` |
| `--list-style-type-` | `list-style-type` | `--list-style-type-disc` | `list-disc` |
| `--list-style-image-` | `list-style-image` | `--list-style-image-none` | `list-image-none` |
| `--text-indent-` | `text-indent` | `--text-indent-4` | `indent-4` |
| `--transform-origin-` | `transform-origin` | `--transform-origin-center` | `origin-center` |
| `--border-width-` | `border-width` | `--border-width-2` | `border-2` |
| `--outline-width-` | `outline-width` | `--outline-width-2` | `outline-2` |
| `--outline-offset-` | `outline-offset` | `--outline-offset-2` | `outline-offset-2` |
| `--ring-width-` | `ring-width` | `--ring-width-2` | `ring-2` |
| `--inset-ring-width-` | `inset-ring-width` | `--inset-ring-width-2` | `inset-ring-2` |
| `--ring-offset-width-` | `ring-offset-width` | `--ring-offset-width-2` | `ring-offset-2` |
| `--text-decoration-thickness-` | `text-decoration-thickness` | `--text-decoration-thickness-2` | `decoration-2` |
| `--text-underline-offset-` | `text-underline-offset` | `--text-underline-offset-2` | `underline-offset-2` |

**Namespaces without default theme values:** Some namespaces (such as `z-index`, `opacity`, `order`, `columns`, and many sizing/spacing-related namespaces like `margin`, `padding`, `gap`, `inset`, `width`, `height`, etc.) have no tokens defined in the default `theme.css`. These utilities work via bare numeric values (e.g., `z-50` → `50`, `opacity-50` → `0.5`), the spacing scale (e.g., `p-4` resolves via `--spacing`), or arbitrary values (e.g., `z-[100]`). Users may add custom theme tokens to these namespaces via `@theme`.

**Note on namespace naming:** In TailwindCSS v4, certain theme namespaces use the utility prefix rather than the CSS property name. Notably:
- Font sizes use `--text-*` (not `--font-size-*`) — matching the `text-*` utility
- Letter spacing uses `--tracking-*` (not `--letter-spacing-*`) — matching the `tracking-*` utility
- Line heights use `--leading-*` (not `--line-height-*`) — matching the `leading-*` utility

These namespace names mirror the utility names users write in their classes, keeping the theme and utility namespaces consistent.

##### Default Value Tokens

In addition to namespaced theme tokens, TailwindCSS v4 defines special `--default-*` tokens that provide implicit default values used by certain utilities:

| Token | Purpose | Default Value |
|-------|---------|---------------|
| `--default-transition-duration` | Default duration for `transition-*` utilities | `150ms` |
| `--default-transition-timing-function` | Default easing for `transition-*` utilities | `cubic-bezier(0.4, 0, 0.2, 1)` |
| `--default-border-width` | Default width for `border` (without explicit width) | `1px` |
| `--default-outline-width` | Default width for `outline` (without explicit width) | `1px` |
| `--default-ring-width` | Default width for `ring` (without explicit width) | `1px` |
| `--default-inset-ring-width` | Default width for `inset-ring` (without explicit width) | `1px` |

These tokens are defined in the theme and can be overridden via `@theme`. They are not namespaced — they are standalone variables that utilities reference directly. For example, `transition-colors` applies `transition-duration: var(--default-transition-duration)` unless an explicit duration utility is also present.

#### 5.1.2 `@utility` Blocks

##### Static Utilities

```css
@utility flex {
  display: flex;
}

@utility sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border-width: 0;
}
```

Static utilities have a fixed name with no wildcard. They produce the exact declarations specified in their block.

##### Dynamic Utilities

```css
@utility w-* {
  width: --value(--spacing);
  width: --value(--width);
  width: --value(length, percentage);
}

@utility bg-* {
  background-color: --value(--color);
}

@utility translate-x-* {
  --tw-translate-x: --value(--spacing);
  --tw-translate-x: --value(length, percentage);
  translate: var(--tw-translate-x) var(--tw-translate-y);
}
```

Dynamic utilities end with `-*` in their `@utility` declaration. The `*` is a wildcard that matches any value suffix. The pattern name is everything before `-*`.

##### The `--value()` Function

`--value()` is a placeholder in utility declarations that gets replaced with the resolved value at generation time. Its arguments specify what types of values the utility accepts:

| Syntax | Meaning |
|--------|---------|
| `--value(--color)` | Resolve against the `color` theme namespace |
| `--value(--spacing)` | Resolve against the `spacing` theme namespace (including computed scale) |
| `--value(--width)` | Resolve against the `width` theme namespace |
| `--value(length)` | Accept arbitrary CSS length values |
| `--value(percentage)` | Accept arbitrary CSS percentage values |
| `--value(length, percentage)` | Accept either length or percentage |
| `--value(number)` | Accept arbitrary numbers |
| `--value(integer)` | Accept arbitrary integers |
| `--value(ratio)` | Accept aspect-ratio values like `16/9` |
| `--value(url)` | Accept URL values |
| `--value(color)` | Accept arbitrary CSS color values |
| `--value(shadow)` | Accept arbitrary box-shadow values |
| `--value(any)` | Accept any value |

When a utility has **multiple declarations with different `--value()` types**, they represent a resolution **priority chain**. The engine tries each in order and uses the first that resolves:

```css
@utility w-* {
  width: --value(--spacing);         /* 1st: try spacing scale */
  width: --value(--width);           /* 2nd: try width tokens */
  width: --value(length, percentage); /* 3rd: accept raw CSS values */
}
```

For `w-4`: the spacing namespace resolves → `calc(4 * 0.25rem)` → use first declaration.
For `w-prose`: spacing doesn't match, `--width-prose` resolves → use second declaration.
For `w-[300px]`: arbitrary value → use third declaration.

**Implementation:** When multiple declarations share the same property name but have different `--value()` arguments, they represent alternatives. The engine keeps all of them and tries each during resolution. The first `--value()` that successfully resolves wins, and only that declaration's property-value pair is emitted.

##### Multiple Properties in a Utility

A utility may produce multiple CSS declarations:

```css
@utility translate-x-* {
  --tw-translate-x: --value(--spacing);
  translate: var(--tw-translate-x) var(--tw-translate-y);
}
```

Only declarations containing `--value()` are subject to resolution. Declarations without `--value()` are emitted verbatim.

##### Functional Utilities

Some utilities produce declarations that call CSS functions:

```css
@utility blur-* {
  filter: blur(--value(--blur));
  filter: blur(--value(length));
}
```

The `--value()` appears inside a CSS function call. The engine substitutes the resolved value at the exact position of `--value(...)`.

#### 5.1.3 `@variant` and `@custom-variant` Directives

TailwindCSS v4 uses two directives for defining variants:

- **`@variant`** — Used internally in Tailwind's own CSS source to define built-in variants. This is the syntax the engine parses from the upstream Tailwind CSS file.
- **`@custom-variant`** — Used in user-authored CSS to define custom variants. This is the syntax users write in their project CSS files loaded via `LoadCSS`.

Both directives have identical parsing behavior and register variants in the same way. The engine treats them interchangeably.

##### Built-in Variants (from Tailwind's CSS source)

```css
@variant hover — special: generates &:hover wrapped in @media (hover: hover)
@variant focus (&:focus);
@variant focus-visible (&:focus-visible);
@variant active (&:active);
@variant visited (&:visited);
@variant disabled (&:disabled);
@variant first (&:first-child);
@variant last (&:last-child);
@variant odd (&:nth-child(odd));
@variant even (&:nth-child(even));
@variant empty (&:empty);
@variant placeholder (&::placeholder);
@variant before (&::before);
@variant after (&::after);
@variant selection (& ::selection);
@variant marker (& ::marker);
@variant file (&::file-selector-button);
@variant first-letter (&::first-letter);
@variant first-line (&::first-line);
@variant backdrop (&::backdrop);

@variant first-of-type (&:first-of-type);
@variant last-of-type (&:last-of-type);

@variant sm (@media (width >= 40rem));
@variant md (@media (width >= 48rem));
@variant lg (@media (width >= 64rem));
@variant xl (@media (width >= 80rem));
@variant 2xl (@media (width >= 96rem));

@variant dark (@media (prefers-color-scheme: dark));
@variant motion-safe (@media (prefers-reduced-motion: no-preference));
@variant motion-reduce (@media (prefers-reduced-motion: reduce));
@variant contrast-more (@media (prefers-contrast: more));
@variant contrast-less (@media (prefers-contrast: less));
@variant portrait (@media (orientation: portrait));
@variant landscape (@media (orientation: landscape));
@variant print (@media print);

@variant forced-colors (@media (forced-colors: active));

@variant rtl (&:where(:dir(rtl), [dir="rtl"], [dir="rtl"] *));
@variant ltr (&:where(:dir(ltr), [dir="ltr"], [dir="ltr"] *));

@variant open (&:is([open], :popover-open, :open));
@variant inert (&:is([inert], [inert] *));
```

##### User-Defined Custom Variants

Users define custom variants using `@custom-variant` in their project CSS:

```css
/* Parenthesized form (simple selector or media query): */
@custom-variant theme-midnight (&:where([data-theme="midnight"] *));

/* Block form (for more complex transformations): */
@custom-variant supports-grid {
  @supports (display: grid) {
    @slot;
  }
}
```

The block form uses `@slot` to mark where the utility declarations should be inserted, identical to the block form of `@variant`.

Each `@variant` / `@custom-variant` declaration specifies:

1. **Name** — the prefix used in class names (e.g., `hover` in `hover:bg-blue-500`)
2. **Definition** — either:
   - A **selector pattern** containing `&` as a placeholder for the subject selector (e.g., `&:hover`)
   - A **media/feature query** wrapped in `@media (...)`, `@supports (...)`, or `@container (...)` notation

##### Variant Types

| Type | Detection | Example |
|------|-----------|---------|
| Selector variant | Definition contains `&` but no `@media`/`@supports`/`@container` | `@variant hover (&:hover)` |
| Media query variant | Definition starts with `@media` | `@variant md (@media (width >= 48rem))` |
| Feature query variant | Definition starts with `@supports` | `@variant supports-grid (@supports (display: grid))` |
| Container query variant | Definition starts with `@container` | `@variant @md (@container (width >= 48rem))` |

##### Block-Form Variants

Some variants use a block body instead of a parenthesized definition:

```css
@variant group-hover {
  :merge(.group):hover & {
    @slot;
  }
}
```

The `@slot` directive marks where the utility's declarations should be inserted. The `:merge()` function indicates that the selector should be merged/deduplicated when multiple utilities use the same group variant.

##### Compound Variants

Tailwind v4 supports compound variant definitions:

```css
@variant group-* {
  :merge(.group)\:{value} & {
    @slot;
  }
}
```

Here `{value}` is replaced with the specific variant condition (e.g., `group-hover` → `.group:hover &`).

#### 5.1.4 `@layer` Directives

```css
@layer theme, base, components, utilities;
```

Tailwind uses CSS layers to establish cascade ordering. The engine must respect this ordering when emitting CSS (see §10).

#### 5.1.5 `@import` and `@config`

```css
@import "tailwindcss";
@config "./tailwind.config.js";
```

These are directives that Tailwind's build tool processes. The Go engine ignores them (it receives already-resolved CSS, not unprocessed source).

#### 5.1.6 `@source`

```css
@source "../src/**/*.html";
```

Specifies content paths for class scanning. The Go engine ignores this — scanning is done via the `io.Writer` interface, not file globs.

#### 5.1.7 Standard CSS Rules

Any CSS rules that aren't Tailwind-specific at-rules (e.g., `:root` blocks, reset styles, `@keyframes`) are collected as **base rules**. These may be included in the output verbatim (see §10.1).

### 5.2 CSS Tokenizer

The tokenizer converts raw CSS bytes into a stream of typed tokens. It follows the CSS Tokenization specification (https://www.w3.org/TR/css-syntax-3/#tokenization) with simplifications appropriate for the Tailwind dialect.

#### Token Types

| Type | Examples | Notes |
|------|----------|-------|
| `Whitespace` | space, tab, newline | |
| `Ident` | `color`, `width`, `--spacing` | Includes custom properties |
| `AtKeyword` | `@theme`, `@utility`, `@variant` | `@` followed by an ident |
| `Hash` | `#fff`, `#3b82f6` | `#` followed by hex/ident chars |
| `String` | `"hello"`, `'world'` | Handles escape sequences |
| `Number` | `42`, `3.14`, `-1` | Optional sign, integer or float |
| `Dimension` | `48rem`, `0.25rem`, `100vw` | Number followed by a unit ident |
| `Percentage` | `50%`, `100%` | Number followed by `%` |
| `Function` | `calc(`, `var(`, `--value(` | Ident followed by `(` |
| `Colon` | `:` | |
| `Semicolon` | `;` | |
| `Comma` | `,` | |
| `BraceOpen` | `{` | |
| `BraceClose` | `}` | |
| `BracketOpen` | `[` | |
| `BracketClose` | `]` | |
| `ParenOpen` | `(` | |
| `ParenClose` | `)` | |
| `Delim` | any single char | Catch-all for unmatched characters |
| `Comment` | `/* ... */` | Stripped before parsing |
| `CDC` | `-->` | Ignored |
| `CDO` | `<!--` | Ignored |

#### Tokenizer Behavior

- Operates on `[]byte` input, producing a `[]token` slice.
- Comments are consumed but **excluded** from the output token stream.
- Whitespace tokens are preserved (the parser skips them as needed).
- Custom properties (starting with `--`) are tokenized as `Ident` tokens.
- The `Function` token includes the opening `(` in its value (e.g., `"calc("`).
- Sign characters (`+`, `-`) at the start of a number are consumed as part of the number, **unless** the `-` starts a valid ident (e.g., `-webkit-`).

### 5.3 Parser

The parser consumes the token stream and produces a `Stylesheet` structure:

```go
type Stylesheet struct {
    Theme     ThemeConfig      // Merged theme tokens
    Utilities []*UtilityDef    // Ordered list of utility definitions
    Variants  []*VariantDef    // Ordered list of variant definitions
    BaseRules []Rule           // Non-utility CSS rules (resets, keyframes, etc.)
    Layers    []string         // Layer ordering from @layer declarations
}
```

#### Parser Behavior

- The parser is **lenient**. Unknown at-rules, malformed declarations, and unexpected tokens are skipped rather than causing errors.
- `@theme` blocks may appear multiple times; their tokens are merged.
- `@utility` and `@variant` definitions are collected in source order. The `Order` field on each definition reflects this ordering and is used for CSS output sorting.
- Nested at-rules inside `@utility` blocks (e.g., `@media`, `@supports` for responsive utility definitions) are parsed and preserved as part of the utility definition.
- The parser handles both the parenthesized form (`@variant hover (&:hover);`) and block form (`@variant group-hover { ... }`) of variant definitions.


## 6. Class String Parsing

### 6.1 Structure of a Tailwind Class

A Tailwind class has this general structure:

```
[variant:]...[variant:]  [!]  [-]  utility  [-value]  [/modifier]
|___________________|   |_|  |_|  |_____|  |______|  |________|
     variants         imp  neg   name     value     opacity/
                                                    modifier
```

Examples:
```
flex                          → static utility
p-4                           → utility "p", value "4"
bg-blue-500                   → utility "bg", value "blue-500"
hover:bg-blue-600             → variant "hover", utility "bg", value "blue-600"
dark:md:hover:bg-blue-600     → variants ["dark","md","hover"], utility "bg", value "blue-600"
!font-bold                    → important, utility "font", value "bold"
-translate-x-4                → negative, utility "translate-x", value "4"
w-[300px]                     → utility "w", arbitrary value "300px"
w-1/2                         → utility "w", value "1/2" (fraction)
bg-blue-500/75                → utility "bg", value "blue-500", opacity modifier "75"
[mask-type:alpha]             → arbitrary property "mask-type", value "alpha"
text-[length:--my-size]       → utility "text", arbitrary with type hint "length", value "--my-size"
[&:nth-child(3)]:bg-red-500   → arbitrary variant, utility "bg", value "red-500"
group-hover:text-white        → compound variant "group-hover", utility "text", value "white"
bg-(--my-color)               → utility "bg", custom property shorthand "var(--my-color)"
w-(--sidebar-width)           → utility "w", custom property shorthand "var(--sidebar-width)"
```

### 6.2 Parsing Algorithm

#### Step 1: Split Variants

Split the class string on `:` while respecting bracket depth (a `[` opens, `]` closes; don't split inside brackets). The last segment is the utility-with-value; all preceding segments are variant names.

```
dark:md:hover:bg-blue-500
  → variants: ["dark", "md", "hover"]
  → remainder: "bg-blue-500"

[&:nth-child(3)]:bg-red-500
  → variants: ["[&:nth-child(3)]"]
  → remainder: "bg-red-500"
```

#### Step 2: Detect Arbitrary Property

If the remainder (after variant extraction) starts with `[` and ends with `]`, and contains a `:` not inside nested brackets, it's an arbitrary property:

```
[mask-type:alpha]
  → property: "mask-type"
  → value: "alpha"
```

The content between `[` and `]` has `_` replaced with spaces.

#### Step 3: Extract Modifiers

**Important (`!`):** If the remainder starts with `!`, set the important flag and remove it.

**Negative (`-`):** If the remainder (after removing `!`) starts with `-`, set the negative flag and remove it.

**Opacity modifier (`/`):** If the remainder contains a `/` that is NOT inside brackets, the part after `/` is the opacity modifier:

```
bg-blue-500/75    → value "blue-500", opacity modifier "75"
bg-blue-500/[.5]  → value "blue-500", opacity modifier "[.5]" (arbitrary)
```

#### Step 4: Detect Arbitrary Value

If the remainder contains `-[` followed by `]` at the end:

```
w-[300px]           → utility "w", arbitrary "300px"
grid-cols-[1fr_auto_2fr] → utility "grid-cols", arbitrary "1fr auto 2fr"
text-[#ff0000]      → utility "text", arbitrary "#ff0000"
```

The `_` inside brackets is replaced with space. The utility name is everything before `-[`.

##### Type-Hinted Arbitrary Values

Arbitrary values may include a type hint prefix:

```
text-[length:--my-font-size]  → type hint "length", value "var(--my-font-size)"
bg-[color:--my-bg]            → type hint "color", value "var(--my-bg)"
```

The type hint helps disambiguate when a utility accepts multiple value types. After the type hint is extracted, it's used to select the appropriate `--value()` branch.

If the arbitrary value starts with `--`, it's a custom property reference and should be wrapped in `var()`:

```
w-[--sidebar-width]  → "var(--sidebar-width)"
```

##### Parenthesized Custom Property Syntax

TailwindCSS v4 supports a shorthand syntax using parentheses for CSS custom properties, as an alternative to the bracket syntax:

```
bg-(--my-color)     → background-color: var(--my-color)
w-(--sidebar-width) → width: var(--sidebar-width)
p-(--spacing-lg)    → padding: var(--spacing-lg)
```

This is equivalent to the bracket syntax `bg-[--my-color]` → `background-color: var(--my-color)`, but provides a more ergonomic way to reference custom properties.

**Parsing rules:**
- If the value portion of a utility starts with `(` and ends with `)`, and the content inside starts with `--`, treat it as a custom property reference.
- The content inside the parentheses is wrapped in `var()` just like bare `--` references in bracket syntax.
- Parenthesized syntax does NOT support type hints or arbitrary CSS expressions — it is exclusively for custom property references.
- Underscore-to-space replacement does NOT apply inside parentheses (unlike bracket syntax).

```
bg-(--my-color)      → utility "bg", value "var(--my-color)"
text-(--heading-size) → utility "text", value "var(--heading-size)"
```

#### Step 5: Split Utility Name and Value

This is the ambiguous step. Given `bg-blue-500`, is this:
- utility `bg` with value `blue-500`?
- utility `bg-blue` with value `500`?
- utility `bg-blue-500` (static)?

**Resolution strategy:** The engine tries each possible split point against the utility index, longest-prefix-first. The first pattern that matches **and** can resolve the value wins.

```
bg-blue-500
  Try: "bg-blue-500" as static → not found
  Try: "bg-blue" + "500" → "bg-blue" not in dynamic index
  Try: "bg" + "blue-500" → "bg" is in dynamic index, "blue-500" resolves via --color namespace → match!
```

This is performed during generation (§9), not during class parsing. The class parser produces a preliminary split using heuristics (rightmost hyphen where the right side starts with a digit or is a known keyword), but the generator re-evaluates this against the actual utility registry.


## 7. Theme Resolution

### 7.1 Direct Token Lookup

The simplest resolution: look up `--{namespace}-{value}` in the theme tokens.

```
bg-blue-500  →  look up "--color-blue-500"  →  "#3b82f6"
text-lg      →  look up "--text-lg"          →  "1.125rem"
rounded-md   →  look up "--radius-md"       →  "0.375rem"
```

### 7.2 Spacing Scale Computation

The spacing namespace has a special **computed scale** behavior. The theme defines a base spacing unit:

```css
@theme {
  --spacing: 0.25rem;
}
```

Numeric spacing values are computed by multiplying:

```
p-4    →  "calc(4 * 0.25rem)"  →  or pre-computed "1rem"
p-0.5  →  "calc(0.5 * 0.25rem)"
p-px   →  "1px" (special keyword)
```

The engine should pre-compute simple multiples when possible (e.g., `4 * 0.25rem = 1rem`), but fall back to `calc()` for complex cases.

Named spacing overrides take precedence over computed values:

```css
@theme {
  --spacing: 0.25rem;
  --spacing-18: 4.5rem;  /* overrides the computed 18 * 0.25 = 4.5rem */
}
```

Look up `--spacing-{N}` first; if not found, compute `N * base`.

### 7.3 Keyword Values

Certain value strings map directly to CSS values without theme lookup:

| Tailwind Value | CSS Value |
|----------------|-----------|
| `full` | `100%` |
| `screen` | `100vw` or `100vh` (context-dependent) |
| `svw`, `svh`, `lvw`, `lvh`, `dvw`, `dvh` | `100svw`, `100svh`, etc. |
| `min` | `min-content` |
| `max` | `max-content` |
| `fit` | `fit-content` |
| `auto` | `auto` |
| `none` | `none` |
| `px` | `1px` |
| `inherit` | `inherit` |
| `initial` | `initial` |
| `revert` | `revert` |
| `unset` | `unset` |
| `current` | `currentColor` |
| `transparent` | `transparent` |

### 7.4 Fraction Values

Values containing `/` where both sides are numeric are treated as fractions:

```
w-1/2    →  50%
w-2/3    →  66.666667%
w-3/4    →  75%
w-5/12   →  41.666667%
```

Computed as: `(numerator / denominator) * 100` with sufficient decimal precision.

### 7.5 Arbitrary Value Passthrough

Values wrapped in `[...]` are passed through as raw CSS after:
1. Replacing `_` with space
2. Wrapping bare `--custom-property` references in `var()`
3. Extracting and applying any type hint prefix

```
w-[300px]           →  "300px"
w-[calc(100%-2rem)] →  "calc(100% - 2rem)"  (after _ → space)
w-[--sidebar]       →  "var(--sidebar)"
text-[length:1.5em] →  "1.5em" (type hint "length" used for disambiguation)
```

### 7.6 Negative Values

When the negative modifier is set (`-translate-x-4`), the resolved value is negated:

- If the value is a `calc()` expression: wrap as `calc(-1 * <expression>)`
- If the value is a simple dimension: prepend `-` (e.g., `1rem` → `-1rem`)
- If the value is `0`: remains `0` (no negation needed)

### 7.7 Resolution Priority

When resolving a value for a dynamic utility, try in this order:

1. **Arbitrary value** (`[...]`) — pass through directly
2. **Direct theme token** — `--{namespace}-{value}` exact match
3. **Computed spacing scale** — for `spacing` namespace, multiply by base
4. **Keyword mapping** — check the keyword table
5. **Fraction** — if value contains `/`
6. **Bare value** — some utilities accept raw values (e.g., `z-10` → `10`)

If none resolve, the candidate is discarded (no CSS generated).


## 8. Variant Resolution

### 8.1 Selector Variants

Selector variants modify the CSS selector by replacing `&` in the variant definition with the base selector:

```
focus:bg-blue-500
  base selector: .focus\:bg-blue-500
  variant definition: &:focus
  result: .focus\:bg-blue-500:focus
```

The `hover` variant is special: it applies both a selector transform (`&:hover`) and wraps the rule in `@media (hover: hover)`:

```
hover:bg-blue-500
  base selector: .hover\:bg-blue-500
  result:
    @media (hover: hover) {
      .hover\:bg-blue-500:hover {
        background-color: #3b82f6;
      }
    }
```

Multiple selector variants compose by successive substitution (inner-to-outer):

```
group-hover:focus:text-white
  base selector: .group-hover\:focus\:text-white
  apply "focus": .group-hover\:focus\:text-white:focus
  apply "group-hover": .group:hover .group-hover\:focus\:text-white:focus
```

### 8.2 Media Query Variants

Media query variants wrap the rule in an `@media` block:

```
md:bg-blue-500  →
  @media (width >= 48rem) {
    .md\:bg-blue-500 {
      background-color: #3b82f6;
    }
  }
```

### 8.3 Feature Query Variants

Feature query variants wrap the rule in `@supports`:

```
supports-grid:flex  →
  @supports (display: grid) {
    .supports-grid\:flex {
      display: flex;
    }
  }
```

### 8.4 Container Query Variants

Container query variants wrap the rule in `@container`:

```
@md:p-4  →
  @container (width >= 48rem) {
    .\@md\:p-4 {
      padding: 1rem;
    }
  }
```

### 8.5 Variant Stacking

Variants compose by nesting. The ordering is: **outermost wrapping is the leftmost variant in the class name**:

```
dark:md:hover:bg-blue-500

→ @media (prefers-color-scheme: dark) {           ← dark (outermost)
    @media (width >= 48rem) {                     ← md
      @media (hover: hover) {                     ← hover (media)
        .dark\:md\:hover\:bg-blue-500:hover {     ← hover (selector)
          background-color: #3b82f6;
        }
      }
    }
  }
```

When multiple media queries appear, they may optionally be merged with `and`:

```
@media (prefers-color-scheme: dark) and (width >= 48rem) { ... }
```

However, nesting is also valid CSS and simpler to implement. The engine should use nesting.

### 8.6 Arbitrary Variants

Variants wrapped in `[...]` are arbitrary selector or media patterns:

```
[&:nth-child(3)]:bg-red-500
  → .\\[\\&\\:nth-child\\(3\\)\\]\\:bg-red-500:nth-child(3)

[@media(min-width:900px)]:bg-red-500
  → @media (min-width: 900px) {
       .\[\@media\(min-width\:900px\)\]\:bg-red-500 { ... }
     }
```

The content between `[` and `]` is parsed as:
- If it starts with `@media`, `@supports`, or `@container`: treat as an at-rule wrapper
- Otherwise: treat as a selector pattern where `&` is replaced with the base selector

### 8.7 `group-*` and `peer-*` Variants

These are compound variants that relate one element's state to another:

```
group-hover:text-white
  → .group:hover .group-hover\:text-white { color: white; }

peer-focus:ring-2
  → .peer:focus ~ .peer-focus\:ring-2 { ... }
```

The `group-*` variant definition is:
```css
@variant group-* {
  :merge(.group):{value} & {
    @slot;
  }
}
```

Where `{value}` is replaced with the specific pseudo-class/state from the variant name. So `group-hover` replaces `{value}` with `hover`, producing `:merge(.group):hover &`.

Similarly, `peer-*` uses `~` (sibling combinator):
```css
@variant peer-* {
  :merge(.peer):{value} ~ & {
    @slot;
  }
}
```

The `:merge()` function is handled during selector construction — it means that if multiple utilities use the same group variant, the `.group:hover` portion should appear only once.

##### Named Groups and Peers

Tailwind supports named groups with `/`:

```
group/sidebar              → marks an element as a named group
group-hover/sidebar:flex   → utility activates on hover of the named group

→ .group\/sidebar:hover .group-hover\/sidebar\:flex { display: flex; }
```

### 8.8 `not-*`, `has-*`, `in-*` Variants

```
not-hover:opacity-100   → &:not(:hover)
has-checked:bg-gray-50  → &:has(:checked)
in-data-current:font-bold → [data-current] &
```

These are parameterized variants that wrap the value in a CSS pseudo-class.

### 8.9 Responsive Variant Ordering

Responsive variants (`sm`, `md`, `lg`, `xl`, `2xl`) must appear in the CSS output in ascending breakpoint order so that larger breakpoints override smaller ones. This is ensured by the `Order` field from their definition sequence in the parsed CSS.

### 8.10 `@starting-style` Variant

```css
@variant starting (@starting-style);
```

Wraps the rule in a `@starting-style` block for CSS transition entry animations.

### 8.11 ARIA Variants

ARIA state variants allow styling based on ARIA attributes:

```
aria-checked:bg-blue-500    → &[aria-checked="true"]
aria-disabled:opacity-50    → &[aria-disabled="true"]
aria-expanded:rotate-180    → &[aria-expanded="true"]
aria-hidden:hidden          → &[aria-hidden="true"]
aria-pressed:ring-2         → &[aria-pressed="true"]
aria-readonly:bg-gray-100   → &[aria-readonly="true"]
aria-required:border-red    → &[aria-required="true"]
aria-selected:bg-blue-100   → &[aria-selected="true"]
aria-busy:animate-pulse     → &[aria-busy="true"]
```

Arbitrary ARIA variants are also supported:

```
aria-[sort=ascending]:underline  → &[aria-sort="ascending"]
aria-[labelledby=title]:font-bold → &[aria-labelledby="title"]
```

These are parameterized variants: the `aria-*` prefix is followed by the ARIA attribute name, and the variant generates an attribute selector matching the `"true"` value by default.

### 8.12 `data-*` Variants

Data attribute variants allow styling based on `data-*` attributes:

```
data-active:bg-blue-500      → &[data-active]
data-[size=large]:text-lg    → &[data-size="large"]
data-[loading]:animate-pulse → &[data-loading]
```

The `data-*` variants work similarly to ARIA variants. Without a value, they check for attribute presence. With a bracket value, they check for a specific attribute value.

### 8.13 `supports-*` Variants

Feature query variants using `@supports`:

```
supports-[display:grid]:grid        → @supports (display: grid) { ... }
supports-[backdrop-filter]:backdrop-blur → @supports (backdrop-filter: blur()) { ... }
```

These generate `@supports` at-rule wrappers based on the CSS property and value specified in the brackets.

### 8.14 `max-*` Responsive Variants

Maximum-width responsive variants, the inverse of standard responsive variants:

```
max-sm:hidden   → @media (width < 40rem) { ... }
max-md:flex     → @media (width < 48rem) { ... }
max-lg:block    → @media (width < 64rem) { ... }
max-xl:hidden   → @media (width < 80rem) { ... }
max-2xl:grid    → @media (width < 96rem) { ... }
```

These use `<` (less than) rather than `>=` (greater than or equal), creating a maximum-width constraint. Combined with the standard `min-*` variants, they enable range-based responsive design.

### 8.15 Pointer and Input Device Variants

```
pointer-coarse:p-4      → @media (pointer: coarse) { ... }
pointer-fine:p-2         → @media (pointer: fine) { ... }
pointer-none:hidden      → @media (pointer: none) { ... }
any-pointer-coarse:p-4   → @media (any-pointer: coarse) { ... }
any-pointer-fine:p-2     → @media (any-pointer: fine) { ... }
hover-hover:underline    → @media (hover: hover) { ... }
hover-none:no-underline  → @media (hover: none) { ... }
any-hover-hover:underline → @media (any-hover: hover) { ... }
any-hover-none:no-underline → @media (any-hover: none) { ... }
```

### 8.16 `nth-*` Parameterized Variants

```
nth-[3]:bg-red-500         → &:nth-child(3)
nth-[3n+1]:bg-blue-500     → &:nth-child(3n+1)
nth-last-[2]:bg-green-500  → &:nth-last-child(2)
```

Parameterized variants that generate `:nth-child()` and `:nth-last-child()` pseudo-class selectors with arbitrary arguments.

### 8.17 Child and Descendant Selector Variants

TailwindCSS v4 provides special selector variants for targeting children and descendants:

```
*:p-4           → & > * { padding: 1rem; }
**:text-red-500 → & * { color: ... }
```

- **`*` (child variant):** Applies styles to direct children using `& > *`.
- **`**` (descendant variant):** Applies styles to all descendants using `& *`.

### 8.18 Additional Pseudo-Class Variants

```
user-valid:border-green-500     → &:user-valid
user-invalid:border-red-500     → &:user-invalid
optional:border-gray-300        → &:optional
in-range:border-green-500       → &:in-range
out-of-range:border-red-500     → &:out-of-range
details-content:p-4             → &::details-content
```

These pseudo-classes extend the set of form/element state variants beyond the basic ones listed in §5.1.3.

### 8.19 Additional Media Query Variants

```
noscript:hidden                        → @media (scripting: none) { ... }
inverted-colors:filter-none            → @media (inverted-colors: inverted) { ... }
not-forced-colors:shadow-md            → @media (forced-colors: none) { ... }
```

These media query variants provide additional environment detection beyond the standard set.


## 9. Utility Resolution and CSS Generation

### 9.1 Resolution Pipeline

For each candidate class extracted from the byte stream:

```
candidate string
    │
    ▼
┌──────────────┐
│ Parse class   │  → ParsedClass struct
│ (§6)          │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Match utility │  → UtilityDef + resolved value string
│ (§9.2)        │
└──────┬───────┘
       │  (no match → discard candidate)
       ▼
┌──────────────┐
│ Resolve value │  → CSS value string
│ (§7)          │
└──────┬───────┘
       │  (no resolution → discard candidate)
       ▼
┌──────────────┐
│ Build decls   │  → []Declaration with substituted values
│ (§9.3)        │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Build selector│  → escaped CSS selector with variant transforms
│ (§9.4)        │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Wrap variants │  → media queries, @supports, @container
│ (§8)          │
└──────┬───────┘
       │
       ▼
  generatedRule
```

### 9.2 Utility Matching

Given the full utility-value string (e.g., `bg-blue-500`), the engine attempts to match it against registered utilities:

1. **Static exact match.** Check if the full string matches a static utility name.
2. **Dynamic prefix match.** For each dynamic utility pattern, sorted **longest-first**, check if the full string starts with `{pattern}-`. If so, the remainder after the prefix is the value string.

Longest-first ordering is critical for disambiguation:

```
Registered patterns (sorted by length):
  "translate-x"   (length 11)
  "translate"      (length 9)
  "border-t"       (length 8)
  "border"         (length 6)
  "bg"             (length 2)

Candidate "translate-x-4":
  Try "translate-x" + "4" → match!   (not "translate" + "x-4")

Candidate "border-t-2":
  Try "translate-x" → no prefix match
  Try "translate"   → no prefix match
  Try "border-t" + "2" → match!       (not "border" + "t-2")
```

### 9.3 Declaration Building

Once a utility and value are matched:

1. For each declaration in the utility definition:
   - If the declaration contains `--value(...)`, substitute the resolved CSS value.
   - If the declaration does NOT contain `--value(...)`, emit it verbatim.
2. If multiple declarations have the same property (value resolution alternatives), only the first successfully-resolved one is emitted.
3. Apply the important flag: append ` !important` to each declaration's value.

### 9.4 Selector Construction

The CSS selector is built from the raw class string with special characters escaped:

```
bg-blue-500               → .bg-blue-500
hover:bg-blue-500         → .hover\:bg-blue-500
w-[300px]                 → .w-\[300px\]
!-translate-x-4           → .\!-translate-x-4
text-[#ff0000]            → .text-\[\#ff0000\]
bg-blue-500/75            → .bg-blue-500\/75
[mask-type:alpha]         → .\[mask-type\:alpha\]
```

Characters that must be escaped with `\`: `:`, `[`, `]`, `/`, `(`, `)`, `#`, `.`, `,`, `!`, `+`, `*`, `%`, `@`, `&`, `>`, `~`, space.

After escaping, variant selector transformations are applied (§8.1).


## 10. CSS Output

### 10.1 Layer Ordering

The generated CSS is organized into layers following Tailwind's cascade:

```css
/* 1. Theme variables (custom properties on :root) */
/* 2. Base/reset styles */
/* 3. Component styles */
/* 4. Utility styles (the bulk of generated output) */
```

Within the utility layer, rules are ordered by:
1. The utility definition's `Order` field (source order from the CSS)
2. Within the same utility, variant-wrapped rules come after unwrapped ones
3. Responsive variants are ordered by breakpoint size (ascending)

### 10.2 Deduplication

If the same class appears multiple times in the candidate set, it produces only one rule. The candidate set is a `map[string]struct{}`.

If two different classes produce identical CSS (same selector, same declarations), both are emitted — deduplication is by class name, not by CSS content.

### 10.3 Output Format

Generated CSS uses:
- 2-space indentation for declarations inside rules
- Newline between rules
- Nested `@media` blocks for variant wrapping (not merged)
- No trailing newline after the last rule

Example output:

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

.hover\:bg-blue-500:hover {
  background-color: #3b82f6;
}

@media (width >= 48rem) {
.md\:text-lg {
  font-size: 1.125rem;
}
}
```


## 11. Opacity Modifier

### 11.1 Syntax

```
bg-blue-500/75     → 75% opacity on the color
bg-blue-500/[.5]   → arbitrary opacity value 0.5
text-white/50      → 50% opacity on white
```

### 11.2 Resolution

The opacity modifier changes how color values are emitted. Instead of:

```css
background-color: #3b82f6;
```

The engine emits:

```css
background-color: oklch(from #3b82f6 l c h / 75%);
```

Or using the `color-mix()` function for broader compatibility:

```css
background-color: color-mix(in oklch, #3b82f6 75%, transparent);
```

The engine should use the Tailwind v4 approach, which is `oklch()` with the `from` keyword.

For arbitrary opacity values:
```
bg-blue-500/[.5] → background-color: oklch(from #3b82f6 l c h / 0.5);
```

Theme-defined opacity values are checked first:
```css
@theme {
  --opacity-50: 0.5;
}
```

So `bg-blue-500/50` first looks up `--opacity-50` → `0.5`, then applies it.


## 12. `@apply` Directive

### 12.1 Syntax

```css
.btn {
  @apply bg-blue-500 text-white font-bold py-2 px-4 rounded;
}
```

### 12.2 Behavior

`@apply` expands Tailwind utility classes inline within a custom CSS rule. The engine must:

1. Parse the `@apply` directive from CSS input (either in `LoadCSS` or in separate user-provided CSS).
2. Resolve each class name to its declarations using the same pipeline as §9.
3. Substitute the `@apply` line with the resolved declarations.
4. Flatten: variants in `@apply`'d classes are applied to the containing rule's selector.

### 12.3 Processing Order

`@apply` is processed **after** all utility definitions are loaded and **before** final CSS output. This allows `@apply` to reference any utility, including those defined in later `@utility` blocks.

### 12.4 `@apply` with Variants

```css
.btn {
  @apply hover:bg-blue-700;
}
```

This should produce:

```css
.btn:hover {
  background-color: #1d4ed8;
}
```

The variant modifies the selector of the containing rule, not of the utility class.


## 13. Custom Utilities via `@utility`

Users may define their own utilities in CSS that they pass to `LoadCSS`:

```css
@utility tab-* {
  tab-size: --value(integer);
}

@utility content-auto {
  content-visibility: auto;
}
```

These are parsed and registered identically to Tailwind's built-in utilities. There is no distinction between "built-in" and "custom" — all utilities come from the CSS source.


## 14. `@keyframes` Support

Tailwind defines keyframes for animation utilities:

```css
@keyframes spin {
  to { transform: rotate(360deg); }
}
```

When the `animate-spin` utility is used, the engine must include both:
1. The utility's declarations (`animation: spin 1s linear infinite`)
2. The corresponding `@keyframes` block

The engine should collect `@keyframes` blocks during CSS parsing and include them in the output when any utility references them.


## 15. Gradient Utilities

### 15.1 Linear Gradients

TailwindCSS v4 renames the gradient utilities from v3:

| v3 Syntax | v4 Syntax |
|-----------|-----------|
| `bg-gradient-to-r` | `bg-linear-to-r` |
| `bg-gradient-to-t` | `bg-linear-to-t` |
| `bg-gradient-to-br` | `bg-linear-to-br` |

The `bg-linear-to-*` utilities generate `background-image: linear-gradient(to <direction>, ...)` declarations.

Additionally, v4 introduces angle-based linear gradients:

```
bg-linear-45      → background-image: linear-gradient(45deg, ...)
bg-linear-90      → background-image: linear-gradient(90deg, ...)
bg-linear-[137deg] → background-image: linear-gradient(137deg, ...)
```

### 15.2 Radial Gradients

```
bg-radial          → background-image: radial-gradient(...)
bg-radial-[at_top] → background-image: radial-gradient(at top, ...)
```

### 15.3 Conic Gradients

```
bg-conic           → background-image: conic-gradient(...)
bg-conic-[from_45deg] → background-image: conic-gradient(from 45deg, ...)
```

### 15.4 Gradient Color Stops

Gradient color stops are defined using `from-*`, `via-*`, and `to-*` utilities:

```
from-blue-500      → starting color
via-purple-500     → middle color (optional)
to-pink-500        → ending color
```

**Stop positions** allow specifying where each color stop occurs:

```
from-blue-500 from-10%    → start at 10%
via-purple-500 via-30%    → middle at 30%
to-pink-500 to-90%        → end at 90%
```

### 15.5 Gradient Color Interpolation

TailwindCSS v4 supports color interpolation modifiers that control how colors blend across the gradient:

```
bg-linear-to-r/srgb     → linear-gradient(in srgb, ...)
bg-linear-to-r/oklab    → linear-gradient(in oklab, ...)
bg-linear-to-r/oklch    → linear-gradient(in oklch, ...)
```

The interpolation modifier appears after a `/` on the gradient utility. Supported color spaces include `srgb`, `srgb-linear`, `lab`, `oklab`, `lch`, `oklch`, `hsl`, and `hwb`.


## 16. Additional Utility Categories

### 16.1 Text Shadow Utilities

Text shadow utilities parallel the box shadow system:

```
text-shadow-sm    → text-shadow: 0 1px 1px rgb(0 0 0 / 0.05)
text-shadow       → text-shadow: 0 1px 3px rgb(0 0 0 / 0.1)
text-shadow-md    → text-shadow: 0 2px 4px rgb(0 0 0 / 0.1)
text-shadow-lg    → text-shadow: 0 4px 8px rgb(0 0 0 / 0.1)
text-shadow-none  → text-shadow: none
```

Values are resolved from the `--text-shadow-*` theme namespace.

### 16.2 Mask Utilities

CSS mask utilities for controlling element masking:

```
mask-clip-border    → mask-clip: border-box
mask-clip-padding   → mask-clip: padding-box
mask-clip-content   → mask-clip: content-box
mask-clip-fill      → mask-clip: fill-box
mask-clip-stroke    → mask-clip: stroke-box
mask-clip-view      → mask-clip: view-box
mask-clip-none      → mask-clip: no-clip

mask-composite-add        → mask-composite: add
mask-composite-subtract   → mask-composite: subtract
mask-composite-intersect  → mask-composite: intersect
mask-composite-exclude    → mask-composite: exclude

mask-image-none           → mask-image: none
mask-image-[url(...)]     → mask-image: url(...)

mask-mode-alpha           → mask-mode: alpha
mask-mode-luminance       → mask-mode: luminance
mask-mode-match           → mask-mode: match-source

mask-origin-border    → mask-origin: border-box
mask-origin-padding   → mask-origin: padding-box
mask-origin-content   → mask-origin: content-box
mask-origin-fill      → mask-origin: fill-box
mask-origin-stroke    → mask-origin: stroke-box
mask-origin-view      → mask-origin: view-box

mask-position-*       → mask-position: center, top, etc.
mask-repeat-*         → mask-repeat: repeat, no-repeat, etc.
mask-size-*           → mask-size: auto, cover, contain, etc.

mask-type-alpha       → mask-type: alpha
mask-type-luminance   → mask-type: luminance
```

### 16.3 Font Stretch Utilities

```
font-stretch-ultra-condensed  → font-stretch: ultra-condensed
font-stretch-extra-condensed  → font-stretch: extra-condensed
font-stretch-condensed        → font-stretch: condensed
font-stretch-semi-condensed   → font-stretch: semi-condensed
font-stretch-normal           → font-stretch: normal
font-stretch-semi-expanded    → font-stretch: semi-expanded
font-stretch-expanded         → font-stretch: expanded
font-stretch-extra-expanded   → font-stretch: extra-expanded
font-stretch-ultra-expanded   → font-stretch: ultra-expanded
font-stretch-[75%]            → font-stretch: 75%
```

### 16.3.1 Font Feature Settings

The `font-features` utility sets OpenType font feature settings via arbitrary values:

```
font-features-[smcp]        → font-feature-settings: "smcp"
font-features-["liga"_0]    → font-feature-settings: "liga" 0
```

Bare feature tags are automatically quoted. Already-quoted values are passed through.

### 16.4 Color Scheme Utilities

```
color-scheme-normal           → color-scheme: normal
color-scheme-light            → color-scheme: light
color-scheme-dark             → color-scheme: dark
color-scheme-light-dark       → color-scheme: light dark
```

Controls the preferred color scheme for form controls and UI elements.

### 16.5 Field Sizing Utilities

```
field-sizing-content  → field-sizing: content
field-sizing-fixed    → field-sizing: fixed
```

Controls how form fields size themselves based on their content.

### 16.6 Logical Property Utilities

Logical property utilities map to writing-mode-aware CSS properties.

#### Inline and Block Size

These sizing utilities mirror `w-*`/`h-*` but use logical axes:

```
inline-*        → inline-size: ...     (width in horizontal writing mode)
min-inline-*    → min-inline-size: ...
max-inline-*    → max-inline-size: ...

block-*         → block-size: ...      (height in horizontal writing mode)
min-block-*     → min-block-size: ...
max-block-*     → max-block-size: ...
```

Value resolution follows the same rules as width/height utilities: the spacing scale, the `--width-*`/`--height-*` and `--container-*` theme namespaces (for inline-size variants), and arbitrary length/percentage values.

#### Logical Positioning

```
start-*         → inset-inline-start: ...
end-*           → inset-inline-end: ...
inset-bs-*      → inset-block-start: ...
inset-be-*      → inset-block-end: ...

ms-*            → margin-inline-start: ...
me-*            → margin-inline-end: ...
mbs-*           → margin-block-start: ...
mbe-*           → margin-block-end: ...
ps-*            → padding-inline-start: ...
pe-*            → padding-inline-end: ...
pbs-*           → padding-block-start: ...
pbe-*           → padding-block-end: ...

border-s-*      → border-inline-start-width / border-inline-start-color
border-e-*      → border-inline-end-width / border-inline-end-color
border-bs-*     → border-block-start-width / border-block-start-color
border-be-*     → border-block-end-width / border-block-end-color
rounded-s-*     → border-start-start-radius + border-end-start-radius
rounded-e-*     → border-start-end-radius + border-end-end-radius

rounded-tl-*    → border-top-left-radius
rounded-tr-*    → border-top-right-radius
rounded-br-*    → border-bottom-right-radius
rounded-bl-*    → border-bottom-left-radius

rounded-ss-*    → border-start-start-radius
rounded-se-*    → border-start-end-radius
rounded-ee-*    → border-end-end-radius
rounded-es-*    → border-end-start-radius

scroll-ms-*     → scroll-margin-inline-start: ...
scroll-me-*     → scroll-margin-inline-end: ...
scroll-mbs-*    → scroll-margin-block-start: ...
scroll-mbe-*    → scroll-margin-block-end: ...
scroll-ps-*     → scroll-padding-inline-start: ...
scroll-pe-*     → scroll-padding-inline-end: ...
scroll-pbs-*    → scroll-padding-block-start: ...
scroll-pbe-*    → scroll-padding-block-end: ...
```

These utilities use the spacing theme namespace for value resolution.

### 16.7 Container Utility

The `@container` utility marks an element as a container query context:

```
@container         → container-type: inline-size
@container/sidebar → container-type: inline-size; container-name: sidebar
@container-normal  → container-type: normal
@container-size    → container-type: size
```

Container query variants (§8.4) target these containers.

### 16.8 Backface Visibility Utilities

```
backface-visible  → backface-visibility: visible
backface-hidden   → backface-visibility: hidden
```

### 16.9 Perspective Origin Utilities

```
perspective-origin-center       → perspective-origin: center
perspective-origin-top          → perspective-origin: top
perspective-origin-top-right    → perspective-origin: top right
perspective-origin-right        → perspective-origin: right
perspective-origin-bottom-right → perspective-origin: bottom right
perspective-origin-bottom       → perspective-origin: bottom
perspective-origin-bottom-left  → perspective-origin: bottom left
perspective-origin-left         → perspective-origin: left
perspective-origin-top-left     → perspective-origin: top left
perspective-origin-[25%_75%]    → perspective-origin: 25% 75%
```

### 16.10 Transform Style Utilities

```
transform-3d    → transform-style: preserve-3d
transform-flat  → transform-style: flat
```

### 16.11 3D Transform Utilities

#### Translate Z-axis

`translate-z-*` sets the Z-axis translation using the individual `translate` CSS property:

```css
@utility translate-z-* {
  --tw-translate-z: --value(--spacing);
  --tw-translate-z: --value(length, percentage);
  translate: var(--tw-translate-x) var(--tw-translate-y) var(--tw-translate-z);
}
```

Static utilities:
```
translate-none  → translate: none
translate-3d    → --tw-translate-z: 0px; translate: var(--tw-translate-x) var(--tw-translate-y) var(--tw-translate-z)
```

#### Rotate Per-Axis

`rotate-x-*`, `rotate-y-*`, and `rotate-z-*` set rotation around individual axes:

```css
@utility rotate-x-* {
  rotate: x --value(--rotate);
  rotate: x --value(number);
}

@utility rotate-y-* {
  rotate: y --value(--rotate);
  rotate: y --value(number);
}

@utility rotate-z-* {
  rotate: z --value(--rotate);
  rotate: z --value(number);
}
```

Static utility:
```
rotate-none → rotate: none
```

#### Scale Z-axis

`scale-z-*` sets the Z-axis scale:

```css
@utility scale-z-* {
  --tw-scale-z: --value(--scale);
  --tw-scale-z: --value(percentage, number);
  scale: var(--tw-scale-x, 1) var(--tw-scale-y, 1) var(--tw-scale-z);
}
```

### 16.12 Inset Ring Utilities

Inset ring utilities generate inner ring effects using box shadows:

```
inset-ring-0     → box-shadow: inset 0 0 0 0px ...
inset-ring-1     → box-shadow: inset 0 0 0 1px ...
inset-ring-2     → box-shadow: inset 0 0 0 2px ...
inset-ring       → box-shadow: inset 0 0 0 1px ... (default)

inset-ring-blue-500  → sets the inset ring color
inset-ring-[3px]     → arbitrary width
```

These compose with box shadow and ring utilities using CSS custom properties for shadow stacking.


### 16.13 Font Smoothing Utilities

```
antialiased          → -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;
subpixel-antialiased → -webkit-font-smoothing: auto; -moz-osx-font-smoothing: auto;
```

### 16.14 Wrap Utilities

```
wrap-anywhere  → overflow-wrap: anywhere;
wrap-break-word → overflow-wrap: break-word;
wrap-normal    → overflow-wrap: normal;
```

### 16.15 Safe Alignment Utilities

Safe alignment variants prevent content from becoming inaccessible by falling back when the container is too small:

```
items-center-safe          → align-items: safe center;
items-end-safe             → align-items: safe end;
justify-center-safe        → justify-content: safe center;
justify-end-safe           → justify-content: safe end;
place-items-center-safe    → place-items: safe center;
place-items-end-safe       → place-items: safe end;
place-content-center-safe  → place-content: safe center;
place-content-end-safe     → place-content: safe end;
content-center-safe        → align-content: safe center;
content-end-safe           → align-content: safe end;
self-center-safe           → align-self: safe center;
self-end-safe              → align-self: safe end;
justify-self-center-safe   → justify-self: safe center;
justify-self-end-safe      → justify-self: safe end;
place-self-center-safe     → place-self: safe center;
place-self-end-safe        → place-self: safe end;
```

### 16.16 Baseline Last Utilities

```
items-baseline-last → align-items: baseline last;
self-baseline-last  → align-self: baseline last;
```

### 16.17 Transform Box Utilities

```
transform-content → transform-box: content-box;
transform-border  → transform-box: border-box;
transform-fill    → transform-box: fill-box;
transform-stroke  → transform-box: stroke-box;
transform-view    → transform-box: view-box;
```

### 16.18 Font Variant Numeric Utilities

Font variant numeric utilities control OpenType numeric glyph alternates:

```
normal-nums        → font-variant-numeric: normal
ordinal            → font-variant-numeric: ordinal
slashed-zero       → font-variant-numeric: slashed-zero
lining-nums        → font-variant-numeric: lining-nums
oldstyle-nums      → font-variant-numeric: oldstyle-nums
proportional-nums  → font-variant-numeric: proportional-nums
tabular-nums       → font-variant-numeric: tabular-nums
diagonal-fractions → font-variant-numeric: diagonal-fractions
stacked-fractions  → font-variant-numeric: stacked-fractions
```

These are static utilities. `normal-nums` resets all numeric variants to the default. The other utilities each enable a specific OpenType feature.

### 16.19 Size Utilities

The `size-*` utility sets both `width` and `height` simultaneously:

```
size-4        → width: calc(4 * var(--spacing)); height: calc(4 * var(--spacing))
size-px       → width: 1px; height: 1px
size-0.5      → width: calc(0.5 * var(--spacing)); height: calc(0.5 * var(--spacing))
size-full     → width: 100%; height: 100%
size-auto     → width: auto; height: auto
size-min      → width: min-content; height: min-content
size-max      → width: max-content; height: max-content
size-fit      → width: fit-content; height: fit-content
size-[200px]  → width: 200px; height: 200px
```

Value resolution uses the spacing scale, the `--size-*` theme namespace, and accepts arbitrary length/percentage values. Static keyword variants (`auto`, `full`, `min`, `max`, `fit`) are registered separately.

### 16.20 Forced Color Adjust Utilities

```
forced-color-adjust-auto → forced-color-adjust: auto
forced-color-adjust-none → forced-color-adjust: none
```

Controls whether the browser should force colors when the user has enabled a forced-color mode (e.g., Windows High Contrast).

### 16.21 Hyphens Utilities

```
hyphens-none   → -webkit-hyphens: none; hyphens: none
hyphens-manual → -webkit-hyphens: manual; hyphens: manual
hyphens-auto   → -webkit-hyphens: auto; hyphens: auto
```

Controls how words are hyphenated when wrapping across multiple lines. Includes the `-webkit-hyphens` vendor prefix for Safari compatibility.

### 16.22 Text Indent Utility

The `indent-*` utility sets the `text-indent` CSS property:

```
indent-4       → text-indent: calc(4 * var(--spacing))
indent-px      → text-indent: 1px
indent-[50px]  → text-indent: 50px
```

Value resolution uses the spacing scale and accepts arbitrary length/percentage values. Supports negative values via the `-indent-*` syntax.

### 16.23 Word Break Utilities

```
break-normal → overflow-wrap: normal; word-break: normal
break-words  → overflow-wrap: break-word
break-all    → word-break: break-all
break-keep   → word-break: keep-all
```

Controls how words break when reaching the end of a line. `break-normal` resets both `overflow-wrap` and `word-break` to their default behavior.

### 16.24 Vertical Align Utilities

The `align-*` utility sets the `vertical-align` CSS property:

```
align-baseline    → vertical-align: baseline
align-top         → vertical-align: top
align-middle      → vertical-align: middle
align-bottom      → vertical-align: bottom
align-text-top    → vertical-align: text-top
align-text-bottom → vertical-align: text-bottom
align-sub         → vertical-align: sub
align-super       → vertical-align: super
```

These are static utilities for controlling the vertical alignment of inline or table-cell elements.

### 16.25 Contain Utilities

```
contain-none        → contain: none
contain-content     → contain: content
contain-strict      → contain: strict
contain-size        → contain: size
contain-inline-size → contain: inline-size
contain-layout      → contain: layout
contain-paint       → contain: paint
contain-style       → contain: style
```

Controls the CSS containment model, which lets browsers optimize rendering performance by isolating parts of the page.

### 16.26 Will Change Utilities

```
will-change-auto      → will-change: auto
will-change-scroll    → will-change: scroll-position
will-change-contents  → will-change: contents
will-change-transform → will-change: transform
```

Hints to the browser which properties are expected to change, allowing it to set up optimizations ahead of time. These are static utilities with fixed keyword values.

### 16.27 Transition Utilities

The `transition-*` utilities control the `transition-property` CSS property and apply default duration and timing function from the theme:

```
transition-none      → transition-property: none
transition-all       → transition-property: all; transition-timing-function: ...; transition-duration: ...
transition           → transition-property: color, background-color, border-color, ... (common properties); transition-timing-function: ...; transition-duration: ...
transition-colors    → transition-property: color, background-color, border-color, text-decoration-color, fill, stroke; transition-timing-function: ...; transition-duration: ...
transition-opacity   → transition-property: opacity; transition-timing-function: ...; transition-duration: ...
transition-shadow    → transition-property: box-shadow; transition-timing-function: ...; transition-duration: ...
transition-transform → transition-property: transform, translate, scale, rotate; transition-timing-function: ...; transition-duration: ...
```

All transition utilities except `transition-none` also set:
- `transition-timing-function: var(--tw-ease, var(--default-transition-timing-function))`
- `transition-duration: var(--tw-duration, var(--default-transition-duration))`

This means using any `transition-*` utility (other than `none`) automatically applies the default 150ms duration and ease curve, unless overridden by explicit `duration-*` or `ease-*` utilities.

The bare `transition` utility covers a comprehensive set of commonly animated properties including colors, opacity, box-shadow, transform, filters, and layout-related properties like `display`, `content-visibility`, `overlay`, and `pointer-events`.

### 16.28 Scroll Snap Utilities

Scroll snap utilities control CSS scroll snapping behavior:

```
snap-none       → scroll-snap-type: none
snap-x          → scroll-snap-type: x var(--tw-scroll-snap-strictness, proximity)
snap-y          → scroll-snap-type: y var(--tw-scroll-snap-strictness, proximity)
snap-both       → scroll-snap-type: both var(--tw-scroll-snap-strictness, proximity)
snap-mandatory  → --tw-scroll-snap-strictness: mandatory
snap-proximity  → --tw-scroll-snap-strictness: proximity
```

The `snap-x`, `snap-y`, and `snap-both` utilities compose with the strictness utilities via the `--tw-scroll-snap-strictness` custom property, defaulting to `proximity` if no strictness utility is specified.

Scroll snap alignment:
```
snap-start      → scroll-snap-align: start
snap-end        → scroll-snap-align: end
snap-center     → scroll-snap-align: center
snap-align-none → scroll-snap-align: none
```

Scroll snap stop:
```
snap-normal → scroll-snap-stop: normal
snap-always → scroll-snap-stop: always
```

### 16.29 Object Fit and Position Utilities

Object fit utilities control how replaced elements (images, videos) are resized:

```
object-contain    → object-fit: contain
object-cover      → object-fit: cover
object-fill       → object-fit: fill
object-none       → object-fit: none
object-scale-down → object-fit: scale-down
```

Object position utilities control the alignment of replaced elements within their container:

```
object-bottom       → object-position: bottom
object-center       → object-position: center
object-left         → object-position: left
object-left-bottom  → object-position: left bottom
object-left-top     → object-position: left top
object-right        → object-position: right
object-right-bottom → object-position: right bottom
object-right-top    → object-position: right top
object-top          → object-position: top
```

All object utilities are static with fixed keyword values.

## 17. Dark Mode

### 17.1 Media-Based (Default)

```css
@variant dark (@media (prefers-color-scheme: dark));
```

Uses `@media (prefers-color-scheme: dark)` wrapping.

### 17.2 Selector-Based

Projects may override the dark variant to use a selector strategy:

```css
@variant dark (&:where(.dark, .dark *));
```

This applies the utility when an ancestor has the `.dark` class.

The engine doesn't need to know which strategy is in use — it simply applies whatever variant definition is registered for `dark`.


## 18. Important Modifier

### 18.1 Per-Utility Important

```
!font-bold  →  font-weight: 700 !important
!p-4        →  padding: 1rem !important
```

The `!` prefix causes all declarations in the generated rule to have `!important` appended.

### 18.2 Interaction with Variants

```
hover:!bg-blue-500  →
  .hover\:\!bg-blue-500:hover {
    background-color: #3b82f6 !important;
  }
```

The `!` flag is independent of variants and always affects the declarations, not the selector or media query.


## 19. Negative Values

### 19.1 Syntax

```
-m-4           → margin: -1rem
-translate-x-4 → --tw-translate-x: -1rem
-top-2         → top: -0.5rem
```

The `-` prefix before the utility name negates the resolved value.

### 19.2 Negation Rules

- Simple dimensions: prepend `-` (e.g., `1rem` → `-1rem`)
- `calc()` expressions: wrap as `calc(-1 * <original>)`
- Zero: no negation needed
- Non-numeric values: negation is a no-op (the class is discarded)
- `auto`, `none`, color values: cannot be negated (class is discarded)


## 20. Concurrency

### 20.1 Thread Safety Guarantees

The Engine is safe for concurrent use:

- **`Write()`**: Multiple goroutines may call `Write` concurrently. The candidate map is protected by a mutex. The scanner's cross-chunk buffering assumes sequential delivery within a single stream, but concurrent writes from different streams are safe (worst case: a token at a stream boundary is split, producing two non-matching candidates instead of one matching one — harmless by §4.5's candidate filtering principle).
- **`CSS()`**: May be called while `Write` is in progress. Takes a snapshot of the current candidate set under a read lock. Does not modify engine state.
- **`LoadCSS()`**: Must not be called concurrently with `CSS()`. Typically called during initialization before any writes begin.
- **`Reset()`**: Must not be called concurrently with `Write()` or `CSS()`.

### 20.2 Recommended Usage Patterns

**Build-time tool:**
```go
engine := tailwind.New()
engine.LoadCSS(tailwindSource)

// Scan source files (can be parallelized):
for _, file := range sourceFiles {
    f, _ := os.Open(file)
    io.Copy(engine, f)
    f.Close()
}

css := engine.CSS()
os.WriteFile("output.css", []byte(css), 0644)
```

**HTTP middleware:**
```go
engine := tailwind.New()
engine.LoadCSS(tailwindSource)

// Per-request: use passthrough to observe template output
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    rw := tailwind.NewPassthrough(w)
    tmpl.Execute(rw, data)
    // Note: CSS is accumulated across all requests.
    // For per-request CSS, use a separate engine per request.
})
```

**Server with shared accumulation:**
```go
// Shared engine accumulates classes from all requests.
// CSS is regenerated periodically or on demand.
var engine = tailwind.New()
engine.LoadCSS(tailwindSource)

// In request handlers:
engine.Write(templateOutput)

// In a background goroutine or admin endpoint:
css := engine.CSS()
```


## 21. Error Handling

### 21.1 `LoadCSS` Errors

`LoadCSS` returns an error only for fundamental parse failures (e.g., completely unparseable input). Individual malformed rules or unknown constructs are silently skipped. This is by design: the engine should work with partial or future CSS syntax it doesn't fully understand.

### 21.2 `Write` Errors

`Write` only returns errors from the passthrough writer. The scanner itself cannot fail — any byte sequence is valid input (it may not produce useful candidates, but that's not an error).

### 21.3 `CSS` Errors

`CSS` never returns an error. Unresolvable candidates are silently dropped. If no candidates match any utilities, the output is an empty string.

### 21.4 Diagnostics

For debugging, the engine should provide optional diagnostic methods:

```go
func (e *Engine) Diagnostics() Diagnostics
```

```go
type Diagnostics struct {
    TotalCandidates   int      // Number of unique candidates extracted
    MatchedCandidates int      // Number that resolved to utilities
    DroppedCandidates []string // Candidates that didn't match anything
    UtilityCount      int      // Number of registered utility patterns
    VariantCount      int      // Number of registered variants
    ThemeTokenCount   int      // Number of theme tokens
}
```

This is informational only and has no effect on behavior.


## 22. Edge Cases

### 22.1 Empty Input

- `LoadCSS([]byte{})` — no error, no registries populated, `CSS()` returns `""`.
- `Write([]byte{})` — no-op, returns `(0, nil)`.
- `CSS()` with no writes — returns `""`.

### 22.2 Unknown Classes

Classes that don't match any utility are silently dropped. No warning, no error.

### 22.3 Conflicting Utility Definitions

If two `@utility` blocks define the same pattern, the later one wins (last-write-wins during `LoadCSS`).

### 22.4 Very Long Class Names

No length limit on class names. The scanner will accumulate tokens of any length.

### 22.5 Binary Content

If binary content is written to the engine (e.g., an image accidentally piped through), the scanner will extract nonsensical tokens. These won't match any utilities and are harmless. Performance may degrade slightly due to the large number of false-positive candidates.

### 22.6 Nested `@theme` in `@utility`

Invalid per the Tailwind spec. The parser should skip the nested `@theme` and parse the `@utility` body normally.

### 22.7 Recursive `@apply`

If a class referenced in `@apply` itself uses `@apply`, the engine must detect the recursion and stop. Maximum recursion depth: 10. Beyond that, the `@apply` directive is left unresolved and a diagnostic is emitted.


## 23. Performance Expectations

### 23.1 Scanning

The scanner should process bytes at close to memory-copy speed. There are no allocations in the hot path (token bytes are accumulated in a reusable buffer). Target: **>500 MB/s** on modern hardware.

### 23.2 CSS Generation

Generation is proportional to the number of matched candidates, not total candidates. For a typical project with 500-2000 matched utilities:
- Utility matching: O(candidates × utility_count) with early exit on match
- Value resolution: O(1) per candidate (hash map lookups)
- Target: **<10ms** for 2000 candidates

### 23.3 Memory

The engine's memory footprint is:
- Theme tokens: ~100KB for Tailwind's full default theme
- Utility definitions: ~200KB for all Tailwind utilities
- Variant definitions: ~10KB
- Candidate set: proportional to unique tokens in scanned content
- Scanner buffer: negligible (one partial token at a time)

Total baseline: **<1MB** for a fully-loaded engine.


## 24. Testing Strategy

### 24.1 Unit Tests

Each component is tested in isolation:

- **Tokenizer**: Known CSS strings → expected token sequences
- **Parser**: `@theme`, `@utility`, `@variant` blocks → expected data structures
- **Class parser**: Class strings → expected `ParsedClass` fields
- **Scanner**: Byte streams → expected candidate sets (including chunk-split cases)
- **Theme resolver**: Namespace + key → expected CSS values
- **Generator**: Candidate + definitions → expected CSS output

### 24.2 Integration Tests

End-to-end tests: HTML/template bytes in → CSS string out. These use representative subsets of Tailwind's actual CSS source as the engine's input.

### 24.3 Compatibility Tests

Parse Tailwind's actual full CSS source (downloaded from npm or CDN) and run it through the engine with known class sets. Compare output against Tailwind's own CLI output for the same classes.

This is the definitive correctness test. If the engine produces different CSS than Tailwind's CLI for the same input, the engine has a bug.

### 24.4 Fuzz Testing

Use Go's built-in fuzzing (`go test -fuzz`) on:
- The CSS tokenizer (arbitrary byte input should never panic)
- The class parser (arbitrary string input should never panic)
- The scanner (arbitrary byte input should never panic)
- The full pipeline (arbitrary bytes through Write → CSS should never panic)


## 25. Future Considerations

### 25.1 Incremental Generation

For watch-mode use cases, the engine should support incremental updates: when new candidates appear, generate only the new rules rather than regenerating everything. This requires tracking which candidates have already been generated.

### 25.2 Source Maps

For debugging, the engine could emit source maps linking generated CSS selectors back to the utility class names that produced them.

### 25.3 CSS Nesting Output

Modern CSS supports nesting. The engine could optionally emit nested CSS for smaller output:

```css
.bg-blue-500 {
  background-color: #3b82f6;
  &:hover { background-color: #2563eb; }
}
```

### 25.4 Updating the Bundled CSS

When a new Tailwind CSS v4 release is published, running `go generate ./...` fetches the latest version. The embedded default CSS (§3) should be updated periodically to track upstream releases.

### 25.5 Plugin System

Allow Go functions to register custom utility generators that go beyond what `@utility` can express in CSS. This would enable plugins equivalent to Tailwind's JS plugin system.
