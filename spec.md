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
- **Preflight / reset styles.** The engine generates utility CSS. Base/reset styles from the Tailwind source are passed through verbatim if present, but the engine does not generate them.


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

The output is a single string containing valid CSS. Rules are ordered according to the Tailwind layer system (see §9).

Calls `Flush()` implicitly before generation.

Thread safety: may be called concurrently with `Write()`. Takes a snapshot of the current candidate set and reads (but does not modify) the utility/variant/theme registries.

#### 2.1.7 Resetting

```go
func (e *Engine) Reset()
```

Clears all accumulated candidates and resets the scanner state. Theme, utility, and variant definitions are preserved. Use this to re-scan a different set of source files without reloading the CSS.


## 3. Byte Stream Scanner

### 3.1 Overview

The scanner converts a raw byte stream into a set of candidate class name strings. It has **no knowledge** of any markup language, template syntax, or file format. It operates purely on bytes.

### 3.2 Tokenization Rules

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

### 3.3 Bracket Depth Tracking

Square brackets `[` and `]` receive special handling. When the scanner encounters `[`, it increments a depth counter. While depth > 0, **all bytes** (including what would normally be delimiters) are accumulated into the current token. This is necessary because arbitrary values can contain characters like spaces (encoded as `_`) and commas:

```
grid-cols-[1fr_auto_2fr]    ← single token
bg-[rgb(255,0,0)]           ← single token
[mask-type:alpha]            ← single token
```

The depth counter decrements on `]`. Normal tokenization resumes when depth returns to 0.

### 3.4 Cross-Chunk Token Reconstruction

The scanner maintains a byte buffer (`[]byte`) across `Write` calls. When a non-delimiter byte arrives, it's appended to the buffer. When a delimiter byte arrives, the buffer is flushed as a completed token. This correctly handles tokens split across chunk boundaries.

### 3.5 Candidate Filtering

Before accepting a completed token as a candidate, the scanner applies lightweight rejection filters:

1. **Must start with a letter, `!`, `-`, or `[`.** Rejects tokens starting with digits, punctuation, etc.
2. **Single non-letter characters are rejected.** A lone `-` or `!` is not a class.
3. **URLs are rejected.** Any token containing `://` is discarded.
4. **Unbalanced brackets are rejected.** If `[` and `]` counts don't match, discard.

These filters are deliberately **conservative** (reject obvious non-classes) rather than aggressive. The scanner should never reject a valid Tailwind class. Accepting non-classes is fine — they'll be discarded during generation.

### 3.6 What the Scanner Does NOT Do

- It does not parse HTML attributes to find `class="..."`.
- It does not understand Go template syntax (`{{.Field}}`).
- It does not look for `className=` (JSX).
- It does not handle string interpolation or concatenation.
- It does not detect or extract classes from JavaScript/TypeScript source.

All of these would couple the scanner to specific formats. Instead, the scanner extracts **every plausible token** from the byte stream. The tradeoff is a slightly larger candidate set (more misses during generation), which has negligible performance impact.


## 4. CSS Source Parsing

### 4.1 Supported Tailwind v4 Dialect

The parser understands the following CSS constructs:

#### 4.1.1 `@theme` Blocks

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

##### Theme Namespaces

Theme tokens are organized into namespaces by their property name prefix. The engine recognizes these namespaces for value resolution:

| Prefix | Namespace | Example |
|--------|-----------|---------|
| `--color-` | `color` | `--color-blue-500` |
| `--spacing` | `spacing` | `--spacing` (base), `--spacing-*` (overrides) |
| `--breakpoint-` | `breakpoint` | `--breakpoint-md` |
| `--font-family-` | `font-family` | `--font-family-sans` |
| `--font-size-` | `font-size` | `--font-size-lg` |
| `--font-weight-` | `font-weight` | `--font-weight-bold` |
| `--line-height-` | `line-height` | `--line-height-tight` |
| `--letter-spacing-` | `letter-spacing` | `--letter-spacing-wide` |
| `--radius-` | `radius` | `--radius-lg` |
| `--shadow-` | `shadow` | `--shadow-md` |
| `--inset-shadow-` | `inset-shadow` | `--inset-shadow-sm` |
| `--drop-shadow-` | `drop-shadow` | `--drop-shadow-lg` |
| `--blur-` | `blur` | `--blur-md` |
| `--opacity-` | `opacity` | `--opacity-50` |
| `--transition-property-` | `transition-property` | `--transition-property-all` |
| `--ease-` | `ease` | `--ease-in-out` |
| `--animate-` | `animate` | `--animate-spin` |
| `--perspective-` | `perspective` | `--perspective-dramatic` |
| `--aspect-` | `aspect` | `--aspect-video` |
| `--container-` | `container` | `--container-3xs` |
| `--width-` | `width` | `--width-prose` |
| `--z-` | `z` | `--z-50` |

