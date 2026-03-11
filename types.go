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
	if namespace == "spacing" && isNumeric(key) {
		if base, ok := tc.Tokens["--spacing"]; ok {
			return "calc(" + key + " * " + base + ")", true
		}
	}

	return "", false
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
type VariantDef struct {
	Name     string // Variant name, e.g., "hover", "md", "dark".
	Selector string // Selector variant, e.g., "&:hover", "&:first-child".
	Media    string // Media query variant, e.g., "(width >= 48rem)".
	AtRule   string // At-rule variant, e.g., "supports", "container".
	Order    int    // Definition order for stable CSS output sorting.
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
