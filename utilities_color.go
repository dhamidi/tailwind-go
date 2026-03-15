package tailwind

import "strings"

// resolveColorValue resolves a CSS color value for a candidate.
// It tries special keywords, arbitrary values, and theme namespaces in order.
// The opacity modifier is applied when present.
func resolveColorValue(c ResolvedCandidate, themeKeys ...string) string {
	// Special keywords
	switch c.Value {
	case "current":
		return applyColorModifier("currentcolor", c)
	case "inherit":
		return "inherit"
	case "transparent":
		return applyColorModifier("transparent", c)
	}

	// Arbitrary value
	if c.Arbitrary != "" {
		val := c.Arbitrary
		if c.Modifier != "" {
			val = applyModifier(val, c.Modifier, c.Theme)
		}
		return val
	}

	if c.Value == "" {
		return ""
	}

	// Theme resolution: try each namespace in order
	for _, ns := range themeKeys {
		if _, ok := c.Theme.Resolve(ns, c.Value); ok {
			varRef := "var(--" + ns + "-" + c.Value + ")"
			if c.Modifier != "" {
				// With opacity modifier, use color-mix with the CSS variable reference.
				return applyModifier(varRef, c.Modifier, c.Theme)
			}
			return varRef
		}
	}
	return ""
}

// applyColorModifier applies the opacity modifier to a resolved color value.
func applyColorModifier(val string, c ResolvedCandidate) string {
	if c.Modifier != "" {
		return applyModifier(val, c.Modifier, c.Theme)
	}
	return val
}

// resolveGradientInterpolation maps a modifier string to a CSS color interpolation clause.
// Returns an empty string if no modifier is specified (caller decides default).
func resolveGradientInterpolation(modifier string) string {
	switch modifier {
	case "srgb":
		return " in srgb"
	case "srgb-linear":
		return " in srgb-linear"
	case "hsl":
		return " in hsl"
	case "hwb":
		return " in hwb"
	case "oklab":
		return " in oklab"
	case "oklch":
		return " in oklch"
	case "lab":
		return " in lab"
	case "lch":
		return " in lch"
	case "longer-hue":
		return " in oklch longer hue"
	case "shorter-hue":
		return " in oklch shorter hue"
	case "increasing-hue":
		return " in oklch increasing hue"
	case "decreasing-hue":
		return " in oklch decreasing hue"
	default:
		return ""
	}
}

// hasActiveModifier returns true if the candidate has an opacity modifier that
// will produce a color-mix() wrapper (i.e., modifier is present and not 100%).
func hasActiveModifier(c ResolvedCandidate) bool {
	if c.Modifier == "" {
		return false
	}
	// Check if the modifier resolves to 100% (identity - no wrapping)
	if isNumeric(c.Modifier) && c.Modifier == "100" {
		return false
	}
	return true
}

// resolveGradientStopPosition checks if a candidate value is a percentage position
// for gradient stops. Returns the CSS value and true if it's a position, or empty and false.
func resolveGradientStopPosition(c ResolvedCandidate) (string, bool) {
	if c.Arbitrary != "" {
		// Arbitrary values with percentage type hint
		if c.TypeHint == "percentage" || c.TypeHint == "length" {
			return c.Arbitrary, true
		}
		return "", false
	}
	if c.Value == "" {
		return "", false
	}
	// Check for percentage: e.g., "10%" from from-10%
	if strings.HasSuffix(c.Value, "%") {
		numPart := c.Value[:len(c.Value)-1]
		if isNumeric(numPart) {
			return c.Value, true
		}
	}
	return "", false
}

// makeColorCompileFn creates a CompileFn for a simple single-property color utility.
func makeColorCompileFn(property string, themeKeys ...string) CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		val := resolveColorValue(c, themeKeys...)
		if val == "" {
			return nil
		}
		return decls(property, val)
	}
}

