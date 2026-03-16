package tailwind

import (
	"strings"
	"testing"
)

func TestDarkModeMediaStrategy(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@variant dark (@media (prefers-color-scheme: dark));
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="dark:text-white"`))
	result := e.CSS()
	if !strings.Contains(result, "@media (prefers-color-scheme: dark)") {
		t.Errorf("missing media query, got: %s", result)
	}
	if !strings.Contains(result, "color: #fff") {
		t.Errorf("missing color declaration, got: %s", result)
	}
}

func TestDarkModeClassStrategy(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@variant dark (&:where(.dark, .dark *));
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="dark:text-white"`))
	result := e.CSS()
	if strings.Contains(result, "@media") {
		t.Errorf("should not have media query for class strategy, got: %s", result)
	}
	if !strings.Contains(result, ".dark") {
		t.Errorf("missing .dark class selector, got: %s", result)
	}
	if !strings.Contains(result, "color: #fff") {
		t.Errorf("missing color declaration, got: %s", result)
	}
}

func TestDarkModeOverrideFromDefault(t *testing.T) {
	// Start with media strategy, then override with class strategy.
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@variant dark (@media (prefers-color-scheme: dark));
`)
	e := New()
	e.LoadCSS(css)
	// Override with class strategy.
	e.LoadCSS([]byte(`@variant dark (&:where(.dark, .dark *));`))
	e.Write([]byte(`class="dark:text-white"`))
	result := e.CSS()
	if strings.Contains(result, "prefers-color-scheme") {
		t.Errorf("override failed, still using media strategy: %s", result)
	}
	if !strings.Contains(result, ".dark") {
		t.Errorf("missing .dark class selector after override, got: %s", result)
	}
}

func TestCustomVariantDirective(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@custom-variant dark (&:where(.dark, .dark *));
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="dark:text-white"`))
	result := e.CSS()
	if strings.Contains(result, "@media") {
		t.Errorf("should not have media query for class strategy via @custom-variant, got: %s", result)
	}
	if !strings.Contains(result, ".dark") {
		t.Errorf("missing .dark class selector, got: %s", result)
	}
	if !strings.Contains(result, "color: #fff") {
		t.Errorf("missing color declaration, got: %s", result)
	}
}

func TestCustomVariantNewVariant(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@custom-variant theme-midnight (&:where([data-theme="midnight"] *));
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="theme-midnight:text-white"`))
	result := e.CSS()
	if !strings.Contains(result, `data-theme="midnight"`) {
		t.Errorf("missing data-theme selector, got: %s", result)
	}
	if !strings.Contains(result, "color: #fff") {
		t.Errorf("missing color declaration, got: %s", result)
	}
}

func TestDarkModeClassStrategyOrdering(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@theme { --color-white: #fff; --color-gray-900: #111827; }
@utility text-* { color: --value(--color); }
@variant dark (&:where(.dark, .dark *));
`))
	e.Write([]byte(`class="text-gray-900 dark:text-white"`))
	css := e.CSS()

	// dark:text-white must come AFTER text-gray-900 in the output
	gray900Idx := strings.Index(css, ".text-gray-900")
	darkWhiteIdx := strings.Index(css, `dark\:text-white`)
	if gray900Idx < 0 || darkWhiteIdx < 0 {
		t.Fatalf("missing rules in output:\n%s", css)
	}
	if darkWhiteIdx < gray900Idx {
		t.Errorf("dark:text-white (pos %d) must come after text-gray-900 (pos %d) for correct cascade:\n%s", darkWhiteIdx, gray900Idx, css)
	}
}

func TestSelectorVariantOrderingHoverFocus(t *testing.T) {
	e := New()
	e.Write([]byte(`class="bg-blue-500 hover:bg-blue-600 focus:bg-blue-700"`))
	css := e.CSS()

	baseIdx := strings.Index(css, ".bg-blue-500")
	hoverIdx := strings.Index(css, `hover\:bg-blue-600`)
	focusIdx := strings.Index(css, `focus\:bg-blue-700`)
	if baseIdx < 0 || hoverIdx < 0 || focusIdx < 0 {
		t.Fatalf("missing rules in output:\n%s", css)
	}
	if hoverIdx < baseIdx {
		t.Errorf("hover:bg-blue-600 (pos %d) must come after bg-blue-500 (pos %d)", hoverIdx, baseIdx)
	}
	if focusIdx < baseIdx {
		t.Errorf("focus:bg-blue-700 (pos %d) must come after bg-blue-500 (pos %d)", focusIdx, baseIdx)
	}
}

func TestDarkModeClassStrategyParseVariant(t *testing.T) {
	css := []byte(`@variant dark (&:where(.dark, .dark *));`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Variants) != 1 {
		t.Fatalf("got %d variants, want 1", len(ss.Variants))
	}
	v := ss.Variants[0]
	if v.Name != "dark" {
		t.Errorf("name = %q, want dark", v.Name)
	}
	if v.Selector != "&:where(.dark, .dark *)" {
		t.Errorf("selector = %q, want %q", v.Selector, "&:where(.dark, .dark *)")
	}
	if v.Media != "" {
		t.Errorf("media should be empty, got %q", v.Media)
	}
}