#### 4.1.2 `@utility` Blocks

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
  clip: rect(0, 0, 0, 0);
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

#### 4.1.3 `@variant` Directives

```css
@variant hover (&:hover);
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

@variant open (&[open]);
@variant inert (&:is([inert], [inert] *));
```

Each `@variant` declaration specifies:

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

#### 4.1.4 `@layer` Directives

```css
@layer theme, base, components, utilities;
```

Tailwind uses CSS layers to establish cascade ordering. The engine must respect this ordering when emitting CSS (see §9).

#### 4.1.5 `@import` and `@config`

```css
@import "tailwindcss";
@config "./tailwind.config.js";
```

These are directives that Tailwind's build tool processes. The Go engine ignores them (it receives already-resolved CSS, not unprocessed source).

#### 4.1.6 `@source`

```css
@source "../src/**/*.html";
```

Specifies content paths for class scanning. The Go engine ignores this — scanning is done via the `io.Writer` interface, not file globs.

#### 4.1.7 Standard CSS Rules

Any CSS rules that aren't Tailwind-specific at-rules (e.g., `:root` blocks, reset styles, `@keyframes`) are collected as **base rules**. These may be included in the output verbatim (see §9.1).

### 4.2 CSS Tokenizer

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

### 4.3 Parser

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


## 5. Class String Parsing

### 5.1 Structure of a Tailwind Class

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
```

### 5.2 Parsing Algorithm

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

This is performed during generation (§8), not during class parsing. The class parser produces a preliminary split using heuristics (rightmost hyphen where the right side starts with a digit or is a known keyword), but the generator re-evaluates this against the actual utility registry.


## 6. Theme Resolution

### 6.1 Direct Token Lookup

The simplest resolution: look up `--{namespace}-{value}` in the theme tokens.

```
bg-blue-500  →  look up "--color-blue-500"  →  "#3b82f6"
text-lg      →  look up "--font-size-lg"    →  "1.125rem"
rounded-md   →  look up "--radius-md"       →  "0.375rem"
```

### 6.2 Spacing Scale Computation

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

### 6.3 Keyword Values

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

### 6.4 Fraction Values

Values containing `/` where both sides are numeric are treated as fractions:

```
w-1/2    →  50%
w-2/3    →  66.666667%
w-3/4    →  75%
w-5/12   →  41.666667%
```

Computed as: `(numerator / denominator) * 100` with sufficient decimal precision.

### 6.5 Arbitrary Value Passthrough

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

### 6.6 Negative Values

When the negative modifier is set (`-translate-x-4`), the resolved value is negated:

- If the value is a `calc()` expression: wrap as `calc(-1 * <expression>)`
- If the value is a simple dimension: prepend `-` (e.g., `1rem` → `-1rem`)
- If the value is `0`: remains `0` (no negation needed)

### 6.7 Resolution Priority

When resolving a value for a dynamic utility, try in this order:

1. **Arbitrary value** (`[...]`) — pass through directly
2. **Direct theme token** — `--{namespace}-{value}` exact match
3. **Computed spacing scale** — for `spacing` namespace, multiply by base
4. **Keyword mapping** — check the keyword table
5. **Fraction** — if value contains `/`
6. **Bare value** — some utilities accept raw values (e.g., `z-10` → `10`)

If none resolve, the candidate is discarded (no CSS generated).


## 7. Variant Resolution

### 7.1 Selector Variants

Selector variants modify the CSS selector by replacing `&` in the variant definition with the base selector:

```
hover:bg-blue-500
  base selector: .hover\:bg-blue-500
  variant definition: &:hover
  result: .hover\:bg-blue-500:hover
