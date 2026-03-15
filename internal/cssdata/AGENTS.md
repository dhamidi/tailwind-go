# Tailwind CSS Data — Update Guide

## Overview

The tailwind-go project embeds three CSS files from Tailwind CSS in `internal/cssdata/`:

| File | Source | Update Method |
|------|--------|---------------|
| `theme.css` | Downloaded from Tailwind npm package | `go generate ./...` runs `download.sh` |
| `preflight.css` | Downloaded from Tailwind npm package | `go generate ./...` runs `download.sh` |
| `utilities.css` | **Project-owned**, manually translated from Tailwind TypeScript source | Manual update required |

These files are embedded into the Go binary via `internal/cssdata/cssdata.go` using `//go:embed` directives. The `go:generate` directive that triggers the download lives in `engine.go` line 1:

```go
//go:generate bash internal/cssdata/download.sh
```

### Data Flow

```
Tailwind npm package (tailwindcss@VERSION)
  ├── theme.css ──────────► internal/cssdata/theme.css (auto-downloaded)
  ├── preflight.css ──────► internal/cssdata/preflight.css (auto-downloaded)
  └── src/ (TypeScript) ──► internal/cssdata/utilities.css (MANUAL translation)
```

As of this writing, the current embedded version is **4.2.1**.

## Prerequisites

- Node.js and npm (for downloading the Tailwind package via `npm pack`)
- Go toolchain (for running `go generate` and `go test`)

## Step-by-Step Update Procedure

### Step 1: Determine the Target Version

Check https://github.com/tailwindlabs/tailwindcss/releases for the latest v4.x release. Note the version number (e.g., `4.3.0`). Review the release notes for any breaking changes, new utilities, removed utilities, or changed default values.

### Step 2: Update the Download Script Version

Edit `internal/cssdata/download.sh` line 6 and change the `VERSION` variable to the target version:

```bash
# Before
VERSION=4.2.1
# After
VERSION=X.Y.Z
```

### Step 3: Run go generate to Download theme.css and preflight.css

From the project root, run:

```bash
go generate ./...
```

This executes `download.sh`, which uses `npm pack` to download the Tailwind npm package and copies `theme.css` and `preflight.css` into `internal/cssdata/`. The script does **not** touch `utilities.css`.

### Step 4: Review theme.css Changes

Diff the newly downloaded `theme.css` against the previous version:

```bash
git diff internal/cssdata/theme.css
```

Look for:
- **New theme tokens** (e.g., new `--color-*`, `--spacing-*`, `--font-*` custom properties)
- **Removed tokens** that utilities.css or Go code may reference
- **Changed values** (e.g., a breakpoint value changing from `40rem` to `42rem`)
- **New namespaces** (e.g., a new `--blur-*` or `--ease-*` family of tokens)
- **Changed breakpoint values** in `--breakpoint-*` tokens — these MUST match the `@variant` media queries in `utilities.css` (see Step 5)

Also diff `preflight.css`:

```bash
git diff internal/cssdata/preflight.css
```

Preflight changes are usually straightforward and rarely require manual follow-up.

### Step 5: Update utilities.css (MANUAL — Most Critical Step)

This is the highest-risk part of the update. The file `internal/cssdata/utilities.css` (~1589 lines, ~931 `@utility` + ~64 `@variant` definitions) is manually maintained and must be synchronized with Tailwind's upstream utility definitions.

#### 5a: Find the Upstream Utility Definitions

The canonical Tailwind utility definitions are in the Tailwind CSS source repository:
- Repository: https://github.com/tailwindlabs/tailwindcss
- Utility plugins: `packages/tailwindcss/src/utilities.ts`
- Variant definitions: `packages/tailwindcss/src/variants.ts`

Clone or browse the repository at the target version tag (e.g., `v4.3.0`).

#### 5b: Identify Added, Removed, and Changed Utilities

Compare the upstream `utilities.ts` at the old version tag vs. the new version tag:

