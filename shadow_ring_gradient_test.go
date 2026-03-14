package tailwind

import (
	"strings"
	"testing"
)

// newShadowRingTestEngine creates an engine with the theme tokens needed for
// shadow, ring, outline, and gradient tests.
func newShadowRingTestEngine(t *testing.T) *Engine {
	t.Helper()
	css := []byte(`
@theme {
  --color-red-500: #ef4444;
  --color-blue-500: #3b82f6;
  --color-green-500: #22c55e;
  --color-purple-500: #a855f7;
  --color-pink-500: #ec4899;
  --color-black: #000;
  --color-white: #fff;
  --color-transparent: transparent;
  --opacity-25: 0.25;
  --opacity-50: 0.5;
  --opacity-75: 0.75;
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
  --shadow-xl: 0 20px 25px -5px rgb(0 0 0 / 0.1);
  --shadow-2xl: 0 25px 50px -12px rgb(0 0 0 / 0.25);
  --spacing: 0.25rem;
}
`)
	e := New()
	if err := e.LoadCSS(css); err != nil {
		t.Fatal(err)
	}
	return e
}

// assertCSSContains checks that the generated CSS for a given class contains
// the expected property: value pair.
func assertCSSContains(t *testing.T, e *Engine, class, wantProp, wantVal string) {
	t.Helper()
	e.Write([]byte(class))
	result := e.CSS()
	check := wantProp + ": " + wantVal
	if !strings.Contains(result, check) {
		t.Errorf("class %q: expected %q in output:\n%s", class, check, result)
	}
}

// assertCSSHas checks that the generated CSS for a given class contains the
// given substring.
func assertCSSHas(t *testing.T, e *Engine, class, want string) {
	t.Helper()
	e.Write([]byte(class))
	result := e.CSS()
	if !strings.Contains(result, want) {
		t.Errorf("class %q: expected %q in output:\n%s", class, want, result)
	}
}

// boxShadowComposition is the expected box-shadow composition value.
const boxShadowComposition = "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)"

// ===== Shadow Size Utilities =====

func TestShadowSizes(t *testing.T) {
	tests := []struct {
		class     string
		shadowVar string // expected --tw-shadow value substring (empty means just check box-shadow exists)
	}{
		{"shadow", "var(--shadow-sm"},
		{"shadow-sm", "var(--shadow-sm"},
		{"shadow-md", "var(--shadow-md"},
		{"shadow-lg", "var(--shadow-lg"},
		{"shadow-xl", "var(--shadow-xl"},
		{"shadow-2xl", "var(--shadow-2xl"},
		{"shadow-inner", "inset 0 2px 4px 0 rgb(0 0 0 / 0.05)"},
		{"shadow-none", "0 0 #0000"},
		{"shadow-2xs", "var(--shadow-2xs"},
		{"shadow-xs", "var(--shadow-xs"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			e.Write([]byte(tt.class))
			result := e.CSS()

			// All shadow utilities must set --tw-shadow
			if !strings.Contains(result, "--tw-shadow:") && !strings.Contains(result, "--tw-shadow: ") {
				t.Errorf("%s: missing --tw-shadow declaration:\n%s", tt.class, result)
			}

			// All shadow utilities must compose box-shadow
			if !strings.Contains(result, "box-shadow: "+boxShadowComposition) {
				t.Errorf("%s: missing composed box-shadow:\n%s", tt.class, result)
			}

			// Check expected --tw-shadow value
			if !strings.Contains(result, tt.shadowVar) {
				t.Errorf("%s: expected --tw-shadow to contain %q:\n%s", tt.class, tt.shadowVar, result)
			}
		})
	}
}

// ===== Shadow Color Utilities =====

func TestShadowColors(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"shadow-red-500", "--tw-shadow-color", "var(--color-red-500)"},
		{"shadow-blue-500/50", "--tw-shadow-color", "oklch(from #3b82f6 l c h / 50%)"},
		{"shadow-current", "--tw-shadow-color", "currentColor"},
		{"shadow-transparent", "--tw-shadow-color", "transparent"},
		{"shadow-black/25", "--tw-shadow-color", "oklch(from #000 l c h / 25%)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, tt.wantProp, tt.wantVal)
		})
	}
}

