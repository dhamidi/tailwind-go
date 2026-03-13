package tailwind

import (
	"strings"
	"testing"
)

// filterChainSubstr is the expected filter composition value.
const filterChainSubstr = "var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)"

// backdropChainSubstr is the expected backdrop-filter composition value.
const backdropChainSubstr = "var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,) var(--tw-backdrop-opacity,)"

// ===== Transform — Scale =====

func TestScaleUtilities(t *testing.T) {
	tests := []struct {
		class    string
		scaleVal string // expected --tw-scale-x value
	}{
		{"scale-50", "50%"},
		{"scale-75", "75%"},
		{"scale-90", "90%"},
		{"scale-95", "95%"},
		{"scale-100", "100%"},
		{"scale-105", "105%"},
		{"scale-110", "110%"},
		{"scale-125", "125%"},
		{"scale-150", "150%"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "--tw-scale-x: "+tt.scaleVal) {
				t.Errorf("%s: expected --tw-scale-x: %s in:\n%s", tt.class, tt.scaleVal, result)
			}
			if !strings.Contains(result, "--tw-scale-y: "+tt.scaleVal) {
				t.Errorf("%s: expected --tw-scale-y: %s in:\n%s", tt.class, tt.scaleVal, result)
			}
			if !strings.Contains(result, "scale: var(--tw-scale-x) var(--tw-scale-y)") {
				t.Errorf("%s: missing scale composition in:\n%s", tt.class, result)
			}
		})
	}
}