```bash
# In a clone of tailwindlabs/tailwindcss:
git diff v4.2.1..vX.Y.Z -- packages/tailwindcss/src/utilities.ts
git diff v4.2.1..vX.Y.Z -- packages/tailwindcss/src/variants.ts
```

Identify:
- New utility functions that were added
- Existing utilities whose CSS output changed
- Utilities that were removed or renamed
- New or changed variants

#### 5c: The @utility Syntax Format

Each utility in `utilities.css` is declared with the `@utility` at-rule:

```css
@utility bg-* {
  background-color: var(--value, --color-*);
}
```

Key conventions:
- The utility name uses `*` as a wildcard for the value segment (e.g., `bg-*`, `p-*`, `text-*`).
- Static utilities (no value) omit the `*` (e.g., `@utility flex { display: flex; }`).

#### 5d: The --value() Priority Chain Convention

Utilities that accept theme values use the `var(--value, ...)` pattern with a priority chain declaring which theme namespaces to try:

```css
@utility mx-* {
  margin-left: var(--value, --spacing-*, --spacing);
  margin-right: var(--value, --spacing-*, --spacing);
}
```

The resolution order is:
1. `--value` — an explicit/arbitrary value provided by the user (e.g., `mx-[2rem]`)
2. `--spacing-*` — a direct theme token lookup (e.g., `--spacing-4`)
3. `--spacing` — the base multiplier for computed values (e.g., `4 * 0.25rem`)

When translating a new utility from TypeScript, identify which theme namespace(s) it reads from and construct the appropriate `var(--value, ...)` chain.

#### 5e: Update @variant Definitions if Needed

Variant definitions appear at the end of `utilities.css` (lines 1526–1589). Breakpoint variants (lines 1563–1567) have hard-coded media query values that **MUST match** the `--breakpoint-*` tokens in `theme.css`:

```css
@variant sm (@media (width >= 40rem));
@variant md (@media (width >= 48rem));
@variant lg (@media (width >= 64rem));
@variant xl (@media (width >= 80rem));
@variant 2xl (@media (width >= 96rem));
```

Container query variants (lines 1568–1580) similarly have hard-coded width thresholds. If Tailwind changes any of these values, both `theme.css` (auto-downloaded) and the `@variant` lines in `utilities.css` (manual) must agree.

Also check for new pseudo-class, pseudo-element, or media query variants added upstream.

#### 5f: Verify Breakpoint Consistency

After updating, confirm breakpoint values match between the two files:

```bash
# Extract breakpoints from theme.css
grep 'breakpoint' internal/cssdata/theme.css

# Extract breakpoint variants from utilities.css
grep '@variant.*@media (width' internal/cssdata/utilities.css
```

The rem values must be identical.

### Step 6: Update Go Hard-Coded Values (If Needed)

Several Go source files contain hard-coded Tailwind knowledge that may need updating:

1. **`class.go:181–203` — `isValueKeyword()`**: A switch statement listing ~70 Tailwind value keywords (e.g., "full", "auto", "px", "semibold"). This is used to distinguish value segments from utility name segments during class parsing. If the new Tailwind version adds utilities with new keyword values, add them here.

2. **`generate.go:638–682` — `keywordToCSS()`**: Maps ~20 Tailwind value keywords to their CSS equivalents (e.g., "full" → "100%", "px" → "1px", "current" → "currentColor"). If Tailwind changes what a keyword maps to, or adds new keyword-to-CSS mappings, update this function.

3. **`generate.go` — `applyModifier()`**: Uses `color-mix(in oklab, <color> <opacity>, transparent)` for opacity modifiers, matching Tailwind CSS v4. A `/100` modifier (100% opacity) is treated as identity. Shadow colors are additionally wrapped with `--tw-shadow-alpha` via `wrapShadowAlpha()`.