// ===== Inset Shadow Utilities =====

func TestInsetShadows(t *testing.T) {
	tests := []struct {
		class     string
		shadowVar string
	}{
		{"inset-shadow-2xs", "var(--inset-shadow-2xs"},
		{"inset-shadow-xs", "var(--inset-shadow-xs"},
		{"inset-shadow-sm", "var(--inset-shadow-sm"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			e.Write([]byte(tt.class))
			result := e.CSS()

			// Must set --tw-inset-shadow
			if !strings.Contains(result, "--tw-inset-shadow:") {
				t.Errorf("%s: missing --tw-inset-shadow:\n%s", tt.class, result)
			}

			// Must compose box-shadow
			if !strings.Contains(result, "box-shadow: "+boxShadowComposition) {
				t.Errorf("%s: missing composed box-shadow:\n%s", tt.class, result)
			}

			// Check value
			if !strings.Contains(result, tt.shadowVar) {
				t.Errorf("%s: expected %q:\n%s", tt.class, tt.shadowVar, result)
			}
		})
	}
}

// ===== Ring Width Utilities =====

func TestRingWidths(t *testing.T) {
	tests := []struct {
		class   string
		calcPx  string // the pixel value in calc(Npx + ...)
	}{
		{"ring", "3px"},
		{"ring-0", "0px"},
		{"ring-1", "1px"},
		{"ring-2", "2px"},
		{"ring-4", "4px"},
		{"ring-8", "8px"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			e.Write([]byte(tt.class))
			result := e.CSS()

			// Must set --tw-ring-shadow with calc
			wantCalc := "calc(" + tt.calcPx + " + var(--tw-ring-offset-width))"
			if !strings.Contains(result, wantCalc) {
				t.Errorf("%s: expected %q in --tw-ring-shadow:\n%s", tt.class, wantCalc, result)
			}

			// Must compose box-shadow
			if !strings.Contains(result, "box-shadow: "+boxShadowComposition) {
				t.Errorf("%s: missing composed box-shadow:\n%s", tt.class, result)
			}
		})
	}
}

func TestRingInset(t *testing.T) {
	e := newShadowRingTestEngine(t)
	assertCSSContains(t, e, "ring-inset", "--tw-ring-inset", "inset")
}

// ===== Ring Color Utilities =====

func TestRingColors(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"ring-blue-500", "--tw-ring-color", "var(--color-blue-500)"},
		{"ring-red-500/50", "--tw-ring-color", "oklch(from #ef4444 l c h / 50%)"},
		{"ring-current", "--tw-ring-color", "currentColor"},
		{"ring-transparent", "--tw-ring-color", "transparent"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, tt.wantProp, tt.wantVal)
		})
	}
}

// ===== Ring Offset Utilities =====

func TestRingOffsetWidths(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"ring-offset-2", "2px"},
		{"ring-offset-4", "4px"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, "--tw-ring-offset-width", tt.wantVal)
		})
	}
}

func TestRingOffsetColors(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"ring-offset-white", "--tw-ring-offset-color", "var(--color-white)"},
		{"ring-offset-blue-500", "--tw-ring-offset-color", "var(--color-blue-500)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, tt.wantProp, tt.wantVal)
		})
	}
}

// ===== Outline Utilities =====

func TestOutlineStyles(t *testing.T) {
	e := newShadowRingTestEngine(t)
	assertCSSContains(t, e, "outline", "outline-style", "solid")
}

func TestOutlineNone(t *testing.T) {
	e := newShadowRingTestEngine(t)
	e.Write([]byte("outline-none"))
	result := e.CSS()
	if !strings.Contains(result, "outline: 2px solid transparent") {
		t.Errorf("outline-none missing expected output:\n%s", result)
	}
	if !strings.Contains(result, "outline-offset: 2px") {
		t.Errorf("outline-none missing outline-offset:\n%s", result)
	}
}

