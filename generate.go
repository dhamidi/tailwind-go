package tailwind

import (
	"fmt"
	"sort"
	"strings"
)

// nestedBlock is a conditional block nested within a rule (e.g., @supports).
type nestedBlock struct {
	condition    string        // e.g., "@supports (grid: var(--tw))"
	declarations []Declaration // declarations inside the nested block
}

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
	// nested contains conditional blocks inside the rule (e.g., @supports).
	nested []nestedBlock
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

		// Check for multi-rule utilities (e.g., container).
		entry, _ := resolveUtility(pc, utils)
		if reg, ok := entry.(*UtilityRegistration); ok && reg.GenerateRulesFn != nil {
			multiRules := reg.GenerateRulesFn(pc, theme, variants)
			rules = append(rules, multiRules...)
			continue
		}

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
		if rules[i].order != rules[j].order {
			return rules[i].order < rules[j].order
		}
		return rules[i].selector < rules[j].selector
	})

	// Collect @property declarations for any --tw-* custom properties
	// set by the generated rules.
	propertyDecls := collectPropertyDeclarations(rules)

	return emitCSS(rules, referencedKeyframes, keyframes, propertyDecls)
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
	entry, valueStr := resolveUtility(pc, utils)
	if entry == nil {
		return nil
	}

	// Resolve declarations: dispatch to CompileFn for Go-registered
	// utilities or resolveDeclarations for CSS-parsed ones.
	decls := resolveEntryDeclarations(entry, valueStr, pc, theme)
	if decls == nil {
		return nil
	}

	// When before: or after: pseudo-element variants are used,
	// TailwindCSS injects content: var(--tw-content) into the rule.
	for _, v := range pc.Variants {
		if v == "before" || v == "after" {
			decls = append([]Declaration{{Property: "content", Value: "var(--tw-content)"}}, decls...)
			break
		}
	}

	// Build the selector and apply variant selector transforms.
	selector := buildSelector(pc)
	selector = resolveVariantSelector(selector, pc.Variants, variants)

	// Append child selector suffix if the utility defines one.
	if s := entry.utilitySelector(); s != "" {
		if strings.HasPrefix(s, "&") {
			// Pseudo-element/pseudo-class: replace & with the current selector.
			selector = strings.Replace(s, "&", selector, 1)
		} else {
			selector = selector + " " + s
		}
	}

	// Apply variant media query wrapping.
	mediaQueries := resolveVariants(pc.Variants, variants)

	return &generatedRule{
		selector:     selector,
		declarations: decls,
		important:    pc.Important,
		mediaQueries: mediaQueries,
		order:        entry.utilityOrder(),
	}
}

