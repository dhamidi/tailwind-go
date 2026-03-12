//go:build reference

package tailwind

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCompatLive runs a live comparison against the official tailwindcss CLI.
// This test requires npm/npx to be installed and is gated behind the
// "reference" build tag so it doesn't run by default.
//
// Run with: go test -tags reference -run TestCompatLive -v
func TestCompatLive(t *testing.T) {
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not found; skipping live compatibility test")
	}

	tmpDir := t.TempDir()
	setupTailwindDir(t, tmpDir)

	for _, class := range compatBenchClasses {
		t.Run(class, func(t *testing.T) {
			// Generate CSS using tailwind-go
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="%s"></div>`, class)))
			goCSS := e.CSS()

			// Generate CSS using official tailwindcss CLI
			refCSS := generateReferenceCSS(t, tmpDir, class)

			if goCSS == "" {
				t.Errorf("tailwind-go produced no CSS for class %q", class)
				if refCSS != "" {
					t.Logf("Reference output:\n%s", refCSS)
				}
				return
			}

			// Extract and compare utility rules
			refRules := extractUtilityRules(refCSS, class)
			goNorm := normalizeCSS(goCSS)

			if refRules == "" {
				t.Logf("No utility rules in reference for %q; tailwind-go output:\n%s", class, goCSS)
				return
			}

			refParsed := parseCSSRules(refRules)
			goParsed := parseCSSRules(goNorm)

			refDecls := collectDeclarations(refParsed)
			goDecls := collectDeclarations(goParsed)

			var mismatches []string
			for prop, refVal := range refDecls {
				goVal, ok := goDecls[prop]
				if !ok {
					mismatches = append(mismatches, fmt.Sprintf("  MISSING %s (expected: %s)", prop, refVal))
				} else if goVal != refVal {
					mismatches = append(mismatches, fmt.Sprintf("  DIFF %s:\n    ref:  %s\n    go:   %s", prop, refVal, goVal))
				}
			}
			for prop, goVal := range goDecls {
				if _, ok := refDecls[prop]; !ok {
					mismatches = append(mismatches, fmt.Sprintf("  EXTRA %s: %s", prop, goVal))
				}
			}

			if len(mismatches) > 0 {
				t.Errorf("class %q CSS mismatch:\n%s", class, strings.Join(mismatches, "\n"))
			}
		})
	}
}

// generateReferenceCSS runs the official tailwindcss CLI for a single class.
// The tmpDir must have been set up with setupTailwindDir first.
func generateReferenceCSS(t *testing.T, tmpDir, class string) string {
	t.Helper()

	inputHTML := filepath.Join(tmpDir, "input.html")
	inputCSS := filepath.Join(tmpDir, "input.css")
	outputCSS := filepath.Join(tmpDir, "output.css")

	err := os.WriteFile(inputHTML, []byte(fmt.Sprintf(`<div class="%s"></div>`, class)), 0644)
	if err != nil {
		t.Fatalf("Failed to write input HTML: %v", err)
	}

	// Ensure input.css exists with @import
	if _, err := os.Stat(inputCSS); os.IsNotExist(err) {
		if err := os.WriteFile(inputCSS, []byte("@import \"tailwindcss\";\n"), 0644); err != nil {
			t.Fatalf("Failed to write input CSS: %v", err)
		}
	}

	cmd := exec.Command("npx", "@tailwindcss/cli",
		"--input", inputCSS,
		"--content", inputHTML,
		"--output", outputCSS)
	cmd.Dir = tmpDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("tailwindcss CLI failed for class %q: %v\n%s", class, err, output)
		return ""
	}

	data, err := os.ReadFile(outputCSS)
	if err != nil {
		t.Fatalf("Failed to read CLI output: %v", err)
	}

	return string(data)
}

// setupTailwindDir installs tailwindcss in the given directory.
func setupTailwindDir(t *testing.T, dir string) {
	t.Helper()

	// Initialize npm and install tailwindcss
	cmd := exec.Command("npm", "init", "-y")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm init failed: %v\n%s", err, out)
	}

	cmd = exec.Command("npm", "install", "tailwindcss")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("npm install tailwindcss failed: %v\n%s", err, out)
	}

	// Write the input CSS
	if err := os.WriteFile(filepath.Join(dir, "input.css"), []byte("@import \"tailwindcss\";\n"), 0644); err != nil {
		t.Fatalf("Failed to write input.css: %v", err)
	}
}

// TestGenerateReferenceFiles generates reference CSS files for all bench classes.
// This is an alternative to the shell script, run via:
//
//	go test -tags reference -run TestGenerateReferenceFiles -v
func TestGenerateReferenceFiles(t *testing.T) {
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not found; skipping reference generation")
	}

	refDir := filepath.Join("testdata", "reference")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()
	setupTailwindDir(t, tmpDir)

	for _, class := range compatBenchClasses {
		refCSS := generateReferenceCSS(t, tmpDir, class)
		if refCSS == "" {
			t.Logf("SKIP: no output for %q", class)
			continue
		}

		encoded := urlEncode(class)
		outFile := filepath.Join(refDir, encoded+".css")
		if err := os.WriteFile(outFile, []byte(refCSS), 0644); err != nil {
			t.Errorf("Failed to write %s: %v", outFile, err)
		} else {
			t.Logf("OK: %s -> %s", class, encoded+".css")
		}
	}
}
