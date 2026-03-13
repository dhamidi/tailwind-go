package tailwind

import (
	"strings"
	"testing"
)

// TestSpacingScaleBoundary verifies the full spacing scale with computed values.
func TestSpacingScaleBoundary(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		// Boundary values — zero
		{"p-0", "padding", "calc(var(--spacing) * 0)"},
		// Fractional spacing
		{"p-0.5", "padding", "calc(var(--spacing) * 0.5)"},
		{"p-1", "padding", "calc(var(--spacing) * 1)"},
		{"p-1.5", "padding", "calc(var(--spacing) * 1.5)"},
		{"p-2", "padding", "calc(var(--spacing) * 2)"},
		{"p-2.5", "padding", "calc(var(--spacing) * 2.5)"},
		{"p-3", "padding", "calc(var(--spacing) * 3)"},
		{"p-3.5", "padding", "calc(var(--spacing) * 3.5)"},
		{"p-4", "padding", "calc(var(--spacing) * 4)"},
		{"p-5", "padding", "calc(var(--spacing) * 5)"},
		{"p-6", "padding", "calc(var(--spacing) * 6)"},
		{"p-7", "padding", "calc(var(--spacing) * 7)"},
		{"p-8", "padding", "calc(var(--spacing) * 8)"},
		{"p-9", "padding", "calc(var(--spacing) * 9)"},
		{"p-10", "padding", "calc(var(--spacing) * 10)"},
		{"p-11", "padding", "calc(var(--spacing) * 11)"},
		{"p-12", "padding", "calc(var(--spacing) * 12)"},
		{"p-14", "padding", "calc(var(--spacing) * 14)"},
		{"p-16", "padding", "calc(var(--spacing) * 16)"},
		{"p-20", "padding", "calc(var(--spacing) * 20)"},
		{"p-24", "padding", "calc(var(--spacing) * 24)"},
		{"p-28", "padding", "calc(var(--spacing) * 28)"},
		{"p-32", "padding", "calc(var(--spacing) * 32)"},
		{"p-36", "padding", "calc(var(--spacing) * 36)"},
		{"p-40", "padding", "calc(var(--spacing) * 40)"},
		{"p-44", "padding", "calc(var(--spacing) * 44)"},
		{"p-48", "padding", "calc(var(--spacing) * 48)"},
		{"p-52", "padding", "calc(var(--spacing) * 52)"},
		{"p-56", "padding", "calc(var(--spacing) * 56)"},
		{"p-60", "padding", "calc(var(--spacing) * 60)"},
		{"p-64", "padding", "calc(var(--spacing) * 64)"},
		{"p-72", "padding", "calc(var(--spacing) * 72)"},
		{"p-80", "padding", "calc(var(--spacing) * 80)"},
		{"p-96", "padding", "calc(var(--spacing) * 96)"},
		// px keyword → 1px
		{"p-px", "padding", "1px"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}

// TestLargeComputedSpacing verifies large spacing multipliers work correctly.
func TestLargeComputedSpacing(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		{"m-128", "margin", "calc(var(--spacing) * 128)"},
		{"m-256", "margin", "calc(var(--spacing) * 256)"},
		{"gap-100", "gap", "calc(var(--spacing) * 100)"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}

// TestFractionalSpacing verifies fractional spacing values (decimal points).
func TestFractionalSpacing(t *testing.T) {
	e := New()
	cases := []struct {
		class string
		value string
	}{
		{"p-1.5", "calc(var(--spacing) * 1.5)"},
		{"m-0.5", "calc(var(--spacing) * 0.5)"},
		{"gap-2.5", "calc(var(--spacing) * 2.5)"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}

// TestWidthFractions verifies all common fraction denominators for width.
func TestWidthFractions(t *testing.T) {
	e := New()
	cases := []struct {
		class string
		value string
	}{
		{"w-1/2", "calc(1 / 2 * 100%)"},
		{"w-1/3", "calc(1 / 3 * 100%)"},
		{"w-2/3", "calc(2 / 3 * 100%)"},
		{"w-1/4", "calc(1 / 4 * 100%)"},
		{"w-2/4", "calc(2 / 4 * 100%)"},
		{"w-3/4", "calc(3 / 4 * 100%)"},
		{"w-1/5", "calc(1 / 5 * 100%)"},
		{"w-2/5", "calc(2 / 5 * 100%)"},
		{"w-3/5", "calc(3 / 5 * 100%)"},
		{"w-4/5", "calc(4 / 5 * 100%)"},
		{"w-1/6", "calc(1 / 6 * 100%)"},
		{"w-2/6", "calc(2 / 6 * 100%)"},
		{"w-3/6", "calc(3 / 6 * 100%)"},
		{"w-4/6", "calc(4 / 6 * 100%)"},
		{"w-5/6", "calc(5 / 6 * 100%)"},
		{"w-1/12", "calc(1 / 12 * 100%)"},
		{"w-2/12", "calc(2 / 12 * 100%)"},
		{"w-3/12", "calc(3 / 12 * 100%)"},
		{"w-4/12", "calc(4 / 12 * 100%)"},
		{"w-5/12", "calc(5 / 12 * 100%)"},
		{"w-6/12", "calc(6 / 12 * 100%)"},
		{"w-7/12", "calc(7 / 12 * 100%)"},
		{"w-8/12", "calc(8 / 12 * 100%)"},
		{"w-9/12", "calc(9 / 12 * 100%)"},
		{"w-10/12", "calc(10 / 12 * 100%)"},
		{"w-11/12", "calc(11 / 12 * 100%)"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}

// TestWidthKeywords verifies keyword values for width utilities.
func TestWidthKeywords(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		{"w-auto", "width", "auto"},
		{"w-full", "width", "100%"},
		{"w-screen", "width", "100vw"},
		{"w-svw", "width", "100svw"},
		{"w-lvw", "width", "100lvw"},
		{"w-dvw", "width", "100dvw"},
		{"w-min", "width", "min-content"},
		{"w-max", "width", "max-content"},
		{"w-fit", "width", "fit-content"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.property+": "+tc.value) {
				t.Errorf("%s: expected %q: %q in CSS output:\n%s", tc.class, tc.property, tc.value, css)
			}
		})
	}
}

// TestHeightKeywords verifies keyword values for height utilities.
func TestHeightKeywords(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		{"h-auto", "height", "auto"},
		{"h-full", "height", "100%"},
		{"h-screen", "height", "100vh"},
		{"h-svh", "height", "100svh"},
		{"h-lvh", "height", "100lvh"},
		{"h-dvh", "height", "100dvh"},
		{"h-min", "height", "min-content"},
		{"h-max", "height", "max-content"},
		{"h-fit", "height", "fit-content"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.property+": "+tc.value) {
				t.Errorf("%s: expected %q: %q in CSS output:\n%s", tc.class, tc.property, tc.value, css)
			}
		})
	}
}

