package tailwind

import (
	"sort"
	"strings"
)

// registerContainerUtility registers the "container" utility class.
// The container utility is unique: it produces a base rule (width: 100%)
// plus one responsive media-query rule per --breakpoint-* theme token.
func registerContainerUtility(idx *utilityIndex, register func(*UtilityRegistration)) {
	reg := &UtilityRegistration{
		Name: "container",
		Kind: "static",
		GenerateRulesFn: func(pc ParsedClass, theme *ThemeConfig, variants map[string]*VariantDef) []generatedRule {
			selector := buildSelector(pc)
			selector = resolveVariantSelector(selector, pc.Variants, variants)
			variantMedia := resolveVariants(pc.Variants, variants)

			// Base rule: width: 100%
			rules := []generatedRule{{
				selector:     selector,
				declarations: []Declaration{{Property: "width", Value: "100%"}},
				important:    pc.Important,
				mediaQueries: variantMedia,
				order:        0,
			}}

			// Collect breakpoints from theme tokens.
			type bp struct {
				value string // e.g. "40rem"
				rem   float64
			}
			var breakpoints []bp
			for k, v := range theme.Tokens {
				if !strings.HasPrefix(k, "--breakpoint-") {
					continue
				}
				// Parse rem value for sorting.
				var rem float64
				if strings.HasSuffix(v, "rem") {
					rem = parseFloat(strings.TrimSuffix(v, "rem"))
				}
				breakpoints = append(breakpoints, bp{value: v, rem: rem})
			}

			// Sort breakpoints by rem value ascending.
			sort.Slice(breakpoints, func(i, j int) bool {
				return breakpoints[i].rem < breakpoints[j].rem
			})

			// Generate one media-query rule per breakpoint.
			for _, b := range breakpoints {
				mq := append([]string{}, variantMedia...)
				mq = append(mq, "@media (width >= "+b.value+")")
				rules = append(rules, generatedRule{
					selector:     selector,
					declarations: []Declaration{{Property: "max-width", Value: b.value}},
					important:    pc.Important,
					mediaQueries: mq,
					order:        0,
				})
			}

			return rules
		},
	}
	register(reg)
}