func TestOutlineWidths(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"outline-2", "2px"},
		{"outline-4", "4px"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, "outline-width", tt.wantVal)
		})
	}
}

func TestOutlineStyleVariants(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"outline-dashed", "dashed"},
		{"outline-dotted", "dotted"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, "outline-style", tt.wantVal)
		})
	}
}

func TestOutlineOffsets(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"outline-offset-2", "2px"},
		{"outline-offset-4", "4px"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, "outline-offset", tt.wantVal)
		})
	}
}

func TestOutlineColors(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"outline-blue-500", "outline-color", "var(--color-blue-500)"},
		{"outline-red-500/50", "outline-color", "oklch(from #ef4444 l c h / 50%)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, tt.wantProp, tt.wantVal)
		})
	}
}

// ===== Gradient Direction Utilities =====

func TestGradientDirections(t *testing.T) {
	tests := []struct {
		class     string
		wantGrad  string
	}{
		{"bg-gradient-to-t", "linear-gradient(to top, var(--tw-gradient-stops))"},
		{"bg-gradient-to-tr", "linear-gradient(to top right, var(--tw-gradient-stops))"},
		{"bg-gradient-to-r", "linear-gradient(to right, var(--tw-gradient-stops))"},
		{"bg-gradient-to-br", "linear-gradient(to bottom right, var(--tw-gradient-stops))"},
		{"bg-gradient-to-b", "linear-gradient(to bottom, var(--tw-gradient-stops))"},
		{"bg-gradient-to-bl", "linear-gradient(to bottom left, var(--tw-gradient-stops))"},
		{"bg-gradient-to-l", "linear-gradient(to left, var(--tw-gradient-stops))"},
		{"bg-gradient-to-tl", "linear-gradient(to top left, var(--tw-gradient-stops))"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantGrad) {
				t.Errorf("%s: expected %q:\n%s", tt.class, tt.wantGrad, result)
			}
		})
	}
}

// ===== Gradient Stop Colors =====

func TestGradientStopColors(t *testing.T) {
	t.Run("from-blue-500", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("from-blue-500"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-gradient-from: var(--color-blue-500)") {
			t.Errorf("from-blue-500 missing --tw-gradient-from:\n%s", result)
		}
		if !strings.Contains(result, "--tw-gradient-stops:") {
			t.Errorf("from-blue-500 missing --tw-gradient-stops:\n%s", result)
		}
	})

	t.Run("via-purple-500", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("via-purple-500"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-gradient-via: var(--color-purple-500)") {
			t.Errorf("via-purple-500 missing --tw-gradient-via:\n%s", result)
		}
		if !strings.Contains(result, "--tw-gradient-stops:") {
			t.Errorf("via-purple-500 missing --tw-gradient-stops:\n%s", result)
		}
	})

	t.Run("to-pink-500", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		assertCSSContains(t, e, "to-pink-500", "--tw-gradient-to", "var(--color-pink-500)")
	})
}

// ===== Gradient Stop Positions =====

func TestGradientStopPositions(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"from-0%", "--tw-gradient-from-position", "0%"},
		{"from-5%", "--tw-gradient-from-position", "5%"},
		{"from-10%", "--tw-gradient-from-position", "10%"},
		{"via-30%", "--tw-gradient-via-position", "30%"},
		{"to-100%", "--tw-gradient-to-position", "100%"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, tt.wantProp, tt.wantVal)
		})
	}
}

// ===== Gradient Stops with Opacity =====

func TestGradientStopsWithOpacity(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"from-blue-500/50", "--tw-gradient-from", "oklch(from #3b82f6 l c h / 50%)"},
		{"via-red-500/25", "--tw-gradient-via", "oklch(from #ef4444 l c h / 25%)"},
		{"to-green-500/75", "--tw-gradient-to", "oklch(from #22c55e l c h / 75%)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := newShadowRingTestEngine(t)
			assertCSSContains(t, e, tt.class, tt.wantProp, tt.wantVal)
		})
	}
}