```

Multiple selector variants compose by successive substitution (inner-to-outer):

```
group-hover:focus:text-white
  base selector: .group-hover\:focus\:text-white
  apply "focus": .group-hover\:focus\:text-white:focus
  apply "group-hover": .group:hover .group-hover\:focus\:text-white:focus
```

### 7.2 Media Query Variants

Media query variants wrap the rule in an `@media` block:

```
md:bg-blue-500  →
  @media (width >= 48rem) {
    .md\:bg-blue-500 {
      background-color: #3b82f6;
    }
  }
```

### 7.3 Feature Query Variants

Feature query variants wrap the rule in `@supports`:

```
supports-grid:flex  →
  @supports (display: grid) {
    .supports-grid\:flex {
      display: flex;
    }
  }
```

### 7.4 Container Query Variants

Container query variants wrap the rule in `@container`:

```
@md:p-4  →
  @container (width >= 48rem) {
    .\@md\:p-4 {
      padding: 1rem;
    }
  }
```

### 7.5 Variant Stacking

Variants compose by nesting. The ordering is: **outermost wrapping is the leftmost variant in the class name**:

```
dark:md:hover:bg-blue-500

→ @media (prefers-color-scheme: dark) {      ← dark (outermost)
    @media (width >= 48rem) {                ← md
      .dark\:md\:hover\:bg-blue-500:hover { ← hover (selector)
        background-color: #3b82f6;
      }
    }
  }
```

When multiple media queries appear, they may optionally be merged with `and`:

```
@media (prefers-color-scheme: dark) and (width >= 48rem) { ... }
```

However, nesting is also valid CSS and simpler to implement. The engine should use nesting.

### 7.6 Arbitrary Variants

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

### 7.7 `group-*` and `peer-*` Variants

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

### 7.8 `not-*`, `has-*`, `in-*` Variants

```
not-hover:opacity-100   → &:not(:hover)
has-checked:bg-gray-50  → &:has(:checked)
in-data-current:font-bold → [data-current] &
```

These are parameterized variants that wrap the value in a CSS pseudo-class.

### 7.9 Responsive Variant Ordering

Responsive variants (`sm`, `md`, `lg`, `xl`, `2xl`) must appear in the CSS output in ascending breakpoint order so that larger breakpoints override smaller ones. This is ensured by the `Order` field from their definition sequence in the parsed CSS.

### 7.10 `@starting-style` Variant

```css
@variant starting (@starting-style);
```

Wraps the rule in a `@starting-style` block for CSS transition entry animations.


## 8. Utility Resolution and CSS Generation

### 8.1 Resolution Pipeline

For each candidate class extracted from the byte stream:

```
candidate string
    │
    ▼
┌──────────────┐
│ Parse class   │  → ParsedClass struct
│ (§5)          │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Match utility │  → UtilityDef + resolved value string
│ (§8.2)        │
└──────┬───────┘
       │  (no match → discard candidate)
       ▼
┌──────────────┐
│ Resolve value │  → CSS value string
│ (§6)          │
└──────┬───────┘
       │  (no resolution → discard candidate)
       ▼
┌──────────────┐
│ Build decls   │  → []Declaration with substituted values
│ (§8.3)        │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Build selector│  → escaped CSS selector with variant transforms
│ (§8.4)        │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Wrap variants │  → media queries, @supports, @container
│ (§7)          │
└──────┬───────┘
       │
       ▼
  generatedRule
