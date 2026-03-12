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
	keyframes map[string]*KeyframesRule,
) string {
	var rules []generatedRule
	referencedKeyframes := make(map[string]bool)

	for _, raw := range candidates {
		pc := parseClass(raw)
		rule := resolveClass(pc, theme, utils, variants)
		if rule != nil {
			rules = append(rules, *rule)
			for _, d := range rule.declarations {
				if d.Property == "animation" || d.Property == "animation-name" {
					extractKeyframeNames(d.Value, referencedKeyframes)
				}
			}
		}
	}

	// Sort rules by definition order for stable, correct cascade.
	// Per spec §10.1:
	// 1. Utility definition's Order field (source order from the CSS)
	// 2. Within same utility, variant-wrapped rules come after unwrapped ones
	// 3. Responsive variants are ordered by breakpoint size (ascending)
	// All variant-wrapped rules come after all unwrapped rules to ensure
	// responsive overrides always win in the cascade.
	sort.SliceStable(rules, func(i, j int) bool {
		iHasMedia := len(rules[i].mediaQueries) > 0
		jHasMedia := len(rules[j].mediaQueries) > 0
		if iHasMedia != jHasMedia {
			return !iHasMedia // unwrapped first
		}
		return rules[i].order < rules[j].order
	})

	return emitCSS(rules, referencedKeyframes, keyframes)
}

// extractKeyframeNames extracts animation names from a CSS animation value.
func extractKeyframeNames(animValue string, out map[string]bool) {
	parts := strings.FieldsFunc(animValue, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	})
	if len(parts) > 0 {
		out[parts[0]] = true
	}
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

	// Append child selector suffix if the utility defines one.
	if utilDef.Selector != "" {
		selector = selector + " " + utilDef.Selector
	}

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

	// Exhaustive fallback: try every possible hyphen-split of the full
	// string, longest utility prefix first. This handles cases where the
	// class parser's heuristic split was wrong and lookup didn't find a
	// match (e.g., multi-segment color names or compound utilities).
	for i := len(full) - 1; i > 0; i-- {
		if full[i] != '-' {
			continue
		}
		utilPart := full[:i]
		valPart := full[i+1:]
		if valPart == "" {
			continue
		}

		// Try dynamic patterns for this split.
		for _, u := range utils.dynamic {
			if utilPart == u.Pattern {
				return u, valPart
			}
		}
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
		if pc.Negative {
			var negated []Declaration
			for _, d := range utilDef.Declarations {
				neg := negateValue(d.Value)
				if neg == "" {
					return nil // cannot negate — discard
				}
				negated = append(negated, Declaration{Property: d.Property, Value: neg})
			}
			return negated
		}
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

		// If a type hint is provided, skip declarations that don't match.
		if pc.TypeHint != "" {
			if !matchesTypeHint(pc.TypeHint, extractNamespace(d.Value), extractValueTypes(d.Value)) {
				continue
			}
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

// matchesTypeHint returns true if the type hint matches the declaration's
// namespace or any of its value types.
func matchesTypeHint(hint, namespace string, types []string) bool {
	if namespace != "" && namespace == hint {
		return true
	}
	for _, t := range types {
		if t == hint {
			return true
		}
	}
	return false
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
		// If there are also type-based args (mixed form like --value(--font-size, length)),
		// fall through to type-based resolution below. Otherwise, try next alt.
		valueTypes := extractValueTypes(d.Value)
		if len(valueTypes) == 0 {
			return "" // namespace-only but didn't resolve — try next alt
		}
		// Fall through to type-based resolution with the extracted types.
		return resolveRawValue(valueStr, valueTypes, pc, d)
	}

	// No namespace — this is a type-based --value() like --value(length, percentage).
	// Try keyword mapping.
	if cssVal := keywordToCSS(valueStr, d.Property); cssVal != "" {
		if pc.Negative {
			cssVal = negateValue(cssVal)
			if cssVal == "" {
				return ""
			}
		}
		return cssVal
	}

	// Try type-based resolution.
	valueTypes := extractValueTypes(d.Value)
	return resolveRawValue(valueStr, valueTypes, pc, d)
}

// extractValueTypes parses --value(length, percentage) → ["length", "percentage"]
// For mixed forms like --value(--font-size, length, percentage), it returns
// only the non-namespace parts: ["length", "percentage"].
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
	parts := strings.Split(arg, ",")
	var types []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(p, "--") {
			// This is a namespace reference, skip it.
			continue
		}
		if p != "" {
			types = append(types, p)
		}
	}
	return types
}