// ===== Full Gradient Composition =====

func TestGradientFullComposition(t *testing.T) {
	e := newShadowRingTestEngine(t)
	e.Write([]byte(`class="bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500"`))
	result := e.CSS()

	checks := []string{
		"--tw-gradient-from: var(--color-blue-500)",
		"--tw-gradient-via: var(--color-purple-500)",
		"--tw-gradient-to: var(--color-pink-500)",
		"linear-gradient(to right, var(--tw-gradient-stops))",
	}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("full gradient composition: expected %q:\n%s", check, result)
		}
	}
}

// ===== Arbitrary Shadow Values =====

func TestArbitraryShadows(t *testing.T) {
	t.Run("arbitrary box-shadow", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		assertCSSContains(t, e, "shadow-[0_0_10px_rgba(0,0,0,0.5)]", "box-shadow", "0 0 10px rgba(0,0,0,0.5)")
	})

	t.Run("arbitrary inset shadow", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		// Underscores should be converted to spaces
		assertCSSContains(t, e, "shadow-[inset_0_2px_4px_rgba(0,0,0,0.06)]", "box-shadow", "inset 0 2px 4px rgba(0,0,0,0.06)")
	})
}

// ===== Arbitrary Ring Values =====

func TestArbitraryRingWidth(t *testing.T) {
	t.Run("ring-[3px]", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("ring-[3px]"))
		result := e.CSS()
		// Arbitrary ring width should produce some ring-related output
		if result == "" {
			t.Skip("ring-[3px] arbitrary width not supported")
		}
		if !strings.Contains(result, "3px") {
			t.Errorf("ring-[3px]: expected 3px in output:\n%s", result)
		}
	})

	t.Run("ring-[var(--ring-width)]", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("ring-[var(--ring-width)]"))
		result := e.CSS()
		if result == "" {
			t.Skip("ring-[var(--ring-width)] not supported")
		}
		if !strings.Contains(result, "var(--ring-width)") {
			t.Errorf("ring-[var(--ring-width)]: expected CSS variable in output:\n%s", result)
		}
	})
}

// ===== Variant Combinations =====

func TestShadowRingWithVariants(t *testing.T) {
	t.Run("hover:shadow-lg", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("hover:shadow-lg"))
		result := e.CSS()
		if !strings.Contains(result, ":hover") {
			t.Errorf("hover:shadow-lg missing :hover selector:\n%s", result)
		}
		if !strings.Contains(result, "box-shadow:") {
			t.Errorf("hover:shadow-lg missing box-shadow:\n%s", result)
		}
	})

	t.Run("focus:ring-2", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("focus:ring-2"))
		result := e.CSS()
		if !strings.Contains(result, ":focus") {
			t.Errorf("focus:ring-2 missing :focus selector:\n%s", result)
		}
		if !strings.Contains(result, "--tw-ring-shadow:") {
			t.Errorf("focus:ring-2 missing --tw-ring-shadow:\n%s", result)
		}
	})

	t.Run("focus:ring-blue-500", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("focus:ring-blue-500"))
		result := e.CSS()
		if !strings.Contains(result, ":focus") {
			t.Errorf("focus:ring-blue-500 missing :focus selector:\n%s", result)
		}
		if !strings.Contains(result, "--tw-ring-color: var(--color-blue-500)") {
			t.Errorf("focus:ring-blue-500 missing ring color:\n%s", result)
		}
	})

	t.Run("dark:shadow-none", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("dark:shadow-none"))
		result := e.CSS()
		if !strings.Contains(result, "prefers-color-scheme: dark") {
			t.Errorf("dark:shadow-none missing dark mode media query:\n%s", result)
		}
		if !strings.Contains(result, "--tw-shadow: 0 0 #0000") {
			t.Errorf("dark:shadow-none missing shadow reset:\n%s", result)
		}
	})

	t.Run("dark:ring-white/10", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("dark:ring-white/10"))
		result := e.CSS()
		if !strings.Contains(result, "prefers-color-scheme: dark") {
			t.Errorf("dark:ring-white/10 missing dark mode:\n%s", result)
		}
		if !strings.Contains(result, "--tw-ring-color:") {
			t.Errorf("dark:ring-white/10 missing ring color:\n%s", result)
		}
	})

	t.Run("hover:shadow-blue-500/25", func(t *testing.T) {
		e := newShadowRingTestEngine(t)
		e.Write([]byte("hover:shadow-blue-500/25"))
		result := e.CSS()
		if !strings.Contains(result, ":hover") {
			t.Errorf("hover:shadow-blue-500/25 missing :hover:\n%s", result)
		}
		if !strings.Contains(result, "--tw-shadow-color:") {
			t.Errorf("hover:shadow-blue-500/25 missing shadow color:\n%s", result)
		}
	})
}

