package tailwind

import "strings"

// cssUtility creates a UtilityRegistration that delegates to the CSS-based
// resolveDeclarations logic. This ensures functional utilities ported from
// utilities.css produce identical output. The declPairs are alternating
// property/value strings where values may contain --value() placeholders.
func cssUtility(name string, declarations []Declaration) *UtilityRegistration {
	ud := &UtilityDef{
		Pattern:      name,
		Static:       false,
		Declarations: declarations,
	}
	return &UtilityRegistration{
		Name: name,
		Kind: "functional",
		CompileFn: func(c ResolvedCandidate) []Declaration {
			valueStr := c.Value
			if c.Fraction != "" {
				valueStr = c.Fraction
			}
			pc := ParsedClass{
				Arbitrary: c.Arbitrary,
				Modifier:  c.Modifier,
				Negative:  c.Negative,
				TypeHint:  c.TypeHint,
			}
			return resolveDeclarations(ud, valueStr, pc, c.Theme)
		},
	}
}

// cssUtilityWithSelector creates a cssUtility with a child selector suffix.
func cssUtilityWithSelector(name, selector string, declarations []Declaration) *UtilityRegistration {
	reg := cssUtility(name, declarations)
	reg.Selector = selector
	return reg
}

// registerFunctionalUtilities registers all functional (dynamic) utilities
// that were previously defined as @utility blocks in utilities.css.
// These use --value() resolution via resolveDeclarations for identical output.
func registerFunctionalUtilities(idx *utilityIndex, register func(*UtilityRegistration)) {

	// ===== Aspect Ratio =====
	// Custom handling: fractions like 16/9 produce "16 / 9" (a ratio),
	// not a percentage. The parser treats N/M where N >= M as an opacity
	// modifier, so we recombine value + modifier when both are numeric.
	register(functionalUtility("aspect", func(c ResolvedCandidate) []Declaration {
		// Arbitrary values pass through directly.
		if c.Arbitrary != "" {
			return decls("aspect-ratio", c.Arbitrary)
		}

		// Theme lookup: aspect-video, aspect-square, etc.
		if c.Value != "" && c.Theme != nil {
			if resolved, ok := c.Theme.Resolve("aspect", c.Value); ok {
				return decls("aspect-ratio", resolved)
			}
		}

		// Fraction parsed as value (numerator < denominator, e.g., aspect-1/2).
		if c.Fraction != "" {
			parts := strings.SplitN(c.Fraction, "/", 2)
			if len(parts) == 2 {
				return decls("aspect-ratio", parts[0]+" / "+parts[1])
			}
		}

		// Ratio where numerator >= denominator (e.g., aspect-16/9):
		// parser treats 16/9 as value="16" modifier="9".
		if c.Value != "" && c.Modifier != "" && isNumeric(c.Value) && isNumeric(c.Modifier) {
			return decls("aspect-ratio", c.Value+" / "+c.Modifier)
		}

		// Bare numeric value (e.g., aspect-2 → 2).
		if c.Value != "" && isNumeric(c.Value) {
			return decls("aspect-ratio", c.Value)
		}

		return nil
	}))

	// ===== Columns =====
	register(cssUtility("columns", decls(
		"columns", "--value(--container)",
		"columns", "--value(integer)",
		"columns", "--value(any)",
	)))

	// ===== Inset (positioning) =====
	register(negatableCSSUtility("inset", decls(
		"inset", "--value(--spacing)",
		"inset", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("inset-x", decls(
		"inset-inline", "--value(--spacing)",
		"inset-inline", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("inset-y", decls(
		"inset-block", "--value(--spacing)",
		"inset-block", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("top", decls(
		"top", "--value(--spacing)",
		"top", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("right", decls(
		"right", "--value(--spacing)",
		"right", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("bottom", decls(
		"bottom", "--value(--spacing)",
		"bottom", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("left", decls(
		"left", "--value(--spacing)",
		"left", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("start", decls(
		"inset-inline-start", "--value(--spacing)",
		"inset-inline-start", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("end", decls(
		"inset-inline-end", "--value(--spacing)",
		"inset-inline-end", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("inset-bs", decls(
		"inset-block-start", "--value(--spacing)",
		"inset-block-start", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("inset-be", decls(
		"inset-block-end", "--value(--spacing)",
		"inset-block-end", "--value(length, percentage)",
	)))

	// ===== Z-index =====
	register(negatableCSSUtility("z", decls(
		"z-index", "--value(--z-index)",
		"z-index", "--value(integer)",
	)))

	// ===== Order =====
	register(negatableCSSUtility("order", decls(
		"order", "--value(integer)",
	)))

	// ===== Flex Basis =====
	register(cssUtility("basis", decls(
		"flex-basis", "--value(--spacing)",
		"flex-basis", "--value(--width)",
		"flex-basis", "--value(length, percentage)",
	)))

	// ===== Flexbox =====
	register(cssUtility("flex", decls(
		"flex", "--value(--flex)",
		"flex", "--value(any)",
	)))
	register(cssUtility("shrink", decls(
		"flex-shrink", "--value(number)",
	)))
	register(cssUtility("grow", decls(
		"flex-grow", "--value(number)",
	)))

	// ===== Grid =====
	register(cssUtility("grid-cols", decls(
		"grid-template-columns", "--value(--grid-template-columns)",
		"grid-template-columns", "--value(any)",
	)))
	register(cssUtility("grid-rows", decls(
		"grid-template-rows", "--value(--grid-template-rows)",
		"grid-template-rows", "--value(any)",
	)))
	register(cssUtility("col-span", decls(
		"grid-column", "span --value() / span --value()",
	)))
	register(cssUtility("row-span", decls(
		"grid-row", "span --value() / span --value()",
	)))
	register(cssUtility("col-start", decls(
		"grid-column-start", "--value(integer)",
	)))
	register(cssUtility("col-end", decls(
		"grid-column-end", "--value(integer)",
	)))
	register(cssUtility("row-start", decls(
		"grid-row-start", "--value(integer)",
	)))
	register(cssUtility("row-end", decls(
		"grid-row-end", "--value(integer)",
	)))

	// ===== Gap =====
	register(cssUtility("gap", decls(
		"gap", "--value(--spacing)",
		"gap", "--value(length, percentage)",
	)))
	register(cssUtility("gap-x", decls(
		"column-gap", "--value(--spacing)",
		"column-gap", "--value(length, percentage)",
	)))
	register(cssUtility("gap-y", decls(
		"row-gap", "--value(--spacing)",
		"row-gap", "--value(length, percentage)",
	)))

	// ===== Padding =====
	register(cssUtility("p", decls(
		"padding", "--value(--spacing)",
		"padding", "--value(length, percentage)",
	)))
	register(cssUtility("px", decls(
		"padding-inline", "--value(--spacing)",
		"padding-inline", "--value(length, percentage)",
	)))
	register(cssUtility("py", decls(
		"padding-block", "--value(--spacing)",
		"padding-block", "--value(length, percentage)",
	)))
	register(cssUtility("ps", decls(
		"padding-inline-start", "--value(--spacing)",
		"padding-inline-start", "--value(length, percentage)",
	)))
	register(cssUtility("pe", decls(
		"padding-inline-end", "--value(--spacing)",
		"padding-inline-end", "--value(length, percentage)",
	)))
	register(cssUtility("pbs", decls(
		"padding-block-start", "--value(--spacing)",
		"padding-block-start", "--value(length, percentage)",
	)))
	register(cssUtility("pbe", decls(
		"padding-block-end", "--value(--spacing)",
		"padding-block-end", "--value(length, percentage)",
	)))
	register(cssUtility("pt", decls(
		"padding-top", "--value(--spacing)",
		"padding-top", "--value(length, percentage)",
	)))
	register(cssUtility("pr", decls(
		"padding-right", "--value(--spacing)",
		"padding-right", "--value(length, percentage)",
	)))
	register(cssUtility("pb", decls(
		"padding-bottom", "--value(--spacing)",
		"padding-bottom", "--value(length, percentage)",
	)))
	register(cssUtility("pl", decls(
		"padding-left", "--value(--spacing)",
		"padding-left", "--value(length, percentage)",
	)))

	// ===== Margin =====
	register(negatableCSSUtility("m", decls(
		"margin", "--value(--spacing)",
		"margin", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("mx", decls(
		"margin-inline", "--value(--spacing)",
		"margin-inline", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("my", decls(
		"margin-block", "--value(--spacing)",
		"margin-block", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("ms", decls(
		"margin-inline-start", "--value(--spacing)",
		"margin-inline-start", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("me", decls(
		"margin-inline-end", "--value(--spacing)",
		"margin-inline-end", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("mbs", decls(
		"margin-block-start", "--value(--spacing)",
		"margin-block-start", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("mbe", decls(
		"margin-block-end", "--value(--spacing)",
		"margin-block-end", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("mt", decls(
		"margin-top", "--value(--spacing)",
		"margin-top", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("mr", decls(
		"margin-right", "--value(--spacing)",
		"margin-right", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("mb", decls(
		"margin-bottom", "--value(--spacing)",
		"margin-bottom", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("ml", decls(
		"margin-left", "--value(--spacing)",
		"margin-left", "--value(length, percentage)",
	)))

	// ===== Space Between =====
	childSel := "> :not(:last-child)"
	spaceX := cssUtilityWithSelector("space-x", childSel, decls(
		"--tw-space-x-reverse", "0",
		"margin-inline-end", "calc(--value(--spacing) * calc(1 - var(--tw-space-x-reverse)))",
		"margin-inline-start", "calc(--value(--spacing) * var(--tw-space-x-reverse))",
		"margin-inline-end", "calc(--value(length, percentage) * calc(1 - var(--tw-space-x-reverse)))",
		"margin-inline-start", "calc(--value(length, percentage) * var(--tw-space-x-reverse))",
	))
	spaceX.Negatable = true
	register(spaceX)
	spaceY := cssUtilityWithSelector("space-y", childSel, decls(
		"--tw-space-y-reverse", "0",
		"margin-block-end", "calc(--value(--spacing) * calc(1 - var(--tw-space-y-reverse)))",
		"margin-block-start", "calc(--value(--spacing) * var(--tw-space-y-reverse))",
		"margin-block-end", "calc(--value(length, percentage) * calc(1 - var(--tw-space-y-reverse)))",
		"margin-block-start", "calc(--value(length, percentage) * var(--tw-space-y-reverse))",
	))
	spaceY.Negatable = true
	register(spaceY)

	// ===== Width =====
	register(cssUtility("w", decls(
		"width", "--value(--spacing)",
		"width", "--value(--width)",
		"width", "--value(--container)",
		"width", "--value(length, percentage)",
	)))
	register(cssUtility("min-w", decls(
		"min-width", "--value(--spacing)",
		"min-width", "--value(--width)",
		"min-width", "--value(--container)",
		"min-width", "--value(length, percentage)",
	)))
	register(cssUtility("max-w", decls(
		"max-width", "--value(--spacing)",
		"max-width", "--value(--width)",
		"max-width", "--value(--container)",
		"max-width", "--value(length, percentage)",
	)))

	// ===== Height =====
	register(cssUtility("h", decls(
		"height", "--value(--spacing)",
		"height", "--value(--height)",
		"height", "--value(length, percentage)",
	)))
	register(cssUtility("min-h", decls(
		"min-height", "--value(--spacing)",
		"min-height", "--value(--height)",
		"min-height", "--value(length, percentage)",
	)))
	register(cssUtility("max-h", decls(
		"max-height", "--value(--spacing)",
		"max-height", "--value(--height)",
		"max-height", "--value(length, percentage)",
	)))

	// ===== Inline Size =====
	register(cssUtility("inline", decls(
		"inline-size", "--value(--spacing)",
		"inline-size", "--value(--width)",
		"inline-size", "--value(--container)",
		"inline-size", "--value(length, percentage)",
	)))
	register(cssUtility("min-inline", decls(
		"min-inline-size", "--value(--spacing)",
		"min-inline-size", "--value(--width)",
		"min-inline-size", "--value(--container)",
		"min-inline-size", "--value(length, percentage)",
	)))
	register(cssUtility("max-inline", decls(
		"max-inline-size", "--value(--spacing)",
		"max-inline-size", "--value(--width)",
		"max-inline-size", "--value(--container)",
		"max-inline-size", "--value(length, percentage)",
	)))

	// ===== Block Size =====
	register(cssUtility("block", decls(
		"block-size", "--value(--spacing)",
		"block-size", "--value(--height)",
		"block-size", "--value(length, percentage)",
	)))
	register(cssUtility("min-block", decls(
		"min-block-size", "--value(--spacing)",
		"min-block-size", "--value(--height)",
		"min-block-size", "--value(length, percentage)",
	)))
	register(cssUtility("max-block", decls(
		"max-block-size", "--value(--spacing)",
		"max-block-size", "--value(--height)",
		"max-block-size", "--value(length, percentage)",
	)))

	// ===== Size =====
	register(cssUtility("size", decls(
		"width", "--value(--spacing)",
		"width", "--value(--size)",
		"width", "--value(length, percentage)",
		"height", "--value(--spacing)",
		"height", "--value(--size)",
		"height", "--value(length, percentage)",
	)))

	// ===== Font Family =====
	// These are static utilities (no *) that reference theme variables.
	register(staticUtility("font-sans", decls("font-family", "var(--font-sans)")))
	register(staticUtility("font-serif", decls("font-family", "var(--font-serif)")))
	register(staticUtility("font-mono", decls("font-family", "var(--font-mono)")))

	// ===== Font Weight =====
	register(cssUtility("font", decls(
		"font-weight", "--value(--font-weight)",
		"font-weight", "--value(number)",
	)))

	// ===== Text Indent =====
	register(negatableCSSUtility("indent", decls(
		"text-indent", "--value(--spacing)",
		"text-indent", "--value(length, percentage)",
	)))

	// ===== Line Clamp =====
	register(cssUtility("line-clamp", decls(
		"overflow", "hidden",
		"display", "-webkit-box",
		"-webkit-box-orient", "vertical",
		"-webkit-line-clamp", "--value(integer)",
	)))

	// ===== Line Height =====
	register(cssUtility("leading", decls(
		"line-height", "--value(--leading)",
		"line-height", "--value(--line-height)",
		"line-height", "--value(--spacing)",
		"line-height", "--value(length, number)",
	)))

	// ===== Letter Spacing =====
	register(negatableCSSUtility("tracking", decls(
		"letter-spacing", "--value(--tracking)",
		"letter-spacing", "--value(--letter-spacing)",
		"letter-spacing", "--value(length)",
	)))

	// ===== Border Spacing =====
	register(cssUtility("border-spacing", decls(
		"border-spacing", "--value(--spacing)",
		"border-spacing", "--value(length)",
	)))
	register(cssUtility("border-spacing-x", decls(
		"--tw-border-spacing-x", "--value(--spacing)",
		"border-spacing", "var(--tw-border-spacing-x) var(--tw-border-spacing-y)",
	)))
	register(cssUtility("border-spacing-y", decls(
		"--tw-border-spacing-y", "--value(--spacing)",
		"border-spacing", "var(--tw-border-spacing-x) var(--tw-border-spacing-y)",
	)))

	// ===== Border Radius =====
	register(cssUtility("rounded", decls(
		"border-radius", "--value(--radius)",
		"border-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-t", decls(
		"border-top-left-radius", "--value(--radius)",
		"border-top-left-radius", "--value(length, percentage)",
		"border-top-right-radius", "--value(--radius)",
		"border-top-right-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-r", decls(
		"border-top-right-radius", "--value(--radius)",
		"border-top-right-radius", "--value(length, percentage)",
		"border-bottom-right-radius", "--value(--radius)",
		"border-bottom-right-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-b", decls(
		"border-bottom-left-radius", "--value(--radius)",
		"border-bottom-left-radius", "--value(length, percentage)",
		"border-bottom-right-radius", "--value(--radius)",
		"border-bottom-right-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-l", decls(
		"border-top-left-radius", "--value(--radius)",
		"border-top-left-radius", "--value(length, percentage)",
		"border-bottom-left-radius", "--value(--radius)",
		"border-bottom-left-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-tl", decls(
		"border-top-left-radius", "--value(--radius)",
		"border-top-left-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-tr", decls(
		"border-top-right-radius", "--value(--radius)",
		"border-top-right-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-br", decls(
		"border-bottom-right-radius", "--value(--radius)",
		"border-bottom-right-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-bl", decls(
		"border-bottom-left-radius", "--value(--radius)",
		"border-bottom-left-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-s", decls(
		"border-start-start-radius", "--value(--radius)",
		"border-start-start-radius", "--value(length, percentage)",
		"border-end-start-radius", "--value(--radius)",
		"border-end-start-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-e", decls(
		"border-start-end-radius", "--value(--radius)",
		"border-start-end-radius", "--value(length, percentage)",
		"border-end-end-radius", "--value(--radius)",
		"border-end-end-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-ss", decls(
		"border-start-start-radius", "--value(--radius)",
		"border-start-start-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-se", decls(
		"border-start-end-radius", "--value(--radius)",
		"border-start-end-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-ee", decls(
		"border-end-end-radius", "--value(--radius)",
		"border-end-end-radius", "--value(length, percentage)",
	)))
	register(cssUtility("rounded-es", decls(
		"border-end-start-radius", "--value(--radius)",
		"border-end-start-radius", "--value(length, percentage)",
	)))

	// ===== Opacity =====
	register(cssUtility("opacity", decls(
		"opacity", "--value(--opacity)",
		"opacity", "--value(percentage)",
		"opacity", "--value(number)",
	)))

	// ===== Scroll Margin =====
	register(negatableCSSUtility("scroll-m", decls(
		"scroll-margin", "--value(--spacing)",
		"scroll-margin", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-mx", decls(
		"scroll-margin-inline", "--value(--spacing)",
		"scroll-margin-inline", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-my", decls(
		"scroll-margin-block", "--value(--spacing)",
		"scroll-margin-block", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-ms", decls(
		"scroll-margin-inline-start", "--value(--spacing)",
		"scroll-margin-inline-start", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-me", decls(
		"scroll-margin-inline-end", "--value(--spacing)",
		"scroll-margin-inline-end", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-mbs", decls(
		"scroll-margin-block-start", "--value(--spacing)",
		"scroll-margin-block-start", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-mbe", decls(
		"scroll-margin-block-end", "--value(--spacing)",
		"scroll-margin-block-end", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-mt", decls(
		"scroll-margin-top", "--value(--spacing)",
		"scroll-margin-top", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-mr", decls(
		"scroll-margin-right", "--value(--spacing)",
		"scroll-margin-right", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-mb", decls(
		"scroll-margin-bottom", "--value(--spacing)",
		"scroll-margin-bottom", "--value(length, percentage)",
	)))
	register(negatableCSSUtility("scroll-ml", decls(
		"scroll-margin-left", "--value(--spacing)",
		"scroll-margin-left", "--value(length, percentage)",
	)))

	// ===== Scroll Padding =====
	register(cssUtility("scroll-p", decls(
		"scroll-padding", "--value(--spacing)",
		"scroll-padding", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-px", decls(
		"scroll-padding-inline", "--value(--spacing)",
		"scroll-padding-inline", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-py", decls(
		"scroll-padding-block", "--value(--spacing)",
		"scroll-padding-block", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-ps", decls(
		"scroll-padding-inline-start", "--value(--spacing)",
		"scroll-padding-inline-start", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pe", decls(
		"scroll-padding-inline-end", "--value(--spacing)",
		"scroll-padding-inline-end", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pbs", decls(
		"scroll-padding-block-start", "--value(--spacing)",
		"scroll-padding-block-start", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pbe", decls(
		"scroll-padding-block-end", "--value(--spacing)",
		"scroll-padding-block-end", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pt", decls(
		"scroll-padding-top", "--value(--spacing)",
		"scroll-padding-top", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pr", decls(
		"scroll-padding-right", "--value(--spacing)",
		"scroll-padding-right", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pb", decls(
		"scroll-padding-bottom", "--value(--spacing)",
		"scroll-padding-bottom", "--value(length, percentage)",
	)))
	register(cssUtility("scroll-pl", decls(
		"scroll-padding-left", "--value(--spacing)",
		"scroll-padding-left", "--value(length, percentage)",
	)))

	// ===== Filter =====
	filterChain := "var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)"

	register(cssUtility("blur", decls(
		"--tw-blur", "blur(--value(--blur))",
		"--tw-blur", "blur(--value(length))",
		"filter", filterChain,
	)))
	register(cssUtility("brightness", decls(
		"--tw-brightness", "brightness(--value(percentage, number))",
		"filter", filterChain,
	)))
	register(cssUtility("contrast", decls(
		"--tw-contrast", "contrast(--value(percentage, number))",
		"filter", filterChain,
	)))
	register(cssUtility("saturate", decls(
		"--tw-saturate", "saturate(--value(percentage, number))",
		"filter", filterChain,
	)))
	register(negatableCSSUtility("hue-rotate", decls(
		"--tw-hue-rotate", "hue-rotate(--value(number))",
		"filter", filterChain,
	)))
	register(cssUtility("grayscale", decls(
		"--tw-grayscale", "grayscale(--value(percentage, number))",
		"filter", filterChain,
	)))
	register(cssUtility("invert", decls(
		"--tw-invert", "invert(--value(percentage, number))",
		"filter", filterChain,
	)))
	register(cssUtility("sepia", decls(
		"--tw-sepia", "sepia(--value(percentage, number))",
		"filter", filterChain,
	)))

	// ===== Backdrop Filter =====
	backdropChain := "var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-opacity,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,)"

	register(cssUtility("backdrop-blur", decls(
		"--tw-backdrop-blur", "blur(--value(--blur))",
		"--tw-backdrop-blur", "blur(--value(length))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-brightness", decls(
		"--tw-backdrop-brightness", "brightness(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-contrast", decls(
		"--tw-backdrop-contrast", "contrast(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-saturate", decls(
		"--tw-backdrop-saturate", "saturate(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(negatableCSSUtility("backdrop-hue-rotate", decls(
		"--tw-backdrop-hue-rotate", "hue-rotate(--value(number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-grayscale", decls(
		"--tw-backdrop-grayscale", "grayscale(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-invert", decls(
		"--tw-backdrop-invert", "invert(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-sepia", decls(
		"--tw-backdrop-sepia", "sepia(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))
	register(cssUtility("backdrop-opacity", decls(
		"--tw-backdrop-opacity", "opacity(--value(percentage, number))",
		"-webkit-backdrop-filter", backdropChain,
		"backdrop-filter", backdropChain,
	)))

	// ===== Duration =====
	register(cssUtility("duration", decls(
		"transition-duration", "--value(number)",
	)))

	// ===== Delay =====
	register(cssUtility("delay", decls(
		"transition-delay", "--value(number)",
	)))

	// ===== Ease =====
	register(cssUtility("ease", decls(
		"transition-timing-function", "--value(--ease)",
	)))

	// ===== Animation =====
	register(cssUtility("animate", decls(
		"animation", "--value(--animate)",
		"animation", "--value(any)",
	)))

	// ===== Transform =====
	register(negatableCSSUtility("scale", decls(
		"--tw-scale-x", "--value(--scale)",
		"--tw-scale-x", "--value(percentage, number)",
		"--tw-scale-y", "--value(--scale)",
		"--tw-scale-y", "--value(percentage, number)",
		"--tw-scale-z", "--value(--scale)",
		"--tw-scale-z", "--value(percentage, number)",
		"scale", "var(--tw-scale-x) var(--tw-scale-y)",
	)))
	register(negatableCSSUtility("scale-x", decls(
		"--tw-scale-x", "--value(--scale)",
		"--tw-scale-x", "--value(percentage, number)",
		"scale", "var(--tw-scale-x) var(--tw-scale-y, 1)",
	)))
	register(negatableCSSUtility("scale-y", decls(
		"--tw-scale-y", "--value(--scale)",
		"--tw-scale-y", "--value(percentage, number)",
		"scale", "var(--tw-scale-x, 1) var(--tw-scale-y)",
	)))
	register(negatableCSSUtility("scale-z", decls(
		"--tw-scale-z", "--value(--scale)",
		"--tw-scale-z", "--value(percentage, number)",
		"scale", "var(--tw-scale-x, 1) var(--tw-scale-y, 1) var(--tw-scale-z)",
	)))

	register(negatableCSSUtility("rotate", decls(
		"rotate", "--value(--rotate)",
		"rotate", "--value(number)",
	)))
	register(negatableCSSUtility("rotate-x", decls(
		"rotate", "x --value(--rotate)",
		"rotate", "x --value(number)",
	)))
	register(negatableCSSUtility("rotate-y", decls(
		"rotate", "y --value(--rotate)",
		"rotate", "y --value(number)",
	)))
	register(negatableCSSUtility("rotate-z", decls(
		"rotate", "z --value(--rotate)",
		"rotate", "z --value(number)",
	)))

	register(negatableCSSUtility("translate-x", decls(
		"--tw-translate-x", "--value(--spacing)",
		"--tw-translate-x", "--value(length, percentage)",
		"translate", "var(--tw-translate-x) var(--tw-translate-y)",
	)))
	register(negatableCSSUtility("translate-y", decls(
		"--tw-translate-y", "--value(--spacing)",
		"--tw-translate-y", "--value(length, percentage)",
		"translate", "var(--tw-translate-x) var(--tw-translate-y)",
	)))
	register(negatableCSSUtility("translate-z", decls(
		"--tw-translate-z", "--value(--spacing)",
		"--tw-translate-z", "--value(length, percentage)",
		"translate", "var(--tw-translate-x) var(--tw-translate-y) var(--tw-translate-z)",
	)))

	register(negatableCSSUtility("skew-x", decls(
		"--tw-skew-x", "skewX(--value(--skew))",
		"--tw-skew-x", "skewX(--value(number))",
		"transform", "var(--tw-rotate-x,) var(--tw-rotate-y,) var(--tw-rotate-z,) var(--tw-skew-x,) var(--tw-skew-y,)",
	)))
	register(negatableCSSUtility("skew-y", decls(
		"--tw-skew-y", "skewY(--value(--skew))",
		"--tw-skew-y", "skewY(--value(number))",
		"transform", "var(--tw-rotate-x,) var(--tw-rotate-y,) var(--tw-rotate-z,) var(--tw-skew-x,) var(--tw-skew-y,)",
	)))

	// ===== Perspective =====
	register(cssUtility("perspective", decls(
		"perspective", "--value(--perspective)",
		"perspective", "--value(length)",
	)))

	// ===== Perspective Origin =====
	register(cssUtility("perspective-origin", decls(
		"perspective-origin", "--value(position)",
		"perspective-origin", "--value(any)",
	)))

	// ===== Gradient Utilities =====

	// bg-linear-<angle> → linear-gradient(<angle>deg in oklab, ...)
	bgLinear := functionalUtility("bg-linear", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			interp := resolveGradientInterpolation(c.Modifier)
			if interp == "" {
				interp = " in oklab"
			}
			return []Declaration{
				{Property: "--tw-gradient-position", Value: c.Arbitrary + interp},
				{Property: "background-image", Value: "linear-gradient(var(--tw-gradient-stops))"},
			}
		}
		if c.Value == "" {
			return nil
		}
		if isNumeric(c.Value) {
			angle := c.Value
			if c.Negative {
				angle = "-" + angle
			}
			interp := resolveGradientInterpolation(c.Modifier)
			if interp == "" {
				interp = " in oklab"
			}
			return []Declaration{
				{Property: "--tw-gradient-position", Value: angle + "deg" + interp},
				{Property: "background-image", Value: "linear-gradient(var(--tw-gradient-stops))"},
			}
		}
		return nil
	})
	bgLinear.Negatable = true
	register(bgLinear)

	// bg-radial-[<value>] → radial-gradient(<value> in oklab, ...)
	register(functionalUtility("bg-radial", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			interp := resolveGradientInterpolation(c.Modifier)
			if interp == "" {
				interp = " in oklab"
			}
			return []Declaration{
				{Property: "--tw-gradient-position", Value: c.Arbitrary + interp},
				{Property: "background-image", Value: "radial-gradient(var(--tw-gradient-stops))"},
			}
		}
		return nil
	}))

	// bg-conic-<angle> → conic-gradient(from <angle>deg in oklab, ...)
	bgConic := functionalUtility("bg-conic", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			interp := resolveGradientInterpolation(c.Modifier)
			if interp == "" {
				interp = " in oklab"
			}
			return []Declaration{
				{Property: "--tw-gradient-position", Value: c.Arbitrary + interp},
				{Property: "background-image", Value: "conic-gradient(var(--tw-gradient-stops))"},
			}
		}
		if c.Value == "" {
			return nil
		}
		if isNumeric(c.Value) {
			angle := c.Value
			if c.Negative {
				angle = "-" + angle
			}
			interp := resolveGradientInterpolation(c.Modifier)
			if interp == "" {
				interp = " in oklab"
			}
			return []Declaration{
				{Property: "--tw-gradient-position", Value: "from " + angle + "deg" + interp},
				{Property: "background-image", Value: "conic-gradient(var(--tw-gradient-stops))"},
			}
		}
		return nil
	})
	bgConic.Negatable = true
	register(bgConic)

	// ===== Content =====
	register(cssUtility("content", decls(
		"content", "--value(any)",
	)))

	// ===== Font Feature Settings =====
	register(functionalUtility("font-features", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			// Wrap bare tags in quotes: font-features-[smcp] → "smcp"
			val := c.Arbitrary
			if len(val) > 0 && val[0] != '"' && val[0] != '\'' {
				val = `"` + val + `"`
			}
			return decls("font-feature-settings", val)
		}
		return nil
	}))

	// ===== Font Stretch =====
	register(functionalUtility("font-stretch", func(c ResolvedCandidate) []Declaration {
		if c.Arbitrary != "" {
			return decls("font-stretch", c.Arbitrary)
		}
		if c.Value == "" {
			return nil
		}
		// Accept percentage values like font-stretch-50% or font-stretch-125%
		val := c.Value
		if isNumeric(val) {
			val = val + "%"
		}
		return decls("font-stretch", val)
	}))
}
