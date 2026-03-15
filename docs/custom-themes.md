# Custom Themes

This guide explains how to customize the default Tailwind theme in tailwind-go
using `@theme` blocks in your CSS.

## Overview

The `@theme` directive lets you define or override design tokens — colors,
spacing, fonts, and more. These tokens are stored as CSS custom properties and
referenced by Tailwind's utility classes.

```go
engine := tailwind.New()
engine.LoadCSS([]byte(`
@theme {
  --color-brand-500: oklch(58% 0.18 250);
}
`))
```

After loading, utilities like `bg-brand-500` and `text-brand-500` resolve
against your custom token.

## Tutorial: Adding a brand color palette

### Step 1: Define theme tokens

Create a `@theme` block with your brand colors:

```css
@theme {
  --color-brand-50:  oklch(97% 0.01 250);
  --color-brand-100: oklch(93% 0.03 250);
  --color-brand-200: oklch(87% 0.06 250);
  --color-brand-300: oklch(78% 0.10 250);
  --color-brand-400: oklch(68% 0.14 250);
  --color-brand-500: oklch(58% 0.18 250);
  --color-brand-600: oklch(50% 0.16 250);
  --color-brand-700: oklch(42% 0.13 250);
  --color-brand-800: oklch(34% 0.10 250);
  --color-brand-900: oklch(27% 0.07 250);
  --color-brand-950: oklch(20% 0.04 250);
}
```

### Step 2: Use brand colors in markup

```html
<header class="bg-brand-500 text-white p-4">
  <h1 class="text-2xl">Welcome</h1>
</header>
<main class="bg-brand-50 p-8">
  <p class="text-brand-900">Hello, custom theme!</p>
</main>
```

### Step 3: Custom font families

The built-in font utilities are `font-sans`, `font-serif`, and `font-mono`.
These reference the theme tokens `--font-sans`, `--font-serif`, and
`--font-mono` respectively. You can override their values:

```css
@theme {
  --font-sans: "Inter", system-ui, sans-serif;
}
```

To use a completely custom font family name (e.g., "display"), you need to
both define the theme token and register a static utility with `@utility`:

```css
@theme {
  --font-display: "Cal Sans", "Inter", sans-serif;
}

@utility font-display {
  font-family: var(--font-display);
}
```

After this, the `font-display` class will output `font-family: var(--font-display)`.

> **Note:** Simply adding `--font-display` to `@theme` is not enough — the
> `font-*` dynamic utility resolves against the `--font-weight` namespace,
> not `--font-family`. You must define a `@utility` rule for custom font
> family names beyond sans, serif, and mono.

### Complete example

```go
package main

import (
	"fmt"
	"log"

	tailwind "github.com/dhamidi/tailwind-go"
)

func main() {
	engine := tailwind.New()

	err := engine.LoadCSS([]byte(`
@theme {
  --color-brand-50:  oklch(97% 0.01 250);
  --color-brand-100: oklch(93% 0.03 250);
  --color-brand-200: oklch(87% 0.06 250);
  --color-brand-300: oklch(78% 0.10 250);
  --color-brand-400: oklch(68% 0.14 250);
  --color-brand-500: oklch(58% 0.18 250);
  --color-brand-600: oklch(50% 0.16 250);
  --color-brand-700: oklch(42% 0.13 250);
  --color-brand-800: oklch(34% 0.10 250);
  --color-brand-900: oklch(27% 0.07 250);
  --color-brand-950: oklch(20% 0.04 250);
  --font-sans: "Inter", system-ui, sans-serif;
}

@utility font-display {
  font-family: var(--font-display);
}
`))
	if err != nil {
		log.Fatal("LoadCSS failed: ", err)
	}

	engine.Write([]byte(`
<header class="bg-brand-500 text-white font-sans p-4">
  <h1 class="text-2xl">Welcome</h1>
</header>
<main class="bg-brand-50 p-8">
  <p class="text-brand-900">Hello, custom theme!</p>
</main>
`))

	fmt.Println(engine.CSS())
}
```

## How-to Guides

### Override built-in theme values

You can redefine any built-in token. For example, to change the default
spacing scale:

```css
@theme {
  --spacing: 0.3rem;
}
```

Now `p-4` computes as `calc(0.3rem * 4)` = `1.2rem` instead of the default.

### Define custom colors with semantic names

