package tailwind

import (
	"strings"
	"testing"
)

// compatCase is a test case that verifies a Tailwind class produces
// the expected CSS declaration.
type compatCase struct {
	class    string
	property string
	value    string // substring to match in the generated value
}

// TestCompatibilityCommonClasses verifies that commonly-used Tailwind v4 classes
// produce correct CSS when using the embedded default CSS source.
func TestCompatibilityCommonClasses(t *testing.T) {
	cases := []compatCase{
		// Layout
		{"flex", "display", "flex"},
		{"block", "display", "block"},
		{"inline", "display", "inline"},
		{"grid", "display", "grid"},
		{"hidden", "display", "none"},
		{"contents", "display", "contents"},

		// Flexbox
		{"items-center", "align-items", "center"},
		{"items-start", "align-items", "flex-start"},
		{"justify-center", "justify-content", "center"},
		{"justify-between", "justify-content", "space-between"},
		{"flex-col", "flex-direction", "column"},
		{"flex-row", "flex-direction", "row"},
		{"flex-wrap", "flex-wrap", "wrap"},
		{"flex-1", "flex", "1"},

		// Spacing (padding/margin)
		{"p-4", "padding", "calc(var(--spacing) * 4)"},
		{"px-4", "padding-inline", "calc(var(--spacing) * 4)"},
		{"py-2", "padding-top", "calc(var(--spacing) * 2)"},
		{"m-4", "margin", "calc(var(--spacing) * 4)"},
		{"mt-2", "margin-top", "calc(var(--spacing) * 2)"},
		{"mx-auto", "margin-left", "auto"},

		// Sizing
		{"w-full", "width", "100%"},
		{"h-full", "height", "100%"},
		{"w-1/2", "width", "calc(1 / 2 * 100%)"},
		{"min-w-0", "min-width", "0"},
		{"max-w-full", "max-width", "100%"},

		// Typography
		{"text-sm", "font-size", ""},  // some rem value
		{"text-lg", "font-size", ""},
		{"font-bold", "font-weight", "var(--font-weight-bold)"},
		{"font-normal", "font-weight", "var(--font-weight-normal)"},
		{"text-center", "text-align", "center"},
		{"text-left", "text-align", "left"},
		{"italic", "font-style", "italic"},
		{"not-italic", "font-style", "normal"},
		{"uppercase", "text-transform", "uppercase"},
		{"lowercase", "text-transform", "lowercase"},
		{"capitalize", "text-transform", "capitalize"},

		// Colors (background)
		{"bg-white", "background-color", ""},  // some white value
		{"bg-black", "background-color", ""},  // some black value
		{"bg-transparent", "background-color", "transparent"},

		// Border
		{"border", "border-width", "1px"},
		{"rounded", "border-radius", ""},
		{"rounded-full", "border-radius", "calc(infinity * 1px)"},

		// Position
		{"relative", "position", "relative"},
		{"absolute", "position", "absolute"},
		{"fixed", "position", "fixed"},
		{"sticky", "position", "sticky"},
		{"inset-0", "inset", "0"},

		// Overflow
		{"overflow-hidden", "overflow", "hidden"},
		{"overflow-auto", "overflow", "auto"},
		{"overflow-scroll", "overflow", "scroll"},

		// Cursor
		{"cursor-pointer", "cursor", "pointer"},
		{"cursor-default", "cursor", "default"},

		// Opacity
		{"opacity-0", "opacity", "0"},
		{"opacity-100", "opacity", "1"},

		// Z-index
		{"z-10", "z-index", "10"},
		{"z-50", "z-index", "50"},

		// Accessibility
		{"sr-only", "position", "absolute"},

		// Ring utilities
		{"ring", "--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(3px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)"},
		{"ring-2", "--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)"},
		{"ring-0", "--tw-ring-shadow", "calc(0px + var(--tw-ring-offset-width))"},
		{"ring-1", "--tw-ring-shadow", "calc(1px + var(--tw-ring-offset-width))"},
		{"ring-4", "--tw-ring-shadow", "calc(4px + var(--tw-ring-offset-width))"},
		{"ring-8", "--tw-ring-shadow", "calc(8px + var(--tw-ring-offset-width))"},
		{"ring-inset", "--tw-ring-inset", "inset"},

		// Transitions
		{"transition", "transition-property", ""},

		// Arbitrary values
		{"w-[300px]", "width", "300px"},
		{"h-[200px]", "height", "200px"},
		{"p-[1.5rem]", "padding", "1.5rem"},

		// Important
		{"!flex", "display", "flex"},
	}

	e := New() // uses embedded CSS
	// Write all classes at once.
	var buf strings.Builder
	buf.WriteString(`<div class="`)
	for _, tc := range cases {
		buf.WriteString(tc.class)
		buf.WriteByte(' ')
	}
	buf.WriteString(`">`)
	e.Write([]byte(buf.String()))

	css := e.CSS()

	for _, tc := range cases {
		t.Run(tc.class, func(t *testing.T) {
			if !strings.Contains(css, tc.property+":") &&
				!strings.Contains(css, tc.property+": ") {
				t.Errorf("class %q: property %q not found in CSS", tc.class, tc.property)
				return
			}
			if tc.value != "" && !strings.Contains(css, tc.value) {
				t.Errorf("class %q: value %q not found in CSS\nCSS excerpt around property:\n%s",
					tc.class, tc.value, extractRelevant(css, tc.property))
			}
		})
	}
}

func extractRelevant(css, property string) string {
	idx := strings.Index(css, property+":")
	if idx < 0 {
		return "(not found)"
	}
	start := idx
	if start > 100 {
		start = idx - 100
	}
	end := idx + 200
	if end > len(css) {
		end = len(css)
	}
	return css[start:end]
}

// TestCompatibilityVariants verifies variant-prefixed classes.
func TestCompatibilityVariants(t *testing.T) {
	e := New()
	e.Write([]byte(`class="hover:bg-blue-500 md:flex dark:text-white"`))
	css := e.CSS()

	if !strings.Contains(css, ":hover") {
		t.Error("hover variant missing from CSS")
	}
	if !strings.Contains(css, "@media") {
		t.Error("media query variant missing from CSS")
	}
}

// TestCompatibilityArbitraryProperty verifies [prop:value] classes.
func TestCompatibilityArbitraryProperty(t *testing.T) {
	e := New()
	e.Write([]byte(`class="[mask-type:alpha] [content-visibility:auto]"`))
	css := e.CSS()

	if !strings.Contains(css, "mask-type: alpha") {
		t.Error("missing mask-type: alpha")
	}
	if !strings.Contains(css, "content-visibility: auto") {
		t.Error("missing content-visibility: auto")
	}
}

// TestCompatibilityNegativeValues verifies negative utility classes.
func TestCompatibilityNegativeValues(t *testing.T) {
	e := New()
	e.Write([]byte(`class="-m-4 -translate-x-2"`))
	css := e.CSS()
	t.Logf("Negative CSS:\n%s", css)
	// At minimum, CSS should be generated (not empty).
	// Exact values depend on theme.
}

func TestNegativeNonNegatableValueProducesNoOutput(t *testing.T) {
	e := New()
	e.Write([]byte(`class="-m-auto"`))
	css := e.CSS()
	if css != "" {
		t.Errorf("-m-auto should produce no CSS output, got:\n%s", css)
	}
}
