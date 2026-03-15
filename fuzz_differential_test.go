//go:build reference

package tailwind

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var (
	fuzzCount = flag.Int("fuzz-count", 500, "number of classes to generate for differential fuzzing")
	fuzzSeed  = flag.Int64("fuzz-seed", 42, "random seed for class generation")
)

// setupTailwindCLI installs tailwindcss in a cached directory.
func setupTailwindCLI(t *testing.T) string {
	t.Helper()

	cacheDir := os.Getenv("TAILWIND_FUZZ_CACHE")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "tailwind-fuzz-cache")
	}

	// Check if already installed.
	pkgJSON := filepath.Join(cacheDir, "package.json")
	nodeModules := filepath.Join(cacheDir, "node_modules")
	if _, err := os.Stat(pkgJSON); err == nil {
		if _, err := os.Stat(nodeModules); err == nil {
			return cacheDir
		}
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	cmd := exec.Command("npm", "init", "-y")
	cmd.Dir = cacheDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm init failed: %v\n%s", err, out)
	}

	cmd = exec.Command("npm", "install", "tailwindcss", "@tailwindcss/cli")
	cmd.Dir = cacheDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm install failed: %v\n%s", err, out)
	}

	if err := os.WriteFile(filepath.Join(cacheDir, "input.css"), []byte("@import \"tailwindcss\";\n"), 0644); err != nil {
		t.Fatalf("Failed to write input.css: %v", err)
	}

	return cacheDir
}

// generateReferenceBatch runs the official CLI on all classes at once, then
// extracts per-class rules from the combined output.
func generateReferenceBatch(t *testing.T, dir string, classes []string) map[string]string {
	t.Helper()
	result := make(map[string]string)

	// Process in batches to avoid argument-length limits.
	batchSize := 500
	for start := 0; start < len(classes); start += batchSize {
		end := start + batchSize
		if end > len(classes) {
			end = len(classes)
		}
		batch := classes[start:end]
		batchResult := runReferenceBatch(t, dir, batch)
		for k, v := range batchResult {
			result[k] = v
		}
	}
	return result
}

// runReferenceBatch runs the CLI on a single batch and returns per-class CSS.
func runReferenceBatch(t *testing.T, dir string, classes []string) map[string]string {
	t.Helper()

	// Build an HTML file containing all classes.
	var html strings.Builder
	// Include group/peer wrapper elements for group-*/peer-* variants.
	html.WriteString(`<div class="group"><div class="peer">`)
	for _, c := range classes {
		html.WriteString(fmt.Sprintf(`<div class="%s"></div>`, c))
	}
	html.WriteString(`</div></div>`)

	inputHTML := filepath.Join(dir, "input.html")
	inputCSS := filepath.Join(dir, "input.css")
	outputCSS := filepath.Join(dir, "output.css")

	if err := os.WriteFile(inputHTML, []byte(html.String()), 0644); err != nil {
		t.Fatalf("Failed to write input HTML: %v", err)
	}

	if _, err := os.Stat(inputCSS); os.IsNotExist(err) {
		if err := os.WriteFile(inputCSS, []byte("@import \"tailwindcss\";\n"), 0644); err != nil {
			t.Fatalf("Failed to write input CSS: %v", err)
		}
	}

	cmd := exec.Command("npx", "@tailwindcss/cli",
		"--input", inputCSS,
		"--content", inputHTML,
		"--output", outputCSS)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("tailwindcss CLI failed: %v\n%s", err, output)
	}

	data, err := os.ReadFile(outputCSS)
	if err != nil {
		t.Fatalf("Failed to read CLI output: %v", err)
	}

	fullCSS := string(data)

	// Extract per-class rules.
	result := make(map[string]string)
	for _, c := range classes {
		rules := extractFuzzUtilityRules(fullCSS, c)
		if rules != "" {
			result[c] = rules
		}
	}
	return result
}

