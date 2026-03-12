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

	// ===== Responsive variants =====
	add(&VariantDef{Name: "sm", Media: "(width >= 40rem)"})
	add(&VariantDef{Name: "md", Media: "(width >= 48rem)"})
	add(&VariantDef{Name: "lg", Media: "(width >= 64rem)"})
	add(&VariantDef{Name: "xl", Media: "(width >= 80rem)"})
	add(&VariantDef{Name: "2xl", Media: "(width >= 96rem)"})

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

	// ===== Parameterized variants =====
	add(&VariantDef{Name: "not", Compound: true, Template: "&:not(:{value})"})
	add(&VariantDef{Name: "has", Compound: true, Template: "&:has(:{value})"})
	add(&VariantDef{Name: "in", Compound: true, Template: "[{value}] &"})

	// ===== Starting style variant =====
	add(&VariantDef{Name: "starting", AtRule: "starting-style"})

	return order
}