func TestScaleAxisUtilities(t *testing.T) {
	t.Run("scale-x-50", func(t *testing.T) {
		e := New()
		e.Write([]byte("scale-x-50"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-scale-x: 50%") {
			t.Errorf("missing --tw-scale-x: 50%% in:\n%s", result)
		}
		if !strings.Contains(result, "scale: var(--tw-scale-x) var(--tw-scale-y, 1)") {
			t.Errorf("missing scale composition in:\n%s", result)
		}
	})

	t.Run("scale-y-75", func(t *testing.T) {
		e := New()
		e.Write([]byte("scale-y-75"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-scale-y: 75%") {
			t.Errorf("missing --tw-scale-y: 75%% in:\n%s", result)
		}
		if !strings.Contains(result, "scale: var(--tw-scale-x, 1) var(--tw-scale-y)") {
			t.Errorf("missing scale composition in:\n%s", result)
		}
	})
}

func TestScaleNone(t *testing.T) {
	e := New()
	e.Write([]byte("scale-none"))
	result := e.CSS()
	if !strings.Contains(result, "scale: none") {
		t.Errorf("scale-none: expected scale: none in:\n%s", result)
	}
}

func TestScaleArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte("scale-[1.15]"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-scale-x: 1.15") {
		t.Errorf("scale-[1.15]: expected --tw-scale-x: 1.15 in:\n%s", result)
	}
	if !strings.Contains(result, "--tw-scale-y: 1.15") {
		t.Errorf("scale-[1.15]: expected --tw-scale-y: 1.15 in:\n%s", result)
	}
}

// ===== Transform — Rotate =====

func TestRotateUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"rotate-0", "rotate: 0deg"},
		{"rotate-1", "rotate: 1deg"},
		{"rotate-2", "rotate: 2deg"},
		{"rotate-3", "rotate: 3deg"},
		{"rotate-6", "rotate: 6deg"},
		{"rotate-12", "rotate: 12deg"},
		{"rotate-45", "rotate: 45deg"},
		{"rotate-90", "rotate: 90deg"},
		{"rotate-180", "rotate: 180deg"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantVal) {
				t.Errorf("%s: expected %q in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestNegativeRotate(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"-rotate-45", "calc(45deg * -1)"},
		{"-rotate-90", "calc(90deg * -1)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantVal) {
				t.Errorf("%s: expected %q in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestRotateArbitrary(t *testing.T) {
	t.Run("rotate-[17deg]", func(t *testing.T) {
		e := New()
		e.Write([]byte("rotate-[17deg]"))
		result := e.CSS()
		if !strings.Contains(result, "rotate: 17deg") {
			t.Errorf("expected rotate: 17deg in:\n%s", result)
		}
	})

	t.Run("rotate-[0.5turn]", func(t *testing.T) {
		e := New()
		e.Write([]byte("rotate-[0.5turn]"))
		result := e.CSS()
		if !strings.Contains(result, "rotate: 0.5turn") {
			t.Errorf("expected rotate: 0.5turn in:\n%s", result)
		}
	})
}

// ===== Transform — Translate =====

func TestTranslateUtilities(t *testing.T) {
	tests := []struct {
		class    string
		wantProp string
		wantVal  string
	}{
		{"translate-x-0", "--tw-translate-x", "calc(var(--spacing) * 0)"},
		{"translate-x-1", "--tw-translate-x", "calc(var(--spacing) * 1)"},
		{"translate-x-4", "--tw-translate-x", "calc(var(--spacing) * 4)"},
		{"translate-x-full", "--tw-translate-x", "100%"},
		{"translate-x-1/2", "--tw-translate-x", "calc(1 / 2 * 100%)"},
		{"translate-x-px", "--tw-translate-x", "1px"},
		{"translate-y-4", "--tw-translate-y", "calc(var(--spacing) * 4)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantProp+": "+tt.wantVal) {
				t.Errorf("%s: expected %s: %s in:\n%s", tt.class, tt.wantProp, tt.wantVal, result)
			}
			if !strings.Contains(result, "translate: var(--tw-translate-x) var(--tw-translate-y)") {
				t.Errorf("%s: missing translate composition in:\n%s", tt.class, result)
			}
		})
	}
}

func TestNegativeTranslate(t *testing.T) {
	t.Run("-translate-x-1/2", func(t *testing.T) {
		e := New()
		e.Write([]byte("-translate-x-1/2"))
		result := e.CSS()
		if !strings.Contains(result, "calc(calc(1 / 2 * 100%) * -1)") {
			t.Errorf("expected negative translate in:\n%s", result)
		}
	})

	t.Run("-translate-x-full", func(t *testing.T) {
		e := New()
		e.Write([]byte("-translate-x-full"))
		result := e.CSS()
		if !strings.Contains(result, "calc(100% * -1)") {
			t.Errorf("expected calc(100%% * -1) in:\n%s", result)
		}
	})

	t.Run("-translate-y-1/2", func(t *testing.T) {
		e := New()
		e.Write([]byte("-translate-y-1/2"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-translate-y") {
			t.Errorf("expected --tw-translate-y in:\n%s", result)
		}
	})
}

func TestTranslateArbitrary(t *testing.T) {
	t.Run("translate-x-[30%]", func(t *testing.T) {
		e := New()
		e.Write([]byte("translate-x-[30%]"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-translate-x: 30%") {
			t.Errorf("expected --tw-translate-x: 30%% in:\n%s", result)
		}
	})

	t.Run("translate-x-[var(--offset)]", func(t *testing.T) {
		e := New()
		e.Write([]byte("translate-x-[var(--offset)]"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-translate-x: var(--offset)") {
			t.Errorf("expected --tw-translate-x: var(--offset) in:\n%s", result)
		}
	})
}

// ===== Transform — Skew =====

func TestSkewUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"skew-x-1", "--tw-skew-x: 1deg"},
		{"skew-x-2", "--tw-skew-x: 2deg"},
		{"skew-x-3", "--tw-skew-x: 3deg"},
		{"skew-x-6", "--tw-skew-x: 6deg"},
		{"skew-x-12", "--tw-skew-x: 12deg"},
		{"skew-y-3", "--tw-skew-y: 3deg"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantVal) {
				t.Errorf("%s: expected %q in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestNegativeSkew(t *testing.T) {
	e := New()
	e.Write([]byte("-skew-x-6"))
	result := e.CSS()
	if !strings.Contains(result, "calc(6deg * -1)") {
		t.Errorf("-skew-x-6: expected calc(6deg * -1) in:\n%s", result)
	}
}

func TestSkewArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte("skew-x-[17deg]"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-skew-x: 17deg") {
		t.Errorf("expected --tw-skew-x: 17deg in:\n%s", result)
	}
}

// ===== Transform — Origin =====

func TestOriginUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"origin-center", "transform-origin: center"},
		{"origin-top", "transform-origin: top"},
		{"origin-top-right", "transform-origin: top right"},
		{"origin-right", "transform-origin: right"},
		{"origin-bottom-right", "transform-origin: bottom right"},
		{"origin-bottom", "transform-origin: bottom"},
		{"origin-bottom-left", "transform-origin: bottom left"},
		{"origin-left", "transform-origin: left"},
		{"origin-top-left", "transform-origin: top left"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantVal) {
				t.Errorf("%s: expected %q in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

// ===== Transform — Perspective =====

func TestPerspectiveUtilities(t *testing.T) {
	t.Run("perspective-none", func(t *testing.T) {
		e := New()
		e.Write([]byte("perspective-none"))
		result := e.CSS()
		if !strings.Contains(result, "perspective: none") {
			t.Errorf("expected perspective: none in:\n%s", result)
		}
	})

	t.Run("perspective-[500px]", func(t *testing.T) {
		e := New()
		e.Write([]byte("perspective-[500px]"))
		result := e.CSS()
		if !strings.Contains(result, "perspective: 500px") {
			t.Errorf("expected perspective: 500px in:\n%s", result)
		}
	})
}

// ===== Transition =====

func TestTransitionUtilities(t *testing.T) {
	tests := []struct {
		class        string
		wantProperty string // expected transition-property value substring
	}{
		{"transition", "color, background-color"},
		{"transition-all", "all"},
		{"transition-colors", "color, background-color, border-color"},
		{"transition-opacity", "opacity"},
		{"transition-shadow", "box-shadow"},
		{"transition-transform", "transform, translate, scale, rotate"},
		{"transition-none", "none"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "transition-property: "+tt.wantProperty) {
				t.Errorf("%s: expected transition-property: %s in:\n%s", tt.class, tt.wantProperty, result)
			}
			// All except transition-none should have timing-function and duration
			if tt.class != "transition-none" {
				if !strings.Contains(result, "transition-timing-function:") {
					t.Errorf("%s: missing transition-timing-function in:\n%s", tt.class, result)
				}
				if !strings.Contains(result, "transition-duration:") {
					t.Errorf("%s: missing transition-duration in:\n%s", tt.class, result)
				}
			}
		})
	}
}

// ===== Duration =====

func TestDurationUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"duration-0", "0s"},
		{"duration-75", "75ms"},
		{"duration-100", "100ms"},
		{"duration-150", "150ms"},
		{"duration-200", "200ms"},
		{"duration-300", "300ms"},
		{"duration-500", "500ms"},
		{"duration-700", "700ms"},
		{"duration-1000", "1000ms"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "transition-duration: "+tt.wantVal) {
				t.Errorf("%s: expected transition-duration: %s in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestDurationArbitrary(t *testing.T) {
	t.Run("duration-[2000ms]", func(t *testing.T) {
		e := New()
		e.Write([]byte("duration-[2000ms]"))
		result := e.CSS()
		if !strings.Contains(result, "transition-duration: 2000ms") {
			t.Errorf("expected transition-duration: 2000ms in:\n%s", result)
		}
	})

	t.Run("duration-[.5s]", func(t *testing.T) {
		e := New()
		e.Write([]byte("duration-[.5s]"))
		result := e.CSS()
		if !strings.Contains(result, "transition-duration: .5s") {
			t.Errorf("expected transition-duration: .5s in:\n%s", result)
		}
	})
}

// ===== Delay =====

func TestDelayUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"delay-0", "0s"},
		{"delay-100", "100ms"},
		{"delay-200", "200ms"},
		{"delay-300", "300ms"},
		{"delay-500", "500ms"},
		{"delay-1000", "1000ms"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "transition-delay: "+tt.wantVal) {
				t.Errorf("%s: expected transition-delay: %s in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestDelayArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte("delay-[2000ms]"))
	result := e.CSS()
	if !strings.Contains(result, "transition-delay: 2000ms") {
		t.Errorf("expected transition-delay: 2000ms in:\n%s", result)
	}
}

// ===== Easing =====

func TestEaseUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"ease-linear", "linear"},
		{"ease-in", "var(--ease-in)"},
		{"ease-out", "var(--ease-out)"},
		{"ease-in-out", "var(--ease-in-out)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "transition-timing-function: "+tt.wantVal) {
				t.Errorf("%s: expected transition-timing-function: %s in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestEaseArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte("ease-[cubic-bezier(0.4,0,0.2,1)]"))
	result := e.CSS()
	if !strings.Contains(result, "transition-timing-function: cubic-bezier(0.4,0,0.2,1)") {
		t.Errorf("expected cubic-bezier easing in:\n%s", result)
	}
}

// ===== Animation =====

func TestAnimateUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"animate-none", "animation: none"},
		{"animate-spin", "animation: var(--animate-spin, spin 1s linear infinite)"},
		{"animate-ping", "animation: var(--animate-ping, ping 1s cubic-bezier(0, 0, 0.2, 1) infinite)"},
		{"animate-pulse", "animation: var(--animate-pulse, pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite)"},
		{"animate-bounce", "animation: var(--animate-bounce, bounce 1s infinite)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantVal) {
				t.Errorf("%s: expected %q in:\n%s", tt.class, tt.wantVal, result)
			}
		})
	}
}

func TestAnimateArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte("animate-[wiggle_1s_ease-in-out_infinite]"))
	result := e.CSS()
	// Underscores in arbitrary values are converted to spaces
	if !strings.Contains(result, "animation: wiggle 1s ease-in-out infinite") {
		t.Errorf("expected animation with underscore-to-space conversion in:\n%s", result)
	}
}

// ===== Filter — Blur =====

func TestBlurUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVar string
	}{
		{"blur-none", "blur(0)"},
		{"blur-sm", "blur(var(--blur-sm))"},
		{"blur", "blur(var(--blur-md))"},
		{"blur-md", "blur(var(--blur-md))"},
		{"blur-lg", "blur(var(--blur-lg))"},
		{"blur-xl", "blur(var(--blur-xl))"},
		{"blur-2xl", "blur(var(--blur-2xl))"},
		{"blur-3xl", "blur(var(--blur-3xl))"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "--tw-blur: "+tt.wantVar) {
				t.Errorf("%s: expected --tw-blur: %s in:\n%s", tt.class, tt.wantVar, result)
			}
			if !strings.Contains(result, "filter: "+filterChainSubstr) {
				t.Errorf("%s: missing filter chain in:\n%s", tt.class, result)
			}
		})
	}
}

func TestBlurArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte("blur-[20px]"))
	result := e.CSS()
	if !strings.Contains(result, "--tw-blur: blur(20px)") {
		t.Errorf("expected --tw-blur: blur(20px) in:\n%s", result)
	}
	if !strings.Contains(result, "filter: "+filterChainSubstr) {
		t.Errorf("missing filter chain in:\n%s", result)
	}
}