// resolveRawValue handles type-based --value() resolution for non-arbitrary,
// non-theme values (e.g., --value(length), --value(integer), --value(any)).
func resolveRawValue(valueStr string, types []string, pc ParsedClass, d Declaration) string {
	for _, t := range types {
		switch t {
		case "integer", "number":
			if isNumeric(valueStr) {
				v := valueStr
				if pc.Negative {
					v = "-" + v
				}
				// Apply bare value unit suffix based on property/value context.
				// This matches upstream Tailwind's handleBareValue callbacks.
				v = applyBareValueUnit(v, d)
				return v
			}
		case "percentage":
			// For bare integers in percentage context, append %.
			// This handles scale-50 → 50%, brightness-75 → 75%, etc.
			if isNumeric(valueStr) {
				v := valueStr
				if pc.Negative {
					v = "-" + v
				}
				return v + "%"
			}
			// For non-arbitrary values, try raw value conversion.
			if raw := rawValue(valueStr); raw != "" {
				return raw
			}
		case "any":
			if valueStr != "" {
				return valueStr
			}
		case "length":
			// For non-arbitrary values, try raw value conversion.
			if raw := rawValue(valueStr); raw != "" {
				return raw
			}
		}
	}
	return ""
}

// applyBareValueUnit adds a CSS unit suffix to a bare integer/number value
// based on the declaration's property and value template context.
// This matches upstream Tailwind's handleBareValue callbacks which append
// units like deg, ms, px to bare integer values.
func applyBareValueUnit(v string, d Declaration) string {
	prop := d.Property
	valTemplate := d.Value

	// Degree-based properties: rotate, skew, hue-rotate
	if prop == "rotate" || strings.HasPrefix(prop, "--tw-skew") {
		return v + "deg"
	}
	if strings.Contains(valTemplate, "hue-rotate(") {
		return v + "deg"
	}

	// Time-based properties: transition-duration, transition-delay
	if prop == "transition-duration" || prop == "transition-delay" ||
		prop == "animation-duration" || prop == "animation-delay" {
		return v + "ms"
	}

	return v
}

// extractNamespace pulls the theme namespace from a --value() expression.
// e.g., "--value(--color)" → "color", "--value(--spacing)" → "spacing"
// For mixed forms like "--value(--font-size, length, percentage)" → "font-size"
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

	// Handle comma-separated args: extract only the first if it's a namespace.
	first := arg
	if commaIdx := strings.Index(arg, ","); commaIdx >= 0 {
		first = strings.TrimSpace(arg[:commaIdx])
	}

	// --value(--color) → namespace "color"
	if strings.HasPrefix(first, "--") {
		return strings.TrimPrefix(first, "--")
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

		v := lookupVariant(name, defs)
		if v != nil {
			if v.AtRule != "" && v.Media == "" {
				media = append(media, "@"+v.AtRule)
			} else if v.Media != "" {
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
			// At-rule variants (@media, @supports, @container) are handled by
			// resolveVariants as wrapper blocks, not as selector transforms.
			if strings.HasPrefix(inner, "@media") ||
				strings.HasPrefix(inner, "@supports") ||
				strings.HasPrefix(inner, "@container") {
				continue
			}
			sel = strings.ReplaceAll(inner, "&", sel)
			continue
		}

		if v, ok := defs[name]; ok && v.Selector != "" {
			sel = strings.ReplaceAll(v.Selector, "&", sel)
			continue
		}

		// Try compound variant resolution.
		if v := lookupCompoundVariant(name, defs); v != nil {
			remainder := name[len(v.Name)+1:] // e.g., "group-hover" → "hover", "group-hover/sidebar" → "hover/sidebar"
			value := remainder
			groupName := ""
			if slashIdx := strings.Index(remainder, "/"); slashIdx >= 0 {
				value = remainder[:slashIdx]
				groupName = remainder[slashIdx+1:]
			}
			selector := resolveCompoundTemplate(v.Template, value, groupName)
			sel = strings.ReplaceAll(selector, "&", sel)
		}
	}
	return sel
}

