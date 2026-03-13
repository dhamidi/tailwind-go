package tailwind

import (
	"strings"
	"testing"
)

func newMaskTestEngine(t *testing.T) *Engine {
	t.Helper()
	e := New()
	return e
}

// assertMaskCSS is a helper that writes the class and checks for a property:value pair.
func assertMaskCSS(t *testing.T, e *Engine, class, wantProp, wantVal string) {
	t.Helper()
	e.Write([]byte(class))
	result := e.CSS()
	check := wantProp + ": " + wantVal
	if !strings.Contains(result, check) {
		t.Errorf("class %q: expected %q in output:\n%s", class, check, result)
	}
}

// === Static mask utilities ===

func TestMaskUtility_MaskNone(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-none", "mask-image", "none")
}

func TestMaskUtility_MaskClip(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-clip-border", "mask-clip", "border-box")
	assertMaskCSS(t, e, "mask-clip-padding", "mask-clip", "padding-box")
	assertMaskCSS(t, e, "mask-clip-content", "mask-clip", "content-box")
	assertMaskCSS(t, e, "mask-clip-fill", "mask-clip", "fill-box")
	assertMaskCSS(t, e, "mask-clip-stroke", "mask-clip", "stroke-box")
	assertMaskCSS(t, e, "mask-clip-view", "mask-clip", "view-box")
	assertMaskCSS(t, e, "mask-clip-none", "mask-clip", "no-clip")
}

func TestMaskUtility_MaskComposite(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-composite-add", "mask-composite", "add")
	assertMaskCSS(t, e, "mask-composite-subtract", "mask-composite", "subtract")
	assertMaskCSS(t, e, "mask-composite-intersect", "mask-composite", "intersect")
	assertMaskCSS(t, e, "mask-composite-exclude", "mask-composite", "exclude")
}

func TestMaskUtility_MaskMode(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-alpha", "mask-mode", "alpha")
	assertMaskCSS(t, e, "mask-luminance", "mask-mode", "luminance")
	assertMaskCSS(t, e, "mask-match", "mask-mode", "match-source")
}

func TestMaskUtility_MaskOrigin(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-origin-border", "mask-origin", "border-box")
	assertMaskCSS(t, e, "mask-origin-padding", "mask-origin", "padding-box")
	assertMaskCSS(t, e, "mask-origin-content", "mask-origin", "content-box")
	assertMaskCSS(t, e, "mask-origin-fill", "mask-origin", "fill-box")
	assertMaskCSS(t, e, "mask-origin-stroke", "mask-origin", "stroke-box")
	assertMaskCSS(t, e, "mask-origin-view", "mask-origin", "view-box")
}

func TestMaskUtility_MaskPosition(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-position-center", "mask-position", "center")
	assertMaskCSS(t, e, "mask-position-top", "mask-position", "top")
	assertMaskCSS(t, e, "mask-position-right", "mask-position", "right")
	assertMaskCSS(t, e, "mask-position-bottom", "mask-position", "bottom")
	assertMaskCSS(t, e, "mask-position-left", "mask-position", "left")
	assertMaskCSS(t, e, "mask-position-top-right", "mask-position", "top right")
	assertMaskCSS(t, e, "mask-position-bottom-left", "mask-position", "bottom left")
	assertMaskCSS(t, e, "mask-position-top-left", "mask-position", "top left")
	assertMaskCSS(t, e, "mask-position-bottom-right", "mask-position", "bottom right")
}

func TestMaskUtility_MaskRepeat(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-repeat", "mask-repeat", "repeat")
	assertMaskCSS(t, e, "mask-no-repeat", "mask-repeat", "no-repeat")
	assertMaskCSS(t, e, "mask-repeat-x", "mask-repeat", "repeat-x")
	assertMaskCSS(t, e, "mask-repeat-y", "mask-repeat", "repeat-y")
	assertMaskCSS(t, e, "mask-repeat-round", "mask-repeat", "round")
	assertMaskCSS(t, e, "mask-repeat-space", "mask-repeat", "space")
}

func TestMaskUtility_MaskSize(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-size-auto", "mask-size", "auto")
	assertMaskCSS(t, e, "mask-size-cover", "mask-size", "cover")
	assertMaskCSS(t, e, "mask-size-contain", "mask-size", "contain")
}

func TestMaskUtility_MaskType(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-type-alpha", "mask-type", "alpha")
	assertMaskCSS(t, e, "mask-type-luminance", "mask-type", "luminance")
}

// === Edge mask gradients ===