```

### 8.2 Utility Matching

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

### 8.3 Declaration Building

Once a utility and value are matched:

1. For each declaration in the utility definition:
   - If the declaration contains `--value(...)`, substitute the resolved CSS value.
   - If the declaration does NOT contain `--value(...)`, emit it verbatim.
2. If multiple declarations have the same property (value resolution alternatives), only the first successfully-resolved one is emitted.
3. Apply the important flag: append ` !important` to each declaration's value.

### 8.4 Selector Construction

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

After escaping, variant selector transformations are applied (§7.1).


## 9. CSS Output

### 9.1 Layer Ordering

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

### 9.2 Deduplication

If the same class appears multiple times in the candidate set, it produces only one rule. The candidate set is a `map[string]struct{}`.

If two different classes produce identical CSS (same selector, same declarations), both are emitted — deduplication is by class name, not by CSS content.

### 9.3 Output Format

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


## 10. Opacity Modifier

### 10.1 Syntax

```
bg-blue-500/75     → 75% opacity on the color
bg-blue-500/[.5]   → arbitrary opacity value 0.5
text-white/50      → 50% opacity on white
```

### 10.2 Resolution

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


## 11. `@apply` Directive

### 11.1 Syntax

```css
.btn {
  @apply bg-blue-500 text-white font-bold py-2 px-4 rounded;
}
```

### 11.2 Behavior

`@apply` expands Tailwind utility classes inline within a custom CSS rule. The engine must:

1. Parse the `@apply` directive from CSS input (either in `LoadCSS` or in separate user-provided CSS).
2. Resolve each class name to its declarations using the same pipeline as §8.
3. Substitute the `@apply` line with the resolved declarations.
4. Flatten: variants in `@apply`'d classes are applied to the containing rule's selector.

### 11.3 Processing Order

`@apply` is processed **after** all utility definitions are loaded and **before** final CSS output. This allows `@apply` to reference any utility, including those defined in later `@utility` blocks.

### 11.4 `@apply` with Variants

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


## 12. Custom Utilities via `@utility`

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


## 13. `@keyframes` Support

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


## 14. Dark Mode

### 14.1 Media-Based (Default)

```css
@variant dark (@media (prefers-color-scheme: dark));
```

Uses `@media (prefers-color-scheme: dark)` wrapping.

### 14.2 Selector-Based

Projects may override the dark variant to use a selector strategy:

```css
@variant dark (&:where(.dark, .dark *));
```

This applies the utility when an ancestor has the `.dark` class.

The engine doesn't need to know which strategy is in use — it simply applies whatever variant definition is registered for `dark`.


## 15. Important Modifier

### 15.1 Per-Utility Important

```
!font-bold  →  font-weight: 700 !important
!p-4        →  padding: 1rem !important
```

The `!` prefix causes all declarations in the generated rule to have `!important` appended.

### 15.2 Interaction with Variants

```
hover:!bg-blue-500  →
  .hover\:\!bg-blue-500:hover {
    background-color: #3b82f6 !important;
  }
