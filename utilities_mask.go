package tailwind

// resolveMaskStopValue resolves a percentage-based stop value for mask gradients.
// Accepts numeric values (→ percentage) and arbitrary values.
func resolveMaskStopValue(c ResolvedCandidate) string {
	if c.Arbitrary != "" {
		return c.Arbitrary
	}
	if c.Value == "" {
		return ""
	}
	if isNumeric(c.Value) {
		return c.Value + "%"
	}
	return ""
}

// resolveMaskAngleValue resolves an angle value for mask gradients.
func resolveMaskAngleValue(c ResolvedCandidate) string {
	if c.Arbitrary != "" {
		return c.Arbitrary
	}
	if c.Value == "" {
		return ""
	}
	if isNumeric(c.Value) {
		angle := c.Value
		if c.Negative {
			angle = "-" + angle
		}
		return angle + "deg"
	}
	return ""
}

// makeMaskEdgeFromFn creates a CompileFn for mask-{abbr}-from-* utilities.
// edgeName is the full edge name (top, right, bottom, left) used in CSS variable names.
func makeMaskEdgeFromFn(edgeName, direction string) CompileFn {
	varFromPos := "--tw-mask-" + edgeName + "-from-position"
	varToPos := "--tw-mask-" + edgeName + "-to-position"
	varFromColor := "--tw-mask-" + edgeName + "-from-color"
	varToColor := "--tw-mask-" + edgeName + "-to-color"
	return func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: varFromPos, Value: val},
			{Property: "mask-image", Value: "linear-gradient(" + direction + ", var(" + varFromColor + ", black) var(" + varFromPos + ", 0%), var(" + varToColor + ", transparent) var(" + varToPos + ", 100%))"},
		}
	}
}

// makeMaskEdgeToFn creates a CompileFn for mask-{abbr}-to-* utilities.
// edgeName is the full edge name (top, right, bottom, left) used in CSS variable names.
func makeMaskEdgeToFn(edgeName, direction string) CompileFn {
	varFromPos := "--tw-mask-" + edgeName + "-from-position"
	varToPos := "--tw-mask-" + edgeName + "-to-position"
	varFromColor := "--tw-mask-" + edgeName + "-from-color"
	varToColor := "--tw-mask-" + edgeName + "-to-color"
	return func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: varToPos, Value: val},
			{Property: "mask-image", Value: "linear-gradient(" + direction + ", var(" + varFromColor + ", black) var(" + varFromPos + ", 0%), var(" + varToColor + ", transparent) var(" + varToPos + ", 100%))"},
		}
	}
}

