package tailwind

import (
	"strings"
	"testing"
)

func TestResolveUtilityColorClass(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; --color-red-400: #f87171; }
@utility bg-* { background-color: --value(--color); }
@utility text-* { color: --value(--color); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-blue-500 text-red-400"`))
	result := e.CSS()
	if !strings.Contains(result, "#3b82f6") {
		t.Errorf("bg-blue-500 not resolved: %s", result)
	}
	if !strings.Contains(result, "#f87171") {
		t.Errorf("text-red-400 not resolved: %s", result)
	}
}

func TestSpaceXChildSelector(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility space-x-* > :not(:last-child) {
  --tw-space-x-reverse: 0;
  margin-inline-end: calc(--value(--spacing) * var(--tw-space-x-reverse));
  margin-inline-start: calc(--value(--spacing) * calc(1 - var(--tw-space-x-reverse)));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="space-x-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Must target children, not the parent element.
	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-x-4 should use child selector > :not(:last-child): %s", result)
	}
	// Must use logical properties.
	if !strings.Contains(result, "margin-inline-start") {
		t.Errorf("space-x-4 should use margin-inline-start: %s", result)
	}
	if !strings.Contains(result, "margin-inline-end") {
		t.Errorf("space-x-4 should use margin-inline-end: %s", result)
	}
	// Must NOT use margin-left (old behavior).
	if strings.Contains(result, "margin-left") {
		t.Errorf("space-x-4 should NOT use margin-left: %s", result)
	}
	// Must include the reverse variable declaration.
	if !strings.Contains(result, "--tw-space-x-reverse: 0") {
		t.Errorf("space-x-4 should include --tw-space-x-reverse: 0: %s", result)
	}
}

func TestSpaceYChildSelector(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility space-y-* > :not(:last-child) {
  --tw-space-y-reverse: 0;
  margin-block-end: calc(--value(--spacing) * var(--tw-space-y-reverse));
  margin-block-start: calc(--value(--spacing) * calc(1 - var(--tw-space-y-reverse)));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="space-y-2"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-y-2 should use child selector: %s", result)
	}
	if !strings.Contains(result, "margin-block-start") {
		t.Errorf("space-y-2 should use margin-block-start: %s", result)
	}
	if !strings.Contains(result, "margin-block-end") {
		t.Errorf("space-y-2 should use margin-block-end: %s", result)
	}
	if strings.Contains(result, "margin-top") {
		t.Errorf("space-y-2 should NOT use margin-top: %s", result)
	}
}

func TestSpaceXReverseChildSelector(t *testing.T) {
	css := []byte(`
@utility space-x-reverse > :not(:last-child) { --tw-space-x-reverse: 1; }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="space-x-reverse"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-x-reverse should use child selector: %s", result)
	}
	if !strings.Contains(result, "--tw-space-x-reverse: 1") {
		t.Errorf("space-x-reverse should set --tw-space-x-reverse: 1: %s", result)
	}
}

func TestSpaceXWithEmbeddedCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="space-x-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-x-4 should use child selector: %s", result)
	}
	if !strings.Contains(result, "margin-inline-start") {
		t.Errorf("space-x-4 should use margin-inline-start: %s", result)
	}
	if strings.Contains(result, "margin-left") {
		t.Errorf("space-x-4 should NOT use margin-left: %s", result)
	}
}

func TestDivideXChildSelector(t *testing.T) {
	css := []byte(`
@utility divide-x > :not(:last-child) {
  --tw-divide-x-reverse: 0;
  border-inline-start-width: calc(1px * calc(1 - var(--tw-divide-x-reverse)));
  border-inline-end-width: calc(1px * var(--tw-divide-x-reverse));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="divide-x"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-x should use child selector > :not(:last-child): %s", result)
	}
	if !strings.Contains(result, "border-inline-start-width") {
		t.Errorf("divide-x should use border-inline-start-width: %s", result)
	}
	if !strings.Contains(result, "border-inline-end-width") {
		t.Errorf("divide-x should use border-inline-end-width: %s", result)
	}
	if strings.Contains(result, "border-left-width") {
		t.Errorf("divide-x should NOT use border-left-width: %s", result)
	}
	if !strings.Contains(result, "--tw-divide-x-reverse: 0") {
		t.Errorf("divide-x should include --tw-divide-x-reverse: 0: %s", result)
	}
}

func TestDivideYChildSelector(t *testing.T) {
	css := []byte(`
@utility divide-y > :not(:last-child) {
  --tw-divide-y-reverse: 0;
  border-top-width: calc(1px * calc(1 - var(--tw-divide-y-reverse)));
  border-bottom-width: calc(1px * var(--tw-divide-y-reverse));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="divide-y"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-y should use child selector: %s", result)
	}
	if !strings.Contains(result, "border-top-width") {
		t.Errorf("divide-y should use border-top-width: %s", result)
	}
	if !strings.Contains(result, "border-bottom-width") {
		t.Errorf("divide-y should use border-bottom-width: %s", result)
	}
	if !strings.Contains(result, "--tw-divide-y-reverse: 0") {
		t.Errorf("divide-y should include --tw-divide-y-reverse: 0: %s", result)
	}
}

func TestDivideXReverseChildSelector(t *testing.T) {
	css := []byte(`
@utility divide-x-reverse > :not(:last-child) { --tw-divide-x-reverse: 1; }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="divide-x-reverse"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-x-reverse should use child selector: %s", result)
	}
	if !strings.Contains(result, "--tw-divide-x-reverse: 1") {
		t.Errorf("divide-x-reverse should set --tw-divide-x-reverse: 1: %s", result)
	}
}

func TestDivideXWithEmbeddedCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="divide-x"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-x should use child selector: %s", result)
	}
	if !strings.Contains(result, "border-inline-start-width") {
		t.Errorf("divide-x should use border-inline-start-width: %s", result)
	}
	if strings.Contains(result, "border-left-width") {
		t.Errorf("divide-x should NOT use border-left-width: %s", result)
	}
}

func TestDivideYWithEmbeddedCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="divide-y-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-y-4 should use child selector: %s", result)
	}
	if !strings.Contains(result, "border-top-width") {
		t.Errorf("divide-y-4 should use border-top-width: %s", result)
	}
	if !strings.Contains(result, "border-bottom-width") {
		t.Errorf("divide-y-4 should use border-bottom-width: %s", result)
	}
}

func TestGridColsRepeat(t *testing.T) {
	e := New()
	tests := []struct {
		class string
		want  string
	}{
		{"grid-cols-1", "repeat(1, minmax(0, 1fr))"},
		{"grid-cols-3", "repeat(3, minmax(0, 1fr))"},
		{"grid-cols-12", "repeat(12, minmax(0, 1fr))"},
		{"grid-cols-none", "none"},
		{"grid-cols-subgrid", "subgrid"},
		{"grid-cols-[1fr_auto_2fr]", "1fr auto 2fr"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			if !strings.Contains(result, tt.want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tt.class, tt.want, result)
			}
			if strings.Contains(result, "grid-template-columns") != true {
				t.Errorf("%s: expected grid-template-columns property in CSS output:\n%s", tt.class, result)
			}
		})
	}
}

func TestGridRowsRepeat(t *testing.T) {
	e := New()
	tests := []struct {
		class string
		want  string
	}{
		{"grid-rows-2", "repeat(2, minmax(0, 1fr))"},
		{"grid-rows-4", "repeat(4, minmax(0, 1fr))"},
		{"grid-rows-none", "none"},
		{"grid-rows-subgrid", "subgrid"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			if !strings.Contains(result, tt.want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tt.class, tt.want, result)
			}
			if !strings.Contains(result, "grid-template-rows") {
				t.Errorf("%s: expected grid-template-rows property in CSS output:\n%s", tt.class, result)
			}
		})
	}
}

func TestResolveGradientFromComposition(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; }
@utility from-* {
  --tw-gradient-from: --value(--color);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to, transparent);
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="from-blue-500"`))
	result := e.CSS()
	if !strings.Contains(result, "--tw-gradient-from: #3b82f6") {
		t.Errorf("from-blue-500 should set --tw-gradient-from: %s", result)
	}
	if !strings.Contains(result, "--tw-gradient-stops:") {
		t.Errorf("from-blue-500 should set --tw-gradient-stops: %s", result)
	}
}

func TestResolveGradientViaComposition(t *testing.T) {
	css := []byte(`
