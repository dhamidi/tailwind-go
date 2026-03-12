package tailwind

import (
	"strings"
	"testing"
)

func newColorTestEngine(t *testing.T) *Engine {
	t.Helper()
	css := []byte(`
@theme {
  --color-red-500: #ef4444;
  --color-blue-500: #3b82f6;
  --color-green-300: #86efac;
  --color-black: #000;
  --color-white: #fff;
  --font-size-lg: 1.125rem;
  --font-size-xl: 1.25rem;
  --opacity-50: 0.5;
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1);
  --spacing: 0.25rem;
}
`)
	e := New()
	if err := e.LoadCSS(css); err != nil {
		t.Fatal(err)
	}
	return e
}

func assertCSS(t *testing.T, e *Engine, class, wantProp, wantVal string) {
	t.Helper()
	e.Write([]byte(class))
	result := e.CSS()
	check := wantProp + ": " + wantVal
	if !strings.Contains(result, check) {
		t.Errorf("class %q: expected %q in output:\n%s", class, check, result)
	}
}

func assertNoCSS(t *testing.T, e *Engine, class string) {
	t.Helper()
	e.Write([]byte(class))
	result := e.CSS()
	if result != "" {
		t.Errorf("class %q: expected no output, got:\n%s", class, result)
	}
}

func TestColorUtility_BgColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-red-500", "background-color", "#ef4444")
}

func TestColorUtility_BgColorOpacity(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-red-500/50", "background-color", "color-mix(in srgb, #ef4444 0.5, transparent)")
}

func TestColorUtility_BgColorArbitraryOpacity(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-red-500/[.3]", "background-color", "color-mix(in srgb, #ef4444 .3, transparent)")
}

func TestColorUtility_BgTransparent(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-transparent", "background-color", "transparent")
}

func TestColorUtility_BgCurrent(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-current", "background-color", "currentColor")
}

func TestColorUtility_BgInherit(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-inherit", "background-color", "inherit")
}

func TestColorUtility_BgArbitrary(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-[#ff0000]", "background-color", "#ff0000")
}

func TestColorUtility_BgArbitraryTypeHintColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-[color:red]", "background-color", "red")
}

func TestColorUtility_BgArbitraryImage(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-[url(image.jpg)]", "background-image", "url(image.jpg)")
}

func TestColorUtility_BgArbitraryGradient(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-[linear-gradient(to_right,red,blue)]", "background-image", "linear-gradient(to right,red,blue)")
}

func TestColorUtility_TextColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-red-500", "color", "#ef4444")
}

func TestColorUtility_TextFontSize(t *testing.T) {
	// text-lg is a static utility (registered in utilities_static.go)
	// that includes both font-size and line-height.
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-lg", "font-size", "var(--text-lg, 1.125rem)")
}

func TestColorUtility_TextColorOpacity(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-red-500/50", "color", "color-mix(in srgb, #ef4444 0.5, transparent)")
}

func TestColorUtility_TextCurrent(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-current", "color", "currentColor")
}

func TestColorUtility_TextTransparent(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-transparent", "color", "transparent")
}

func TestColorUtility_TextInherit(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-inherit", "color", "inherit")
}

func TestColorUtility_TextArbitraryColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "text-[color:red]", "color", "red")
}

func TestColorUtility_BorderColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-green-300", "border-color", "#86efac")
}

func TestColorUtility_BorderColorOpacity(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-red-500/50", "border-color", "color-mix(in srgb, #ef4444 0.5, transparent)")
}

func TestColorUtility_BorderTopColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-t-red-500", "border-top-color", "#ef4444")
}

func TestColorUtility_BorderRightColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-r-red-500", "border-right-color", "#ef4444")
}

func TestColorUtility_BorderBottomColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-b-red-500", "border-bottom-color", "#ef4444")
}

func TestColorUtility_BorderLeftColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-l-red-500", "border-left-color", "#ef4444")
}

func TestColorUtility_BorderXColor(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("border-x-red-500"))
	result := e.CSS()
	if !strings.Contains(result, "border-left-color: #ef4444") {
		t.Errorf("border-x-red-500 missing left color:\n%s", result)
	}
	if !strings.Contains(result, "border-right-color: #ef4444") {
		t.Errorf("border-x-red-500 missing right color:\n%s", result)
	}
}

func TestColorUtility_BorderYColor(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("border-y-red-500"))
	result := e.CSS()
	if !strings.Contains(result, "border-top-color: #ef4444") {
		t.Errorf("border-y-red-500 missing top color:\n%s", result)
	}
	if !strings.Contains(result, "border-bottom-color: #ef4444") {
		t.Errorf("border-y-red-500 missing bottom color:\n%s", result)
	}
}

