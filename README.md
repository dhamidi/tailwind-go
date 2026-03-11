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
  color: oklch(62.3% 0.214 259.815);
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

## Explanation

The engine works by ingesting Tailwind's own CSS v4 source as its specification. Rather than hardcoding a list of utilities, it parses `@theme`, `@utility`, `@variant`, and `@keyframes` directives from real Tailwind CSS. This means the engine's behavior is defined entirely by the CSS it loads — the same way Tailwind v4 itself works.

The scanner is format-agnostic: it extracts candidate class-name tokens from raw bytes without parsing HTML, JSX, or any other template syntax. You can feed it any text — markup, source code, Markdown — and it will find anything that looks like a Tailwind class. False positives are harmless; unrecognized candidates are silently discarded during CSS generation.

The `Engine` is safe for concurrent use. It uses `sync.RWMutex` internally, so concurrent `Write` and `CSS` calls are safe. Call `Reset()` between requests to clear accumulated candidates while keeping all loaded definitions intact.

## Supported Tailwind Features

- Static utilities: `flex`, `block`, `hidden`
- Dynamic utilities with values: `p-4`, `bg-blue-500`, `w-1/2`
- Variants: `hover:bg-blue-500`, `md:flex`, `dark:text-white`
- Stacked variants: `dark:md:hover:bg-blue-500`
- Important modifier: `!p-4`
- Negative values: `-translate-x-4`
- Arbitrary values: `w-[300px]`, `text-[#ff0000]`
- Arbitrary properties: `[mask-type:alpha]`
- Opacity modifiers: `bg-blue-500/75`, `text-white/[.5]`
- Fractions: `w-1/2` → `50%`
- Type hints: `text-[length:1.5em]`
- Custom properties: `w-[--sidebar-width]`
- Arbitrary variants: `[@media(min-width:900px)]:bg-red-500`
- `@theme`, `@utility`, `@variant`, `@keyframes` directives

## Current Limitations

The following Tailwind CSS v4 features are **not yet supported**:

- **Compound variants**: `group-*`, `peer-*`, `not-*`, `has-*`, `in-*` (e.g., `group-hover:text-white`, `peer-focus:ring-2`)
- **Named groups/peers**: slash notation like `group/sidebar`, `group-hover/sidebar:flex`
- **`@starting-style` variant**
- **`:merge()` function** in variant selectors
- **Dark mode class strategy**: only `@media (prefers-color-scheme: dark)` is supported; class-based dark mode (`.dark` selector) is not
- **Container queries**: `@container`-based variants

The following are **out of scope by design** (see [spec.md §1.2](spec.md)):

- **Preflight / reset styles**: the engine generates utility CSS only
- **`@import` / `@source` / `@config` directives**: build-tool concerns; the engine uses `io.Writer` for input instead
- **CSS minification**: the engine emits readable CSS; pipe through a minifier if needed
- **File watching / CLI**: the engine is a library, not a build tool

## API Reference

| Symbol | Description |
|--------|-------------|
| `New()` | Create engine pre-loaded with Tailwind v4 CSS |
| `NewPassthrough(w)` | Create engine that also forwards bytes to `w` |
| `Engine.Write(p)` | Scan bytes for candidate classes (`io.Writer`) |
| `Engine.CSS()` | Generate CSS for accumulated candidates |
| `Engine.LoadCSS(css)` | Parse and load additional Tailwind v4 CSS |
| `Engine.Candidates()` | Return extracted candidate class names |
| `Engine.Flush()` | Finalize any buffered partial token |
| `Engine.Reset()` | Clear candidates, keep definitions |