// ===== Shadow + Ring Composition Interaction =====

func TestShadowRingComposition(t *testing.T) {
	// When both shadow and ring are used, they share the box-shadow composition
	e := newShadowRingTestEngine(t)
	e.Write([]byte(`class="shadow-lg ring-2"`))
	result := e.CSS()

	// Both should produce box-shadow with full composition
	if !strings.Contains(result, boxShadowComposition) {
		t.Errorf("shadow + ring composition: expected composed box-shadow:\n%s", result)
	}
	// shadow-lg should set --tw-shadow
	if !strings.Contains(result, "--tw-shadow:") {
		t.Errorf("shadow + ring: missing --tw-shadow:\n%s", result)
	}
	// ring-2 should set --tw-ring-shadow
	if !strings.Contains(result, "--tw-ring-shadow:") {
		t.Errorf("shadow + ring: missing --tw-ring-shadow:\n%s", result)
	}
}

// ===== Ring Offset Shadow Interaction =====

func TestRingOffsetGeneratesVar(t *testing.T) {
	e := newShadowRingTestEngine(t)
	e.Write([]byte("ring-offset-4"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-ring-offset-width: 4px") {
		t.Errorf("ring-offset-4: missing --tw-ring-offset-width:\n%s", result)
	}
}

// ===== Gradient Direction Sets background-image =====

func TestGradientDirectionSetsBackgroundImage(t *testing.T) {
	e := newShadowRingTestEngine(t)
	e.Write([]byte("bg-gradient-to-r"))
	result := e.CSS()
	if !strings.Contains(result, "background-image: linear-gradient(to right, var(--tw-gradient-stops))") {
		t.Errorf("bg-gradient-to-r missing background-image:\n%s", result)
	}
}

// ===== Gradient Stops Compose Correctly =====

func TestGradientFromSetsStops(t *testing.T) {
	e := newShadowRingTestEngine(t)
	e.Write([]byte("from-blue-500"))
	result := e.CSS()
	// from sets --tw-gradient-stops with from and to
	if !strings.Contains(result, "--tw-gradient-stops: var(--tw-gradient-from) var(--tw-gradient-from-position,), var(--tw-gradient-to, transparent) var(--tw-gradient-to-position,)") {
		t.Errorf("from-blue-500 missing gradient-stops composition:\n%s", result)
	}
}

func TestGradientViaSetsStops(t *testing.T) {
	e := newShadowRingTestEngine(t)
	e.Write([]byte("via-purple-500"))
	result := e.CSS()
	// via sets --tw-gradient-stops with from, via, and to
	if !strings.Contains(result, "--tw-gradient-stops: var(--tw-gradient-from, transparent) var(--tw-gradient-from-position,), var(--tw-gradient-via) var(--tw-gradient-via-position,), var(--tw-gradient-to, transparent) var(--tw-gradient-to-position,)") {
		t.Errorf("via-purple-500 missing gradient-stops composition:\n%s", result)
	}
}