@theme { --color-green-500: #22c55e; }
@utility via-* {
  --tw-gradient-via: --value(--color);
  --tw-gradient-stops: var(--tw-gradient-from, transparent), var(--tw-gradient-via), var(--tw-gradient-to, transparent);
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="via-green-500"`))
	result := e.CSS()
	if !strings.Contains(result, "--tw-gradient-via: #22c55e") {
		t.Errorf("via-green-500 should set --tw-gradient-via: %s", result)
	}
	if !strings.Contains(result, "--tw-gradient-stops:") {
		t.Errorf("via-green-500 should set --tw-gradient-stops: %s", result)
	}
	if !strings.Contains(result, "var(--tw-gradient-via)") {
		t.Errorf("via-green-500 stops should include var(--tw-gradient-via): %s", result)
	}
}

func TestResolveGradientFullComposition(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; --color-green-500: #22c55e; --color-red-500: #ef4444; }
@utility from-* {
  --tw-gradient-from: --value(--color);
  --tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to, transparent);
}
@utility via-* {
  --tw-gradient-via: --value(--color);
  --tw-gradient-stops: var(--tw-gradient-from, transparent), var(--tw-gradient-via), var(--tw-gradient-to, transparent);
}
@utility to-* {
  --tw-gradient-to: --value(--color);
}
@utility bg-gradient-to-r { background-image: linear-gradient(to right, var(--tw-gradient-stops)); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="from-blue-500 via-green-500 to-red-500 bg-gradient-to-r"`))
	result := e.CSS()
	if !strings.Contains(result, "--tw-gradient-from: #3b82f6") {
		t.Errorf("should contain --tw-gradient-from: %s", result)
	}
	if !strings.Contains(result, "--tw-gradient-via: #22c55e") {
		t.Errorf("should contain --tw-gradient-via: %s", result)
	}
	if !strings.Contains(result, "--tw-gradient-to: #ef4444") {
		t.Errorf("should contain --tw-gradient-to: %s", result)
	}
	if !strings.Contains(result, "linear-gradient(to right, var(--tw-gradient-stops))") {
		t.Errorf("should contain linear-gradient: %s", result)
	}
}