4. **`types.go:28` — `ThemeConfig.Resolve()`**: Hard-codes "spacing" as a special namespace that uses multiplier-based computation (`calc(key * base)`). If Tailwind adds other computed namespaces (beyond spacing), this function must be extended.

5. **`class.go:290–295` — `processArbitraryValue()`**: Contains 14 CSS type hint keywords (length, percentage, color, number, integer, url, image, shadow, position, angle, time, frequency, ratio, any). These are CSS spec concepts rather than Tailwind-specific, so they are unlikely to change.

### Step 7: Run Tests

Run the full test suite from the project root:

```bash
go test ./...
```

Check for:
- Failing tests that indicate regressions from the update
- Tests that need updated expected output due to intentional Tailwind changes

If there is a compatibility test suite or test fixtures comparing output against Tailwind CLI, run those as well.

### Step 8: Manual Verification

Generate CSS for a representative set of classes and compare against Tailwind CLI output for the same version:

```bash
# Using the Tailwind CLI (install with: npm install -g tailwindcss@X.Y.Z)
# Create a test HTML file with representative classes, then:
npx @tailwindcss/cli -i input.css -o expected.css
```

Compare the Go engine's output against `expected.css` for the same input classes. Pay special attention to:
- New utilities added in the target version
- Utilities that changed behavior
- Color output format (oklch vs other)
- Spacing calculations
- Breakpoint and container query thresholds

## File Reference

| File | Role |
|------|------|
| `internal/cssdata/download.sh` | Downloads `theme.css` and `preflight.css` from npm; contains `VERSION` variable (line 6) |
| `internal/cssdata/theme.css` | Auto-downloaded theme tokens (colors, spacing, fonts, breakpoints, etc.) |
| `internal/cssdata/preflight.css` | Auto-downloaded CSS reset/normalize styles |
| `internal/cssdata/utilities.css` | **Manually maintained** utility and variant definitions (~931 `@utility`, ~64 `@variant`) |
| `internal/cssdata/cssdata.go` | Go embed directives that load the three CSS files into the binary |
| `engine.go` | Contains the `//go:generate` directive (line 1) |
| `class.go` | Class parsing; `isValueKeyword()` at line 181, `processArbitraryValue()` at line 286 |
| `generate.go` | CSS generation; `keywordToCSS()` at line 638, `applyModifier()` at line 723 |
| `types.go` | Theme config; `Resolve()` at line 20 with hard-coded "spacing" namespace |

## Common Pitfalls

1. **utilities.css drift**: This is the highest-risk file. It is manually maintained and must be manually synchronized when Tailwind adds, removes, or modifies utilities. Missing a new utility means the Go engine will silently fail to generate CSS for it.

2. **Keyword list drift**: The `isValueKeyword()` and `keywordToCSS()` functions maintain hard-coded keyword lists. New Tailwind utilities that introduce new keyword values will silently fail to parse correctly if these lists are not updated.

3. **Breakpoint duplication**: Breakpoint values exist in both `theme.css` (as `--breakpoint-*` tokens) and `utilities.css` (as `@variant` media query literals). The theme tokens are auto-downloaded; the variant definitions are manual. A mismatch causes subtle bugs where responsive utilities fire at different widths than expected.

4. **Container query duplication**: Similar to breakpoints, container query width thresholds in `@variant` definitions (lines 1568–1580) are hard-coded and must match any upstream changes.

5. **Color space changes**: The `oklch()` color function is hard-coded in `generate.go:723–736` for opacity modifier application. A color space change in Tailwind would require Go code changes, not just CSS data updates.

6. **Forgetting to run `go generate`**: After changing the version in `download.sh`, you must run `go generate ./...` to actually download the new files. The version bump alone does nothing.

7. **New computed namespaces**: If Tailwind introduces a new namespace that uses multiplier-based computation (like spacing does), `types.go:28` `Resolve()` must be extended to handle it. Without this, theme lookups for the new namespace will silently fail.