// TestSizeUtility verifies that size-* sets both width and height.
func TestSizeUtility(t *testing.T) {
	cases := []struct {
		class string
		width string
		height string
	}{
		{"size-0", "calc(var(--spacing) * 0)", "calc(var(--spacing) * 0)"},
		{"size-4", "calc(var(--spacing) * 4)", "calc(var(--spacing) * 4)"},
		{"size-8", "calc(var(--spacing) * 8)", "calc(var(--spacing) * 8)"},
		{"size-full", "100%", "100%"},
		{"size-min", "min-content", "min-content"},
		{"size-max", "max-content", "max-content"},
		{"size-fit", "fit-content", "fit-content"},
		{"size-auto", "auto", "auto"},
		{"size-1/2", "calc(1 / 2 * 100%)", "calc(1 / 2 * 100%)"},
	}

	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tc.class + `"`))
			css := e.CSS()
			if !strings.Contains(css, "width: "+tc.width) {
				t.Errorf("%s: expected width: %q in CSS output:\n%s", tc.class, tc.width, css)
			}
			if !strings.Contains(css, "height: "+tc.height) {
				t.Errorf("%s: expected height: %q in CSS output:\n%s", tc.class, tc.height, css)
			}
		})
	}
}

// TestMinMaxSizingEdgeCases verifies min/max width and height edge cases.
func TestMinMaxSizingEdgeCases(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		{"min-w-0", "min-width", "calc(var(--spacing) * 0)"},
		{"min-w-full", "min-width", "100%"},
		{"min-w-min", "min-width", "min-content"},
		{"min-w-max", "min-width", "max-content"},
		{"min-w-fit", "min-width", "fit-content"},
		{"max-w-none", "max-width", "none"},
		{"max-w-full", "max-width", "100%"},
		{"max-w-min", "max-width", "min-content"},
		{"max-w-max", "max-width", "max-content"},
		{"max-w-fit", "max-width", "fit-content"},
		{"max-w-prose", "max-width", "65ch"},
		{"max-w-screen", "max-width", "100vw"},
		{"min-h-0", "min-height", "calc(var(--spacing) * 0)"},
		{"min-h-full", "min-height", "100%"},
		{"min-h-screen", "min-height", "100vh"},
		{"min-h-svh", "min-height", "100svh"},
		{"max-h-none", "max-height", "none"},
		{"max-h-full", "max-height", "100%"},
		{"max-h-screen", "max-height", "100vh"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.property+": "+tc.value) {
				t.Errorf("%s: expected %q: %q in CSS output:\n%s", tc.class, tc.property, tc.value, css)
			}
		})
	}
}