// registerMaskUtilities registers all mask-related utilities.
func registerMaskUtilities(idx *utilityIndex, register func(*UtilityRegistration)) {

	// ===== mask-image: none =====
	register(staticUtility("mask-none", decls("mask-image", "none")))

	// ===== mask-clip =====
	register(staticUtility("mask-clip-border", decls("mask-clip", "border-box")))
	register(staticUtility("mask-clip-padding", decls("mask-clip", "padding-box")))
	register(staticUtility("mask-clip-content", decls("mask-clip", "content-box")))
	register(staticUtility("mask-clip-fill", decls("mask-clip", "fill-box")))
	register(staticUtility("mask-clip-stroke", decls("mask-clip", "stroke-box")))
	register(staticUtility("mask-clip-view", decls("mask-clip", "view-box")))
	register(staticUtility("mask-clip-none", decls("mask-clip", "no-clip")))

	// ===== mask-composite =====
	register(staticUtility("mask-composite-add", decls("mask-composite", "add")))
	register(staticUtility("mask-composite-subtract", decls("mask-composite", "subtract")))
	register(staticUtility("mask-composite-intersect", decls("mask-composite", "intersect")))
	register(staticUtility("mask-composite-exclude", decls("mask-composite", "exclude")))

	// ===== mask-mode =====
	register(staticUtility("mask-alpha", decls("mask-mode", "alpha")))
	register(staticUtility("mask-luminance", decls("mask-mode", "luminance")))
	register(staticUtility("mask-match", decls("mask-mode", "match-source")))

	// ===== mask-origin =====
	register(staticUtility("mask-origin-border", decls("mask-origin", "border-box")))
	register(staticUtility("mask-origin-padding", decls("mask-origin", "padding-box")))
	register(staticUtility("mask-origin-content", decls("mask-origin", "content-box")))
	register(staticUtility("mask-origin-fill", decls("mask-origin", "fill-box")))
	register(staticUtility("mask-origin-stroke", decls("mask-origin", "stroke-box")))
	register(staticUtility("mask-origin-view", decls("mask-origin", "view-box")))

	// ===== mask-position =====
	register(staticUtility("mask-position-center", decls("mask-position", "center")))
	register(staticUtility("mask-position-top", decls("mask-position", "top")))
	register(staticUtility("mask-position-top-right", decls("mask-position", "top right")))
	register(staticUtility("mask-position-right", decls("mask-position", "right")))
	register(staticUtility("mask-position-bottom-right", decls("mask-position", "bottom right")))
	register(staticUtility("mask-position-bottom", decls("mask-position", "bottom")))
	register(staticUtility("mask-position-bottom-left", decls("mask-position", "bottom left")))
	register(staticUtility("mask-position-left", decls("mask-position", "left")))
	register(staticUtility("mask-position-top-left", decls("mask-position", "top left")))

	// ===== mask-repeat =====
	register(staticUtility("mask-repeat", decls("mask-repeat", "repeat")))
	register(staticUtility("mask-no-repeat", decls("mask-repeat", "no-repeat")))
	register(staticUtility("mask-repeat-x", decls("mask-repeat", "repeat-x")))
	register(staticUtility("mask-repeat-y", decls("mask-repeat", "repeat-y")))
	register(staticUtility("mask-repeat-round", decls("mask-repeat", "round")))
	register(staticUtility("mask-repeat-space", decls("mask-repeat", "space")))

	// ===== mask-size =====
	register(staticUtility("mask-size-auto", decls("mask-size", "auto")))
	register(staticUtility("mask-size-cover", decls("mask-size", "cover")))
	register(staticUtility("mask-size-contain", decls("mask-size", "contain")))

	// ===== mask-type =====
	register(staticUtility("mask-type-alpha", decls("mask-type", "alpha")))
	register(staticUtility("mask-type-luminance", decls("mask-type", "luminance")))

	// ===== Edge-based mask gradients =====
	// Each edge direction: t(top), r(right), b(bottom), l(left)
	// CSS variables use full edge names (top, right, bottom, left) per upstream convention.
	type edgeDef struct {
		abbr      string // abbreviated form for utility class name
		full      string // full name for CSS variable names
		direction string // gradient direction
	}
	edges := []edgeDef{
		{"t", "top", "to top"},
		{"r", "right", "to right"},
		{"b", "bottom", "to bottom"},
		{"l", "left", "to left"},
	}
	for _, ed := range edges {
		register(functionalUtility("mask-"+ed.abbr+"-from", makeMaskEdgeFromFn(ed.full, ed.direction)))
		register(functionalUtility("mask-"+ed.abbr+"-to", makeMaskEdgeToFn(ed.full, ed.direction)))
	}

	// mask-x (horizontal: both left and right)
	register(functionalUtility("mask-x-from", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-x-from", Value: val},
			{Property: "mask-image", Value: "linear-gradient(to right, transparent var(--tw-mask-x-from, 0%), black var(--tw-mask-x-to, 100%), black calc(100% - var(--tw-mask-x-to, 100%)), transparent calc(100% - var(--tw-mask-x-from, 0%)))"},
		}
	}))
	register(functionalUtility("mask-x-to", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-x-to", Value: val},
			{Property: "mask-image", Value: "linear-gradient(to right, transparent var(--tw-mask-x-from, 0%), black var(--tw-mask-x-to, 100%), black calc(100% - var(--tw-mask-x-to, 100%)), transparent calc(100% - var(--tw-mask-x-from, 0%)))"},
		}
	}))

	// mask-y (vertical: both top and bottom)
	register(functionalUtility("mask-y-from", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-y-from", Value: val},
			{Property: "mask-image", Value: "linear-gradient(to bottom, transparent var(--tw-mask-y-from, 0%), black var(--tw-mask-y-to, 100%), black calc(100% - var(--tw-mask-y-to, 100%)), transparent calc(100% - var(--tw-mask-y-from, 0%)))"},
		}
	}))
	register(functionalUtility("mask-y-to", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-y-to", Value: val},
			{Property: "mask-image", Value: "linear-gradient(to bottom, transparent var(--tw-mask-y-from, 0%), black var(--tw-mask-y-to, 100%), black calc(100% - var(--tw-mask-y-to, 100%)), transparent calc(100% - var(--tw-mask-y-from, 0%)))"},
		}
	}))

	// ===== Linear gradient masks =====
	// mask-linear-<angle> → sets position and mask-image
	maskLinear := functionalUtility("mask-linear", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskAngleValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-linear-position", Value: val},
			{Property: "mask-image", Value: "linear-gradient(var(--tw-mask-linear-position, 0deg), var(--tw-mask-linear-from-color, black) var(--tw-mask-linear-from-position, 0%), var(--tw-mask-linear-to-color, transparent) var(--tw-mask-linear-to-position, 100%))"},
		}
	})
	maskLinear.Negatable = true
	register(maskLinear)

	// mask-linear-from-* → gradient from stop position
	register(functionalUtility("mask-linear-from", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-linear-from-position", Value: val},
			{Property: "mask-image", Value: "linear-gradient(var(--tw-mask-linear-position, 0deg), var(--tw-mask-linear-from-color, black) var(--tw-mask-linear-from-position, 0%), var(--tw-mask-linear-to-color, transparent) var(--tw-mask-linear-to-position, 100%))"},
		}
	}))

	// mask-linear-to-* → gradient to stop position
	register(functionalUtility("mask-linear-to", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-linear-to-position", Value: val},
			{Property: "mask-image", Value: "linear-gradient(var(--tw-mask-linear-position, 0deg), var(--tw-mask-linear-from-color, black) var(--tw-mask-linear-from-position, 0%), var(--tw-mask-linear-to-color, transparent) var(--tw-mask-linear-to-position, 100%))"},
		}
	}))

	// ===== Radial gradient masks =====
	// mask-radial shape
	register(staticUtility("mask-circle", decls("--tw-mask-radial-shape", "circle")))
	register(staticUtility("mask-ellipse", decls("--tw-mask-radial-shape", "ellipse")))

	// mask-radial size
	register(staticUtility("mask-radial-closest-corner", decls("--tw-mask-radial-size", "closest-corner")))
	register(staticUtility("mask-radial-closest-side", decls("--tw-mask-radial-size", "closest-side")))
	register(staticUtility("mask-radial-farthest-corner", decls("--tw-mask-radial-size", "farthest-corner")))
	register(staticUtility("mask-radial-farthest-side", decls("--tw-mask-radial-size", "farthest-side")))

	// mask-radial-at-* position
	register(staticUtility("mask-radial-at-center", decls("--tw-mask-radial-position", "center")))
	register(staticUtility("mask-radial-at-top", decls("--tw-mask-radial-position", "top")))
	register(staticUtility("mask-radial-at-top-right", decls("--tw-mask-radial-position", "top right")))
	register(staticUtility("mask-radial-at-right", decls("--tw-mask-radial-position", "right")))
	register(staticUtility("mask-radial-at-bottom-right", decls("--tw-mask-radial-position", "bottom right")))
	register(staticUtility("mask-radial-at-bottom", decls("--tw-mask-radial-position", "bottom")))
	register(staticUtility("mask-radial-at-bottom-left", decls("--tw-mask-radial-position", "bottom left")))
	register(staticUtility("mask-radial-at-left", decls("--tw-mask-radial-position", "left")))
	register(staticUtility("mask-radial-at-top-left", decls("--tw-mask-radial-position", "top left")))

	// mask-radial-[...] → arbitrary radial gradient
	register(functionalUtility("mask-radial", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			return decls("mask-image", "radial-gradient("+c.Arbitrary+")")
		}
		// Without arbitrary, emit the composed radial gradient
		return []Declaration{
			{Property: "mask-image", Value: "radial-gradient(var(--tw-mask-radial-shape, ellipse) var(--tw-mask-radial-size, farthest-corner) at var(--tw-mask-radial-position, center), var(--tw-mask-radial-from-color, black) var(--tw-mask-radial-from-position, 0%), var(--tw-mask-radial-to-color, transparent) var(--tw-mask-radial-to-position, 100%))"},
		}
	}))

	// mask-radial-from-* → radial gradient from stop position
	register(functionalUtility("mask-radial-from", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-radial-from-position", Value: val},
			{Property: "mask-image", Value: "radial-gradient(var(--tw-mask-radial-shape, ellipse) var(--tw-mask-radial-size, farthest-corner) at var(--tw-mask-radial-position, center), var(--tw-mask-radial-from-color, black) var(--tw-mask-radial-from-position, 0%), var(--tw-mask-radial-to-color, transparent) var(--tw-mask-radial-to-position, 100%))"},
		}
	}))

	// mask-radial-to-* → radial gradient to stop position
	register(functionalUtility("mask-radial-to", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-radial-to-position", Value: val},
			{Property: "mask-image", Value: "radial-gradient(var(--tw-mask-radial-shape, ellipse) var(--tw-mask-radial-size, farthest-corner) at var(--tw-mask-radial-position, center), var(--tw-mask-radial-from-color, black) var(--tw-mask-radial-from-position, 0%), var(--tw-mask-radial-to-color, transparent) var(--tw-mask-radial-to-position, 100%))"},
		}
	}))

	// ===== Conic gradient masks =====
	// mask-conic-<angle> → sets position (angle) and mask-image
	// mask-conic-[...] → arbitrary conic gradient
	maskConic := functionalUtility("mask-conic", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			return decls("mask-image", "conic-gradient("+c.Arbitrary+")")
		}
		if c.Value == "" {
			return nil
		}
		if isNumeric(c.Value) {
			angle := c.Value
			if c.Negative {
				angle = "-" + angle
			}
			return []Declaration{
				{Property: "--tw-mask-conic-position", Value: angle + "deg"},
				{Property: "mask-image", Value: "conic-gradient(from var(--tw-mask-conic-position, 0deg) at center, var(--tw-mask-conic-from-color, black) var(--tw-mask-conic-from-position, 0%), var(--tw-mask-conic-to-color, transparent) var(--tw-mask-conic-to-position, 100%))"},
			}
		}
		return nil
	})
	maskConic.Negatable = true
	register(maskConic)

	// mask-conic-from-* → conic gradient from stop position
	register(functionalUtility("mask-conic-from", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-conic-from-position", Value: val},
			{Property: "mask-image", Value: "conic-gradient(from var(--tw-mask-conic-position, 0deg) at center, var(--tw-mask-conic-from-color, black) var(--tw-mask-conic-from-position, 0%), var(--tw-mask-conic-to-color, transparent) var(--tw-mask-conic-to-position, 100%))"},
		}
	}))

	// mask-conic-to-* → conic gradient to stop position
	register(functionalUtility("mask-conic-to", func(c ResolvedCandidate) []Declaration {
		val := resolveMaskStopValue(c)
		if val == "" {
			return nil
		}
		return []Declaration{
			{Property: "--tw-mask-conic-to-position", Value: val},
			{Property: "mask-image", Value: "conic-gradient(from var(--tw-mask-conic-position, 0deg) at center, var(--tw-mask-conic-from-color, black) var(--tw-mask-conic-from-position, 0%), var(--tw-mask-conic-to-color, transparent) var(--tw-mask-conic-to-position, 100%))"},
		}
	}))

	// ===== Arbitrary mask-image =====
	// mask-[url(/img/mask.png)] → mask-image: url(/img/mask.png)
	register(functionalUtility("mask", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			return decls("mask-image", c.Arbitrary)
		}
		return nil
	}))
}
