package tailwind

import "strings"

// ParsedClass is the decomposed form of a Tailwind class string.
//
// Example: "dark:md:hover:!-translate-x-1/2" becomes:
//
//	ParsedClass{
//	    Raw:       "dark:md:hover:!-translate-x-1/2",
//	    Variants:  ["dark", "md", "hover"],
//	    Important: true,
//	    Negative:  true,
//	    Utility:   "translate-x",
//	    Value:     "1/2",
//	}
type ParsedClass struct {
	Raw       string   // Original class string.
	Variants  []string // Variant prefixes in order.
	Important bool     // ! modifier.
	Negative  bool     // - prefix on value.
	Utility   string   // Utility name (e.g., "translate-x", "bg", "w").
	Value     string   // Value portion (e.g., "4", "blue-500", "1/2").
	Arbitrary string   // Content of [...] if present, with _ → space.
	ArbitraryProperty bool // True for [mask-type:alpha] style classes.
	Modifier  string   // Opacity modifier: "75", "[.5]", etc.
}

// parseClass decomposes a raw Tailwind class string.
func parseClass(raw string) ParsedClass {
	pc := ParsedClass{Raw: raw}

	s := raw

	// 1. Extract variant prefixes.
	// Split on ':' but respect [...] brackets.
	parts := splitVariants(s)
	if len(parts) > 1 {
		pc.Variants = parts[:len(parts)-1]
		s = parts[len(parts)-1]
	}

	// 2. Check for arbitrary property: [property:value]
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := s[1 : len(s)-1]
		inner = strings.ReplaceAll(inner, "_", " ")
		if idx := strings.Index(inner, ":"); idx > 0 {
			pc.ArbitraryProperty = true
			pc.Utility = inner[:idx]
			pc.Arbitrary = inner[idx+1:]
			return pc
		}
	}

	// 3. Important modifier (!).
	if strings.HasPrefix(s, "!") {
		pc.Important = true
		s = s[1:]
	}

	// 4. Negative modifier (-).
	if strings.HasPrefix(s, "-") {
		pc.Negative = true
		s = s[1:]
	}

	// 4b. Extract opacity modifier (e.g., /75, /[.5]).
	s, pc.Modifier = extractModifier(s)

	// 5. Check for arbitrary value: utility-[value]
	if idx := strings.Index(s, "-["); idx > 0 {
		bracket := s[idx+1:]
		if strings.HasSuffix(bracket, "]") {
			pc.Utility = s[:idx]
			inner := bracket[1 : len(bracket)-1]
			pc.Arbitrary = strings.ReplaceAll(inner, "_", " ")
			return pc
		}
	}

	// 6. Split into utility name and value.
	// This is the ambiguous part — we need the utility index to
	// disambiguate. For now, we use a heuristic: the value is
	// everything after the last hyphen that starts with a digit,
	// a known value keyword, or a fraction.
	//
	// The definitive resolution happens in the generator, which
	// tries each possible split against the utility index.
	pc.Utility = s
	pc.Value = ""

	// Try to split off a value: find the rightmost hyphen where
	// the right side looks like a value.
	if idx := findValueSplit(s); idx > 0 {
		pc.Utility = s[:idx]
		pc.Value = s[idx+1:]
	}

	return pc
}

// splitVariants splits a class string on ':' delimiters while respecting
// square bracket groups. "dark:hover:[&>svg]:text-white" becomes
// ["dark", "hover", "[&>svg]", "text-white"].
func splitVariants(s string) []string {
	var parts []string
	depth := 0
	start := 0

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case ':':
			if depth == 0 {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])

	return parts
}

// findValueSplit finds the best position to split "utility-value"
// into the utility name and value. Returns the index of the hyphen,
// or -1 if no good split exists.
//
// Strategy: scan right-to-left, looking for hyphens. The first
// (rightmost) hyphen where the right side starts with a digit or
// is a known value keyword is our split point.
func findValueSplit(s string) int {
	// We try each possible split point from right to left.
	// The caller (generator) will validate against the utility index.
	best := -1

	for i := len(s) - 1; i > 0; i-- {
		if s[i] != '-' {
			continue
		}

		right := s[i+1:]
		if right == "" {
			continue
		}

		// If the right side starts with a digit, it's very likely a value.
		if isDigit(right[0]) {
			best = i
			// Keep scanning — a longer utility name might also work,
			// and we prefer the longest valid utility.
			continue
		}

		// Check for known value-like patterns.
		if isValueKeyword(right) {
			best = i
			continue
		}

		// Check for fraction-like: "1/2", "2/3", etc.
		if strings.Contains(right, "/") {
			parts := strings.SplitN(right, "/", 2)
			if len(parts) == 2 && isNumeric(parts[0]) && isNumeric(parts[1]) {
				best = i
				continue
			}
		}
	}

	return best
}

// isValueKeyword returns true if s is a Tailwind value keyword.
func isValueKeyword(s string) bool {
	switch s {
	case "full", "screen", "min", "max", "fit", "auto", "none",
		"px", "xs", "sm", "md", "lg", "xl", "2xl", "3xl", "4xl", "5xl",
		"inherit", "initial", "revert", "unset", "current",
		"transparent", "black", "white",
		"thin", "extralight", "light", "normal", "medium",
		"semibold", "bold", "extrabold",
		"tight", "snug", "relaxed", "loose",
		"start", "end", "center", "between", "around", "evenly", "stretch",
		"baseline", "wrap", "nowrap", "reverse",
		"hidden", "visible", "scroll", "clip",
		"contain", "cover", "fill",
		"solid", "dashed", "dotted", "double",
		"row", "col", "dense",
		"fixed", "absolute", "relative", "sticky", "static",
		"block", "inline", "flex", "grid", "table", "contents",
		"ease", "linear",
		"pointer", "default", "wait", "text", "move", "not-allowed",
		"top", "bottom", "left", "right", "inset":
		return true
	}
	return false
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, b := range []byte(s) {
		if !isDigit(b) && b != '.' {
			return false
		}
	}
	return true
}

// extractModifier splits the opacity modifier from a class string.
// It finds the last '/' not inside brackets. If the segment before '/'
// (after the last hyphen) and after '/' are both numeric and form a
// valid fraction (numerator < denominator), it's a fraction like w-1/2,
// not an opacity modifier.
func extractModifier(s string) (base, modifier string) {
	depth := 0
	lastSlash := -1
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '/':
			if depth == 0 {
				lastSlash = i
			}
		}
	}
	if lastSlash < 0 {
		return s, ""
	}
	left := s[:lastSlash]
	right := s[lastSlash+1:]
	// Check for fraction pattern: the segment after the last '-' in left
	// and right must both be numeric, with numerator < denominator (e.g., 1/2, 2/3).
	if isNumeric(right) {
		lastHyphen := strings.LastIndexByte(left, '-')
		var segment string
		if lastHyphen >= 0 {
			segment = left[lastHyphen+1:]
		} else {
			segment = left
		}
		if isNumeric(segment) && segment != "" {
			num := parseSimpleInt(segment)
			den := parseSimpleInt(right)
			if den > 0 && num < den {
				return s, ""
			}
		}
	}
	return left, right
}

// parseSimpleInt parses a non-negative integer from a string.
// Returns -1 if the string is not a valid integer.
func parseSimpleInt(s string) int {
	if s == "" {
		return -1
	}
	n := 0
	for _, b := range []byte(s) {
		if !isDigit(b) {
			return -1
		}
		n = n*10 + int(b-'0')
	}
	return n
}