// ===== Filter — Brightness =====

func TestBrightnessUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"brightness-0", "brightness(0%)"},
		{"brightness-50", "brightness(50%)"},
		{"brightness-75", "brightness(75%)"},
		{"brightness-90", "brightness(90%)"},
		{"brightness-95", "brightness(95%)"},
		{"brightness-100", "brightness(100%)"},
		{"brightness-105", "brightness(105%)"},
		{"brightness-110", "brightness(110%)"},
		{"brightness-125", "brightness(125%)"},
		{"brightness-150", "brightness(150%)"},
		{"brightness-200", "brightness(200%)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "--tw-brightness: "+tt.wantVal) {
				t.Errorf("%s: expected --tw-brightness: %s in:\n%s", tt.class, tt.wantVal, result)
			}
			if !strings.Contains(result, "filter: "+filterChainSubstr) {
				t.Errorf("%s: missing filter chain in:\n%s", tt.class, result)
			}
		})
	}
}

// ===== Filter — Contrast, Saturate, Hue-Rotate =====

func TestContrastUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"contrast-0", "contrast(0%)"},
		{"contrast-50", "contrast(50%)"},
		{"contrast-100", "contrast(100%)"},
		{"contrast-200", "contrast(200%)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "--tw-contrast: "+tt.wantVal) {
				t.Errorf("%s: expected --tw-contrast: %s in:\n%s", tt.class, tt.wantVal, result)
			}
			if !strings.Contains(result, "filter:") {
				t.Errorf("%s: missing filter in:\n%s", tt.class, result)
			}
		})
	}
}

func TestSaturateUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"saturate-0", "saturate(0%)"},
		{"saturate-50", "saturate(50%)"},
		{"saturate-100", "saturate(100%)"},
		{"saturate-200", "saturate(200%)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "--tw-saturate: "+tt.wantVal) {
				t.Errorf("%s: expected --tw-saturate: %s in:\n%s", tt.class, tt.wantVal, result)
			}
			if !strings.Contains(result, "filter:") {
				t.Errorf("%s: missing filter in:\n%s", tt.class, result)
			}
		})
	}
}

func TestHueRotateUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVal string
	}{
		{"hue-rotate-0", "hue-rotate(0deg)"},
		{"hue-rotate-15", "hue-rotate(15deg)"},
		{"hue-rotate-30", "hue-rotate(30deg)"},
		{"hue-rotate-60", "hue-rotate(60deg)"},
		{"hue-rotate-90", "hue-rotate(90deg)"},
		{"hue-rotate-180", "hue-rotate(180deg)"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, "--tw-hue-rotate: "+tt.wantVal) {
				t.Errorf("%s: expected --tw-hue-rotate: %s in:\n%s", tt.class, tt.wantVal, result)
			}
			if !strings.Contains(result, "filter:") {
				t.Errorf("%s: missing filter in:\n%s", tt.class, result)
			}
		})
	}
}

// ===== Filter — Grayscale, Invert, Sepia =====