func TestMaskUtility_EdgeBottom(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-b-from-50"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-b-from: 50%") {
		t.Errorf("mask-b-from-50 missing --tw-mask-b-from:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(to bottom") {
		t.Errorf("mask-b-from-50 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_EdgeTop(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-t-to-100"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-t-to: 100%") {
		t.Errorf("mask-t-to-100 missing --tw-mask-t-to:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(to top") {
		t.Errorf("mask-t-to-100 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_EdgeRight(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-r-from-25"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-r-from: 25%") {
		t.Errorf("mask-r-from-25 missing --tw-mask-r-from:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(to right") {
		t.Errorf("mask-r-from-25 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_EdgeLeft(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-l-to-75"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-l-to: 75%") {
		t.Errorf("mask-l-to-75 missing var:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(to left") {
		t.Errorf("mask-l-to-75 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_EdgeX(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-x-from-10"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-x-from: 10%") {
		t.Errorf("mask-x-from-10 missing var:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(to right") {
		t.Errorf("mask-x-from-10 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_EdgeY(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-y-to-90"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-y-to: 90%") {
		t.Errorf("mask-y-to-90 missing var:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(to bottom") {
		t.Errorf("mask-y-to-90 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_EdgeArbitrary(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-b-from-[25%]"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-b-from: 25%") {
		t.Errorf("mask-b-from-[25%%] missing var:\n%s", result)
	}
}

// === Linear gradient masks ===

func TestMaskUtility_LinearAngle(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-linear-180"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-linear-angle: 180deg") {
		t.Errorf("mask-linear-180 missing angle:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(") {
		t.Errorf("mask-linear-180 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_LinearNegativeAngle(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("-mask-linear-45"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-linear-angle: -45deg") {
		t.Errorf("-mask-linear-45 missing negative angle:\n%s", result)
	}
}

func TestMaskUtility_LinearFrom(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-linear-from-50"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-linear-from: 50%") {
		t.Errorf("mask-linear-from-50 missing var:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: linear-gradient(") {
		t.Errorf("mask-linear-from-50 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_LinearTo(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-linear-to-100"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-linear-to: 100%") {
		t.Errorf("mask-linear-to-100 missing var:\n%s", result)
	}
}

// === Radial gradient masks ===

func TestMaskUtility_RadialShape(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-circle", "--tw-mask-radial-shape", "circle")
	assertMaskCSS(t, e, "mask-ellipse", "--tw-mask-radial-shape", "ellipse")
}

func TestMaskUtility_RadialSize(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-radial-closest-corner", "--tw-mask-radial-size", "closest-corner")
	assertMaskCSS(t, e, "mask-radial-closest-side", "--tw-mask-radial-size", "closest-side")
	assertMaskCSS(t, e, "mask-radial-farthest-corner", "--tw-mask-radial-size", "farthest-corner")
	assertMaskCSS(t, e, "mask-radial-farthest-side", "--tw-mask-radial-size", "farthest-side")
}

func TestMaskUtility_RadialPosition(t *testing.T) {
	e := newMaskTestEngine(t)
	assertMaskCSS(t, e, "mask-radial-at-center", "--tw-mask-radial-position", "center")
	assertMaskCSS(t, e, "mask-radial-at-top", "--tw-mask-radial-position", "top")
	assertMaskCSS(t, e, "mask-radial-at-bottom-left", "--tw-mask-radial-position", "bottom left")
}

func TestMaskUtility_RadialArbitrary(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-radial-[circle_at_top]"))
	result := e.CSS()
	if !strings.Contains(result, "mask-image: radial-gradient(circle at top)") {
		t.Errorf("mask-radial-[circle_at_top] missing radial-gradient:\n%s", result)
	}
}

func TestMaskUtility_RadialFrom(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-radial-from-30"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-radial-from: 30%") {
		t.Errorf("mask-radial-from-30 missing var:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: radial-gradient(") {
		t.Errorf("mask-radial-from-30 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_RadialTo(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-radial-to-80"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-radial-to: 80%") {
		t.Errorf("mask-radial-to-80 missing var:\n%s", result)
	}
}

// === Conic gradient masks ===

func TestMaskUtility_ConicAngle(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-conic-90"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-conic-angle: 90deg") {
		t.Errorf("mask-conic-90 missing angle:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: conic-gradient(") {
		t.Errorf("mask-conic-90 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_ConicNegativeAngle(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("-mask-conic-30"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-conic-angle: -30deg") {
		t.Errorf("-mask-conic-30 missing negative angle:\n%s", result)
	}
}

func TestMaskUtility_ConicFrom(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-conic-from-25"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-conic-from: 25%") {
		t.Errorf("mask-conic-from-25 missing var:\n%s", result)
	}
	if !strings.Contains(result, "mask-image: conic-gradient(") {
		t.Errorf("mask-conic-from-25 missing mask-image:\n%s", result)
	}
}

func TestMaskUtility_ConicTo(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-conic-to-75"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-mask-conic-to: 75%") {
		t.Errorf("mask-conic-to-75 missing var:\n%s", result)
	}
}

func TestMaskUtility_ConicArbitrary(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-conic-[from_45deg_at_center]"))
	result := e.CSS()
	if !strings.Contains(result, "mask-image: conic-gradient(from 45deg at center)") {
		t.Errorf("mask-conic arbitrary missing conic-gradient:\n%s", result)
	}
}

// === Arbitrary mask-image ===

func TestMaskUtility_ArbitraryURL(t *testing.T) {
	e := newMaskTestEngine(t)
	e.Write([]byte("mask-[url(/img/mask.png)]"))
	result := e.CSS()
	if !strings.Contains(result, "mask-image: url(/img/mask.png)") {
		t.Errorf("mask-[url(/img/mask.png)] missing:\n%s", result)
	}
}