// lookupVariant finds a variant by exact name or compound match.
func lookupVariant(name string, defs map[string]*VariantDef) *VariantDef {
	if v, ok := defs[name]; ok {
		return v
	}
	return lookupCompoundVariant(name, defs)
}

// lookupCompoundVariant checks if a variant name matches any compound variant
// pattern (e.g., "group-hover" matches "group" compound variant).
func lookupCompoundVariant(name string, defs map[string]*VariantDef) *VariantDef {
	for _, v := range defs {
		if !v.Compound {
			continue
		}
		prefix := v.Name + "-"
		if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
			return v
		}
	}
	return nil
}

// stripMerge removes :merge() wrappers from a selector string.
// For example, ":merge(.group):hover &" becomes ".group:hover &".
func stripMerge(selector string) string {
	for {
		idx := strings.Index(selector, ":merge(")
		if idx < 0 {
			return selector
		}
		start := idx + len(":merge(")
		depth := 1
		end := start
		for end < len(selector) && depth > 0 {
			if selector[end] == '(' {
				depth++
			} else if selector[end] == ')' {
				depth--
			}
			end++
		}
		inner := selector[start : end-1]
		selector = selector[:idx] + inner + selector[end:]
	}
}

// resolveCompoundTemplate substitutes {value} in a compound variant template
// and processes :merge() functions. If groupName is non-empty, the class inside
// :merge() is suffixed with \/ and the group name (e.g., .group → .group\/sidebar).
func resolveCompoundTemplate(template, value, groupName string) string {
	result := strings.ReplaceAll(template, "{value}", value)

	if groupName == "" {
		return stripMerge(result)
	}

	// When groupName is set, we need to suffix the inner content before stripping.
	for {
		idx := strings.Index(result, ":merge(")
		if idx < 0 {
			break
		}
		start := idx + len(":merge(")
		depth := 1
		end := start
		for end < len(result) && depth > 0 {
			if result[end] == '(' {
				depth++
			} else if result[end] == ')' {
				depth--
			}
			end++
		}
		inner := result[start:end-1] + `\/` + groupName
		result = result[:idx] + inner + result[end:]
	}

	return result
}

// emitCSS serializes generated rules into a CSS string.
func emitCSS(rules []generatedRule, referencedKF map[string]bool, keyframes map[string]*KeyframesRule) string {
	var sb strings.Builder

	// Emit referenced @keyframes before utility rules.
	first := true
	for name := range referencedKF {
		if kf, ok := keyframes[name]; ok {
			if !first {
				sb.WriteByte('\n')
			}
			sb.WriteString(kf.Body)
			sb.WriteByte('\n')
			first = false
		}
	}

	for i, r := range rules {
		if i > 0 || !first {
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

// isHeightProperty returns true if the CSS property is height-related.
func isHeightProperty(prop string) bool {
	switch prop {
	case "height", "min-height", "max-height",
		"inset-block", "inset-block-start", "inset-block-end",
		"top", "bottom":
		return true
	}
	return false
}

// keywordToCSS maps Tailwind value keywords to CSS values.
// The property parameter provides context for keywords like "screen"
// that map differently depending on the CSS property.
func keywordToCSS(s string, property string) string {
	switch s {
	case "full":
		return "100%"
	case "screen":
		if isHeightProperty(property) {
			return "100vh"
		}
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

// fractionToPercent converts "1/2" → "calc(1 / 2 * 100%)", etc.
func fractionToPercent(s string) string {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return ""
	}
	num := parts[0]
	den := parts[1]
	if !isPositiveInt(num) || !isPositiveInt(den) {
		return ""
	}
	if den == "0" {
		return ""
	}
	return "calc(" + num + " / " + den + " * 100%)"
}

// isPositiveInt returns true if s is a non-empty string of digits representing a positive integer.
func isPositiveInt(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// negateValue prepends calc(-1 * ...) for numeric/calc values.
// Returns "" if the value cannot be negated (e.g. keywords like "auto").
func negateValue(v string) string {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "calc(") {
		return "calc(-1 * " + v[5:]
	}
	// Only negate values that start with a digit or a dimension-like token.
	if len(v) > 0 && (v[0] >= '0' && v[0] <= '9' || v[0] == '.') {
		return "calc(-1 * " + v + ")"
	}
	// Cannot negate keyword values like "auto", "none", etc.
	return ""
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