func TestGrayscaleUtilities(t *testing.T) {
	t.Run("grayscale", func(t *testing.T) {
		e := New()
		e.Write([]byte("grayscale"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-grayscale: grayscale(100%)") {
			t.Errorf("expected grayscale(100%%) in:\n%s", result)
		}
	})

	t.Run("grayscale-0", func(t *testing.T) {
		e := New()
		e.Write([]byte("grayscale-0"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-grayscale: grayscale(0)") {
			t.Errorf("expected grayscale(0) in:\n%s", result)
		}
	})
}

func TestInvertUtilities(t *testing.T) {
	t.Run("invert", func(t *testing.T) {
		e := New()
		e.Write([]byte("invert"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-invert: invert(100%)") {
			t.Errorf("expected invert(100%%) in:\n%s", result)
		}
	})

	t.Run("invert-0", func(t *testing.T) {
		e := New()
		e.Write([]byte("invert-0"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-invert: invert(0)") {
			t.Errorf("expected invert(0) in:\n%s", result)
		}
	})
}

func TestSepiaUtilities(t *testing.T) {
	t.Run("sepia", func(t *testing.T) {
		e := New()
		e.Write([]byte("sepia"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-sepia: sepia(100%)") {
			t.Errorf("expected sepia(100%%) in:\n%s", result)
		}
	})

	t.Run("sepia-0", func(t *testing.T) {
		e := New()
		e.Write([]byte("sepia-0"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-sepia: sepia(0)") {
			t.Errorf("expected sepia(0) in:\n%s", result)
		}
	})
}

// ===== Drop Shadow =====

func TestDropShadowUtilities(t *testing.T) {
	tests := []struct {
		class   string
		wantVar string // substring in --tw-drop-shadow value
	}{
		{"drop-shadow-none", "drop-shadow(0 0 #0000)"},
		{"drop-shadow-xs", "drop-shadow(var(--drop-shadow-xs"},
		{"drop-shadow-sm", "drop-shadow(var(--drop-shadow-sm"},
		{"drop-shadow", "drop-shadow(var(--drop-shadow-sm"},
		{"drop-shadow-md", "drop-shadow(var(--drop-shadow-md"},
		{"drop-shadow-lg", "drop-shadow(var(--drop-shadow-lg"},
		{"drop-shadow-xl", "drop-shadow(var(--drop-shadow-xl"},
		{"drop-shadow-2xl", "drop-shadow(var(--drop-shadow-2xl"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			result := e.CSS()
			if !strings.Contains(result, tt.wantVar) {
				t.Errorf("%s: expected %q in:\n%s", tt.class, tt.wantVar, result)
			}
			if !strings.Contains(result, "filter: "+filterChainSubstr) {
				t.Errorf("%s: missing filter chain in:\n%s", tt.class, result)
			}
		})
	}
}

// ===== Backdrop Filter =====

func TestBackdropFilterUtilities(t *testing.T) {
	t.Run("backdrop-blur", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-blur"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-blur: blur(var(--blur-md))") {
			t.Errorf("expected backdrop-blur default in:\n%s", result)
		}
		if !strings.Contains(result, "backdrop-filter: "+backdropChainSubstr) {
			t.Errorf("missing backdrop-filter chain in:\n%s", result)
		}
	})

	t.Run("backdrop-blur-sm", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-blur-sm"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-blur: blur(var(--blur-sm))") {
			t.Errorf("expected --tw-backdrop-blur in:\n%s", result)
		}
	})

	t.Run("backdrop-blur-md", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-blur-md"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-blur: blur(var(--blur-md))") {
			t.Errorf("expected --tw-backdrop-blur in:\n%s", result)
		}
	})

	t.Run("backdrop-blur-[10px]", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-blur-[10px]"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-blur: blur(10px)") {
			t.Errorf("expected --tw-backdrop-blur: blur(10px) in:\n%s", result)
		}
		if !strings.Contains(result, "backdrop-filter: "+backdropChainSubstr) {
			t.Errorf("missing backdrop-filter chain in:\n%s", result)
		}
	})

	t.Run("backdrop-brightness-50", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-brightness-50"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-brightness: brightness(50%)") {
			t.Errorf("expected backdrop-brightness in:\n%s", result)
		}
		if !strings.Contains(result, "backdrop-filter:") {
			t.Errorf("missing backdrop-filter in:\n%s", result)
		}
	})

	t.Run("backdrop-contrast-125", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-contrast-125"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-contrast: contrast(125%)") {
			t.Errorf("expected backdrop-contrast in:\n%s", result)
		}
	})

	t.Run("backdrop-grayscale", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-grayscale"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-grayscale: grayscale(100%)") {
			t.Errorf("expected backdrop-grayscale in:\n%s", result)
		}
	})

	t.Run("backdrop-invert", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-invert"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-invert: invert(100%)") {
			t.Errorf("expected backdrop-invert in:\n%s", result)
		}
	})

	t.Run("backdrop-saturate-150", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-saturate-150"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-saturate: saturate(150%)") {
			t.Errorf("expected backdrop-saturate in:\n%s", result)
		}
	})

	t.Run("backdrop-sepia", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-sepia"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-sepia: sepia(100%)") {
			t.Errorf("expected backdrop-sepia in:\n%s", result)
		}
	})

	t.Run("backdrop-hue-rotate-90", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-hue-rotate-90"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-hue-rotate: hue-rotate(90deg)") {
			t.Errorf("expected backdrop-hue-rotate in:\n%s", result)
		}
	})

	t.Run("backdrop-opacity-50", func(t *testing.T) {
		e := New()
		e.Write([]byte("backdrop-opacity-50"))
		result := e.CSS()
		if !strings.Contains(result, "--tw-backdrop-opacity: opacity(50%)") {
			t.Errorf("expected backdrop-opacity in:\n%s", result)
		}
	})
}