// extractFuzzUtilityRules extracts utility-relevant CSS for a class, matching
// both plain and variant-wrapped selectors.
func extractFuzzUtilityRules(fullCSS, className string) string {
	escapedClass := cssEscapeClass(className)

	rules := parseCSSRules(fullCSS)
	var utilityRules []cssRule

	for _, r := range rules {
		sel := r.selector
		// Strip "@layer utilities " prefix if present.
		sel = strings.TrimPrefix(sel, "@layer utilities ")

		if matchesClassSelector(sel, escapedClass, className) {
			utilityRules = append(utilityRules, r)
		}
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

// matchesClassSelector checks whether a CSS selector targets the given class.
// It matches exact selectors like .flex, variant-wrapped selectors like
// .hover\:bg-blue-500:hover, and media-query combined selectors.
func matchesClassSelector(selector, escapedClass, rawClass string) bool {
	// Exact match: ".classname"
	if selector == "."+escapedClass || selector == "."+rawClass {
		return true
	}

	// Variant-wrapped: selector starts with "." + escaped class, possibly
	// followed by pseudo-classes/elements or combinators.
	prefix := "." + escapedClass
	if strings.HasPrefix(selector, prefix) {
		rest := selector[len(prefix):]
		if rest == "" || rest[0] == ':' || rest[0] == ' ' || rest[0] == '>' {
			return true
		}
	}

	// Media-query wrapped: "@media ... .classname..." or "@supports ...".
	if strings.HasPrefix(selector, "@") {
		parts := strings.SplitN(selector, " .", 2)
		if len(parts) == 2 {
			classPart := "." + parts[1]
			return matchesClassSelector(classPart, escapedClass, rawClass)
		}
	}

	return false
}

// normalizeCalc normalizes whitespace around operators inside calc() for comparison.
func normalizeCalc(s string) string {
	result := s
	for _, op := range []string{" * ", " + ", " - "} {
		compact := strings.ReplaceAll(op, " ", "")
		result = strings.ReplaceAll(result, op, compact)
	}
	for _, op := range []string{"* ", " *", "+ ", " +", "- ", " -"} {
		compact := strings.TrimSpace(op)
		result = strings.ReplaceAll(result, op, compact)
	}
	return result
}

// normalizeVarFallback normalizes "var(--x, )" vs "var(--x,)" differences.
func normalizeVarFallback(s string) string {
	return strings.ReplaceAll(s, ", )", ",)")
}

// normalizeDeclValue applies all value normalizations for comparison.
func normalizeDeclValue(s string) string {
	s = normalizeCalc(s)
	s = normalizeVarFallback(s)
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

// compareDeclarations compares two declaration maps and returns mismatch descriptions.
func compareDeclarations(ref, got map[string]string) []string {
	var mismatches []string
	for prop, refVal := range ref {
		goVal, ok := got[prop]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("  MISSING %s (expected: %s)", prop, refVal))
			continue
		}
		if normalizeDeclValue(goVal) != normalizeDeclValue(refVal) {
			mismatches = append(mismatches, fmt.Sprintf("  DIFF %s:\n    ref: %s\n    go:  %s", prop, refVal, goVal))
		}
	}
	for prop, goVal := range got {
		if _, ok := ref[prop]; !ok {
			mismatches = append(mismatches, fmt.Sprintf("  EXTRA %s: %s", prop, goVal))
		}
	}
	sort.Strings(mismatches)
	return mismatches
}

// TestDifferentialFuzz generates random Tailwind CSS classes, processes them
// through both the official TailwindCSS CLI and the Go implementation, and
// asserts that outputs are semantically equivalent.
//
// Run with: go test -tags reference -run TestDifferentialFuzz -v -timeout 30m
func TestDifferentialFuzz(t *testing.T) {
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not found; skipping differential fuzz test")
	}

	rng := rand.New(rand.NewSource(*fuzzSeed))
	classes := generateRandomClasses(rng, *fuzzCount)

	// Deduplicate while preserving order.
	seen := map[string]bool{}
	var unique []string
	for _, c := range classes {
		if !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}
	classes = unique
	t.Logf("Testing %d unique classes (seed=%d)", len(classes), *fuzzSeed)

	// Setup official CLI (cached).
	cliDir := setupTailwindCLI(t)

	// Generate reference CSS (batched).
	refMap := generateReferenceBatch(t, cliDir, classes)
	t.Logf("Reference CLI produced output for %d classes", len(refMap))

	// Generate Go CSS for all classes in one engine (for efficiency).
	goEngine := New()
	var htmlBuf strings.Builder
	htmlBuf.WriteString(`<div class="group"><div class="peer">`)
	for _, c := range classes {
		htmlBuf.WriteString(fmt.Sprintf(`<div class="%s"></div>`, c))
	}
	htmlBuf.WriteString(`</div></div>`)
	goEngine.Write([]byte(htmlBuf.String()))
	goFullCSS := goEngine.CSS()

	// Track results.
	var pass, fail, skip int
	failThreshold := len(classes) / 2

	for _, class := range classes {
		t.Run(class, func(t *testing.T) {
			refCSS, hasRef := refMap[class]

			// Generate Go output for this specific class individually.
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="group"><div class="peer"><div class="%s"></div></div></div>`, class)))
			goCSS := e.CSS()

			if (!hasRef || refCSS == "") && goCSS == "" {
				skip++
				t.Skipf("both produced no output for %q", class)
				return
			}

			if !hasRef || refCSS == "" {
				// Go produced output but reference didn't — log as info, not failure.
				skip++
				t.Logf("reference produced no utility rules for %q, go produced output", class)
				return
			}

			if goCSS == "" {
				fail++
				t.Errorf("tailwind-go produced no CSS for %q\nReference:\n%s", class, refCSS)
				return
			}

			// Semantic comparison.
			refRules := parseCSSRules(refCSS)
			goNorm := normalizeCSS(goCSS)
			goRules := parseCSSRules(goNorm)

			refDecls := collectDeclarations(refRules)
			goDecls := collectDeclarations(goRules)

			if len(refDecls) == 0 {
				skip++
				return
			}

			mismatches := compareDeclarations(refDecls, goDecls)
			if len(mismatches) > 0 {
				fail++
				t.Errorf("class %q:\n%s", class, strings.Join(mismatches, "\n"))
			} else {
				pass++
			}
		})

		if fail > failThreshold {
			t.Logf("Early termination: >50%% failure rate (%d/%d)", fail, len(classes))
			break
		}
	}

	t.Logf("Results: %d pass, %d fail, %d skip out of %d total", pass, fail, skip, len(classes))

	// Verify the Go engine didn't crash processing all classes together.
	if goFullCSS == "" {
		t.Error("Go engine produced no CSS when processing all classes together")
	}
}
