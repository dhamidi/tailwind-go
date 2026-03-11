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