```css
@theme {
  --color-success: oklch(72% 0.20 142);
  --color-warning: oklch(80% 0.16  84);
  --color-danger:  oklch(58% 0.22  27);
}
```

This enables `bg-success`, `text-warning`, `border-danger`, etc.

### Override built-in breakpoint values

The responsive breakpoint variants (`sm`, `md`, `lg`, `xl`, `2xl`) are
built-in. You can override the value of an existing breakpoint token:

```css
@theme {
  --breakpoint-sm: 36rem;
  --breakpoint-md: 44rem;
}
```

> **Current limitation:** Custom breakpoint names (e.g., `--breakpoint-xs`)
> cannot be added via `@theme` alone. The responsive variants (`sm`, `md`,
> `lg`, `xl`, `2xl`) are registered at engine initialization. Adding a new
> `--breakpoint-xs` token stores the value but does **not** register a
> corresponding `xs:` variant. To add entirely new breakpoints, a `@variant`
> definition would be needed.

### Extend the color palette with shades

Define a full shade scale for a custom color:

```css
@theme {
  --color-ocean-50:  oklch(97% 0.01 220);
  --color-ocean-100: oklch(93% 0.03 220);
  --color-ocean-200: oklch(87% 0.06 220);
  --color-ocean-300: oklch(78% 0.10 220);
  --color-ocean-400: oklch(68% 0.14 220);
  --color-ocean-500: oklch(58% 0.18 220);
  --color-ocean-600: oklch(50% 0.16 220);
  --color-ocean-700: oklch(42% 0.13 220);
  --color-ocean-800: oklch(34% 0.10 220);
  --color-ocean-900: oklch(27% 0.07 220);
  --color-ocean-950: oklch(20% 0.04 220);
}
```

All color utilities (`bg-ocean-500`, `text-ocean-200`, `border-ocean-700`,
etc.) are available immediately.

## Explanation

### How theme tokens map to utilities

When you write `--color-brand-500: oklch(58% 0.18 250)` in a `@theme` block,
the engine stores a token named `--color-brand-500`. Color utilities like
`bg-*`, `text-*`, and `border-*` resolve their value by looking up
`--color-<name>` in the theme. So `bg-brand-500` generates:

```css
.bg-brand-500 {
  background-color: var(--color-brand-500);
}
```

The actual color value is emitted in a `:root` block as part of the theme
layer.

### Opacity modifiers and theme colors

You can apply an opacity modifier to any color utility:

```html
<div class="bg-brand-500/50">
```

This produces CSS using `color-mix(in oklab, ...)` to apply opacity:

```css
.bg-brand-500\/50 {
  background-color: color-mix(in oklab, var(--color-brand-500) 50%, transparent);
}
```

The engine keeps the CSS variable reference and uses `color-mix(in oklab, ...)`
to apply the specified opacity. The `/50` modifier maps to `50%` opacity.
A `/100` modifier (full opacity) is treated as identity and produces no wrapping.

### Token namespaces

Theme tokens are organized by prefix:

| Prefix            | Used by                         |
|-------------------|---------------------------------|
| `--color-*`       | `bg-*`, `text-*`, `border-*`    |
| `--spacing`       | `p-*`, `m-*`, `gap-*`, `w-*`    |
| `--font-sans`     | `font-sans`                     |
| `--font-serif`    | `font-serif`                    |
| `--font-mono`     | `font-mono`                     |
| `--font-weight-*` | `font-*` (dynamic)              |
| `--breakpoint-*`  | `sm:`, `md:`, `lg:`, `xl:`, `2xl:` |
| `--text-*`        | `text-*` (font size)            |

### Merging behavior

Multiple `@theme` blocks and multiple `LoadCSS` calls merge tokens
additively. Later values for the same key overwrite earlier ones:

```go
engine.LoadCSS([]byte(`@theme { --color-brand: red; }`))
engine.LoadCSS([]byte(`@theme { --color-brand: blue; }`))
// bg-brand now uses blue
```

## Reference

### Supported token prefixes

See the [Token namespaces](#token-namespaces) table above for the full list
of recognized prefixes and the utilities they feed.

### API usage

```go
// Create engine with defaults
engine := tailwind.New()

// Load custom theme
engine.LoadCSS([]byte(`@theme { --color-primary: #3b82f6; }`))

// Scan markup
engine.Write([]byte(`<div class="bg-primary text-white">`))

// Get generated CSS
css := engine.CSS()
```