// makeBorderColorCompileFn creates a CompileFn for a multi-property border color utility
// (e.g., border-x sets both left and right, border-y sets both top and bottom).
func makeBorderColorCompileFn(properties ...string) CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		val := resolveColorValue(c, "border-color", "color")
		if val == "" {
			return nil
		}
		pairs := make([]string, 0, len(properties)*2)
		for _, p := range properties {
			pairs = append(pairs, p, val)
		}
		return decls(pairs...)
	}
}

// registerColorUtilities registers all color-related functional utilities.
func registerColorUtilities(idx *utilityIndex, register func(*UtilityRegistration)) {
	// === Simple color utilities ===
	register(colorUtility("accent", makeColorCompileFn("accent-color", "accent-color", "color")))
	register(colorUtility("caret", makeColorCompileFn("caret-color", "caret-color", "color")))
	register(colorUtility("fill", makeColorCompileFn("fill", "fill", "color")))
	register(colorUtility("decoration", makeColorCompileFn("text-decoration-color", "text-decoration-color", "color")))

	// === Outline color ===
	register(colorUtility("outline", makeColorCompileFn("outline-color", "outline-color", "color")))

	// === Ring color ===
	register(colorUtility("ring", makeColorCompileFn("--tw-ring-color", "ring-color", "color")))
	register(colorUtility("ring-offset", makeColorCompileFn("--tw-ring-offset-color", "ring-offset-color", "color")))

	// === Inset Ring (color and arbitrary width) ===
	register(functionalUtility("inset-ring", makeInsetRingCompileFn()))

	// === Border color ===
	register(colorUtility("border", makeColorCompileFn("border-color", "border-color", "color")))
	register(colorUtility("border-t", makeColorCompileFn("border-top-color", "border-color", "color")))
	register(colorUtility("border-r", makeColorCompileFn("border-right-color", "border-color", "color")))
	register(colorUtility("border-b", makeColorCompileFn("border-bottom-color", "border-color", "color")))
	register(colorUtility("border-l", makeColorCompileFn("border-left-color", "border-color", "color")))
	register(colorUtility("border-s", makeColorCompileFn("border-inline-start-color", "border-color", "color")))
	register(colorUtility("border-e", makeColorCompileFn("border-inline-end-color", "border-color", "color")))
	register(colorUtility("border-bs", makeColorCompileFn("border-block-start-color", "border-color", "color")))
	register(colorUtility("border-be", makeColorCompileFn("border-block-end-color", "border-color", "color")))
	register(colorUtility("border-x", makeBorderColorCompileFn("border-left-color", "border-right-color")))
	register(colorUtility("border-y", makeBorderColorCompileFn("border-top-color", "border-bottom-color")))

	// === Divide color (targets children) ===
	divideReg := colorUtility("divide", makeColorCompileFn("border-color", "divide-color", "color"))
	divideReg.Selector = "> :not(:last-child)"
	register(divideReg)

	// === Placeholder color (targets ::placeholder pseudo-element) ===
	placeholderReg := colorUtility("placeholder", makeColorCompileFn("color", "placeholder-color", "color"))
	placeholderReg.Selector = "&::placeholder"
	register(placeholderReg)

	// === Background color ===
	register(colorUtility("bg", makeBgCompileFn()))

	// === Text color / font-size ===
	register(functionalUtility("text", makeTextCompileFn()))

	// === Stroke (color and width) ===
	register(colorUtility("stroke", makeStrokeCompileFn()))

	// === Shadow (value and color) ===
	register(functionalUtility("shadow", makeShadowCompileFn()))

	// === Text shadow (value and color) ===
	register(functionalUtility("text-shadow", makeTextShadowCompileFn()))

	// === Gradient color stops ===
	register(colorUtility("from", func(c ResolvedCandidate) []Declaration {
		// Check for position value first (e.g., from-10%)
		if pos, ok := resolveGradientStopPosition(c); ok {
			return []Declaration{
				{Property: "--tw-gradient-from-position", Value: pos},
			}
		}
		val := resolveColorValue(c, "color")
		if val == "" {
			return nil
		}
		decls := []Declaration{
			{Property: "--tw-gradient-from", Value: val},
		}
		if !hasActiveModifier(c) {
			decls = append(decls, Declaration{Property: "--tw-gradient-stops", Value: "var(--tw-gradient-via-stops, var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))"})
		}
		return decls
	}))

	register(colorUtility("via", func(c ResolvedCandidate) []Declaration {
		// Check for position value first (e.g., via-30%)
		if pos, ok := resolveGradientStopPosition(c); ok {
			return []Declaration{
				{Property: "--tw-gradient-via-position", Value: pos},
			}
		}
		val := resolveColorValue(c, "color")
		if val == "" {
			return nil
		}
		decls := []Declaration{
			{Property: "--tw-gradient-via", Value: val},
		}
		if !hasActiveModifier(c) {
			decls = append(decls,
				Declaration{Property: "--tw-gradient-via-stops", Value: "var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-via) var(--tw-gradient-via-position), var(--tw-gradient-to) var(--tw-gradient-to-position)"},
				Declaration{Property: "--tw-gradient-stops", Value: "var(--tw-gradient-via-stops)"},
			)
		}
		return decls
	}))

	register(colorUtility("to", func(c ResolvedCandidate) []Declaration {
		// Check for position value first (e.g., to-90%)
		if pos, ok := resolveGradientStopPosition(c); ok {
			return []Declaration{
				{Property: "--tw-gradient-to-position", Value: pos},
			}
		}
		val := resolveColorValue(c, "color")
		if val == "" {
			return nil
		}
		decls := []Declaration{
			{Property: "--tw-gradient-to", Value: val},
		}
		if !hasActiveModifier(c) {
			decls = append(decls, Declaration{Property: "--tw-gradient-stops", Value: "var(--tw-gradient-via-stops, var(--tw-gradient-position), var(--tw-gradient-from) var(--tw-gradient-from-position), var(--tw-gradient-to) var(--tw-gradient-to-position))"})
		}
		return decls
	}))
}