```

The `!` flag is independent of variants and always affects the declarations, not the selector or media query.


## 16. Negative Values

### 16.1 Syntax

```
-m-4           → margin: -1rem
-translate-x-4 → --tw-translate-x: -1rem
-top-2         → top: -0.5rem
```

The `-` prefix before the utility name negates the resolved value.

### 16.2 Negation Rules

- Simple dimensions: prepend `-` (e.g., `1rem` → `-1rem`)
- `calc()` expressions: wrap as `calc(-1 * <original>)`
- Zero: no negation needed
- Non-numeric values: negation is a no-op (the class is discarded)
- `auto`, `none`, color values: cannot be negated (class is discarded)


## 17. Concurrency

### 17.1 Thread Safety Guarantees

The Engine is safe for concurrent use:

- **`Write()`**: Multiple goroutines may call `Write` concurrently. The candidate map is protected by a mutex. The scanner's cross-chunk buffering assumes sequential delivery within a single stream, but concurrent writes from different streams are safe (worst case: a token at a stream boundary is split, producing two non-matching candidates instead of one matching one — harmless by §3.5's over-extraction principle).
- **`CSS()`**: May be called while `Write` is in progress. Takes a snapshot of the current candidate set under a read lock. Does not modify engine state.
- **`LoadCSS()`**: Must not be called concurrently with `CSS()`. Typically called during initialization before any writes begin.
- **`Reset()`**: Must not be called concurrently with `Write()` or `CSS()`.

### 17.2 Recommended Usage Patterns

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


## 18. Error Handling

### 18.1 `LoadCSS` Errors

`LoadCSS` returns an error only for fundamental parse failures (e.g., completely unparseable input). Individual malformed rules or unknown constructs are silently skipped. This is by design: the engine should work with partial or future CSS syntax it doesn't fully understand.

### 18.2 `Write` Errors

`Write` only returns errors from the passthrough writer. The scanner itself cannot fail — any byte sequence is valid input (it may not produce useful candidates, but that's not an error).

### 18.3 `CSS` Errors

`CSS` never returns an error. Unresolvable candidates are silently dropped. If no candidates match any utilities, the output is an empty string.

### 18.4 Diagnostics

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


## 19. Edge Cases

### 19.1 Empty Input

- `LoadCSS([]byte{})` — no error, no registries populated, `CSS()` returns `""`.
- `Write([]byte{})` — no-op, returns `(0, nil)`.
- `CSS()` with no writes — returns `""`.

### 19.2 Unknown Classes

Classes that don't match any utility are silently dropped. No warning, no error.

### 19.3 Conflicting Utility Definitions

If two `@utility` blocks define the same pattern, the later one wins (last-write-wins during `LoadCSS`).

### 19.4 Very Long Class Names

No length limit on class names. The scanner will accumulate tokens of any length.

### 19.5 Binary Content

If binary content is written to the engine (e.g., an image accidentally piped through), the scanner will extract nonsensical tokens. These won't match any utilities and are harmless. Performance may degrade slightly due to the large number of false-positive candidates.

### 19.6 Nested `@theme` in `@utility`

Invalid per the Tailwind spec. The parser should skip the nested `@theme` and parse the `@utility` body normally.

### 19.7 Recursive `@apply`

If a class referenced in `@apply` itself uses `@apply`, the engine must detect the recursion and stop. Maximum recursion depth: 10. Beyond that, the `@apply` directive is left unresolved and a diagnostic is emitted.


## 20. Performance Expectations

### 20.1 Scanning

The scanner should process bytes at close to memory-copy speed. There are no allocations in the hot path (token bytes are accumulated in a reusable buffer). Target: **>500 MB/s** on modern hardware.

### 20.2 CSS Generation

Generation is proportional to the number of matched candidates, not total candidates. For a typical project with 500-2000 matched utilities:
- Utility matching: O(candidates × utility_count) with early exit on match
- Value resolution: O(1) per candidate (hash map lookups)
- Target: **<10ms** for 2000 candidates

### 20.3 Memory

The engine's memory footprint is:
- Theme tokens: ~100KB for Tailwind's full default theme
- Utility definitions: ~200KB for all Tailwind utilities
- Variant definitions: ~10KB
- Candidate set: proportional to unique tokens in scanned content
- Scanner buffer: negligible (one partial token at a time)

Total baseline: **<1MB** for a fully-loaded engine.


## 21. Testing Strategy

### 21.1 Unit Tests

Each component is tested in isolation:

- **Tokenizer**: Known CSS strings → expected token sequences
- **Parser**: `@theme`, `@utility`, `@variant` blocks → expected data structures
- **Class parser**: Class strings → expected `ParsedClass` fields
- **Scanner**: Byte streams → expected candidate sets (including chunk-split cases)
- **Theme resolver**: Namespace + key → expected CSS values
- **Generator**: Candidate + definitions → expected CSS output

### 21.2 Integration Tests

End-to-end tests: HTML/template bytes in → CSS string out. These use representative subsets of Tailwind's actual CSS source as the engine's input.

### 21.3 Compatibility Tests

Parse Tailwind's actual full CSS source (downloaded from npm or CDN) and run it through the engine with known class sets. Compare output against Tailwind's own CLI output for the same classes.

This is the definitive correctness test. If the engine produces different CSS than Tailwind's CLI for the same input, the engine has a bug.

### 21.4 Fuzz Testing

Use Go's built-in fuzzing (`go test -fuzz`) on:
- The CSS tokenizer (arbitrary byte input should never panic)
- The class parser (arbitrary string input should never panic)
- The scanner (arbitrary byte input should never panic)
- The full pipeline (arbitrary bytes through Write → CSS should never panic)


## 22. Future Considerations

### 22.1 Incremental Generation

For watch-mode use cases, the engine should support incremental updates: when new candidates appear, generate only the new rules rather than regenerating everything. This requires tracking which candidates have already been generated.

### 22.2 Source Maps

For debugging, the engine could emit source maps linking generated CSS selectors back to the utility class names that produced them.

### 22.3 CSS Nesting Output

Modern CSS supports nesting. The engine could optionally emit nested CSS for smaller output:

```css
.bg-blue-500 {
  background-color: #3b82f6;
  &:hover { background-color: #2563eb; }
}
```

### 22.4 Tailwind v4 CSS as Embedded Resource

The engine could embed a recent version of Tailwind's CSS source using `//go:embed`, providing a batteries-included default that can be overridden by loading a newer CSS file.

### 22.5 Plugin System

Allow Go functions to register custom utility generators that go beyond what `@utility` can express in CSS. This would enable plugins equivalent to Tailwind's JS plugin system.
