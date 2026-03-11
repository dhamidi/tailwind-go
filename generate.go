package tailwind

import (
	"fmt"
	"sort"
	"strings"
)

// generatedRule is a single CSS rule ready for output.
type generatedRule struct {
	// selector is the fully-escaped CSS selector.
	selector string
	// declarations are the CSS property: value pairs.
	declarations []Declaration
	// important marks all declarations as !important.
	important bool
	// mediaQueries wraps the rule in @media blocks (outermost first).
	mediaQueries []string
	// order controls sort position in output.
	order int
}

// generate produces a CSS string for the given candidate classes.
func generate(
	candidates []string,
	theme *ThemeConfig,
	utils *utilityIndex,
	variants map[string]*VariantDef,
) string {
	var rules []generatedRule

	for _, raw := range candidates {
		pc := parseClass(raw)
		rule := resolveClass(pc, theme, utils, variants)
		if rule != nil {
			rules = append(rules, *rule)
		}
	}

	// Sort rules by definition order for stable, correct cascade.
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].order < rules[j].order
	})

	return emitCSS(rules)
}

// resolveClass attempts to match a parsed class against the utility
// registry and produce a generated rule.
func resolveClass(
	pc ParsedClass,
	theme *ThemeConfig,
	utils *utilityIndex,
	variants map[string]*VariantDef,
) *generatedRule {

	// Handle arbitrary properties: [mask-type:alpha]
	if pc.ArbitraryProperty {
		return resolveArbitraryProperty(pc, variants)
	}

	// Look up the utility.
	utilDef, valueStr := resolveUtility(pc, utils)
	if utilDef == nil {
		return nil
	}

	// Resolve the declarations by substituting the value.
	decls := resolveDeclarations(utilDef, valueStr, pc, theme)
	if decls == nil {
		return nil
	}

	// Build the selector and apply variant selector transforms.
	selector := buildSelector(pc)
	selector = resolveVariantSelector(selector, pc.Variants, variants)

	// Apply variant media query wrapping.
	mediaQueries := resolveVariants(pc.Variants, variants)

	return &generatedRule{
		selector:     selector,
		declarations: decls,
		important:    pc.Important,
		mediaQueries: mediaQueries,
		order:        utilDef.Order,
	}
}

// resolveUtility finds the matching UtilityDef for a parsed class.
// It reconstructs the full utility-value string and uses the index's
// longest-prefix matching to disambiguate patterns like "bg" vs "bg-x".
func resolveUtility(pc ParsedClass, utils *utilityIndex) (*UtilityDef, string) {
	// Arbitrary values: the class parser already knows the utility name.
	// We just need to match it against a pattern.
	if pc.Arbitrary != "" {
		// Direct pattern match (dynamic utilities).
		for _, u := range utils.dynamic {
			if pc.Utility == u.Pattern {
				return u, pc.Arbitrary
			}
		}
		// Static match (unusual but possible).
		if u, ok := utils.static[pc.Utility]; ok {
			return u, pc.Arbitrary
		}
		return nil, ""
	}

	// Reconstruct the full utility string (e.g., "bg-blue-500", "p-4").
	full := pc.Utility
	if pc.Value != "" {
		full = pc.Utility + "-" + pc.Value
	}

	// Static exact match on the full string.
	if u, ok := utils.static[full]; ok {
		return u, ""
	}

	// Dynamic match: lookup finds the longest matching pattern prefix
	// and extracts the value portion.
	if u, val := utils.lookup(full); u != nil {
		return u, val
	}

	return nil, ""
}