// makeBgCompileFn creates the compile function for bg-* utilities.
// Handles color, and for arbitrary values also image inference.
func makeBgCompileFn() CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		// Type hint handling for arbitrary values
		if c.TypeHint != "" {
			switch c.TypeHint {
			case "color":
				val := resolveColorValue(c, "background-color", "color")
				if val == "" {
					return nil
				}
				return decls("background-color", val)
			case "image", "url":
				if c.Arbitrary != "" {
					return decls("background-image", c.Arbitrary)
				}
			case "position":
				if c.Arbitrary != "" {
					return decls("background-position", c.Arbitrary)
				}
			case "size", "length", "bg-size":
				if c.Arbitrary != "" {
					return decls("background-size", c.Arbitrary)
				}
			}
			return nil
		}

		// Arbitrary value: infer type
		if c.Arbitrary != "" {
			if looksLikeImage(c.Arbitrary) {
				return decls("background-image", c.Arbitrary)
			}
			// Default to color
			val := c.Arbitrary
			if c.Modifier != "" {
				val = applyModifier(val, c.Modifier, c.Theme)
			}
			return decls("background-color", val)
		}

		// Named value: try color themes
		val := resolveColorValue(c, "background-color", "color")
		if val != "" {
			return decls("background-color", val)
		}
		return nil
	}
}

// makeTextCompileFn creates the compile function for text-* utilities.
// Handles both font-size and color disambiguation.
// Font-size is tried first; color is tried second. Both can resolve
// independently (matching the CSS priority chain behavior).
func makeTextCompileFn() CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		var result []Declaration

		// Font-size property group
		if fsVal := resolveTextFontSize(c); fsVal != "" {
			result = append(result, Declaration{Property: "font-size", Value: fsVal})
		}

		// Color property group
		if colVal := resolveTextColorValue(c); colVal != "" {
			result = append(result, Declaration{Property: "color", Value: colVal})
		}

		if len(result) == 0 {
			return nil
		}
		return result
	}
}