// resolveUtility finds the matching utilityEntry for a parsed class.
// It reconstructs the full utility-value string and uses the index's
// longest-prefix matching to disambiguate patterns like "bg" vs "bg-x".
func resolveUtility(pc ParsedClass, utils *utilityIndex) (utilityEntry, string) {
	// Arbitrary values: the class parser already knows the utility name.
	// We just need to match it against a pattern.
	if pc.Arbitrary != "" {
		// Direct pattern match (dynamic utilities).
		for _, u := range utils.dynamic {
			if pc.Utility == u.utilityPattern() {
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
			if utilPart == u.utilityPattern() {
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
	// Track whether a spacing namespace alternative exists and rejected the value.
	// When this happens, bare numeric values must not be accepted by fallback
	// alternatives (e.g., --value(length, percentage)) to prevent bypassing
	// spacing multiplier validation.
	spacingRejected := false

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

		ns := extractNamespace(d.Value)

		// If a spacing namespace already rejected this bare numeric value,
		// skip non-namespace alternatives that would accept it as length/percentage.
		if spacingRejected && ns == "" && pc.Arbitrary == "" && isNumeric(valueStr) {
			continue
		}

		cssValue := resolveValueForDecl(d, valueStr, pc, theme)
		if cssValue != "" {
			resolved := substituteValue(d.Value, cssValue)
			// Apply opacity modifier if the declaration used a color namespace.
			if pc.Modifier != "" && ns == "color" {
				// For theme-resolved colors, use the CSS variable reference
				// instead of the literal value for color-mix output.
				if pc.Arbitrary == "" && valueStr != "" {
					if _, ok := theme.Resolve(ns, valueStr); ok {
						varRef := "var(--" + ns + "-" + valueStr + ")"
						resolved = applyModifier(varRef, pc.Modifier, theme)
					} else {
						resolved = applyModifier(resolved, pc.Modifier, theme)
					}
				} else {
					resolved = applyModifier(resolved, pc.Modifier, theme)
				}
			}
			return &Declaration{Property: d.Property, Value: resolved}
		}

		// Track if the spacing namespace rejected this value.
		if ns == "spacing" && isNumeric(valueStr) {
			spacingRejected = true
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
	// Only apply fraction-to-percent for declarations whose value types include
	// percentage or length (or sizing namespaces). Integer-only utilities like
	// z-index and order must not accept fractions.
	if strings.Contains(valueStr, "/") && declarationAcceptsFraction(d) {
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
	// If --value() has no args at all, accept any value (e.g., col-span-*)
	if len(valueTypes) == 0 && strings.Contains(d.Value, "--value()") {
		if isNumeric(valueStr) {
			v := valueStr
			if pc.Negative {
				v = negateValue(v)
			}
			return v
		}
		return valueStr
	}
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
				// Apply bare value unit suffix first, then negate.
				v = applyBareValueUnit(v, d)
				if pc.Negative {
					v = negateValue(v)
				}
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

// substituteValue replaces all --value(...) occurrences in a declaration with the resolved value.
func substituteValue(template, resolved string) string {
	result := template
	for {
		idx := strings.Index(result, "--value(")
		if idx < 0 {
			return result
		}
		// Find the matching closing paren.
		depth := 0
		end := idx + len("--value(")
		found := false
		for end < len(result) {
			if result[end] == '(' {
				depth++
			} else if result[end] == ')' {
				if depth == 0 {
					result = result[:idx] + resolved + result[end+1:]
					found = true
					break
				}
				depth--
			}
			end++
		}
		if !found {
			result = result[:idx] + resolved
			break
		}
	}
	return result
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

// isAtRuleOnlyVariant returns true if the variant name represents a dynamically
// resolved at-rule variant that should be skipped during selector resolution
// (handled by resolveVariants instead).
func isAtRuleOnlyVariant(name string) bool {
	if strings.HasPrefix(name, "min-[") || strings.HasPrefix(name, "max-[") {
		return true
	}
	if strings.HasPrefix(name, "@min-[") || strings.HasPrefix(name, "@max-[") {
		return true
	}
	return false
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

		// Handle supports-* variants (feature queries).
		// Check for exact registered variant first (e.g., CSS-defined @variant supports-grid).
		if strings.HasPrefix(name, "not-supports-") {
			if _, ok := defs[name]; !ok {
				remainder := name[len("not-supports-"):]
				if strings.HasPrefix(remainder, "[") && strings.HasSuffix(remainder, "]") {
					inner := remainder[1 : len(remainder)-1]
					inner = strings.ReplaceAll(inner, "_", " ")
					inner = normalizeSupportsCondition(inner)
					media = append(media, "@supports not ("+inner+")")
				} else {
					prop := strings.ReplaceAll(remainder, "_", " ")
					media = append(media, "@supports not ("+prop+": var(--tw))")
				}
				continue
			}
		}
		if strings.HasPrefix(name, "supports-") {
			if _, ok := defs[name]; !ok {
				remainder := name[len("supports-"):]
				if strings.HasPrefix(remainder, "[") && strings.HasSuffix(remainder, "]") {
					inner := remainder[1 : len(remainder)-1]
					inner = strings.ReplaceAll(inner, "_", " ")
					// Normalize "property:value" to "property: value" for readability.
					inner = normalizeSupportsCondition(inner)
					media = append(media, "@supports ("+inner+")")
				} else {
					prop := strings.ReplaceAll(remainder, "_", " ")
					media = append(media, "@supports ("+prop+": var(--tw))")
				}
				continue
			}
		}

		// Handle min-[...] and max-[...] arbitrary breakpoints.
		if strings.HasPrefix(name, "min-[") && strings.HasSuffix(name, "]") {
			val := name[len("min-[") : len(name)-1]
			val = strings.ReplaceAll(val, "_", " ")
			media = append(media, "@media (width >= "+val+")")
			continue
		}
		if strings.HasPrefix(name, "max-[") && strings.HasSuffix(name, "]") {
			val := name[len("max-[") : len(name)-1]
			val = strings.ReplaceAll(val, "_", " ")
			media = append(media, "@media (width < "+val+")")
			continue
		}

		// Handle @min-[...] and @max-[...] arbitrary container queries.
		if strings.HasPrefix(name, "@min-[") && strings.HasSuffix(name, "]") {
			val := name[len("@min-[") : len(name)-1]
			val = strings.ReplaceAll(val, "_", " ")
			media = append(media, "@container (width >= "+val+")")
			continue
		}
		if strings.HasPrefix(name, "@max-[") && strings.HasSuffix(name, "]") {
			val := name[len("@max-[") : len(name)-1]
			val = strings.ReplaceAll(val, "_", " ")
			media = append(media, "@container (width < "+val+")")
			continue
		}

		// Handle not-<media-variant> by negating the inner variant's media query,
		// but only if no explicit "not-*" variant is registered.
		if strings.HasPrefix(name, "not-") {
			if _, hasExplicit := defs[name]; !hasExplicit {
				innerName := name[4:]
				if innerV, ok := defs[innerName]; ok && innerV.Media != "" && innerV.Selector == "" {
					if innerV.AtRule != "" {
						media = append(media, "@"+innerV.AtRule+" not "+innerV.Media)
					} else {
						media = append(media, "@media not "+innerV.Media)
					}
					continue
				}
			}
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

		if v, ok := defs[name]; ok {
			if v.Selector != "" {
				sel = resolveMultiSelector(v.Selector, sel)
			}
			// Exact match found — skip compound variant resolution
			// even for media-only variants.
			continue
		}

		// Skip variants handled as at-rules (not selector transforms).
		if isAtRuleOnlyVariant(name) {
			continue
		}

		// Skip supports-* and not-supports-* dynamic variants (at-rule only).
		if strings.HasPrefix(name, "supports-") || strings.HasPrefix(name, "not-supports-") {
			if _, ok := defs[name]; !ok {
				continue
			}
		}

		// Skip not-<media-variant> — handled as media wrapper in resolveVariants.
		if strings.HasPrefix(name, "not-") {
			if _, hasExplicit := defs[name]; !hasExplicit {
				innerName := name[4:]
				if innerV, ok := defs[innerName]; ok && innerV.Media != "" && innerV.Selector == "" {
					continue
				}
			}
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

			// Reject compound variants with incompatible inner variants:
			// pseudo-elements (before, after, etc.) and * / ** cannot be compounded.
			if v.Name == "not" || v.Name == "has" {
				if value == "*" || value == "**" {
					continue
				}
				if innerDef, ok := defs[value]; ok && innerDef.Selector != "" {
					if strings.Contains(innerDef.Selector, "&::") {
						continue
					}
				}
			}

			template := v.Template
			isArbitrary := strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]")
			// Strip brackets from arbitrary values and adjust template.
			if isArbitrary {
				value = value[1 : len(value)-1]
				value = strings.ReplaceAll(value, "_", " ")
				// Remove ="true" for arbitrary ARIA values.
				template = strings.ReplaceAll(template, `="true"`, "")
				// For templates with :{value}, drop the template's colon prefix
				// so the arbitrary value is used verbatim. This avoids double colons
				// (e.g., has-[:checked] → :has(:checked)) and spurious colons
				// (e.g., has-[>img] → :has(>img)).
				template = strings.ReplaceAll(template, ":{value}", "{value}")
			} else if (v.Name == "not" || v.Name == "has" || v.Name == "group" || v.Name == "peer") && !isArbitrary {
				// For not-*, has-*, group-*, and peer-* with non-arbitrary values, resolve the inner
				// variant name through the registry (e.g., "first" → "first-child",
				// "open" → "is([open], :popover-open, :open)").
				if innerDef, ok := defs[value]; ok && innerDef.Selector != "" {
					// Extract the pseudo-class from the selector (e.g., "&:first-child" → "first-child").
					innerSel := innerDef.Selector
					if idx := strings.Index(innerSel, "&:"); idx >= 0 {
						value = innerSel[idx+2:]
					}
				}
			}
			selector := resolveCompoundTemplate(template, value, groupName)
			sel = strings.ReplaceAll(selector, "&", sel)
		}
	}
	return sel
}

// resolveMultiSelector handles comma-separated selector templates.
// For example, "& *::marker, &::marker" with base ".foo" produces
// ".foo *::marker, .foo::marker".
// Commas inside parentheses (e.g., ":is([open], :popover-open)") are preserved.
func resolveMultiSelector(template, base string) string {
	// Split on commas that are NOT inside parentheses.
	var parts []string
	depth := 0
	start := 0
	for i, ch := range template {
		switch ch {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, template[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, template[start:])

	for i, p := range parts {
		parts[i] = strings.ReplaceAll(strings.TrimSpace(p), "&", base)
	}
	return strings.Join(parts, ",\n")
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
// Returns the longest matching compound variant to handle nested compounds
// like "group-aria-checked" matching "group-aria" over "group".
func lookupCompoundVariant(name string, defs map[string]*VariantDef) *VariantDef {
	var best *VariantDef
	for _, v := range defs {
		if !v.Compound {
			continue
		}
		prefix := v.Name + "-"
		if strings.HasPrefix(name, prefix) && len(name) > len(prefix) {
			if best == nil || len(v.Name) > len(best.Name) {
				best = v
			}
		}
	}
	return best
}

// normalizeSupportsCondition adds a space after colons in @supports conditions
// that are missing one (e.g., "display:grid" → "display: grid").
func normalizeSupportsCondition(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		result.WriteByte(s[i])
		if s[i] == ':' && i+1 < len(s) && s[i+1] != ' ' {
			result.WriteByte(' ')
		}
	}
	return result.String()
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

// twPropertyDefs maps --tw-* custom property names to their @property
// declaration blocks. Used by collectPropertyDeclarations to emit
// @property rules for custom properties referenced by generated utilities.
var twPropertyDefs = map[string]string{
	"--tw-translate-x":         "@property --tw-translate-x {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-translate-y":         "@property --tw-translate-y {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-translate-z":         "@property --tw-translate-z {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-scale-x":             "@property --tw-scale-x {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 1;\n}\n",
	"--tw-scale-y":             "@property --tw-scale-y {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 1;\n}\n",
	"--tw-scale-z":             "@property --tw-scale-z {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 1;\n}\n",
	"--tw-rotate-x":            "@property --tw-rotate-x {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-rotate-y":            "@property --tw-rotate-y {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-rotate-z":            "@property --tw-rotate-z {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-skew-x":              "@property --tw-skew-x {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-skew-y":              "@property --tw-skew-y {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-shadow":              "@property --tw-shadow {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0 0 #0000;\n}\n",
	"--tw-shadow-color":        "@property --tw-shadow-color {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-shadow-alpha":        "@property --tw-shadow-alpha {\n  syntax: \"<percentage>\";\n  inherits: false;\n  initial-value: 100%;\n}\n",
	"--tw-inset-shadow":        "@property --tw-inset-shadow {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0 0 #0000;\n}\n",
	"--tw-inset-shadow-color":  "@property --tw-inset-shadow-color {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-inset-shadow-alpha":  "@property --tw-inset-shadow-alpha {\n  syntax: \"<percentage>\";\n  inherits: false;\n  initial-value: 100%;\n}\n",
	"--tw-ring-color":          "@property --tw-ring-color {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-ring-shadow":         "@property --tw-ring-shadow {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0 0 #0000;\n}\n",
	"--tw-inset-ring-color":    "@property --tw-inset-ring-color {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-inset-ring-shadow":   "@property --tw-inset-ring-shadow {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0 0 #0000;\n}\n",
	"--tw-ring-inset":          "@property --tw-ring-inset {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-ring-offset-width":   "@property --tw-ring-offset-width {\n  syntax: \"<length>\";\n  inherits: false;\n  initial-value: 0px;\n}\n",
	"--tw-ring-offset-color":   "@property --tw-ring-offset-color {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: #fff;\n}\n",
	"--tw-ring-offset-shadow":  "@property --tw-ring-offset-shadow {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0 0 #0000;\n}\n",
	"--tw-blur":                "@property --tw-blur {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-brightness":          "@property --tw-brightness {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-contrast":            "@property --tw-contrast {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-grayscale":           "@property --tw-grayscale {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-hue-rotate":          "@property --tw-hue-rotate {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-invert":              "@property --tw-invert {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-opacity":             "@property --tw-opacity {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-saturate":            "@property --tw-saturate {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-sepia":               "@property --tw-sepia {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-drop-shadow":         "@property --tw-drop-shadow {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-drop-shadow-color":   "@property --tw-drop-shadow-color {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-drop-shadow-alpha":   "@property --tw-drop-shadow-alpha {\n  syntax: \"<percentage>\";\n  inherits: false;\n  initial-value: 100%;\n}\n",
	"--tw-drop-shadow-size":    "@property --tw-drop-shadow-size {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-space-y-reverse":     "@property --tw-space-y-reverse {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-space-x-reverse":     "@property --tw-space-x-reverse {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-border-style":        "@property --tw-border-style {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: solid;\n}\n",
	"--tw-leading":             "@property --tw-leading {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-font-weight":         "@property --tw-font-weight {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-tracking":            "@property --tw-tracking {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-duration":            "@property --tw-duration {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-ease":                "@property --tw-ease {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-outline-style":       "@property --tw-outline-style {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-border-spacing-x":    "@property --tw-border-spacing-x {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-border-spacing-y":    "@property --tw-border-spacing-y {\n  syntax: \"*\";\n  inherits: false;\n  initial-value: 0;\n}\n",
	"--tw-gradient-from":       "@property --tw-gradient-from {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-gradient-via":        "@property --tw-gradient-via {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-gradient-to":         "@property --tw-gradient-to {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-gradient-stops":      "@property --tw-gradient-stops {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-gradient-from-position": "@property --tw-gradient-from-position {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-gradient-via-position":  "@property --tw-gradient-via-position {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-gradient-to-position":   "@property --tw-gradient-to-position {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-ordinal":             "@property --tw-ordinal {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-slashed-zero":        "@property --tw-slashed-zero {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-numeric-figure":      "@property --tw-numeric-figure {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-numeric-spacing":     "@property --tw-numeric-spacing {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-numeric-fraction":    "@property --tw-numeric-fraction {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-pan-x":               "@property --tw-pan-x {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-pan-y":               "@property --tw-pan-y {\n  syntax: \"*\";\n  inherits: false;\n}\n",
	"--tw-pinch-zoom":          "@property --tw-pinch-zoom {\n  syntax: \"*\";\n  inherits: false;\n}\n",
}

// collectPropertyDeclarations scans generated rules for --tw-* custom
// property declarations and returns the corresponding sorted @property
// declaration strings.
func collectPropertyDeclarations(rules []generatedRule) []string {
	seen := make(map[string]bool)
	for _, r := range rules {
		for _, d := range r.declarations {
			if strings.HasPrefix(d.Property, "--tw-") {
				if _, ok := twPropertyDefs[d.Property]; ok {
					seen[d.Property] = true
				}
			}
			// Also scan values for var(--tw-*) references that need @property.
			scanVarRefs(d.Value, seen)
		}
	}
	if len(seen) == 0 {
		return nil
	}
	props := make([]string, 0, len(seen))
	for p := range seen {
		props = append(props, p)
	}
	sort.Strings(props)
	var result []string
	for _, p := range props {
		result = append(result, twPropertyDefs[p])
	}
	return result
}

// scanVarRefs finds var(--tw-*) references in a CSS value and marks
// the referenced properties in the seen map if they have @property defs.
func scanVarRefs(value string, seen map[string]bool) {
	s := value
	for {
		idx := strings.Index(s, "var(--tw-")
		if idx < 0 {
			return
		}
		start := idx + len("var(")
		// Find the end of the property name (comma or closing paren).
		end := start
		for end < len(s) && s[end] != ',' && s[end] != ')' {
			end++
		}
		prop := s[start:end]
		if _, ok := twPropertyDefs[prop]; ok {
			seen[prop] = true
		}
		s = s[end:]
	}
}

// emitCSS serializes generated rules into a CSS string.
func emitCSS(rules []generatedRule, referencedKF map[string]bool, keyframes map[string]*KeyframesRule, propertyDecls []string) string {
	var sb strings.Builder

	// Emit @property declarations before everything else.
	for i, pd := range propertyDecls {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(pd)
	}

	// Emit referenced @keyframes before utility rules.
	first := len(propertyDecls) == 0
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

		// Nested blocks (e.g., @supports).
		for _, nb := range r.nested {
			sb.WriteString("  ")
			sb.WriteString(nb.condition)
			sb.WriteString(" {\n")
			for _, d := range nb.declarations {
				sb.WriteString("    ")
				sb.WriteString(d.Property)
				sb.WriteString(": ")
				sb.WriteString(d.Value)
				if r.important {
					sb.WriteString(" !important")
				}
				sb.WriteString(";\n")
			}
			sb.WriteString("  }\n")
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

	decls := []Declaration{{
		Property: pc.Utility,
		Value:    pc.Arbitrary,
	}}

	// Inject content: var(--tw-content) for before:/after: variants.
	for _, v := range pc.Variants {
		if v == "before" || v == "after" {
			decls = append([]Declaration{{Property: "content", Value: "var(--tw-content)"}}, decls...)
			break
		}
	}

	return &generatedRule{
		selector:     selector,
		declarations: decls,
		important:    pc.Important,
		mediaQueries: mediaQueries,
		order:        9999, // arbitrary properties sort last
	}
}

// isHeightProperty returns true if the CSS property is height-related.
func isHeightProperty(prop string) bool {
	switch prop {
	case "height", "min-height", "max-height",
		"block-size", "min-block-size", "max-block-size",
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

// declarationAcceptsFraction returns true if a declaration's --value() type
// is compatible with fraction values (percentage-based). Declarations with
// integer-only types, or namespace-only declarations for non-sizing namespaces
// (like --z-index, --opacity), do not accept fractions.
func declarationAcceptsFraction(d Declaration) bool {
	types := extractValueTypes(d.Value)
	ns := extractNamespace(d.Value)

	// Sizing namespaces accept fractions (they produce percentages).
	switch ns {
	case "spacing", "width", "height", "size", "container":
		return true
	}

	// If no explicit types and just a namespace (e.g., --value(--z-index)),
	// do not accept fractions.
	if len(types) == 0 {
		return ns == ""
	}

	// Check if any explicit type is fraction-compatible.
	for _, t := range types {
		if t == "percentage" || t == "length" {
			return true
		}
	}
	return false
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

// negateValue negates a CSS value.
// Returns "" if the value cannot be negated (e.g. keywords like "auto").
func negateValue(v string) string {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "calc(") && strings.HasSuffix(v, ")") {
		inner := v[5 : len(v)-1]
		// For "var(--X) * N" patterns (e.g. spacing calc), negate the multiplier directly.
		if idx := strings.LastIndex(inner, " * "); idx >= 0 {
			base := inner[:idx]
			multiplier := inner[idx+3:]
			if strings.HasPrefix(base, "var(") && len(multiplier) > 0 && (multiplier[0] >= '0' && multiplier[0] <= '9' || multiplier[0] == '.') {
				return "calc(" + base + " * -" + multiplier + ")"
			}
		}
		return "calc(" + v + " * -1)"
	}
	// Only negate values that start with a digit or a dimension-like token.
	if len(v) > 0 && (v[0] >= '0' && v[0] <= '9' || v[0] == '.') {
		return "calc(" + v + " * -1)"
	}
	// Cannot negate keyword values like "auto", "none", etc.
	return ""
}

// applyModifier wraps a CSS color value with color-mix opacity modifier.
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
	// 100% opacity is identity — no wrapping needed.
	if opacityStr == "100%" {
		return cssValue
	}
	return "color-mix(in oklab, " + cssValue + " " + opacityStr + ", transparent)"
}

// resolveModifierOpacity resolves an opacity modifier value.
// Numeric modifiers are always treated as percentages.
// Non-numeric modifiers check the theme for --opacity-{modifier}.
func resolveModifierOpacity(modifier string, theme *ThemeConfig) string {
	if isNumeric(modifier) {
		return modifier + "%"
	}
	if v, ok := theme.Resolve("opacity", modifier); ok {
		return v
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
