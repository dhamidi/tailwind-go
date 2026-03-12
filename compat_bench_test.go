package tailwind

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// normalizeCSS normalizes CSS output for comparison:
//   - Strips comments.
//   - Normalizes whitespace (single spaces, no trailing).
//   - Sorts declarations within each rule alphabetically by property name.
//   - Sorts rules by selector.
func normalizeCSS(input string) string {
	// Strip CSS comments (/* ... */)
	reComment := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	s := reComment.ReplaceAllString(input, "")

	// Parse into rules and normalize
	rules := parseCSSRules(s)
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].selector < rules[j].selector
	})

	var out strings.Builder
	for _, rule := range rules {
		sort.Strings(rule.declarations)
		out.WriteString(rule.selector)
		out.WriteString(" {\n")
		for _, decl := range rule.declarations {
			out.WriteString("  ")
			out.WriteString(decl)
			out.WriteString(";\n")
		}
		out.WriteString("}\n")
	}
	return out.String()
}

type cssRule struct {
	selector     string
	declarations []string
}

// parseCSSRules is a simple CSS rule parser for normalization purposes.
// It handles basic selectors and declarations, and nested @-rules (one level).
func parseCSSRules(css string) []cssRule {
	var rules []cssRule
	css = strings.TrimSpace(css)
	i := 0
	for i < len(css) {
		// Skip whitespace
		for i < len(css) && (css[i] == ' ' || css[i] == '\t' || css[i] == '\n' || css[i] == '\r') {
			i++
		}
		if i >= len(css) {
			break
		}

		// Find opening brace
		braceIdx := strings.IndexByte(css[i:], '{')
		if braceIdx < 0 {
			break
		}
		selector := strings.TrimSpace(css[i : i+braceIdx])
		i = i + braceIdx + 1

		// Find matching closing brace (handle nesting)
		depth := 1
		bodyStart := i
		for i < len(css) && depth > 0 {
			if css[i] == '{' {
				depth++
			} else if css[i] == '}' {
				depth--
			}
			if depth > 0 {
				i++
			}
		}
		body := css[bodyStart:i]
		if i < len(css) {
			i++ // skip closing '}'
		}

		// If selector starts with '@' and body contains '{', recurse
		if strings.HasPrefix(selector, "@") && strings.Contains(body, "{") {
			innerRules := parseCSSRules(body)
			for _, ir := range innerRules {
				rules = append(rules, cssRule{
					selector:     selector + " " + ir.selector,
					declarations: ir.declarations,
				})
			}
		} else {
			// Parse declarations
			decls := parseDeclarations(body)
			if len(decls) > 0 {
				rules = append(rules, cssRule{
					selector:     normalizeSelector(selector),
					declarations: decls,
				})
			}
		}
	}
	return rules
}

// normalizeSelector normalizes whitespace in a CSS selector.
func normalizeSelector(sel string) string {
	fields := strings.Fields(sel)
	return strings.Join(fields, " ")
}

// parseDeclarations parses "property: value;" strings from a CSS block body.
func parseDeclarations(body string) []string {
	var decls []string
	for _, part := range strings.Split(body, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		colonIdx := strings.IndexByte(part, ':')
		if colonIdx < 0 {
			continue
		}
		prop := strings.TrimSpace(part[:colonIdx])
		val := strings.TrimSpace(part[colonIdx+1:])
		if prop != "" && val != "" {
			decls = append(decls, prop+": "+val)
		}
	}
	return decls
}

// urlDecode decodes a URL-encoded filename back to the original class name.
func urlDecode(encoded string) string {
	s := encoded
	s = strings.ReplaceAll(s, "%21", "!")
	s = strings.ReplaceAll(s, "%23", "#")
	s = strings.ReplaceAll(s, "%25", "%")
	s = strings.ReplaceAll(s, "%2E", ".")
	s = strings.ReplaceAll(s, "%2F", "/")
	s = strings.ReplaceAll(s, "%3A", ":")
	s = strings.ReplaceAll(s, "%40", "@")
	s = strings.ReplaceAll(s, "%5B", "[")
	s = strings.ReplaceAll(s, "%5D", "]")
	s = strings.ReplaceAll(s, "%20", " ")
	return s
}

// urlEncode encodes a class name for filesystem-safe usage.
func urlEncode(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "/", "%2F")
	s = strings.ReplaceAll(s, "[", "%5B")
	s = strings.ReplaceAll(s, "]", "%5D")
	s = strings.ReplaceAll(s, "#", "%23")
	s = strings.ReplaceAll(s, " ", "%20")
	s = strings.ReplaceAll(s, "!", "%21")
	s = strings.ReplaceAll(s, ":", "%3A")
	s = strings.ReplaceAll(s, "@", "%40")
	s = strings.ReplaceAll(s, ".", "%2E")
	return s
}