// resolveTextFontSize resolves the font-size portion of text-*.
// Matches the CSS behavior: --value(--font-size, length, percentage).
func resolveTextFontSize(c ResolvedCandidate) string {
	// Type hint: skip if explicitly color
	if c.TypeHint == "color" {
		return ""
	}

	// Arbitrary value (accepted for font-size unless type hint is color)
	if c.Arbitrary != "" {
		return c.Arbitrary
	}

	if c.Value == "" {
		return ""
	}

	// Try --font-size theme namespace
	if val, ok := c.Theme.Resolve("font-size", c.Value); ok {
		return val
	}

	// Type-based fallback: percentage for numeric values
	if isNumeric(c.Value) {
		return c.Value + "%"
	}

	return ""
}

// resolveTextColorValue resolves the color portion of text-*.
// Matches the CSS behavior: --value(--color) then --value(color).
func resolveTextColorValue(c ResolvedCandidate) string {
	// Type hint: skip if not color
	if c.TypeHint != "" && c.TypeHint != "color" {
		return ""
	}

	// Arbitrary value
	if c.Arbitrary != "" {
		val := c.Arbitrary
		if c.Modifier != "" {
			val = applyModifier(val, c.Modifier, c.Theme)
		}
		return val
	}

	// Special keywords (resolved via the no-namespace --value(color) alternative)
	switch c.Value {
	case "current":
		return applyColorModifier("currentcolor", c)
	case "inherit":
		return "inherit"
	case "transparent":
		return applyColorModifier("transparent", c)
	}

	if c.Value == "" {
		return ""
	}

	// Theme resolution: --text-color, then --color
	for _, ns := range []string{"text-color", "color"} {
		if _, ok := c.Theme.Resolve(ns, c.Value); ok {
			varRef := "var(--" + ns + "-" + c.Value + ")"
			if c.Modifier != "" {
				return applyModifier(varRef, c.Modifier, c.Theme)
			}
			return varRef
		}
	}
	return ""
}

// makeStrokeCompileFn creates the compile function for stroke-*.
// Handles both color (stroke property) and width (stroke-width property).
func makeStrokeCompileFn() CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		// Type hint dispatching
		if c.TypeHint == "color" {
			val := resolveColorValue(c, "stroke", "color")
			if val == "" {
				return nil
			}
			return decls("stroke", val)
		}
		if c.TypeHint == "number" || c.TypeHint == "length" {
			if c.Arbitrary != "" {
				return decls("stroke-width", c.Arbitrary)
			}
			if isNumeric(c.Value) {
				return decls("stroke-width", c.Value)
			}
			return nil
		}

		// Try color first (matching current CSS behavior)
		val := resolveColorValue(c, "stroke", "color")
		if val != "" {
			return decls("stroke", val)
		}

		// Fallback: arbitrary → stroke property, numeric → stroke-width
		if c.Arbitrary != "" {
			return decls("stroke", c.Arbitrary)
		}
		if isNumeric(c.Value) {
			return decls("stroke-width", c.Value)
		}
		return nil
	}
}

// makeInsetRingCompileFn creates the compile function for inset-ring-*.
// Handles both color (--tw-inset-ring-color) and arbitrary width (--tw-inset-ring-shadow).
func makeInsetRingCompileFn() CompileFn {
	const boxShadow = "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)"
	return func(c ResolvedCandidate) []Declaration {
		// Type hint dispatching
		if c.TypeHint == "color" {
			val := resolveColorValue(c, "inset-ring-color", "color")
			if val == "" {
				return nil
			}
			return decls("--tw-inset-ring-color", val)
		}
		if c.TypeHint == "length" || c.TypeHint == "number" {
			if c.Arbitrary != "" {
				return decls(
					"--tw-inset-ring-shadow", "inset 0 0 0 "+c.Arbitrary+" var(--tw-inset-ring-color, currentcolor)",
					"box-shadow", boxShadow,
				)
			}
			return nil
		}

		// Arbitrary value without type hint: infer length vs color
		if c.Arbitrary != "" {
			if looksLikeCSSLength(c.Arbitrary) {
				return decls(
					"--tw-inset-ring-shadow", "inset 0 0 0 "+c.Arbitrary+" var(--tw-inset-ring-color, currentcolor)",
					"box-shadow", boxShadow,
				)
			}
			// Treat as color
			val := c.Arbitrary
			if c.Modifier != "" {
				val = applyModifier(val, c.Modifier, c.Theme)
			}
			return decls("--tw-inset-ring-color", val)
		}

		// Named value: try color resolution
		val := resolveColorValue(c, "inset-ring-color", "color")
		if val != "" {
			return decls("--tw-inset-ring-color", val)
		}
		return nil
	}
}

