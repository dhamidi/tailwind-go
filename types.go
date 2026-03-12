package tailwind

import (
	"sort"
	"strings"
)

// ThemeConfig holds the design token registry parsed from @theme blocks.
type ThemeConfig struct {
	// Tokens maps custom property names to their values.
	// e.g., "--color-blue-500" → "#3b82f6"
	//       "--spacing" → "0.25rem"
	//       "--breakpoint-md" → "48rem"
	Tokens map[string]string
}

// Resolve looks up a theme token by namespace and key.
// For example, Resolve("color", "blue-500") checks "--color-blue-500".
// For the spacing scale, Resolve("spacing", "4") computes 4 * base.
func (tc *ThemeConfig) Resolve(namespace, key string) (string, bool) {
	// Direct token lookup: --{namespace}-{key}
	prop := "--" + namespace + "-" + key
	if v, ok := tc.Tokens[prop]; ok {
		return v, true
	}

	// For spacing, compute from the base multiplier (numeric keys only).
	// Use var(--spacing) to reference the CSS custom property directly.
	// Only accept multiples of 0.25 to match Tailwind's spacing scale.
	if namespace == "spacing" && isNumeric(key) {
		if !isValidSpacingMultiplier(key) {
			return "", false
		}
		if _, ok := tc.Tokens["--spacing"]; ok {
			return "calc(var(--spacing) * " + key + ")", true
		}
	}

	// For grid-template-columns and grid-template-rows, positive integers
	// produce repeat(N, minmax(0, 1fr)). This matches upstream Tailwind's
	// handleBareValue for grid-cols-* / grid-rows-*.
	if (namespace == "grid-template-columns" || namespace == "grid-template-rows") && isPositiveInteger(key) {
		return "repeat(" + key + ", minmax(0, 1fr))", true
	}

	return "", false
}

// isPositiveInteger returns true if s is an integer > 0 with no leading zeros.
func isPositiveInteger(s string) bool {
	if s == "" || s == "0" {
		return false
	}
	for _, b := range []byte(s) {
		if !isDigit(b) {
			return false
		}
	}
	return true
}

// isValidSpacingMultiplier returns true if s is a numeric value that is
// a multiple of 0.25. This validates Tailwind's spacing scale where only
// values like 0, 0.25, 0.5, 0.75, 1, 1.25, ... are valid.
func isValidSpacingMultiplier(s string) bool {
	// Parse the numeric value by splitting on decimal point.
	parts := strings.SplitN(s, ".", 2)
	if len(parts) == 1 {
		// Integer — always a valid multiple of 0.25.
		return true
	}
	// Has a fractional part. Valid fractional parts for multiples of 0.25
	// are: .25, .5, .75, .0 (and equivalents like .50, .250, etc.)
	frac := parts[1]
	// Normalize by removing trailing zeros.
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		// e.g., "1.0" — integer, valid.
		return true
	}
	// Valid fractional parts after trimming trailing zeros: "25", "5", "75"
	return frac == "25" || frac == "5" || frac == "75"
}

// NamespaceValues returns all tokens under a given namespace prefix.
// For namespace "color", returns {"blue-500": "#3b82f6", ...}.
func (tc *ThemeConfig) NamespaceValues(namespace string) map[string]string {
	prefix := "--" + namespace + "-"
	out := make(map[string]string)
	for k, v := range tc.Tokens {
		if strings.HasPrefix(k, prefix) {
			out[strings.TrimPrefix(k, prefix)] = v
		}
	}
	return out
}

// UtilityDef defines a utility pattern parsed from @utility blocks.
type UtilityDef struct {
	// Pattern is the utility name pattern, e.g., "w", "bg", "translate-x".
	// A pattern like "w" matches classes like w-4, w-full, w-[300px].
	Pattern string

	// Static is true when the utility takes no value (e.g., "flex", "block").
	Static bool

	// Declarations are the CSS declarations this utility produces.
	// Each may contain ValueRef placeholders for dynamic resolution.
	Declarations []Declaration

	// Order tracks definition order for stable CSS output sorting.
	Order int

	// Selector is an optional child selector suffix appended to the
	// generated CSS selector. For example, "> :not(:last-child)" causes
	// the rule to target children rather than the element itself.
	Selector string
}

