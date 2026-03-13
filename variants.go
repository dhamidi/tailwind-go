package tailwind

// registerGoVariants registers all built-in variant definitions in Go.
// These replace the equivalent @variant directives previously in utilities.css.
func registerGoVariants(variants map[string]*VariantDef, startOrder int) int {
	order := startOrder

	add := func(v *VariantDef) {
		order++
		v.Order = order
		variants[v.Name] = v
	}

	// ===== Pseudo-class variants =====
	add(&VariantDef{Name: "hover", Selector: "&:hover", Media: "(hover: hover)"})
	add(&VariantDef{Name: "focus", Selector: "&:focus"})
	add(&VariantDef{Name: "focus-within", Selector: "&:focus-within"})
	add(&VariantDef{Name: "focus-visible", Selector: "&:focus-visible"})
	add(&VariantDef{Name: "active", Selector: "&:active"})
	add(&VariantDef{Name: "visited", Selector: "&:visited"})
	add(&VariantDef{Name: "target", Selector: "&:target"})
	add(&VariantDef{Name: "first", Selector: "&:first-child"})
	add(&VariantDef{Name: "last", Selector: "&:last-child"})
	add(&VariantDef{Name: "only", Selector: "&:only-child"})
	add(&VariantDef{Name: "odd", Selector: "&:nth-child(odd)"})
	add(&VariantDef{Name: "even", Selector: "&:nth-child(even)"})
	add(&VariantDef{Name: "first-of-type", Selector: "&:first-of-type"})
	add(&VariantDef{Name: "last-of-type", Selector: "&:last-of-type"})
	add(&VariantDef{Name: "only-of-type", Selector: "&:only-of-type"})
	add(&VariantDef{Name: "empty", Selector: "&:empty"})
	add(&VariantDef{Name: "disabled", Selector: "&:disabled"})
	add(&VariantDef{Name: "enabled", Selector: "&:enabled"})
	add(&VariantDef{Name: "checked", Selector: "&:checked"})
	add(&VariantDef{Name: "indeterminate", Selector: "&:indeterminate"})
	add(&VariantDef{Name: "default", Selector: "&:default"})
	add(&VariantDef{Name: "required", Selector: "&:required"})
	add(&VariantDef{Name: "valid", Selector: "&:valid"})
	add(&VariantDef{Name: "invalid", Selector: "&:invalid"})
	add(&VariantDef{Name: "placeholder-shown", Selector: "&:placeholder-shown"})
	add(&VariantDef{Name: "autofill", Selector: "&:autofill"})
	add(&VariantDef{Name: "read-only", Selector: "&:read-only"})
	add(&VariantDef{Name: "open", Selector: "&[open]"})
	add(&VariantDef{Name: "user-valid", Selector: "&:user-valid"})
	add(&VariantDef{Name: "user-invalid", Selector: "&:user-invalid"})
	add(&VariantDef{Name: "optional", Selector: "&:optional"})
	add(&VariantDef{Name: "in-range", Selector: "&:in-range"})
	add(&VariantDef{Name: "out-of-range", Selector: "&:out-of-range"})

	// ===== Pseudo-element variants =====
	add(&VariantDef{Name: "before", Selector: "&::before"})
	add(&VariantDef{Name: "after", Selector: "&::after"})
	add(&VariantDef{Name: "placeholder", Selector: "&::placeholder"})
	add(&VariantDef{Name: "file", Selector: "&::file-selector-button"})
	add(&VariantDef{Name: "marker", Selector: "& *::marker, & *::-webkit-details-marker, &::marker, &::-webkit-details-marker"})
	add(&VariantDef{Name: "selection", Selector: "& *::selection, &::selection"})
	add(&VariantDef{Name: "first-line", Selector: "&::first-line"})
	add(&VariantDef{Name: "first-letter", Selector: "&::first-letter"})
	add(&VariantDef{Name: "backdrop", Selector: "&::backdrop"})
	add(&VariantDef{Name: "details-content", Selector: "&::details-content"})

	// ===== Child/descendant selector variants =====
	add(&VariantDef{Name: "*", Selector: "& > *"})
	add(&VariantDef{Name: "**", Selector: "& *"})

	// ===== ARIA attribute variants =====
	add(&VariantDef{Name: "aria-busy", Selector: `&[aria-busy="true"]`})
	add(&VariantDef{Name: "aria-checked", Selector: `&[aria-checked="true"]`})
	add(&VariantDef{Name: "aria-disabled", Selector: `&[aria-disabled="true"]`})
	add(&VariantDef{Name: "aria-expanded", Selector: `&[aria-expanded="true"]`})
	add(&VariantDef{Name: "aria-hidden", Selector: `&[aria-hidden="true"]`})
	add(&VariantDef{Name: "aria-pressed", Selector: `&[aria-pressed="true"]`})
	add(&VariantDef{Name: "aria-readonly", Selector: `&[aria-readonly="true"]`})
	add(&VariantDef{Name: "aria-required", Selector: `&[aria-required="true"]`})
	add(&VariantDef{Name: "aria-selected", Selector: `&[aria-selected="true"]`})

	// ===== Responsive variants =====
	add(&VariantDef{Name: "sm", Media: "(width >= 40rem)"})
	add(&VariantDef{Name: "md", Media: "(width >= 48rem)"})
	add(&VariantDef{Name: "lg", Media: "(width >= 64rem)"})
	add(&VariantDef{Name: "xl", Media: "(width >= 80rem)"})
	add(&VariantDef{Name: "2xl", Media: "(width >= 96rem)"})

	// ===== max-* responsive variants =====
	add(&VariantDef{Name: "max-sm", Media: "(width < 40rem)"})
	add(&VariantDef{Name: "max-md", Media: "(width < 48rem)"})
	add(&VariantDef{Name: "max-lg", Media: "(width < 64rem)"})
	add(&VariantDef{Name: "max-xl", Media: "(width < 80rem)"})
	add(&VariantDef{Name: "max-2xl", Media: "(width < 96rem)"})

	// ===== Container query variants =====
	add(&VariantDef{Name: "@3xs", AtRule: "container", Media: "(width >= 16rem)"})
	add(&VariantDef{Name: "@2xs", AtRule: "container", Media: "(width >= 18rem)"})
	add(&VariantDef{Name: "@xs", AtRule: "container", Media: "(width >= 20rem)"})
	add(&VariantDef{Name: "@sm", AtRule: "container", Media: "(width >= 24rem)"})
	add(&VariantDef{Name: "@md", AtRule: "container", Media: "(width >= 28rem)"})
	add(&VariantDef{Name: "@lg", AtRule: "container", Media: "(width >= 32rem)"})
	add(&VariantDef{Name: "@xl", AtRule: "container", Media: "(width >= 36rem)"})
	add(&VariantDef{Name: "@2xl", AtRule: "container", Media: "(width >= 42rem)"})
	add(&VariantDef{Name: "@3xl", AtRule: "container", Media: "(width >= 48rem)"})
	add(&VariantDef{Name: "@4xl", AtRule: "container", Media: "(width >= 56rem)"})
	add(&VariantDef{Name: "@5xl", AtRule: "container", Media: "(width >= 64rem)"})
	add(&VariantDef{Name: "@6xl", AtRule: "container", Media: "(width >= 72rem)"})
	add(&VariantDef{Name: "@7xl", AtRule: "container", Media: "(width >= 80rem)"})

	// ===== Container query max-* variants =====
	add(&VariantDef{Name: "@max-3xs", AtRule: "container", Media: "(width < 16rem)"})
	add(&VariantDef{Name: "@max-2xs", AtRule: "container", Media: "(width < 18rem)"})
	add(&VariantDef{Name: "@max-xs", AtRule: "container", Media: "(width < 20rem)"})
	add(&VariantDef{Name: "@max-sm", AtRule: "container", Media: "(width < 24rem)"})
	add(&VariantDef{Name: "@max-md", AtRule: "container", Media: "(width < 28rem)"})
	add(&VariantDef{Name: "@max-lg", AtRule: "container", Media: "(width < 32rem)"})
	add(&VariantDef{Name: "@max-xl", AtRule: "container", Media: "(width < 36rem)"})
	add(&VariantDef{Name: "@max-2xl", AtRule: "container", Media: "(width < 42rem)"})
	add(&VariantDef{Name: "@max-3xl", AtRule: "container", Media: "(width < 48rem)"})
	add(&VariantDef{Name: "@max-4xl", AtRule: "container", Media: "(width < 56rem)"})
	add(&VariantDef{Name: "@max-5xl", AtRule: "container", Media: "(width < 64rem)"})
	add(&VariantDef{Name: "@max-6xl", AtRule: "container", Media: "(width < 72rem)"})
	add(&VariantDef{Name: "@max-7xl", AtRule: "container", Media: "(width < 80rem)"})

	// ===== Media query variants =====
	add(&VariantDef{Name: "dark", Media: "(prefers-color-scheme: dark)"})
	add(&VariantDef{Name: "print", Media: "print"})
	add(&VariantDef{Name: "portrait", Media: "(orientation: portrait)"})
	add(&VariantDef{Name: "landscape", Media: "(orientation: landscape)"})
	add(&VariantDef{Name: "motion-safe", Media: "(prefers-reduced-motion: no-preference)"})
	add(&VariantDef{Name: "motion-reduce", Media: "(prefers-reduced-motion: reduce)"})
	add(&VariantDef{Name: "contrast-more", Media: "(prefers-contrast: more)"})
	add(&VariantDef{Name: "contrast-less", Media: "(prefers-contrast: less)"})
	add(&VariantDef{Name: "forced-colors", Media: "(forced-colors: active)"})
	add(&VariantDef{Name: "not-forced-colors", Media: "(forced-colors: none)"})

	// ===== Pointer/input device variants =====
	add(&VariantDef{Name: "pointer-fine", Media: "(pointer: fine)"})
	add(&VariantDef{Name: "pointer-coarse", Media: "(pointer: coarse)"})
	add(&VariantDef{Name: "pointer-none", Media: "(pointer: none)"})
	add(&VariantDef{Name: "any-pointer-fine", Media: "(any-pointer: fine)"})
	add(&VariantDef{Name: "any-pointer-coarse", Media: "(any-pointer: coarse)"})
	add(&VariantDef{Name: "any-pointer-none", Media: "(any-pointer: none)"})

	// ===== Additional media query variants =====
	add(&VariantDef{Name: "noscript", Media: "(scripting: none)"})
	add(&VariantDef{Name: "inverted-colors", Media: "(inverted-colors: inverted)"})

	// ===== Parameterized variants =====
	add(&VariantDef{Name: "not", Compound: true, Template: "&:not(:{value})"})
	add(&VariantDef{Name: "has", Compound: true, Template: "&:has(:{value})"})
	add(&VariantDef{Name: "in", Compound: true, Template: "[{value}] &"})

	// ARIA/data compound variants (for arbitrary values like aria-[...], data-[...])
	add(&VariantDef{Name: "aria", Compound: true, Template: `&[aria-{value}="true"]`})
	add(&VariantDef{Name: "data", Compound: true, Template: "&[data-{value}]"})

	// Group and peer compound variants
	add(&VariantDef{Name: "group", Compound: true, Template: ":merge(.group):{value} &"})
	add(&VariantDef{Name: "peer", Compound: true, Template: ":merge(.peer):{value} ~ &"})

	// Group/peer + ARIA compound variants
	add(&VariantDef{Name: "group-aria", Compound: true, Template: `:merge(.group)[aria-{value}="true"] &`})
	add(&VariantDef{Name: "peer-aria", Compound: true, Template: `:merge(.peer)[aria-{value}="true"] ~ &`})

	// Group/peer + data compound variants
	add(&VariantDef{Name: "group-data", Compound: true, Template: ":merge(.group)[data-{value}] &"})
	add(&VariantDef{Name: "peer-data", Compound: true, Template: ":merge(.peer)[data-{value}] ~ &"})

	// Parameterized nth-* variants
	add(&VariantDef{Name: "nth", Compound: true, Template: "&:nth-child({value})"})
	add(&VariantDef{Name: "nth-last", Compound: true, Template: "&:nth-last-child({value})"})
	add(&VariantDef{Name: "nth-of-type", Compound: true, Template: "&:nth-of-type({value})"})
	add(&VariantDef{Name: "nth-last-of-type", Compound: true, Template: "&:nth-last-of-type({value})"})

	// ===== Starting style variant =====
	add(&VariantDef{Name: "starting", AtRule: "starting-style"})

	return order
}