// TestNegativeMargins verifies negative margin values comprehensively.
func TestNegativeMargins(t *testing.T) {
	cases := []struct {
		class    string
		property string
		value    string
	}{
		// -m-0 negating zero: calc(var(--spacing) * 0) negated → calc(var(--spacing) * -0)
		{"-m-0", "margin", "calc(var(--spacing) * -0)"},
		{"-m-0.5", "margin", "calc(var(--spacing) * -0.5)"},
		{"-m-1", "margin", "calc(var(--spacing) * -1)"},
		{"-m-2", "margin", "calc(var(--spacing) * -2)"},
		{"-m-4", "margin", "calc(var(--spacing) * -4)"},
		{"-m-8", "margin", "calc(var(--spacing) * -8)"},
		{"-m-px", "margin", "calc(1px * -1)"},
		{"-mx-4", "margin-left", "calc(var(--spacing) * -4)"},
		{"-my-2", "margin-top", "calc(var(--spacing) * -2)"},
		{"-mt-8", "margin-top", "calc(var(--spacing) * -8)"},
	}

	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tc.class + `"`))
			css := e.CSS()
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}

// TestNegativeMarginAutoNoOutput verifies that -ml-auto produces no output.
func TestNegativeMarginAutoNoOutput(t *testing.T) {
	e := New()
	e.Write([]byte(`class="-ml-auto"`))
	css := e.CSS()
	if css != "" {
		t.Errorf("-ml-auto should produce no CSS output, got:\n%s", css)
	}
}

// TestInsetPositioning verifies inset (positioning) utilities.
func TestInsetPositioning(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		{"inset-0", "inset", "calc(var(--spacing) * 0)"},
		{"inset-auto", "inset", "auto"},
		{"inset-1/2", "inset", "calc(1 / 2 * 100%)"},
		{"inset-full", "inset", "100%"},
		{"top-0", "top", "calc(var(--spacing) * 0)"},
		{"top-auto", "top", "auto"},
		{"top-1/2", "top", "calc(1 / 2 * 100%)"},
		{"top-full", "top", "100%"},
		{"right-0", "right", "calc(var(--spacing) * 0)"},
		{"bottom-0", "bottom", "calc(var(--spacing) * 0)"},
		{"left-0", "left", "calc(var(--spacing) * 0)"},
		{"-top-4", "top", "calc(var(--spacing) * -4)"},
		{"-left-1/2", "left", "calc(calc(1 / 2 * 100%) * -1)"},
		{"start-0", "inset-inline-start", "calc(var(--spacing) * 0)"},
		{"end-0", "inset-inline-end", "calc(var(--spacing) * 0)"},
		{"start-auto", "inset-inline-start", "auto"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.property+": "+tc.value) {
				t.Errorf("%s: expected %q: %q in CSS output:\n%s", tc.class, tc.property, tc.value, css)
			}
		})
	}
}