func TestResolveUtilityBorderVariants(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility border-* { border-width: --value(--spacing); }
@utility border-t-* { border-top-width: --value(--spacing); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="border-2 border-t-4"`))
	result := e.CSS()
	if !strings.Contains(result, "border-width") {
		t.Errorf("border-2 not resolved: %s", result)
	}
	if !strings.Contains(result, "border-top-width") {
		t.Errorf("border-t-4 not resolved: %s", result)
	}
}

func TestFilterComposition(t *testing.T) {
	e := New()
	e.Write([]byte(`class="blur-md brightness-75"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Both utilities should set their own CSS variable.
	if !strings.Contains(result, "--tw-blur: blur(") {
		t.Errorf("blur-md should set --tw-blur variable: %s", result)
	}
	if !strings.Contains(result, "--tw-brightness: brightness(75%)") {
		t.Errorf("brightness-75 should set --tw-brightness variable: %s", result)
	}

	// Both utilities should output the composed filter property.
	composedFilter := "var(--tw-blur,) var(--tw-brightness,)"
	if !strings.Contains(result, composedFilter) {
		t.Errorf("filter utilities should output composed filter with CSS variables: %s", result)
	}

	// The filter property should NOT be a bare blur() or brightness() value.
	if strings.Contains(result, "filter: blur(") {
		t.Errorf("filter should use CSS variable composition, not bare filter functions: %s", result)
	}
	if strings.Contains(result, "filter: brightness(") {
		t.Errorf("filter should use CSS variable composition, not bare filter functions: %s", result)
	}
}

func TestFilterCompositionSingleUtility(t *testing.T) {
	e := New()
	e.Write([]byte(`class="blur-md"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Single filter utility should still use composed pattern.
	if !strings.Contains(result, "--tw-blur:") {
		t.Errorf("blur-md should set --tw-blur: %s", result)
	}
	if !strings.Contains(result, "var(--tw-blur,)") {
		t.Errorf("blur-md should use composed filter: %s", result)
	}
}

func TestBackdropFilterComposition(t *testing.T) {
	e := New()
	e.Write([]byte(`class="backdrop-blur-lg backdrop-brightness-50"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Both utilities should set their own CSS variable.
	if !strings.Contains(result, "--tw-backdrop-blur: blur(") {
		t.Errorf("backdrop-blur-lg should set --tw-backdrop-blur variable: %s", result)
	}
	if !strings.Contains(result, "--tw-backdrop-brightness: brightness(50%)") {
		t.Errorf("backdrop-brightness-50 should set --tw-backdrop-brightness variable: %s", result)
	}

	// Both should use composed backdrop-filter.
	composedFilter := "var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,)"
	if !strings.Contains(result, composedFilter) {
		t.Errorf("backdrop-filter utilities should output composed filter: %s", result)
	}
}

func TestFontVariantNumericComposition(t *testing.T) {
	e := New()
	e.Write([]byte(`class="ordinal tabular-nums"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Each utility should set its own custom property.
	if !strings.Contains(result, "--tw-ordinal: ordinal") {
		t.Errorf("ordinal should set --tw-ordinal: %s", result)
	}
	if !strings.Contains(result, "--tw-numeric-spacing: tabular-nums") {
		t.Errorf("tabular-nums should set --tw-numeric-spacing: %s", result)
	}

	// Both should use the composed font-variant-numeric value.
	composedFVN := "var(--tw-ordinal,) var(--tw-slashed-zero,) var(--tw-numeric-figure,) var(--tw-numeric-spacing,) var(--tw-numeric-fraction,)"
	if !strings.Contains(result, composedFVN) {
		t.Errorf("font-variant-numeric should use composed pattern: %s", result)
	}

	// Should NOT set font-variant-numeric directly to a bare value.
	if strings.Contains(result, "font-variant-numeric: ordinal;") {
		t.Errorf("ordinal should use composite pattern, not bare value: %s", result)
	}
	if strings.Contains(result, "font-variant-numeric: tabular-nums;") {
		t.Errorf("tabular-nums should use composite pattern, not bare value: %s", result)
	}
}

func TestFontVariantNumericSingle(t *testing.T) {
	e := New()
	e.Write([]byte(`class="ordinal"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "--tw-ordinal: ordinal") {
		t.Errorf("ordinal should set --tw-ordinal: %s", result)
	}
	if !strings.Contains(result, "var(--tw-ordinal,)") {
		t.Errorf("ordinal should use composed font-variant-numeric: %s", result)
	}
}

func TestFontVariantNumericNormalNums(t *testing.T) {
	e := New()
	e.Write([]byte(`class="normal-nums"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// normal-nums should NOT use the composite pattern — it resets directly.
	if !strings.Contains(result, "font-variant-numeric: normal") {
		t.Errorf("normal-nums should set font-variant-numeric: normal: %s", result)
	}
	if strings.Contains(result, "--tw-ordinal") {
		t.Errorf("normal-nums should not set custom properties: %s", result)
	}
}

func TestFontVariantNumericAllCategories(t *testing.T) {
	tests := []struct {
		class    string
		property string
		value    string
	}{
		{"ordinal", "--tw-ordinal", "ordinal"},
		{"slashed-zero", "--tw-slashed-zero", "slashed-zero"},
		{"lining-nums", "--tw-numeric-figure", "lining-nums"},
		{"oldstyle-nums", "--tw-numeric-figure", "oldstyle-nums"},
		{"proportional-nums", "--tw-numeric-spacing", "proportional-nums"},
		{"tabular-nums", "--tw-numeric-spacing", "tabular-nums"},
		{"diagonal-fractions", "--tw-numeric-fraction", "diagonal-fractions"},
		{"stacked-fractions", "--tw-numeric-fraction", "stacked-fractions"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			expected := tt.property + ": " + tt.value
			if !strings.Contains(result, expected) {
				t.Errorf("%s should contain %q:\n%s", tt.class, expected, result)
			}
		})
	}
}

func TestTouchActionComposition(t *testing.T) {
	e := New()
	e.Write([]byte(`class="touch-pan-x touch-pan-y"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Each utility should set its own custom property.
	if !strings.Contains(result, "--tw-pan-x: pan-x") {
		t.Errorf("touch-pan-x should set --tw-pan-x: %s", result)
	}
	if !strings.Contains(result, "--tw-pan-y: pan-y") {
		t.Errorf("touch-pan-y should set --tw-pan-y: %s", result)
	}

	// Both should use the composed touch-action value.
	composedTouch := "var(--tw-pan-x,) var(--tw-pan-y,) var(--tw-pinch-zoom,)"
	if !strings.Contains(result, composedTouch) {
		t.Errorf("touch-action should use composed pattern: %s", result)
	}

	// Should NOT set touch-action directly to a bare value.
	if strings.Contains(result, "touch-action: pan-x;") {
		t.Errorf("touch-pan-x should use composite pattern, not bare value: %s", result)
	}
}

func TestTouchActionCompositionAllThree(t *testing.T) {
	e := New()
	e.Write([]byte(`class="touch-pan-x touch-pan-y touch-pinch-zoom"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "--tw-pan-x: pan-x") {
		t.Errorf("touch-pan-x should set --tw-pan-x: %s", result)
	}
	if !strings.Contains(result, "--tw-pan-y: pan-y") {
		t.Errorf("touch-pan-y should set --tw-pan-y: %s", result)
	}
	if !strings.Contains(result, "--tw-pinch-zoom: pinch-zoom") {
		t.Errorf("touch-pinch-zoom should set --tw-pinch-zoom: %s", result)
	}
}

func TestTouchActionDirectUtilities(t *testing.T) {
	// touch-auto, touch-none, touch-manipulation should NOT use composite.
	tests := []struct {
		class string
		value string
	}{
		{"touch-auto", "auto"},
		{"touch-none", "none"},
		{"touch-manipulation", "manipulation"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			if !strings.Contains(result, "touch-action: "+tt.value) {
				t.Errorf("%s should set touch-action: %s:\n%s", tt.class, tt.value, result)
			}
			if strings.Contains(result, "--tw-pan") {
				t.Errorf("%s should not use custom properties: %s", tt.class, result)
			}
		})
	}
}

func TestTouchActionDirectionVariants(t *testing.T) {
	tests := []struct {
		class    string
		property string
		value    string
	}{
		{"touch-pan-left", "--tw-pan-x", "pan-left"},
		{"touch-pan-right", "--tw-pan-x", "pan-right"},
		{"touch-pan-up", "--tw-pan-y", "pan-up"},
		{"touch-pan-down", "--tw-pan-y", "pan-down"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			expected := tt.property + ": " + tt.value
			if !strings.Contains(result, expected) {
				t.Errorf("%s should contain %q:\n%s", tt.class, expected, result)
			}
		})
	}
}

func TestPropertyDeclarationsInCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="translate-x-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// CSS() should emit @property for custom properties used.
	if !strings.Contains(result, "@property --tw-translate-x") {
		t.Errorf("CSS() should emit @property for --tw-translate-x:\n%s", result)
	}
	if !strings.Contains(result, "@property --tw-translate-y") {
		t.Errorf("CSS() should emit @property for --tw-translate-y:\n%s", result)
	}
}

func TestPropertyDeclarationsForShadow(t *testing.T) {
	e := New()
	e.Write([]byte(`class="shadow-md"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "@property --tw-shadow") {
		t.Errorf("CSS() should emit @property for --tw-shadow:\n%s", result)
	}
}

func TestPropertyDeclarationsForComposite(t *testing.T) {
	e := New()
	e.Write([]byte(`class="ordinal"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "@property --tw-ordinal") {
		t.Errorf("CSS() should emit @property for --tw-ordinal:\n%s", result)
	}
}

func TestNoPropertyDeclarationsForSimpleUtility(t *testing.T) {
	e := New()
	e.Write([]byte(`class="flex"`))
	result := e.CSS()

	if strings.Contains(result, "@property") {
		t.Errorf("flex should not emit any @property declarations:\n%s", result)
	}
}
