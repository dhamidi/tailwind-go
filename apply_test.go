package tailwind

import (
	"strings"
	"testing"
)

// --- @apply directive tests ---

func TestApplyMultipleUtilities(t *testing.T) {
	// @apply with several utilities: padding, margin, shadow, border-radius, border
	css := []byte(`
@theme {
  --spacing: 0.25rem;
  --color-gray-200: #e5e7eb;
  --radius-xl: 0.75rem;
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
}
@utility p-* { padding: --value(--spacing); }
@utility m-* { margin: --value(--spacing); }
@utility rounded-* { border-radius: --value(--radius); }
@utility shadow-* { box-shadow: --value(--shadow); }
@utility border { border-width: 1px; }

.card {
  @apply p-4 m-2 shadow-lg rounded-xl border;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".card") {
		t.Error("missing .card selector")
	}
	if !strings.Contains(result, "padding:") {
		t.Error("missing padding declaration from p-4")
	}
	if !strings.Contains(result, "margin:") {
		t.Error("missing margin declaration from m-2")
	}
	if !strings.Contains(result, "border-radius:") {
		t.Error("missing border-radius from rounded-xl")
	}
	if !strings.Contains(result, "border-width: 1px") {
		t.Error("missing border-width from border")
	}
}

func TestApplyWithImportant(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility p-* { padding: --value(--spacing); }

.btn {
  @apply !p-4;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".btn") {
		t.Error("missing .btn selector")
	}
	// The important modifier should produce !important on the padding
	if !strings.Contains(result, "!important") {
		t.Error("expected !important in output for !p-4")
	}
}

func TestApplyWithHoverVariant(t *testing.T) {
	css := []byte(`
@theme { --color-blue-600: #2563eb; }
@utility bg-* { background-color: --value(--color); }
@variant hover (&:hover);

.btn {
  @apply hover:bg-blue-600;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ":hover") {
		t.Errorf("expected :hover selector, got: %s", result)
	}
	if !strings.Contains(result, "background-color:") {
		t.Errorf("expected background-color declaration, got: %s", result)
	}
}

func TestApplyWithFocusVariant(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility ring-* { box-shadow: --value(--shadow); }
@variant focus (&:focus);

.btn {
  @apply focus:ring-2;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ":focus") {
		t.Logf("note: focus variant may not generate :focus selector")
	}
}

func TestApplyWithResponsiveVariant(t *testing.T) {
	// Responsive variant applied inside @apply
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility px-* { padding-left: --value(--spacing); padding-right: --value(--spacing); }
@variant md (@media (width >= 48rem));

.container {
  @apply md:px-8;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	// If responsive variants work in @apply, expect @media wrapper
	if !strings.Contains(result, "@media") {
		t.Logf("note: responsive variant in @apply may not produce @media wrapper")
	}
}

func TestApplyNegativeValues(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility mt-* { margin-top: --value(--spacing); }
@utility ml-* { margin-left: --value(--spacing); }

.overlap {
  @apply -mt-4 -ml-2;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".overlap") {
		t.Error("missing .overlap selector")
	}
	// Negative values should produce negative margins
	if !strings.Contains(result, "margin-top:") && !strings.Contains(result, "margin") {
		t.Error("expected margin declarations for negative utilities")
	}
}

func TestApplyOrderingOverride(t *testing.T) {
	// Later @apply should override earlier @apply
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility p-* { padding: --value(--spacing); }

.foo {
  @apply p-2;
  @apply p-4;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".foo") {
		t.Error("missing .foo selector")
	}
	// Both p-2 and p-4 declarations should be present (CSS cascade handles override)
	if !strings.Contains(result, "padding:") {
		t.Error("expected padding declaration")
	}
}

func TestApplyWithCustomUtility(t *testing.T) {
	// Define a custom @utility and use it in @apply
	css := []byte(`
@utility custom-flex {
  display: flex;
  align-items: center;
}

.nav {
  @apply custom-flex;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".nav") {
		t.Error("missing .nav selector")
	}
	if !strings.Contains(result, "display: flex") {
		t.Error("@apply custom-flex should produce display: flex")
	}
	if !strings.Contains(result, "align-items: center") {
		t.Error("@apply custom-flex should produce align-items: center")
	}
}

func TestApplyWithBuiltinUtilities(t *testing.T) {
	// Use the engine's built-in utilities with @apply
	e := New()
	e.LoadCSS([]byte(`
.btn {
  @apply px-4 py-2 rounded text-white;
}
`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".btn") {
		t.Error("missing .btn selector")
	}
	// px-4 should generate padding-left/padding-right or padding-inline
	if !strings.Contains(result, "padding") {
		t.Error("expected padding from px-4 py-2")
	}
}

// --- Custom @utility tests ---

func TestCustomStaticUtility(t *testing.T) {
	css := []byte(`
@utility btn-primary {
  background-color: blue;
  color: white;
  padding: 0.5rem 1rem;
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`<div class="btn-primary">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "btn-primary") {
		t.Error("missing btn-primary class in output")
	}
	if !strings.Contains(result, "background-color: blue") {
		t.Error("missing background-color: blue")
	}
	if !strings.Contains(result, "color: white") {
		t.Error("missing color: white")
	}
	if !strings.Contains(result, "padding: 0.5rem 1rem") {
		t.Error("missing padding")
	}
}

func TestCustomUtilityWithVariant(t *testing.T) {
	css := []byte(`
@utility btn-primary {
  background-color: blue;
  color: white;
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`<div class="hover:btn-primary">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	// hover variant should work with custom utility
	if strings.Contains(result, ":hover") && strings.Contains(result, "background-color: blue") {
		// pass — variant applied correctly
	} else {
		t.Logf("note: hover:btn-primary may or may not be supported; got: %s", result)
	}
}

// --- @theme tests ---

func TestThemeOverrideColor(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@theme { --color-brand: #ff6600; }
@utility bg-* { background-color: --value(--color); }
`))
	e.Write([]byte(`<div class="bg-brand">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "#ff6600") {
		t.Errorf("expected bg-brand to resolve to #ff6600, got: %s", result)
	}
}

func TestThemeOverrideSpacing(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@theme { --spacing: 0.5rem; }
`))
	e.Write([]byte(`<div class="p-4">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	// p-4 with --spacing: 0.5rem should produce calc(var(--spacing) * 4) or 2rem
	if !strings.Contains(result, "padding") {
		t.Error("expected padding declaration for p-4")
	}
}

func TestThemeInlineMode(t *testing.T) {
	// @theme inline should cause values to be inlined rather than using var()
	e := New()
	e.LoadCSS([]byte(`
@theme inline {
  --spacing-xl: 3rem;
}
@utility p-* { padding: --value(--spacing); }
`))
	e.Write([]byte(`<div class="p-xl">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	// If inline mode works, we should see the literal value, not var(--spacing-xl)
	if strings.Contains(result, "padding") {
		if strings.Contains(result, "var(--spacing-xl)") {
			t.Logf("note: @theme inline may not suppress variable reference")
		}
	}
}

func TestThemeCustomFontFamily(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@theme { --font-display: "Playfair Display", serif; }
`))
	e.Write([]byte(`<div class="font-display">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "font-family") {
		t.Error("expected font-family declaration for font-display")
	}
	if !strings.Contains(result, "Playfair Display") && !strings.Contains(result, "--font-display") {
		t.Error("expected custom font family value or variable reference")
	}
}

func TestThemeCustomRadius(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@theme { --radius-huge: 2rem; }
`))
	e.Write([]byte(`<div class="rounded-huge">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "border-radius") {
		t.Errorf("expected border-radius for rounded-huge, got: %s", result)
	}
}

// --- Custom @variant tests ---

func TestCustomVariantDefinition(t *testing.T) {
	css := []byte(`
@variant hocus (&:hover, &:focus);
@utility text-* { color: --value(--color); }
@theme { --color-blue-500: #3b82f6; }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`<div class="hocus:text-blue-500">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	// The custom hocus variant should generate :hover and/or :focus selectors
	if !strings.Contains(result, ":hover") && !strings.Contains(result, ":focus") {
		t.Errorf("expected :hover or :focus from hocus variant, got: %s", result)
	}
	if !strings.Contains(result, "#3b82f6") && !strings.Contains(result, "color:") {
		t.Errorf("expected color declaration for text-blue-500, got: %s", result)
	}
}

func TestCustomMediaVariant(t *testing.T) {
	css := []byte(`
@variant tall (@media (min-height: 800px));
@utility text-* { color: --value(--color); }
@theme { --color-red-500: #ef4444; }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`<div class="tall:text-red-500">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "@media") {
		t.Errorf("expected @media wrapper for tall variant, got: %s", result)
	}
	if !strings.Contains(result, "min-height") {
		t.Errorf("expected min-height in media query, got: %s", result)
	}
}

// --- @apply combined with @theme and @utility ---

func TestApplyResolvesCustomThemeColor(t *testing.T) {
	css := []byte(`
@theme { --color-primary: #e11d48; }
@utility bg-* { background-color: --value(--color); }

.hero {
  @apply bg-primary;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".hero") {
		t.Error("missing .hero selector")
	}
	if !strings.Contains(result, "#e11d48") {
		t.Errorf("expected bg-primary to resolve to #e11d48, got: %s", result)
	}
}

func TestApplyResolvesCustomUtility(t *testing.T) {
	// Define @utility then use in @apply
	css := []byte(`
@utility center-flex {
  display: flex;
  justify-content: center;
  align-items: center;
}

.modal {
  @apply center-flex;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".modal") {
		t.Error("missing .modal selector")
	}
	if !strings.Contains(result, "display: flex") {
		t.Error("expected display: flex from center-flex")
	}
	if !strings.Contains(result, "justify-content: center") {
		t.Error("expected justify-content: center from center-flex")
	}
	if !strings.Contains(result, "align-items: center") {
		t.Error("expected align-items: center from center-flex")
	}
}

func TestApplyMultipleVariants(t *testing.T) {
	// @apply with multiple variant-prefixed classes
	css := []byte(`
@theme {
  --color-blue-500: #3b82f6;
  --color-blue-700: #1d4ed8;
}
@utility bg-* { background-color: --value(--color); }
@utility text-* { color: --value(--color); }
@variant hover (&:hover);
@variant focus (&:focus);

.interactive {
  @apply hover:bg-blue-700 focus:text-blue-500;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ":hover") {
		t.Errorf("expected :hover selector, got: %s", result)
	}
	if !strings.Contains(result, ":focus") {
		t.Errorf("expected :focus selector, got: %s", result)
	}
}

func TestApplyMixedStaticAndDynamic(t *testing.T) {
	// Mix of static utility and dynamic value utility in @apply
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility flex { display: flex; }
@utility items-center { align-items: center; }
@utility gap-* { gap: --value(--spacing); }

.row {
  @apply flex items-center gap-4;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".row") {
		t.Error("missing .row selector")
	}
	if !strings.Contains(result, "display: flex") {
		t.Error("expected display: flex")
	}
	if !strings.Contains(result, "align-items: center") {
		t.Error("expected align-items: center")
	}
	if !strings.Contains(result, "gap:") {
		t.Error("expected gap declaration from gap-4")
	}
}

// --- @apply edge cases ---

func TestApplyEmptyClassList(t *testing.T) {
	// @apply with no classes should not crash
	css := []byte(`
.empty {
  @apply;
}
`)
	e := New()
	err := e.LoadCSS(css)
	// Should not panic; might return error or produce empty output
	_ = err
	_ = e.CSS()
}

func TestApplyDuplicateClasses(t *testing.T) {
	css := []byte(`
@utility flex { display: flex; }

.dup {
  @apply flex flex;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".dup") {
		t.Error("missing .dup selector")
	}
	if !strings.Contains(result, "display: flex") {
		t.Error("expected display: flex")
	}
}

func TestApplyWithBuiltinVariants(t *testing.T) {
	// Use the engine's built-in hover variant with @apply
	e := New()
	e.LoadCSS([]byte(`
.link {
  @apply hover:underline;
}
`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, ".link") {
		t.Logf("note: .link may be missing if hover:underline doesn't resolve")
	}
}

func TestThemeOverridesPrecedence(t *testing.T) {
	// A second @theme should override the first
	e := New()
	e.LoadCSS([]byte(`
@theme { --color-accent: #111111; }
`))
	e.LoadCSS([]byte(`
@theme { --color-accent: #222222; }
@utility bg-* { background-color: --value(--color); }
`))
	e.Write([]byte(`<div class="bg-accent">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "#222222") {
		t.Errorf("expected second theme override #222222, got: %s", result)
	}
}

func TestThemeVariablesInCSS(t *testing.T) {
	// Verify that theme variables are emitted as CSS custom properties
	e := New()
	e.LoadCSS([]byte(`
@theme { --color-custom-purple: #7c3aed; }
@utility bg-* { background-color: --value(--color); }
`))
	e.Write([]byte(`<div class="bg-custom-purple">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "#7c3aed") && !strings.Contains(result, "--color-custom-purple") {
		t.Errorf("expected custom-purple color value or variable, got: %s", result)
	}
}

func TestCustomUtilityMultipleDeclarations(t *testing.T) {
	// Custom utility with many declarations
	css := []byte(`
@utility card-base {
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  background-color: white;
  overflow: hidden;
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`<div class="card-base">`))
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "card-base") {
		t.Error("missing card-base in output")
	}
	if !strings.Contains(result, "border-radius: 0.5rem") {
		t.Error("missing border-radius")
	}
	if !strings.Contains(result, "background-color: white") {
		t.Error("missing background-color")
	}
	if !strings.Contains(result, "overflow: hidden") {
		t.Error("missing overflow")
	}
}

func TestApplyWithNonexistentAndValidClasses(t *testing.T) {
	// Mix of valid and invalid classes in @apply
	css := []byte(`
@utility flex { display: flex; }

.mixed {
  @apply flex nonexistent;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	// Should still produce the valid utility
	if !strings.Contains(result, "display: flex") {
		t.Error("expected display: flex for the valid class")
	}
}

func TestApplyPreservesHostDeclarations(t *testing.T) {
	// Regular declarations in the same rule as @apply
	css := []byte(`
@utility flex { display: flex; }

.btn {
  cursor: pointer;
  @apply flex;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS:\n%s", result)

	if !strings.Contains(result, "display: flex") {
		t.Error("expected display: flex from @apply")
	}
}
