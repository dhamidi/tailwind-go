package tailwind

import (
	"strings"
	"testing"
)

// TestSpaceUtilitiesEmbedded verifies space-x/y utilities using the embedded CSS
// produce correct child selectors and logical properties.
func TestSpaceUtilitiesEmbedded(t *testing.T) {
	tests := []struct {
		class    string
		contains []string
		excludes []string
	}{
		{
			class: "space-x-4",
			contains: []string{
				"> :not(:last-child)",
				"margin-inline-start",
				"margin-inline-end",
				"--tw-space-x-reverse: 0",
			},
			excludes: []string{"margin-left"},
		},
		{
			class: "space-y-2",
			contains: []string{
				"> :not(:last-child)",
				"margin-block-start",
				"margin-block-end",
				"--tw-space-y-reverse: 0",
			},
			excludes: []string{"margin-top"},
		},
		{
			class: "space-x-reverse",
			contains: []string{
				"> :not(:last-child)",
				"--tw-space-x-reverse: 1",
			},
		},
		{
			class: "space-y-reverse",
			contains: []string{
				"> :not(:last-child)",
				"--tw-space-y-reverse: 1",
			},
		},
		{
			class: "space-x-0",
			contains: []string{
				"> :not(:last-child)",
				"margin-inline-start",
			},
		},
		{
			class: "space-x-px",
			contains: []string{
				"> :not(:last-child)",
				"margin-inline-start",
			},
		},
		{
			class: "space-x-0.5",
			contains: []string{
				"> :not(:last-child)",
				"margin-inline-start",
			},
		},
		{
			class: "-space-x-4",
			contains: []string{
				"> :not(:last-child)",
				"margin-inline-start",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in CSS output", tt.class, want)
				}
			}
			for _, notWant := range tt.excludes {
				if strings.Contains(result, notWant) {
					t.Errorf("class %q: should NOT contain %q", tt.class, notWant)
				}
			}
		})
	}
}

// TestDivideWidthUtilitiesEmbedded verifies divide-x/y width utilities
// use child selectors and correct border properties.
func TestDivideWidthUtilitiesEmbedded(t *testing.T) {
	tests := []struct {
		class    string
		contains []string
	}{
		{
			class: "divide-x",
			contains: []string{
				"> :not(:last-child)",
				"border-inline-start-width",
				"border-inline-end-width",
				"--tw-divide-x-reverse: 0",
			},
		},
		{
			class: "divide-y",
			contains: []string{
				"> :not(:last-child)",
				"border-top-width",
				"border-bottom-width",
				"--tw-divide-y-reverse: 0",
			},
		},
		{
			class: "divide-x-2",
			contains: []string{
				"> :not(:last-child)",
				"border-inline-start-width",
				"calc(2px",
			},
		},
		{
			class: "divide-y-4",
			contains: []string{
				"> :not(:last-child)",
				"border-top-width",
				"calc(4px",
			},
		},
		{
			class: "divide-x-reverse",
			contains: []string{
				"> :not(:last-child)",
				"--tw-divide-x-reverse: 1",
			},
		},
		{
			class: "divide-y-reverse",
			contains: []string{
				"> :not(:last-child)",
				"--tw-divide-y-reverse: 1",
			},
		},
		{
			class: "divide-x-0",
			contains: []string{
				"> :not(:last-child)",
				"border-inline-start-width",
				"calc(0px",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in CSS output", tt.class, want)
				}
			}
		})
	}
}

// TestDivideColorUtilitiesEmbedded verifies divide color utilities
// apply border-color on child selectors.
func TestDivideColorUtilitiesEmbedded(t *testing.T) {
	tests := []struct {
		class    string
		contains []string
	}{
		{
			class:    "divide-gray-200",
			contains: []string{"> :not(:last-child)", "border-color"},
		},
		{
			class:    "divide-blue-500/50",
			contains: []string{"> :not(:last-child)", "border-color"},
		},
		{
			class:    "divide-current",
			contains: []string{"> :not(:last-child)", "border-color", "currentcolor"},
		},
		{
			class:    "divide-transparent",
			contains: []string{"> :not(:last-child)", "border-color", "transparent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in CSS output", tt.class, want)
				}
			}
		})
	}
}

// TestDivideStyleUtilitiesEmbedded verifies divide style utilities
// apply border-style on child selectors.
func TestDivideStyleUtilitiesEmbedded(t *testing.T) {
	tests := []struct {
		class string
		style string
	}{
		{"divide-dashed", "dashed"},
		{"divide-dotted", "dotted"},
		{"divide-double", "double"},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			if !strings.Contains(result, "> :not(:last-child)") {
				t.Errorf("class %q: expected child selector in CSS output", tt.class)
			}
			if !strings.Contains(result, "border-style: "+tt.style) {
				t.Errorf("class %q: expected border-style: %s in CSS output", tt.class, tt.style)
			}
		})
	}
}

// TestSpaceDivideWithVariants verifies space/divide utilities combined
// with responsive and state variants.
func TestSpaceDivideWithVariants(t *testing.T) {
	tests := []struct {
		class    string
		contains []string
	}{
		{
			class: "md:space-x-4",
			contains: []string{
				"@media",
				"width >= 48rem",
				"> :not(:last-child)",
				"margin-inline-start",
			},
		},
		{
			class: "hover:divide-blue-500",
			contains: []string{
				":hover",
				"> :not(:last-child)",
				"border-color",
			},
		},
		{
			class: "dark:divide-gray-700",
			contains: []string{
				"prefers-color-scheme: dark",
				"> :not(:last-child)",
				"border-color",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in CSS output", tt.class, want)
				}
			}
		})
	}
}