// TestCompatReference reads each testdata/reference/*.css file, extracts
// the class name from the filename, generates tailwind-go output, and compares.
// This test does NOT require npm/node and can run with zero external dependencies.
func TestCompatReference(t *testing.T) {
	refDir := filepath.Join("testdata", "reference")
	entries, err := os.ReadDir(refDir)
	if err != nil {
		t.Skipf("Reference directory not found: %v (run testdata/generate_reference.sh first)", err)
		return
	}

	if len(entries) == 0 {
		t.Skip("No reference files found in testdata/reference/")
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".css") {
			continue
		}

		encoded := strings.TrimSuffix(entry.Name(), ".css")
		className := urlDecode(encoded)

		t.Run(className, func(t *testing.T) {
			// Read reference CSS
			refData, err := os.ReadFile(filepath.Join(refDir, entry.Name()))
			if err != nil {
				t.Fatalf("Failed to read reference file: %v", err)
			}

			refCSS := string(refData)

			// Generate CSS using tailwind-go
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="%s"></div>`, className)))
			goCSS := e.CSS()

			// Extract only utility-relevant rules from reference CSS
			// (skip @layer base, preflight, etc.)
			refRules := extractUtilityRules(refCSS, className)
			goRules := normalizeCSS(goCSS)

			if goCSS == "" {
				t.Errorf("tailwind-go produced no CSS for class %q", className)
				t.Logf("Reference CSS (utility rules):\n%s", refRules)
				return
			}

			// Compare normalized outputs
			if refRules == "" {
				// Reference file may not have utility rules (e.g., preflight-only output)
				t.Logf("No utility rules found in reference for class %q, skipping comparison", className)
				t.Logf("tailwind-go output:\n%s", goCSS)
				return
			}

			// Compare declaration-by-declaration
			refParsed := parseCSSRules(refRules)
			goParsed := parseCSSRules(goRules)

			refDeclMap := collectDeclarations(refParsed)
			goDeclMap := collectDeclarations(goParsed)

			var mismatches []string
			for prop, refVal := range refDeclMap {
				goVal, ok := goDeclMap[prop]
				if !ok {
					mismatches = append(mismatches, fmt.Sprintf("  MISSING property %q (expected: %s)", prop, refVal))
				} else if goVal != refVal {
					mismatches = append(mismatches, fmt.Sprintf("  DIFF property %q:\n    reference: %s\n    tailwind-go: %s", prop, refVal, goVal))
				}
			}
			for prop, goVal := range goDeclMap {
				if _, ok := refDeclMap[prop]; !ok {
					mismatches = append(mismatches, fmt.Sprintf("  EXTRA property %q: %s", prop, goVal))
				}
			}

			if len(mismatches) > 0 {
				sort.Strings(mismatches)
				t.Errorf("class %q: CSS mismatch:\n%s\n\nReference (normalized):\n%s\ntailwind-go (normalized):\n%s",
					className, strings.Join(mismatches, "\n"), refRules, goRules)
			}
		})
	}
}

// extractUtilityRules attempts to extract utility-relevant CSS rules from
// a full tailwindcss CLI output (which includes preflight/base layers).
// It only matches the exact class selector (not substrings or variant forms).
func extractUtilityRules(fullCSS, className string) string {
	escapedClass := cssEscapeClass(className)

	// Build exact selector patterns to match.
	// We want ".classname {" but not ".hover\:classname {" or ".!classname {"
	exactSelectors := []string{
		"." + escapedClass,
		"." + className,
	}

	rules := parseCSSRules(fullCSS)
	var utilityRules []cssRule
	for _, r := range rules {
		sel := r.selector
		// Strip "@layer utilities " prefix if present
		sel = strings.TrimPrefix(sel, "@layer utilities ")

		// Check if the selector is exactly our class (not a variant/modifier form)
		matched := false
		for _, exact := range exactSelectors {
			if sel == exact {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		utilityRules = append(utilityRules, r)
	}

	if len(utilityRules) == 0 {
		return ""
	}

	var out strings.Builder
	for _, r := range utilityRules {
		sort.Strings(r.declarations)
		out.WriteString(r.selector)
		out.WriteString(" {\n")
		for _, d := range r.declarations {
			out.WriteString("  ")
			out.WriteString(d)
			out.WriteString(";\n")
		}
		out.WriteString("}\n")
	}
	return out.String()
}

// cssEscapeClass produces a CSS-escaped version of a class name.
// Tailwind CSS escapes special characters with backslashes in selectors.
func cssEscapeClass(class string) string {
	var b strings.Builder
	for _, ch := range class {
		switch ch {
		case '/', ':', '[', ']', '#', '.', '!', '@', '(', ')', ',', '%':
			b.WriteByte('\\')
			b.WriteRune(ch)
		default:
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// collectDeclarations flattens all declarations from parsed CSS rules into
// a map of property -> value for comparison.
func collectDeclarations(rules []cssRule) map[string]string {
	m := make(map[string]string)
	for _, r := range rules {
		for _, d := range r.declarations {
			colonIdx := strings.IndexByte(d, ':')
			if colonIdx < 0 {
				continue
			}
			prop := strings.TrimSpace(d[:colonIdx])
			val := strings.TrimSpace(d[colonIdx+1:])
			m[prop] = val
		}
	}
	return m
}

// compatBenchClasses is the comprehensive list of utility classes to test.
var compatBenchClasses = []string{
	// Static utilities
	"flex", "hidden", "sr-only", "truncate", "italic",

	// Spacing
	"p-4", "m-2", "px-8", "-mt-4", "p-0", "p-px", "p-0.5",

	// Sizing
	"w-1/2", "w-full", "w-screen", "h-64", "size-4",

	// Colors
	"bg-blue-500", "text-red-500/75", "border-green-300",

	// Typography
	"text-lg", "text-xl", "font-bold", "leading-tight", "tracking-wide",

	// Borders
	"rounded-lg", "rounded-full", "border", "border-2",

	// Flexbox/Grid
	"flex-1", "grid-cols-3", "gap-4", "col-span-2",

	// Transforms
	"translate-x-4", "rotate-45", "scale-50", "-translate-y-1/2",

	// Transitions
	"transition", "duration-300", "ease-in-out",

	// Filters
	"blur-md", "brightness-75", "grayscale",

	// Variants
	"hover:bg-blue-500", "md:flex", "dark:text-white", "focus:ring-2",

	// Arbitrary values
	"w-[300px]", "bg-[#ff0000]", "grid-cols-[1fr_auto_2fr]",

	// Modifiers
	"bg-blue-500/50", "text-black/[.5]",

	// Negative
	"-m-4", "-translate-x-4", "-rotate-12",

	// Fractions
	"w-2/3", "w-1/4", "translate-x-1/2",
}

// TestCompatBenchSummary generates a summary of tailwind-go output for all
// test bench classes. This test runs without npm and is useful for quick
// verification that classes produce output.
func TestCompatBenchSummary(t *testing.T) {
	e := New()
	var buf strings.Builder
	buf.WriteString(`<div class="`)
	for _, class := range compatBenchClasses {
		buf.WriteString(class)
		buf.WriteByte(' ')
	}
	buf.WriteString(`">`)
	e.Write([]byte(buf.String()))
	css := e.CSS()

	var noOutput []string
	for _, class := range compatBenchClasses {
		eachEngine := New()
		eachEngine.Write([]byte(fmt.Sprintf(`<div class="%s"></div>`, class)))
		eachCSS := eachEngine.CSS()
		if eachCSS == "" {
			noOutput = append(noOutput, class)
		}
	}

	if len(noOutput) > 0 {
		t.Errorf("Classes that produced no CSS output:\n  %s", strings.Join(noOutput, "\n  "))
	}

	t.Logf("Total CSS length: %d bytes for %d classes", len(css), len(compatBenchClasses))
}

// TestNormalizeCSS verifies the normalizeCSS helper.
func TestNormalizeCSS(t *testing.T) {
	input := `
/* A comment */
.foo {
  z-index: 10;
  color: red;
  background: blue;
}
.bar {
  display: flex;
}
`
	result := normalizeCSS(input)

	// Should strip comments
	if strings.Contains(result, "comment") {
		t.Error("normalizeCSS did not strip comments")
	}

	// Should sort declarations within rules
	colorIdx := strings.Index(result, "background: blue")
	zIdx := strings.Index(result, "z-index: 10")
	if colorIdx < 0 || zIdx < 0 {
		t.Fatal("missing declarations in normalized output")
	}
	if colorIdx > zIdx {
		t.Error("declarations not sorted alphabetically within rule")
	}

	// Should sort rules by selector
	barIdx := strings.Index(result, ".bar")
	fooIdx := strings.Index(result, ".foo")
	if barIdx < 0 || fooIdx < 0 {
		t.Fatal("missing selectors in normalized output")
	}
	if barIdx > fooIdx {
		t.Error("rules not sorted by selector")
	}
}