// resolveDeclarations substitutes values into utility declarations.
// When multiple declarations share the same CSS property (a priority chain),
// only the first successfully-resolved alternative is emitted.
func resolveDeclarations(
	utilDef *UtilityDef,
	valueStr string,
	pc ParsedClass,
	theme *ThemeConfig,
) []Declaration {
	if utilDef.Static {
		return utilDef.Declarations
	}

	// Group declarations by property name (preserving order of first appearance).
	type propGroup struct {
		property string
		alts     []Declaration
	}
	var groups []propGroup
	seenProp := make(map[string]int) // property → index in groups

	for _, d := range utilDef.Declarations {
		if idx, ok := seenProp[d.Property]; ok {
			groups[idx].alts = append(groups[idx].alts, d)
		} else {
			seenProp[d.Property] = len(groups)
			groups = append(groups, propGroup{property: d.Property, alts: []Declaration{d}})
		}
	}

	// Resolve each property group: try alternatives in order.
	var result []Declaration
	for _, g := range groups {
		resolved := resolvePropertyGroup(g.alts, valueStr, pc, theme)
		if resolved != nil {
			result = append(result, *resolved)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// resolvePropertyGroup tries each alternative declaration for a property
// and returns the first one that successfully resolves.
func resolvePropertyGroup(alts []Declaration, valueStr string, pc ParsedClass, theme *ThemeConfig) *Declaration {
	for _, d := range alts {
		if !strings.Contains(d.Value, "--value(") {
			// No placeholder — emit verbatim.
			return &Declaration{Property: d.Property, Value: d.Value}
		}
		cssValue := resolveValueForDecl(d, valueStr, pc, theme)
		if cssValue != "" {
			resolved := substituteValue(d.Value, cssValue)
			// Apply opacity modifier if the declaration used a color namespace.
			if pc.Modifier != "" && extractNamespace(d.Value) == "color" {
				resolved = applyModifier(resolved, pc.Modifier, theme)
			}
			return &Declaration{Property: d.Property, Value: resolved}
		}
	}
	return nil
}

// resolveValueForDecl resolves the CSS value for a specific declaration's --value() args.
func resolveValueForDecl(d Declaration, valueStr string, pc ParsedClass, theme *ThemeConfig) string {
	// Arbitrary values pass through for any --value() type.
	if pc.Arbitrary != "" {
		v := pc.Arbitrary
		if pc.Negative {
			v = negateValue(v)
		}
		return v
	}

	if valueStr == "" {
		return ""
	}

	// Fraction support: "1/2" → "50%". Check before theme resolution so that
	// fractions are not erroneously computed as spacing multipliers.
	if strings.Contains(valueStr, "/") {
		if pct := fractionToPercent(valueStr); pct != "" {
			if pc.Negative {
				return negateValue(pct)
			}
			return pct
		}
	}

	ns := extractNamespace(d.Value)
	if ns != "" {
		// Namespace specified (e.g., --value(--spacing)). Try theme resolution.
		if resolved, ok := theme.Resolve(ns, valueStr); ok {
			if pc.Negative {
				resolved = negateValue(resolved)
			}
			return resolved
		}
		return "" // namespace specified but didn't resolve — try next alt
	}

	// No namespace — this is a type-based --value() like --value(length, percentage).
	// Try keyword mapping.
	if cssVal := keywordToCSS(valueStr); cssVal != "" {
		return cssVal
	}

	// Try type-based resolution.
	valueTypes := extractValueTypes(d.Value)
	return resolveRawValue(valueStr, valueTypes, pc)
}

// extractValueTypes parses --value(length, percentage) → ["length", "percentage"]
func extractValueTypes(value string) []string {
	idx := strings.Index(value, "--value(")
	if idx < 0 {
		return nil
	}
	start := idx + len("--value(")
	end := strings.Index(value[start:], ")")
	if end < 0 {
		return nil
	}
	arg := strings.TrimSpace(value[start : start+end])
	if strings.HasPrefix(arg, "--") {
		// This is a namespace reference, not type list.
		return nil
	}
	parts := strings.Split(arg, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// resolveRawValue handles type-based --value() resolution for non-arbitrary,
// non-theme values (e.g., --value(length), --value(integer), --value(any)).
func resolveRawValue(valueStr string, types []string, pc ParsedClass) string {
	for _, t := range types {
		switch t {
		case "integer", "number":
			if isNumeric(valueStr) {
				v := valueStr
				if pc.Negative {
					v = "-" + v
				}
				return v
			}
		case "any":
			if valueStr != "" {
				return valueStr
			}
		case "length", "percentage":
			// For non-arbitrary values, try raw value conversion.
			if raw := rawValue(valueStr); raw != "" {
				return raw
			}
		}
	}
	return ""
}

// extractNamespace pulls the theme namespace from a --value() expression.
// e.g., "--value(--color)" → "color", "--value(--spacing)" → "spacing"
func extractNamespace(value string) string {
	idx := strings.Index(value, "--value(")
	if idx < 0 {
		return ""
	}
	start := idx + len("--value(")
	end := strings.Index(value[start:], ")")
	if end < 0 {
		return ""
	}
	arg := strings.TrimSpace(value[start : start+end])

	// --value(--color) → namespace "color"
	if strings.HasPrefix(arg, "--") {
		return strings.TrimPrefix(arg, "--")
	}

	// --value(length, percentage) → no namespace, accepts arbitrary CSS types
	return ""
}

// substituteValue replaces --value(...) in a declaration with the resolved value.
func substituteValue(template, resolved string) string {
	idx := strings.Index(template, "--value(")
	if idx < 0 {
		// No placeholder — return as-is (shouldn't happen for dynamic utils).
		return template
	}
	// Find the matching closing paren.
	depth := 0
	end := idx + len("--value(")
	for end < len(template) {
		if template[end] == '(' {
			depth++
		} else if template[end] == ')' {
			if depth == 0 {
				return template[:idx] + resolved + template[end+1:]
			}
			depth--
		}
		end++
	}
	return template[:idx] + resolved
}

// buildSelector constructs the CSS selector for a class.
// Escapes special characters in the class name.
func buildSelector(pc ParsedClass) string {
	escaped := escapeSelector(pc.Raw)
	return "." + escaped
}

// escapeSelector escapes special CSS characters in a selector.
func escapeSelector(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case ':', '[', ']', '/', '(', ')', '#', '.', ',',
			'!', '+', '*', '%', '@', '&', '>', '~', ' ':
			sb.WriteByte('\\')
			sb.WriteRune(c)
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// resolveVariants converts variant names to media queries / selector transforms.
func resolveVariants(names []string, defs map[string]*VariantDef) []string {
	var media []string
	for _, name := range names {
		// Handle arbitrary variants: [&:nth-child(3)] or [@media(min-width:900px)]
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			inner := name[1 : len(name)-1]
			inner = strings.ReplaceAll(inner, "_", " ")
			if strings.HasPrefix(inner, "@media") ||
				strings.HasPrefix(inner, "@supports") ||
				strings.HasPrefix(inner, "@container") {
				media = append(media, inner)
			}
			continue
		}

		if v, ok := defs[name]; ok {
			if v.Media != "" {
				if v.AtRule != "" {
					media = append(media, "@"+v.AtRule+" "+v.Media)
				} else {
					media = append(media, "@media "+v.Media)
				}
			}
		}
	}
	return media
}

// resolveVariantSelector applies variant selector transformations.
func resolveVariantSelector(base string, names []string, defs map[string]*VariantDef) string {
	sel := base
	for _, name := range names {
		// Arbitrary variant
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			inner := name[1 : len(name)-1]
			inner = strings.ReplaceAll(inner, "_", " ")
			sel = strings.ReplaceAll(inner, "&", sel)
			continue
		}

		if v, ok := defs[name]; ok && v.Selector != "" {
			sel = strings.ReplaceAll(v.Selector, "&", sel)
		}
	}
	return sel
}

// emitCSS serializes generated rules into a CSS string.
func emitCSS(rules []generatedRule) string {
	var sb strings.Builder

	for i, r := range rules {
		if i > 0 {
			sb.WriteByte('\n')
		}

		// Open media query wrappers (outermost first).
		for _, mq := range r.mediaQueries {
			sb.WriteString(mq)
			sb.WriteString(" {\n")
		}

		// Selector.
		sb.WriteString(r.selector)
		sb.WriteString(" {\n")

		// Declarations.
		for _, d := range r.declarations {
			sb.WriteString("  ")
			sb.WriteString(d.Property)
			sb.WriteString(": ")
			sb.WriteString(d.Value)
			if r.important {
				sb.WriteString(" !important")
			}
			sb.WriteString(";\n")
		}

		sb.WriteString("}\n")

		// Close media query wrappers.
		for range r.mediaQueries {
			sb.WriteString("}\n")
		}
	}

	return sb.String()
}

// resolveArbitraryProperty handles [property:value] classes.
func resolveArbitraryProperty(pc ParsedClass, variants map[string]*VariantDef) *generatedRule {
	selector := buildSelector(pc)
	mediaQueries := resolveVariants(pc.Variants, variants)

	return &generatedRule{
		selector: selector,
		declarations: []Declaration{{
			Property: pc.Utility,
			Value:    pc.Arbitrary,
		}},
		important:    pc.Important,
		mediaQueries: mediaQueries,
		order:        9999, // arbitrary properties sort last
	}
}

// keywordToCSS maps Tailwind value keywords to CSS values.
func keywordToCSS(s string) string {
	switch s {
	case "full":
		return "100%"
	case "screen":
		return "100vw"
	case "svw":
		return "100svw"
	case "lvw":
		return "100lvw"
	case "dvw":
		return "100dvw"
	case "svh":
		return "100svh"
	case "lvh":
		return "100lvh"
	case "dvh":
		return "100dvh"
	case "min":
		return "min-content"
	case "max":
		return "max-content"
	case "fit":
		return "fit-content"
	case "auto":
		return "auto"
	case "none":
		return "none"
	case "px":
		return "1px"
	case "inherit":
		return "inherit"
	case "initial":
		return "initial"
	case "revert":
		return "revert"
	case "unset":
		return "unset"
	case "current":
		return "currentColor"
	case "transparent":
		return "transparent"
	}
	return ""
}

// fractionToPercent converts "1/2" → "50%", "2/3" → "66.666667%", etc.
func fractionToPercent(s string) string {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	num := parseFloat(parts[0])
	den := parseFloat(parts[1])
	if den == 0 {
		return ""
	}
	pct := (num / den) * 100
	return fmt.Sprintf("%.6g%%", pct)
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// negateValue prepends calc(-1 * ...) or a minus sign.
func negateValue(v string) string {
	if strings.HasPrefix(v, "calc(") {
		return "calc(-1 * " + v[5:]
	}
	return "calc(-1 * " + v + ")"
}

// applyModifier wraps a CSS color value with oklch opacity modifier.
func applyModifier(cssValue, modifier string, theme *ThemeConfig) string {
	if modifier == "" {
		return cssValue
	}
	var opacityStr string
	if strings.HasPrefix(modifier, "[") && strings.HasSuffix(modifier, "]") {
		opacityStr = modifier[1 : len(modifier)-1]
		opacityStr = strings.ReplaceAll(opacityStr, "_", " ")
	} else {
		opacityStr = resolveModifierOpacity(modifier, theme)
	}
	return "oklch(from " + cssValue + " l c h / " + opacityStr + ")"
}

// resolveModifierOpacity resolves an opacity modifier value.
// It checks the theme for --opacity-{modifier} first, then falls back
// to treating numeric modifiers as percentages.
func resolveModifierOpacity(modifier string, theme *ThemeConfig) string {
	if v, ok := theme.Resolve("opacity", modifier); ok {
		return v
	}
	if isNumeric(modifier) {
		return modifier + "%"
	}
	return modifier
}

// rawValue handles raw Tailwind value tokens like "px", "0", "0.5".
func rawValue(s string) string {
	// "0" → "0px" or just "0"
	if s == "0" {
		return "0px"
	}
	// "0.5" → "0.125rem" (spacing scale)
	// These should have been caught by theme resolution.
	return ""
}