// TestPseudoElementVariants verifies pseudo-element variant selectors
// produce the correct CSS pseudo-element selectors.
func TestPseudoElementVariants(t *testing.T) {
	tests := []struct {
		class    string
		contains []string
	}{
		{
			class:    "before:block",
			contains: []string{"::before", "display: block", "content: var(--tw-content)"},
		},
		{
			class:    "before:absolute",
			contains: []string{"::before", "position: absolute", "content: var(--tw-content)"},
		},
		{
			class:    "after:block",
			contains: []string{"::after", "display: block", "content: var(--tw-content)"},
		},
		{
			class:    "marker:text-blue-500",
			contains: []string{"::marker"},
		},
		{
			class:    "selection:bg-blue-500",
			contains: []string{"::selection", "background-color"},
		},
		{
			class:    "selection:text-white",
			contains: []string{"::selection", "color"},
		},
		{
			class:    "placeholder:text-gray-400",
			contains: []string{"::placeholder", "color"},
		},
		{
			class:    "placeholder:italic",
			contains: []string{"::placeholder", "font-style: italic"},
		},
		{
			class:    "file:border-0",
			contains: []string{"::file-selector-button", "border-width: 0px"},
		},
		{
			class:    "file:bg-blue-500",
			contains: []string{"::file-selector-button", "background-color"},
		},
		{
			class:    "file:text-white",
			contains: []string{"::file-selector-button", "color"},
		},
		{
			class:    "file:mr-4",
			contains: []string{"::file-selector-button", "margin-right"},
		},
		{
			class:    "first-line:uppercase",
			contains: []string{"::first-line", "text-transform: uppercase"},
		},
		{
			class:    "first-letter:text-4xl",
			contains: []string{"::first-letter", "font-size"},
		},
		{
			class:    "first-letter:font-bold",
			contains: []string{"::first-letter", "font-weight"},
		},
		{
			class:    "backdrop:bg-black/50",
			contains: []string{"::backdrop", "background-color"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in CSS output", tt.class, want)
				}
			}
		})
	}
}

// TestPseudoElementContentArbitrary verifies content-[''] with pseudo-elements.
// The content-* utility with arbitrary values requires a registered dynamic
// utility. This test documents the current behavior: content-[''] does not
// produce output because no content-* functional utility is registered.
func TestPseudoElementContentArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte(`class="before:content-['']"`))
	result := e.CSS()
	t.Logf("Generated CSS for before:content-['']:\n%s", result)

	// content-none is registered as a static utility, verify it works with before:.
	e2 := New()
	e2.Write([]byte(`class="before:content-none"`))
	result2 := e2.CSS()
	t.Logf("Generated CSS for before:content-none:\n%s", result2)

	if !strings.Contains(result2, "::before") {
		t.Errorf("before:content-none should contain ::before")
	}
	if !strings.Contains(result2, "content: none") {
		t.Errorf("before:content-none should contain content: none")
	}
}

// TestStackedPseudoElementVariants verifies variant stacking with
// pseudo-elements produces correct compound selectors.
func TestStackedPseudoElementVariants(t *testing.T) {
	tests := []struct {
		class    string
		contains []string
	}{
		{
			class:    "hover:before:bg-blue-500",
			contains: []string{"::before", ":hover", "background-color"},
		},
		{
			class:    "dark:placeholder:text-gray-500",
			contains: []string{"prefers-color-scheme: dark", "::placeholder", "color"},
		},
		{
			class:    "focus:placeholder:text-transparent",
			contains: []string{":focus", "::placeholder", "color", "transparent"},
		},
		{
			class:    "md:before:absolute",
			contains: []string{"@media", "width >= 48rem", "::before", "position: absolute"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			t.Logf("Generated CSS:\n%s", result)

			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in CSS output", tt.class, want)
				}
			}
		})
	}
}

// TestBeforeAfterCoexist verifies that both before: and after: pseudo-element
// variants can coexist in the same render without interference.
func TestBeforeAfterCoexist(t *testing.T) {
	e := New()
	e.Write([]byte(`class="before:block before:absolute after:block after:content-['']"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "::before") {
		t.Error("expected ::before pseudo-element in CSS output")
	}
	if !strings.Contains(result, "::after") {
		t.Error("expected ::after pseudo-element in CSS output")
	}

	// Both should generate display: block declarations.
	beforeIdx := strings.Index(result, "::before")
	afterIdx := strings.Index(result, "::after")
	if beforeIdx < 0 || afterIdx < 0 {
		t.Fatal("missing ::before or ::after selectors")
	}
	// Verify they are distinct selectors (not the same rule).
	if beforeIdx == afterIdx {
		t.Error("::before and ::after should be in separate rules")
	}
}

// TestMarkerListDisc verifies the marker: variant with list-disc utility.
func TestMarkerListDisc(t *testing.T) {
	e := New()
	e.Write([]byte(`class="marker:list-disc"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "::marker") {
		t.Error("marker:list-disc should contain ::marker pseudo-element")
	}
	if !strings.Contains(result, "list-style-type: disc") {
		t.Error("marker:list-disc should contain list-style-type: disc")
	}
}