// Declaration is a single CSS property: value pair.
// The Value field may contain "--value(...)" placeholders that are resolved
// at generation time against the [ThemeConfig] or by value type.
type Declaration struct {
	Property string // CSS property name, e.g., "display", "padding".
	Value    string // CSS value, possibly containing "--value(...)" placeholders.
}

// VariantDef defines a variant parsed from @variant directives.
// There are three variant types: selector-based (Selector is set),
// media query (Media is set), or at-rule (AtRule is set).
// Compound variants (e.g., group-*, peer-*, not-*) use Template
// with a {value} placeholder.
type VariantDef struct {
	Name     string // Variant name, e.g., "hover", "md", "dark", "group" (for group-*).
	Selector string // Selector variant, e.g., "&:hover", "&:first-child".
	Media    string // Media query variant, e.g., "(width >= 48rem)".
	AtRule   string // At-rule variant, e.g., "supports", "container".
	Order    int    // Definition order for stable CSS output sorting.
	Compound bool   // True for wildcard variants like group-*, peer-*, not-*.
	Template string // Selector template with {value} and & placeholders for compound variants.
}

// KeyframesRule represents a @keyframes block, used by animation utilities.
// When a utility references an animation name that matches a KeyframesRule,
// the keyframes block is included in the generated CSS output.
type KeyframesRule struct {
	Name string // Animation name, e.g., "spin", "ping", "bounce".
	Body string // Raw CSS body including @keyframes name { ... }.
}

// ApplyRule represents an @apply directive found inside a CSS rule.
type ApplyRule struct {
	Selector string   // the parent CSS selector (e.g., ".btn")
	Classes  []string // the class names to apply
	Order    int
}

// Stylesheet is the parsed representation of Tailwind CSS input.
// It is produced by parsing CSS containing @theme, @utility, @variant,
// @keyframes, and @apply directives.
type Stylesheet struct {
	Theme      ThemeConfig      // Design tokens from @theme blocks.
	Utilities  []*UtilityDef    // Utility definitions from @utility blocks.
	Variants   []*VariantDef    // Variant definitions from @variant directives.
	Keyframes  []*KeyframesRule // Animation keyframes from @keyframes blocks.
	ApplyRules []*ApplyRule     // Resolved @apply directives.
}

// utilityIndex provides fast lookup of utility definitions by class prefix.
type utilityIndex struct {
	// static maps exact utility names to their definitions.
	// e.g., "flex" → &UtilityDef{...}
	static map[string]*UtilityDef

	// dynamic holds pattern-based utilities, sorted longest-first.
	// e.g., "translate-x" before "translate" before "t".
	dynamic []*UtilityDef
}

func newUtilityIndex() *utilityIndex {
	return &utilityIndex{
		static: make(map[string]*UtilityDef),
	}
}

func (idx *utilityIndex) add(u *UtilityDef) {
	if u.Static {
		idx.static[u.Pattern] = u
	} else {
		// Replace existing dynamic utility with the same pattern.
		replaced := false
		for i, existing := range idx.dynamic {
			if existing.Pattern == u.Pattern {
				idx.dynamic[i] = u
				replaced = true
				break
			}
		}
		if !replaced {
			idx.dynamic = append(idx.dynamic, u)
		}
		// Keep sorted longest-pattern-first for greedy matching.
		sort.Slice(idx.dynamic, func(i, j int) bool {
			return len(idx.dynamic[i].Pattern) > len(idx.dynamic[j].Pattern)
		})
	}
}

// lookup finds the utility definition for a given utility name and
// value string. Returns the definition and the extracted value portion.
func (idx *utilityIndex) lookup(utility string) (*UtilityDef, string) {
	// Try static first (exact match, no value).
	if u, ok := idx.static[utility]; ok {
		return u, ""
	}

	// Try dynamic patterns (longest prefix match).
	for _, u := range idx.dynamic {
		if utility == u.Pattern {
			// Exact match on pattern name but it's dynamic — no value.
			// This shouldn't normally happen; skip.
			continue
		}
		prefix := u.Pattern + "-"
		if strings.HasPrefix(utility, prefix) {
			value := strings.TrimPrefix(utility, prefix)
			return u, value
		}
	}

	return nil, ""
}