// ===== Variant Combinations =====

func TestTransformFilterVariants(t *testing.T) {
	t.Run("hover:scale-105", func(t *testing.T) {
		e := New()
		e.Write([]byte("hover:scale-105"))
		result := e.CSS()
		if !strings.Contains(result, ":hover") {
			t.Errorf("missing :hover selector in:\n%s", result)
		}
		if !strings.Contains(result, "--tw-scale-x: 105%") {
			t.Errorf("missing scale value in:\n%s", result)
		}
	})

	t.Run("hover:brightness-110", func(t *testing.T) {
		e := New()
		e.Write([]byte("hover:brightness-110"))
		result := e.CSS()
		if !strings.Contains(result, ":hover") {
			t.Errorf("missing :hover selector in:\n%s", result)
		}
		if !strings.Contains(result, "--tw-brightness: brightness(110%)") {
			t.Errorf("missing brightness value in:\n%s", result)
		}
	})

	t.Run("active:scale-95", func(t *testing.T) {
		e := New()
		e.Write([]byte("active:scale-95"))
		result := e.CSS()
		if !strings.Contains(result, ":active") {
			t.Errorf("missing :active selector in:\n%s", result)
		}
		if !strings.Contains(result, "--tw-scale-x: 95%") {
			t.Errorf("missing scale value in:\n%s", result)
		}
	})

	t.Run("motion-safe:transition", func(t *testing.T) {
		e := New()
		e.Write([]byte("motion-safe:transition"))
		result := e.CSS()
		if !strings.Contains(result, "prefers-reduced-motion: no-preference") {
			t.Errorf("missing motion-safe media query in:\n%s", result)
		}
		if !strings.Contains(result, "transition-property:") {
			t.Errorf("missing transition-property in:\n%s", result)
		}
	})

	t.Run("motion-safe:duration-300", func(t *testing.T) {
		e := New()
		e.Write([]byte("motion-safe:duration-300"))
		result := e.CSS()
		if !strings.Contains(result, "prefers-reduced-motion: no-preference") {
			t.Errorf("missing motion-safe media query in:\n%s", result)
		}
		if !strings.Contains(result, "transition-duration: 300ms") {
			t.Errorf("missing duration value in:\n%s", result)
		}
	})

	t.Run("motion-reduce:transition-none", func(t *testing.T) {
		e := New()
		e.Write([]byte("motion-reduce:transition-none"))
		result := e.CSS()
		if !strings.Contains(result, "prefers-reduced-motion: reduce") {
			t.Errorf("missing motion-reduce media query in:\n%s", result)
		}
		if !strings.Contains(result, "transition-property: none") {
			t.Errorf("missing transition-property: none in:\n%s", result)
		}
	})

	t.Run("dark:backdrop-blur-lg", func(t *testing.T) {
		e := New()
		e.Write([]byte("dark:backdrop-blur-lg"))
		result := e.CSS()
		if !strings.Contains(result, "prefers-color-scheme: dark") {
			t.Errorf("missing dark mode media query in:\n%s", result)
		}
		if !strings.Contains(result, "--tw-backdrop-blur: blur(var(--blur-lg))") {
			t.Errorf("missing backdrop-blur value in:\n%s", result)
		}
	})
}