func TestColorUtility_BorderStartColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-s-red-500", "border-inline-start-color", "#ef4444")
}

func TestColorUtility_BorderEndColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "border-e-red-500", "border-inline-end-color", "#ef4444")
}

func TestColorUtility_OutlineColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "outline-red-500", "outline-color", "#ef4444")
}

func TestColorUtility_RingColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "ring-red-500", "--tw-ring-color", "#ef4444")
}

func TestColorUtility_RingOffsetColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "ring-offset-red-500", "--tw-ring-offset-color", "#ef4444")
}

func TestColorUtility_AccentColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "accent-red-500", "accent-color", "#ef4444")
}

func TestColorUtility_CaretColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "caret-red-500", "caret-color", "#ef4444")
}

func TestColorUtility_FillColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "fill-red-500", "fill", "#ef4444")
}

func TestColorUtility_StrokeColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "stroke-red-500", "stroke", "#ef4444")
}

func TestColorUtility_StrokeWidth(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "stroke-2", "stroke-width", "2")
}

func TestColorUtility_DecorationColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "decoration-red-500", "text-decoration-color", "#ef4444")
}

func TestColorUtility_DivideColor(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("divide-red-500"))
	result := e.CSS()
	if !strings.Contains(result, "border-color: #ef4444") {
		t.Errorf("divide-red-500 missing border-color:\n%s", result)
	}
	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-red-500 missing child selector:\n%s", result)
	}
}

func TestColorUtility_ShadowValue(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("shadow-lg"))
	result := e.CSS()
	if !strings.Contains(result, "box-shadow:") {
		t.Errorf("shadow-lg missing box-shadow:\n%s", result)
	}
}

func TestColorUtility_ShadowColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "shadow-red-500", "--tw-shadow-color", "#ef4444")
}

func TestColorUtility_GradientFrom(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("from-red-500"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-gradient-from: #ef4444") {
		t.Errorf("from-red-500 missing gradient-from:\n%s", result)
	}
	if !strings.Contains(result, "--tw-gradient-stops: var(--tw-gradient-from), var(--tw-gradient-to, transparent)") {
		t.Errorf("from-red-500 missing gradient-stops:\n%s", result)
	}
}

func TestColorUtility_GradientVia(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("via-red-500"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-gradient-via: #ef4444") {
		t.Errorf("via-red-500 missing gradient-via:\n%s", result)
	}
	if !strings.Contains(result, "--tw-gradient-stops: var(--tw-gradient-from, transparent), var(--tw-gradient-via), var(--tw-gradient-to, transparent)") {
		t.Errorf("via-red-500 missing gradient-stops:\n%s", result)
	}
}

func TestColorUtility_GradientTo(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "to-red-500", "--tw-gradient-to", "#ef4444")
}

func TestColorUtility_GradientFromOpacity(t *testing.T) {
	e := newColorTestEngine(t)
	e.Write([]byte("from-red-500/50"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-gradient-from: color-mix(in srgb, #ef4444 0.5, transparent)") {
		t.Errorf("from-red-500/50 missing opacity:\n%s", result)
	}
}

func TestColorUtility_ArbitraryColorTypeHint(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-[color:red]", "background-color", "red")
}

func TestColorUtility_FillCurrent(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "fill-current", "fill", "currentColor")
}

func TestColorUtility_StrokeCurrent(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "stroke-current", "stroke", "currentColor")
}

func TestColorUtility_BgArbitraryWithOpacity(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "bg-[#ff0000]/50", "background-color", "color-mix(in srgb, #ff0000 0.5, transparent)")
}

func TestColorUtility_ShadowArbitrary(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "shadow-[0_4px_8px_rgba(0,0,0,0.1)]", "box-shadow", "0 4px 8px rgba(0,0,0,0.1)")
}

func TestColorUtility_ShadowColorTypeHint(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "shadow-[color:red]", "--tw-shadow-color", "red")
}

func TestColorUtility_StrokeArbitraryWidth(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "stroke-[2px]", "stroke", "2px")
}

func TestColorUtility_StrokeTypeHintColor(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "stroke-[color:red]", "stroke", "red")
}

func TestColorUtility_StrokeTypeHintNumber(t *testing.T) {
	e := newColorTestEngine(t)
	assertCSS(t, e, "stroke-[number:3]", "stroke-width", "3")
}