// looksLikeCSSLength returns true if the value appears to be a CSS length
// (e.g., "3px", "0.5rem", "1em") rather than a color.
func looksLikeCSSLength(val string) bool {
	if val == "" {
		return false
	}
	// CSS functions like calc(), var() are treated as lengths
	if strings.HasPrefix(val, "calc(") || strings.HasPrefix(val, "var(") {
		return true
	}
	// Starts with a digit or decimal point → likely a length
	return (val[0] >= '0' && val[0] <= '9') || val[0] == '.'
}

// wrapShadowAlpha wraps a color value with --tw-shadow-alpha using color-mix.
func wrapShadowAlpha(colorValue string) string {
	return "color-mix(in oklab, " + colorValue + " var(--tw-shadow-alpha), transparent)"
}

// makeShadowCompileFn creates the compile function for shadow-*.
// Handles both shadow values (box-shadow) and shadow color (--tw-shadow-color).
func makeShadowCompileFn() CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		// Type hint: color → shadow color
		if c.TypeHint == "color" {
			val := resolveColorValue(c, "shadow-color", "color")
			if val == "" {
				return nil
			}
			return decls("--tw-shadow-color", wrapShadowAlpha(val))
		}

		// Arbitrary value without type hint → box-shadow (matching current CSS)
		if c.Arbitrary != "" {
			return decls("box-shadow", c.Arbitrary)
		}

		if c.Value == "" {
			return nil
		}

		// Named: try --shadow namespace first (box-shadow value)
		if val, ok := c.Theme.Resolve("shadow", c.Value); ok {
			return decls("box-shadow", val)
		}

		// Named: try shadow color themes
		val := resolveColorValue(c, "shadow-color", "color")
		if val != "" {
			return decls("--tw-shadow-color", wrapShadowAlpha(val))
		}

		return nil
	}
}

// makeTextShadowCompileFn creates the compile function for text-shadow-* utilities.
// Handles both text-shadow values (from theme) and text-shadow color.
func makeTextShadowCompileFn() CompileFn {
	return func(c ResolvedCandidate) []Declaration {
		// Type hint: color → text shadow color
		if c.TypeHint == "color" {
			val := resolveColorValue(c, "color")
			if val == "" {
				return nil
			}
			return decls("--tw-text-shadow-color", val, "text-shadow", "0px 1px 0px var(--tw-text-shadow-color)")
		}

		// Arbitrary value without type hint → text-shadow
		if c.Arbitrary != "" {
			return decls("text-shadow", c.Arbitrary)
		}

		if c.Value == "" {
			return nil
		}

		// Named: try --text-shadow namespace first (text-shadow value)
		if val, ok := c.Theme.Resolve("text-shadow", c.Value); ok {
			return decls("text-shadow", val)
		}

		// Named: try color themes for text-shadow color
		val := resolveColorValue(c, "color")
		if val != "" {
			return decls("--tw-text-shadow-color", val, "text-shadow", "0px 1px 0px var(--tw-text-shadow-color)")
		}

		return nil
	}
}

// looksLikeImage returns true if the CSS value looks like a background-image value.
func looksLikeImage(val string) bool {
	lower := strings.ToLower(val)
	return strings.HasPrefix(lower, "url(") ||
		strings.HasPrefix(lower, "linear-gradient(") ||
		strings.HasPrefix(lower, "radial-gradient(") ||
		strings.HasPrefix(lower, "conic-gradient(") ||
		strings.HasPrefix(lower, "repeating-linear-gradient(") ||
		strings.HasPrefix(lower, "repeating-radial-gradient(") ||
		strings.HasPrefix(lower, "repeating-conic-gradient(") ||
		strings.HasPrefix(lower, "image-set(")
}