// TestInsetXY verifies inset-x and inset-y set both sides.
func TestInsetXY(t *testing.T) {
	e := New()
	e.Write([]byte(`class="inset-x-0 inset-y-0"`))
	css := e.CSS()

	// inset-x-0 sets left and right
	if !strings.Contains(css, "left:") || !strings.Contains(css, "right:") {
		t.Errorf("inset-x-0 should set left and right:\n%s", css)
	}
	// inset-y-0 sets top and bottom
	if !strings.Contains(css, "top:") || !strings.Contains(css, "bottom:") {
		t.Errorf("inset-y-0 should set top and bottom:\n%s", css)
	}
}

// TestGapUtilities verifies gap, gap-x, and gap-y utilities.
func TestGapUtilities(t *testing.T) {
	e := New()
	cases := []struct {
		class    string
		property string
		value    string
	}{
		{"gap-0", "gap", "calc(var(--spacing) * 0)"},
		{"gap-1", "gap", "calc(var(--spacing) * 1)"},
		{"gap-4", "gap", "calc(var(--spacing) * 4)"},
		{"gap-8", "gap", "calc(var(--spacing) * 8)"},
		{"gap-px", "gap", "1px"},
		{"gap-0.5", "gap", "calc(var(--spacing) * 0.5)"},
		{"gap-x-4", "column-gap", "calc(var(--spacing) * 4)"},
		{"gap-y-8", "row-gap", "calc(var(--spacing) * 8)"},
	}

	for _, tc := range cases {
		e.Write([]byte(`class="` + tc.class + `"`))
	}
	css := e.CSS()
	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.property+": "+tc.value) {
				t.Errorf("%s: expected %q: %q in CSS output:\n%s", tc.class, tc.property, tc.value, css)
			}
		})
	}
}

// TestColumnsUtility verifies columns utility with integers and named widths.
func TestColumnsUtility(t *testing.T) {
	cases := []struct {
		class string
		value string
	}{
		{"columns-1", "columns: 1"},
		{"columns-2", "columns: 2"},
		{"columns-3", "columns: 3"},
		{"columns-4", "columns: 4"},
		{"columns-auto", "columns: auto"},
		{"columns-3xs", "columns: 16rem"},
		{"columns-2xs", "columns: 18rem"},
		{"columns-xs", "columns: 20rem"},
		{"columns-sm", "columns: 24rem"},
		{"columns-md", "columns: 28rem"},
		{"columns-lg", "columns: 32rem"},
		{"columns-xl", "columns: 36rem"},
		{"columns-2xl", "columns: 42rem"},
	}

	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tc.class + `"`))
			css := e.CSS()
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}

// TestBasisUtility verifies flex-basis utilities.
func TestBasisUtility(t *testing.T) {
	cases := []struct {
		class string
		value string
	}{
		{"basis-0", "flex-basis: calc(var(--spacing) * 0)"},
		{"basis-1", "flex-basis: calc(var(--spacing) * 1)"},
		{"basis-1/2", "flex-basis: calc(1 / 2 * 100%)"},
		{"basis-1/3", "flex-basis: calc(1 / 3 * 100%)"},
		{"basis-full", "flex-basis: 100%"},
		{"basis-auto", "flex-basis: auto"},
		{"basis-[300px]", "flex-basis: 300px"},
	}

	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tc.class + `"`))
			css := e.CSS()
			if !strings.Contains(css, tc.value) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.value, css)
			}
		})
	}
}
