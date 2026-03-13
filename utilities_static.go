package tailwind

// decls is a convenience constructor for a []Declaration from alternating
// property/value strings: decls("display", "flex") → [{Property:"display", Value:"flex"}].
func decls(pairs ...string) []Declaration {
	out := make([]Declaration, len(pairs)/2)
	for i := 0; i < len(pairs)-1; i += 2 {
		out[i/2] = Declaration{Property: pairs[i], Value: pairs[i+1]}
	}
	return out
}

// staticUtilityWithSelector creates a static utility that also carries
// a child-selector suffix (e.g., "> :not(:last-child)").
func staticUtilityWithSelector(name, selector string, d []Declaration) *UtilityRegistration {
	reg := staticUtility(name, d)
	reg.Selector = selector
	return reg
}

// registerStaticUtilities registers all static (fixed-declaration) utilities.
// These replace the equivalent @utility blocks in utilities.css and produce
// identical CSS output.
func registerStaticUtilities(idx *utilityIndex, register func(*UtilityRegistration)) {
	// ===== Screen Reader =====
	register(staticUtility("sr-only", decls(
		"position", "absolute",
		"width", "1px",
		"height", "1px",
		"padding", "0",
		"margin", "-1px",
		"overflow", "hidden",
		"clip-path", "inset(50%)",
		"white-space", "nowrap",
		"border-width", "0",
	)))
	register(staticUtility("not-sr-only", decls(
		"position", "static",
		"width", "auto",
		"height", "auto",
		"padding", "0",
		"margin", "0",
		"overflow", "visible",
		"clip-path", "none",
		"white-space", "normal",
	)))

	// ===== Forced Color Adjust =====
	register(staticUtility("forced-color-adjust-auto", decls("forced-color-adjust", "auto")))
	register(staticUtility("forced-color-adjust-none", decls("forced-color-adjust", "none")))

	// ===== Display =====
	register(staticUtility("block", decls("display", "block")))
	register(staticUtility("inline-block", decls("display", "inline-block")))
	register(staticUtility("inline", decls("display", "inline")))
	register(staticUtility("flex", decls("display", "flex")))
	register(staticUtility("inline-flex", decls("display", "inline-flex")))
	register(staticUtility("grid", decls("display", "grid")))
	register(staticUtility("inline-grid", decls("display", "inline-grid")))
	register(staticUtility("contents", decls("display", "contents")))
	register(staticUtility("list-item", decls("display", "list-item")))
	register(staticUtility("hidden", decls("display", "none")))
	register(staticUtility("table", decls("display", "table")))
	register(staticUtility("table-caption", decls("display", "table-caption")))
	register(staticUtility("table-cell", decls("display", "table-cell")))
	register(staticUtility("table-column", decls("display", "table-column")))
	register(staticUtility("table-column-group", decls("display", "table-column-group")))
	register(staticUtility("table-footer-group", decls("display", "table-footer-group")))
	register(staticUtility("table-header-group", decls("display", "table-header-group")))
	register(staticUtility("table-row-group", decls("display", "table-row-group")))
	register(staticUtility("table-row", decls("display", "table-row")))
	register(staticUtility("flow-root", decls("display", "flow-root")))

	// ===== Visibility =====
	register(staticUtility("visible", decls("visibility", "visible")))
	register(staticUtility("invisible", decls("visibility", "hidden")))
	register(staticUtility("collapse", decls("visibility", "collapse")))

	// ===== Position =====
	register(staticUtility("static", decls("position", "static")))
	register(staticUtility("fixed", decls("position", "fixed")))
	register(staticUtility("absolute", decls("position", "absolute")))
	register(staticUtility("relative", decls("position", "relative")))
	register(staticUtility("sticky", decls("position", "sticky")))

	// ===== Isolation =====
	register(staticUtility("isolate", decls("isolation", "isolate")))
	register(staticUtility("isolation-auto", decls("isolation", "auto")))

	// ===== Overflow =====
	register(staticUtility("overflow-auto", decls("overflow", "auto")))
	register(staticUtility("overflow-hidden", decls("overflow", "hidden")))
	register(staticUtility("overflow-clip", decls("overflow", "clip")))
	register(staticUtility("overflow-visible", decls("overflow", "visible")))
	register(staticUtility("overflow-scroll", decls("overflow", "scroll")))
	register(staticUtility("overflow-x-auto", decls("overflow-x", "auto")))
	register(staticUtility("overflow-x-hidden", decls("overflow-x", "hidden")))
	register(staticUtility("overflow-x-clip", decls("overflow-x", "clip")))
	register(staticUtility("overflow-x-visible", decls("overflow-x", "visible")))
	register(staticUtility("overflow-x-scroll", decls("overflow-x", "scroll")))
	register(staticUtility("overflow-y-auto", decls("overflow-y", "auto")))
	register(staticUtility("overflow-y-hidden", decls("overflow-y", "hidden")))
	register(staticUtility("overflow-y-clip", decls("overflow-y", "clip")))
	register(staticUtility("overflow-y-visible", decls("overflow-y", "visible")))
	register(staticUtility("overflow-y-scroll", decls("overflow-y", "scroll")))
	register(staticUtility("overscroll-auto", decls("overscroll-behavior", "auto")))
	register(staticUtility("overscroll-contain", decls("overscroll-behavior", "contain")))
	register(staticUtility("overscroll-none", decls("overscroll-behavior", "none")))
	register(staticUtility("overscroll-x-auto", decls("overscroll-behavior-x", "auto")))
	register(staticUtility("overscroll-x-contain", decls("overscroll-behavior-x", "contain")))
	register(staticUtility("overscroll-x-none", decls("overscroll-behavior-x", "none")))
	register(staticUtility("overscroll-y-auto", decls("overscroll-behavior-y", "auto")))
	register(staticUtility("overscroll-y-contain", decls("overscroll-behavior-y", "contain")))
	register(staticUtility("overscroll-y-none", decls("overscroll-behavior-y", "none")))

	// ===== Float =====
	register(staticUtility("float-right", decls("float", "right")))
	register(staticUtility("float-left", decls("float", "left")))
	register(staticUtility("float-none", decls("float", "none")))
	register(staticUtility("float-start", decls("float", "inline-start")))
	register(staticUtility("float-end", decls("float", "inline-end")))

	// ===== Clear =====
	register(staticUtility("clear-right", decls("clear", "right")))
	register(staticUtility("clear-left", decls("clear", "left")))
	register(staticUtility("clear-both", decls("clear", "both")))
	register(staticUtility("clear-none", decls("clear", "none")))
	register(staticUtility("clear-start", decls("clear", "inline-start")))
	register(staticUtility("clear-end", decls("clear", "inline-end")))

	// ===== Box Sizing =====
	register(staticUtility("box-border", decls("box-sizing", "border-box")))
	register(staticUtility("box-content", decls("box-sizing", "content-box")))

	// ===== Box Decoration Break =====
	register(staticUtility("box-decoration-clone", decls("box-decoration-break", "clone")))
	register(staticUtility("box-decoration-slice", decls("box-decoration-break", "slice")))

	// ===== Aspect Ratio =====
	register(staticUtility("aspect-auto", decls("aspect-ratio", "auto")))
	register(staticUtility("aspect-square", decls("aspect-ratio", "1 / 1")))
	register(staticUtility("aspect-video", decls("aspect-ratio", "16 / 9")))

	// ===== Columns =====
	register(staticUtility("columns-auto", decls("columns", "auto")))

	// ===== Break Before / After / Inside =====
	register(staticUtility("break-before-auto", decls("break-before", "auto")))
	register(staticUtility("break-before-avoid", decls("break-before", "avoid")))
	register(staticUtility("break-before-all", decls("break-before", "all")))
	register(staticUtility("break-before-avoid-page", decls("break-before", "avoid-page")))
	register(staticUtility("break-before-page", decls("break-before", "page")))
	register(staticUtility("break-before-left", decls("break-before", "left")))
	register(staticUtility("break-before-right", decls("break-before", "right")))
	register(staticUtility("break-before-column", decls("break-before", "column")))
	register(staticUtility("break-after-auto", decls("break-after", "auto")))
	register(staticUtility("break-after-avoid", decls("break-after", "avoid")))
	register(staticUtility("break-after-all", decls("break-after", "all")))
	register(staticUtility("break-after-avoid-page", decls("break-after", "avoid-page")))
	register(staticUtility("break-after-page", decls("break-after", "page")))
	register(staticUtility("break-after-left", decls("break-after", "left")))
	register(staticUtility("break-after-right", decls("break-after", "right")))
	register(staticUtility("break-after-column", decls("break-after", "column")))
	register(staticUtility("break-inside-auto", decls("break-inside", "auto")))
	register(staticUtility("break-inside-avoid", decls("break-inside", "avoid")))
	register(staticUtility("break-inside-avoid-page", decls("break-inside", "avoid-page")))
	register(staticUtility("break-inside-avoid-column", decls("break-inside", "avoid-column")))

	// ===== Z-index =====
	register(staticUtility("z-auto", decls("z-index", "auto")))

	// ===== Order =====
	register(staticUtility("order-first", decls("order", "calc(-infinity)")))
	register(staticUtility("order-last", decls("order", "calc(infinity)")))
	register(staticUtility("order-none", decls("order", "0")))

	// ===== Flex Basis =====
	register(staticUtility("basis-auto", decls("flex-basis", "auto")))
	register(staticUtility("basis-full", decls("flex-basis", "100%")))

	// ===== Flexbox =====
	register(staticUtility("flex-row", decls("flex-direction", "row")))
	register(staticUtility("flex-row-reverse", decls("flex-direction", "row-reverse")))
	register(staticUtility("flex-col", decls("flex-direction", "column")))
	register(staticUtility("flex-col-reverse", decls("flex-direction", "column-reverse")))
	register(staticUtility("flex-wrap", decls("flex-wrap", "wrap")))
	register(staticUtility("flex-wrap-reverse", decls("flex-wrap", "wrap-reverse")))
	register(staticUtility("flex-nowrap", decls("flex-wrap", "nowrap")))
	register(staticUtility("flex-1", decls("flex", "1")))
	register(staticUtility("flex-auto", decls("flex", "1 1 auto")))
	register(staticUtility("flex-initial", decls("flex", "0 1 auto")))
	register(staticUtility("flex-none", decls("flex", "none")))
	register(staticUtility("shrink", decls("flex-shrink", "1")))
	register(staticUtility("shrink-0", decls("flex-shrink", "0")))
	register(staticUtility("grow", decls("flex-grow", "1")))
	register(staticUtility("grow-0", decls("flex-grow", "0")))

	// ===== Grid =====
	register(staticUtility("grid-cols-none", decls("grid-template-columns", "none")))
	register(staticUtility("grid-cols-subgrid", decls("grid-template-columns", "subgrid")))
	register(staticUtility("grid-rows-none", decls("grid-template-rows", "none")))
	register(staticUtility("grid-rows-subgrid", decls("grid-template-rows", "subgrid")))
	register(staticUtility("col-span-full", decls("grid-column", "1 / -1")))
	register(staticUtility("row-span-full", decls("grid-row", "1 / -1")))
	register(staticUtility("col-auto", decls("grid-column", "auto")))
	register(staticUtility("row-auto", decls("grid-row", "auto")))
	register(staticUtility("auto-cols-auto", decls("grid-auto-columns", "auto")))
	register(staticUtility("auto-cols-min", decls("grid-auto-columns", "min-content")))
	register(staticUtility("auto-cols-max", decls("grid-auto-columns", "max-content")))
	register(staticUtility("auto-cols-fr", decls("grid-auto-columns", "minmax(0, 1fr)")))
	register(staticUtility("auto-rows-auto", decls("grid-auto-rows", "auto")))
	register(staticUtility("auto-rows-min", decls("grid-auto-rows", "min-content")))
	register(staticUtility("auto-rows-max", decls("grid-auto-rows", "max-content")))
	register(staticUtility("auto-rows-fr", decls("grid-auto-rows", "minmax(0, 1fr)")))
	register(staticUtility("grid-flow-row", decls("grid-auto-flow", "row")))
	register(staticUtility("grid-flow-col", decls("grid-auto-flow", "column")))
	register(staticUtility("grid-flow-dense", decls("grid-auto-flow", "dense")))
	register(staticUtility("grid-flow-row-dense", decls("grid-auto-flow", "row dense")))
	register(staticUtility("grid-flow-col-dense", decls("grid-auto-flow", "column dense")))

	// ===== Justify Content =====
	register(staticUtility("justify-normal", decls("justify-content", "normal")))
	register(staticUtility("justify-start", decls("justify-content", "flex-start")))
	register(staticUtility("justify-end", decls("justify-content", "flex-end")))
	register(staticUtility("justify-center", decls("justify-content", "center")))
	register(staticUtility("justify-between", decls("justify-content", "space-between")))
	register(staticUtility("justify-around", decls("justify-content", "space-around")))
	register(staticUtility("justify-evenly", decls("justify-content", "space-evenly")))
	register(staticUtility("justify-stretch", decls("justify-content", "stretch")))

	// ===== Justify Items =====
	register(staticUtility("justify-items-start", decls("justify-items", "start")))
	register(staticUtility("justify-items-end", decls("justify-items", "end")))
	register(staticUtility("justify-items-center", decls("justify-items", "center")))
	register(staticUtility("justify-items-stretch", decls("justify-items", "stretch")))
	register(staticUtility("justify-items-normal", decls("justify-items", "normal")))

	// ===== Justify Self =====
	register(staticUtility("justify-self-auto", decls("justify-self", "auto")))
	register(staticUtility("justify-self-start", decls("justify-self", "start")))
	register(staticUtility("justify-self-end", decls("justify-self", "end")))
	register(staticUtility("justify-self-center", decls("justify-self", "center")))
	register(staticUtility("justify-self-stretch", decls("justify-self", "stretch")))

	// ===== Align Content =====
	register(staticUtility("content-normal", decls("align-content", "normal")))
	register(staticUtility("content-center", decls("align-content", "center")))
	register(staticUtility("content-start", decls("align-content", "flex-start")))
	register(staticUtility("content-end", decls("align-content", "flex-end")))
	register(staticUtility("content-between", decls("align-content", "space-between")))
	register(staticUtility("content-around", decls("align-content", "space-around")))
	register(staticUtility("content-evenly", decls("align-content", "space-evenly")))
	register(staticUtility("content-baseline", decls("align-content", "baseline")))
	register(staticUtility("content-stretch", decls("align-content", "stretch")))

	// ===== Align Items =====
	register(staticUtility("items-start", decls("align-items", "flex-start")))
	register(staticUtility("items-end", decls("align-items", "flex-end")))
	register(staticUtility("items-center", decls("align-items", "center")))
	register(staticUtility("items-baseline", decls("align-items", "baseline")))
	register(staticUtility("items-stretch", decls("align-items", "stretch")))

	// ===== Align Self =====
	register(staticUtility("self-auto", decls("align-self", "auto")))
	register(staticUtility("self-start", decls("align-self", "flex-start")))
	register(staticUtility("self-end", decls("align-self", "flex-end")))
	register(staticUtility("self-center", decls("align-self", "center")))
	register(staticUtility("self-stretch", decls("align-self", "stretch")))
	register(staticUtility("self-baseline", decls("align-self", "baseline")))

	// ===== Place Content =====
	register(staticUtility("place-content-center", decls("place-content", "center")))
	register(staticUtility("place-content-start", decls("place-content", "start")))
	register(staticUtility("place-content-end", decls("place-content", "end")))
	register(staticUtility("place-content-between", decls("place-content", "space-between")))
	register(staticUtility("place-content-around", decls("place-content", "space-around")))
	register(staticUtility("place-content-evenly", decls("place-content", "space-evenly")))
	register(staticUtility("place-content-baseline", decls("place-content", "baseline")))
	register(staticUtility("place-content-stretch", decls("place-content", "stretch")))

	// ===== Place Items =====
	register(staticUtility("place-items-start", decls("place-items", "start")))
	register(staticUtility("place-items-end", decls("place-items", "end")))
	register(staticUtility("place-items-center", decls("place-items", "center")))
	register(staticUtility("place-items-baseline", decls("place-items", "baseline")))
	register(staticUtility("place-items-stretch", decls("place-items", "stretch")))

	// ===== Place Self =====
	register(staticUtility("place-self-auto", decls("place-self", "auto")))
	register(staticUtility("place-self-start", decls("place-self", "start")))
	register(staticUtility("place-self-end", decls("place-self", "end")))
	register(staticUtility("place-self-center", decls("place-self", "center")))
	register(staticUtility("place-self-stretch", decls("place-self", "stretch")))

	// ===== Margin (auto) =====
	register(staticUtility("m-auto", decls("margin", "auto")))
	register(staticUtility("mx-auto", decls("margin-left", "auto", "margin-right", "auto")))
	register(staticUtility("my-auto", decls("margin-top", "auto", "margin-bottom", "auto")))
	register(staticUtility("ms-auto", decls("margin-inline-start", "auto")))
	register(staticUtility("me-auto", decls("margin-inline-end", "auto")))
	register(staticUtility("mbs-auto", decls("margin-block-start", "auto")))
	register(staticUtility("mbe-auto", decls("margin-block-end", "auto")))
	register(staticUtility("mt-auto", decls("margin-top", "auto")))
	register(staticUtility("mr-auto", decls("margin-right", "auto")))
	register(staticUtility("mb-auto", decls("margin-bottom", "auto")))
	register(staticUtility("ml-auto", decls("margin-left", "auto")))

	// ===== Space Between (child selector) =====
	childSel := "> :not(:last-child)"
	register(staticUtilityWithSelector("space-x-reverse", childSel, decls("--tw-space-x-reverse", "1")))
	register(staticUtilityWithSelector("space-y-reverse", childSel, decls("--tw-space-y-reverse", "1")))

	// ===== Width =====
	register(staticUtility("w-auto", decls("width", "auto")))
	register(staticUtility("w-screen", decls("width", "100vw")))
	register(staticUtility("w-svw", decls("width", "100svw")))
	register(staticUtility("w-lvw", decls("width", "100lvw")))
	register(staticUtility("w-dvw", decls("width", "100dvw")))
	register(staticUtility("w-min", decls("width", "min-content")))
	register(staticUtility("w-max", decls("width", "max-content")))
	register(staticUtility("w-fit", decls("width", "fit-content")))
	register(staticUtility("w-full", decls("width", "100%")))

	// ===== Min Width =====
	register(staticUtility("min-w-full", decls("min-width", "100%")))
	register(staticUtility("min-w-min", decls("min-width", "min-content")))
	register(staticUtility("min-w-max", decls("min-width", "max-content")))
	register(staticUtility("min-w-fit", decls("min-width", "fit-content")))

	// ===== Max Width =====
	register(staticUtility("max-w-none", decls("max-width", "none")))
	register(staticUtility("max-w-full", decls("max-width", "100%")))
	register(staticUtility("max-w-min", decls("max-width", "min-content")))
	register(staticUtility("max-w-max", decls("max-width", "max-content")))
	register(staticUtility("max-w-fit", decls("max-width", "fit-content")))
	register(staticUtility("max-w-prose", decls("max-width", "65ch")))
	register(staticUtility("max-w-screen", decls("max-width", "100vw")))

	// ===== Height =====
	register(staticUtility("h-auto", decls("height", "auto")))
	register(staticUtility("h-screen", decls("height", "100vh")))
	register(staticUtility("h-svh", decls("height", "100svh")))
	register(staticUtility("h-lvh", decls("height", "100lvh")))
	register(staticUtility("h-dvh", decls("height", "100dvh")))
	register(staticUtility("h-min", decls("height", "min-content")))
	register(staticUtility("h-max", decls("height", "max-content")))
	register(staticUtility("h-fit", decls("height", "fit-content")))
	register(staticUtility("h-full", decls("height", "100%")))

	// ===== Min Height =====
	register(staticUtility("min-h-full", decls("min-height", "100%")))
	register(staticUtility("min-h-screen", decls("min-height", "100vh")))
	register(staticUtility("min-h-svh", decls("min-height", "100svh")))
	register(staticUtility("min-h-lvh", decls("min-height", "100lvh")))
	register(staticUtility("min-h-dvh", decls("min-height", "100dvh")))
	register(staticUtility("min-h-min", decls("min-height", "min-content")))
	register(staticUtility("min-h-max", decls("min-height", "max-content")))
	register(staticUtility("min-h-fit", decls("min-height", "fit-content")))

	// ===== Max Height =====
	register(staticUtility("max-h-none", decls("max-height", "none")))
	register(staticUtility("max-h-full", decls("max-height", "100%")))
	register(staticUtility("max-h-screen", decls("max-height", "100vh")))
	register(staticUtility("max-h-svh", decls("max-height", "100svh")))
	register(staticUtility("max-h-lvh", decls("max-height", "100lvh")))
	register(staticUtility("max-h-dvh", decls("max-height", "100dvh")))
	register(staticUtility("max-h-min", decls("max-height", "min-content")))
	register(staticUtility("max-h-max", decls("max-height", "max-content")))
	register(staticUtility("max-h-fit", decls("max-height", "fit-content")))

	// ===== Size =====
	register(staticUtility("size-auto", decls("width", "auto", "height", "auto")))
	register(staticUtility("size-full", decls("width", "100%", "height", "100%")))
	register(staticUtility("size-min", decls("width", "min-content", "height", "min-content")))
	register(staticUtility("size-max", decls("width", "max-content", "height", "max-content")))
	register(staticUtility("size-fit", decls("width", "fit-content", "height", "fit-content")))

	// ===== Font Size =====
	register(staticUtility("text-xs", decls("font-size", "var(--text-xs)", "line-height", "var(--tw-leading, var(--text-xs--line-height))")))
	register(staticUtility("text-sm", decls("font-size", "var(--text-sm)", "line-height", "var(--tw-leading, var(--text-sm--line-height))")))
	register(staticUtility("text-base", decls("font-size", "var(--text-base)", "line-height", "var(--tw-leading, var(--text-base--line-height))")))
	register(staticUtility("text-lg", decls("font-size", "var(--text-lg)", "line-height", "var(--tw-leading, var(--text-lg--line-height))")))
	register(staticUtility("text-xl", decls("font-size", "var(--text-xl)", "line-height", "var(--tw-leading, var(--text-xl--line-height))")))
	register(staticUtility("text-2xl", decls("font-size", "var(--text-2xl)", "line-height", "var(--tw-leading, var(--text-2xl--line-height))")))
	register(staticUtility("text-3xl", decls("font-size", "var(--text-3xl)", "line-height", "var(--tw-leading, var(--text-3xl--line-height))")))
	register(staticUtility("text-4xl", decls("font-size", "var(--text-4xl)", "line-height", "var(--tw-leading, var(--text-4xl--line-height))")))
	register(staticUtility("text-5xl", decls("font-size", "var(--text-5xl)", "line-height", "var(--tw-leading, var(--text-5xl--line-height))")))
	register(staticUtility("text-6xl", decls("font-size", "var(--text-6xl)", "line-height", "var(--tw-leading, var(--text-6xl--line-height))")))
	register(staticUtility("text-7xl", decls("font-size", "var(--text-7xl)", "line-height", "var(--tw-leading, var(--text-7xl--line-height))")))
	register(staticUtility("text-8xl", decls("font-size", "var(--text-8xl)", "line-height", "var(--tw-leading, var(--text-8xl--line-height))")))
	register(staticUtility("text-9xl", decls("font-size", "var(--text-9xl)", "line-height", "var(--tw-leading, var(--text-9xl--line-height))")))

	// ===== Font Weight =====
	register(staticUtility("font-thin", decls("--tw-font-weight", "var(--font-weight-thin)", "font-weight", "var(--font-weight-thin)")))
	register(staticUtility("font-extralight", decls("--tw-font-weight", "var(--font-weight-extralight)", "font-weight", "var(--font-weight-extralight)")))
	register(staticUtility("font-light", decls("--tw-font-weight", "var(--font-weight-light)", "font-weight", "var(--font-weight-light)")))
	register(staticUtility("font-normal", decls("--tw-font-weight", "var(--font-weight-normal)", "font-weight", "var(--font-weight-normal)")))
	register(staticUtility("font-medium", decls("--tw-font-weight", "var(--font-weight-medium)", "font-weight", "var(--font-weight-medium)")))
	register(staticUtility("font-semibold", decls("--tw-font-weight", "var(--font-weight-semibold)", "font-weight", "var(--font-weight-semibold)")))
	register(staticUtility("font-bold", decls("--tw-font-weight", "var(--font-weight-bold)", "font-weight", "var(--font-weight-bold)")))
	register(staticUtility("font-extrabold", decls("--tw-font-weight", "var(--font-weight-extrabold)", "font-weight", "var(--font-weight-extrabold)")))
	register(staticUtility("font-black", decls("--tw-font-weight", "var(--font-weight-black)", "font-weight", "var(--font-weight-black)")))

	// ===== Font Style =====
	register(staticUtility("italic", decls("font-style", "italic")))
	register(staticUtility("not-italic", decls("font-style", "normal")))

	// ===== Font Variant Numeric =====
	register(staticUtility("normal-nums", decls("font-variant-numeric", "normal")))
	register(staticUtility("ordinal", decls("font-variant-numeric", "ordinal")))
	register(staticUtility("slashed-zero", decls("font-variant-numeric", "slashed-zero")))
	register(staticUtility("lining-nums", decls("font-variant-numeric", "lining-nums")))
	register(staticUtility("oldstyle-nums", decls("font-variant-numeric", "oldstyle-nums")))
	register(staticUtility("proportional-nums", decls("font-variant-numeric", "proportional-nums")))
	register(staticUtility("tabular-nums", decls("font-variant-numeric", "tabular-nums")))
	register(staticUtility("diagonal-fractions", decls("font-variant-numeric", "diagonal-fractions")))
	register(staticUtility("stacked-fractions", decls("font-variant-numeric", "stacked-fractions")))

	// ===== Text Alignment =====
	register(staticUtility("text-left", decls("text-align", "left")))
	register(staticUtility("text-center", decls("text-align", "center")))
	register(staticUtility("text-right", decls("text-align", "right")))
	register(staticUtility("text-justify", decls("text-align", "justify")))
	register(staticUtility("text-start", decls("text-align", "start")))
	register(staticUtility("text-end", decls("text-align", "end")))

	// ===== Text Decoration =====
	register(staticUtility("underline", decls("text-decoration-line", "underline")))
	register(staticUtility("overline", decls("text-decoration-line", "overline")))
	register(staticUtility("line-through", decls("text-decoration-line", "line-through")))
	register(staticUtility("no-underline", decls("text-decoration-line", "none")))

	// ===== Text Decoration Color =====
	register(staticUtility("decoration-inherit", decls("text-decoration-color", "inherit")))

	register(staticUtility("decoration-transparent", decls("text-decoration-color", "transparent")))

	// ===== Text Decoration Style =====
	register(staticUtility("decoration-solid", decls("text-decoration-style", "solid")))
	register(staticUtility("decoration-double", decls("text-decoration-style", "double")))
	register(staticUtility("decoration-dotted", decls("text-decoration-style", "dotted")))
	register(staticUtility("decoration-dashed", decls("text-decoration-style", "dashed")))
	register(staticUtility("decoration-wavy", decls("text-decoration-style", "wavy")))

	// ===== Text Decoration Thickness =====
	register(staticUtility("decoration-auto", decls("text-decoration-thickness", "auto")))
	register(staticUtility("decoration-from-font", decls("text-decoration-thickness", "from-font")))
	register(staticUtility("decoration-0", decls("text-decoration-thickness", "0px")))
	register(staticUtility("decoration-1", decls("text-decoration-thickness", "1px")))
	register(staticUtility("decoration-2", decls("text-decoration-thickness", "2px")))
	register(staticUtility("decoration-4", decls("text-decoration-thickness", "4px")))
	register(staticUtility("decoration-8", decls("text-decoration-thickness", "8px")))

	// ===== Text Underline Offset =====
	register(staticUtility("underline-offset-auto", decls("text-underline-offset", "auto")))
	register(staticUtility("underline-offset-0", decls("text-underline-offset", "0px")))
	register(staticUtility("underline-offset-1", decls("text-underline-offset", "1px")))
	register(staticUtility("underline-offset-2", decls("text-underline-offset", "2px")))
	register(staticUtility("underline-offset-4", decls("text-underline-offset", "4px")))
	register(staticUtility("underline-offset-8", decls("text-underline-offset", "8px")))

	// ===== Text Transform =====
	register(staticUtility("uppercase", decls("text-transform", "uppercase")))
	register(staticUtility("lowercase", decls("text-transform", "lowercase")))
	register(staticUtility("capitalize", decls("text-transform", "capitalize")))
	register(staticUtility("normal-case", decls("text-transform", "none")))

	// ===== Text Overflow =====
	register(staticUtility("truncate", decls("overflow", "hidden", "text-overflow", "ellipsis", "white-space", "nowrap")))
	register(staticUtility("text-ellipsis", decls("text-overflow", "ellipsis")))
	register(staticUtility("text-clip", decls("text-overflow", "clip")))

	// ===== Text Wrap =====
	register(staticUtility("text-wrap", decls("text-wrap", "wrap")))
	register(staticUtility("text-nowrap", decls("text-wrap", "nowrap")))
	register(staticUtility("text-balance", decls("text-wrap", "balance")))
	register(staticUtility("text-pretty", decls("text-wrap", "pretty")))

	// ===== Line Clamp =====
	register(staticUtility("line-clamp-none", decls("-webkit-line-clamp", "unset", "overflow", "visible", "display", "block")))

	// ===== Line Height =====
	register(staticUtility("leading-none", decls("--tw-leading", "1", "line-height", "1")))
	register(staticUtility("leading-tight", decls("--tw-leading", "var(--leading-tight)", "line-height", "var(--leading-tight)")))
	register(staticUtility("leading-snug", decls("--tw-leading", "var(--leading-snug)", "line-height", "var(--leading-snug)")))
	register(staticUtility("leading-normal", decls("--tw-leading", "var(--leading-normal)", "line-height", "var(--leading-normal)")))
	register(staticUtility("leading-relaxed", decls("--tw-leading", "var(--leading-relaxed)", "line-height", "var(--leading-relaxed)")))
	register(staticUtility("leading-loose", decls("--tw-leading", "var(--leading-loose)", "line-height", "var(--leading-loose)")))

	// ===== Letter Spacing =====
	register(staticUtility("tracking-tighter", decls("--tw-tracking", "var(--tracking-tighter)", "letter-spacing", "var(--tracking-tighter)")))
	register(staticUtility("tracking-tight", decls("--tw-tracking", "var(--tracking-tight)", "letter-spacing", "var(--tracking-tight)")))
	register(staticUtility("tracking-normal", decls("--tw-tracking", "var(--tracking-normal)", "letter-spacing", "var(--tracking-normal)")))
	register(staticUtility("tracking-wide", decls("--tw-tracking", "var(--tracking-wide)", "letter-spacing", "var(--tracking-wide)")))
	register(staticUtility("tracking-wider", decls("--tw-tracking", "var(--tracking-wider)", "letter-spacing", "var(--tracking-wider)")))
	register(staticUtility("tracking-widest", decls("--tw-tracking", "var(--tracking-widest)", "letter-spacing", "var(--tracking-widest)")))

	// ===== Hyphens =====
	register(staticUtility("hyphens-none", decls("hyphens", "none")))
	register(staticUtility("hyphens-manual", decls("hyphens", "manual")))
	register(staticUtility("hyphens-auto", decls("hyphens", "auto")))

	// ===== Background Color =====
	register(staticUtility("bg-inherit", decls("background-color", "inherit")))

	register(staticUtility("bg-transparent", decls("background-color", "transparent")))

	// ===== Background Position =====
	register(staticUtility("bg-bottom", decls("background-position", "bottom")))
	register(staticUtility("bg-center", decls("background-position", "center")))
	register(staticUtility("bg-left", decls("background-position", "left")))
	register(staticUtility("bg-left-bottom", decls("background-position", "left bottom")))
	register(staticUtility("bg-left-top", decls("background-position", "left top")))
	register(staticUtility("bg-right", decls("background-position", "right")))
	register(staticUtility("bg-right-bottom", decls("background-position", "right bottom")))
	register(staticUtility("bg-right-top", decls("background-position", "right top")))
	register(staticUtility("bg-top", decls("background-position", "top")))

	// ===== Background Repeat =====
	register(staticUtility("bg-repeat", decls("background-repeat", "repeat")))
	register(staticUtility("bg-no-repeat", decls("background-repeat", "no-repeat")))
	register(staticUtility("bg-repeat-x", decls("background-repeat", "repeat-x")))
	register(staticUtility("bg-repeat-y", decls("background-repeat", "repeat-y")))
	register(staticUtility("bg-repeat-round", decls("background-repeat", "round")))
	register(staticUtility("bg-repeat-space", decls("background-repeat", "space")))

	// ===== Background Size =====
	register(staticUtility("bg-auto", decls("background-size", "auto")))
	register(staticUtility("bg-cover", decls("background-size", "cover")))
	register(staticUtility("bg-contain", decls("background-size", "contain")))

	// ===== Background Clip =====
	register(staticUtility("bg-clip-border", decls("background-clip", "border-box")))
	register(staticUtility("bg-clip-padding", decls("background-clip", "padding-box")))
	register(staticUtility("bg-clip-content", decls("background-clip", "content-box")))
	register(staticUtility("bg-clip-text", decls("background-clip", "text")))

	// ===== Background Origin =====
	register(staticUtility("bg-origin-border", decls("background-origin", "border-box")))
	register(staticUtility("bg-origin-padding", decls("background-origin", "padding-box")))
	register(staticUtility("bg-origin-content", decls("background-origin", "content-box")))

	// ===== Background Attachment =====
	register(staticUtility("bg-fixed", decls("background-attachment", "fixed")))
	register(staticUtility("bg-local", decls("background-attachment", "local")))
	register(staticUtility("bg-scroll", decls("background-attachment", "scroll")))

	// ===== Gradient Direction =====
	// Legacy bg-gradient-to-* (backward compat, no color interpolation)
	register(staticUtility("bg-gradient-to-t", decls("background-image", "linear-gradient(to top, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-tr", decls("background-image", "linear-gradient(to top right, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-r", decls("background-image", "linear-gradient(to right, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-br", decls("background-image", "linear-gradient(to bottom right, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-b", decls("background-image", "linear-gradient(to bottom, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-bl", decls("background-image", "linear-gradient(to bottom left, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-l", decls("background-image", "linear-gradient(to left, var(--tw-gradient-stops))")))
	register(staticUtility("bg-gradient-to-tl", decls("background-image", "linear-gradient(to top left, var(--tw-gradient-stops))")))

	// V4 bg-linear-to-* (with color interpolation modifier support)
	for _, gd := range []struct{ suffix, dir string }{
		{"t", "to top"}, {"tr", "to top right"}, {"r", "to right"}, {"br", "to bottom right"},
		{"b", "to bottom"}, {"bl", "to bottom left"}, {"l", "to left"}, {"tl", "to top left"},
	} {
		dir := gd.dir
		register(&UtilityRegistration{
			Name: "bg-linear-to-" + gd.suffix,
			Kind: "static",
			CompileFn: func(c ResolvedCandidate) []Declaration {
				interp := resolveGradientInterpolation(c.Modifier)
				return decls("background-image", "linear-gradient("+dir+interp+", var(--tw-gradient-stops))")
			},
		})
	}

	// bg-radial (static, default radial gradient with modifier support)
	register(&UtilityRegistration{
		Name: "bg-radial",
		Kind: "static",
		CompileFn: func(c ResolvedCandidate) []Declaration {
			interp := resolveGradientInterpolation(c.Modifier)
			if interp == "" {
				interp = " in oklab"
			}
			return decls("background-image", "radial-gradient("+interp[1:]+", var(--tw-gradient-stops))")
		},
	})

	register(staticUtility("bg-none", decls("background-image", "none")))

	// ===== Border Width =====
	register(staticUtility("border", decls("border-style", "var(--tw-border-style)", "border-width", "1px")))
	register(staticUtility("border-0", decls("border-style", "var(--tw-border-style)", "border-width", "0px")))
	register(staticUtility("border-2", decls("border-style", "var(--tw-border-style)", "border-width", "2px")))
	register(staticUtility("border-4", decls("border-style", "var(--tw-border-style)", "border-width", "4px")))
	register(staticUtility("border-8", decls("border-style", "var(--tw-border-style)", "border-width", "8px")))
	register(staticUtility("border-x", decls("border-left-width", "1px", "border-right-width", "1px")))
	register(staticUtility("border-x-0", decls("border-left-width", "0px", "border-right-width", "0px")))
	register(staticUtility("border-x-2", decls("border-left-width", "2px", "border-right-width", "2px")))
	register(staticUtility("border-x-4", decls("border-left-width", "4px", "border-right-width", "4px")))
	register(staticUtility("border-x-8", decls("border-left-width", "8px", "border-right-width", "8px")))
	register(staticUtility("border-y", decls("border-top-width", "1px", "border-bottom-width", "1px")))
	register(staticUtility("border-y-0", decls("border-top-width", "0px", "border-bottom-width", "0px")))
	register(staticUtility("border-y-2", decls("border-top-width", "2px", "border-bottom-width", "2px")))
	register(staticUtility("border-y-4", decls("border-top-width", "4px", "border-bottom-width", "4px")))
	register(staticUtility("border-y-8", decls("border-top-width", "8px", "border-bottom-width", "8px")))
	register(staticUtility("border-t", decls("border-top-width", "1px")))
	register(staticUtility("border-t-0", decls("border-top-width", "0px")))
	register(staticUtility("border-t-2", decls("border-top-width", "2px")))
	register(staticUtility("border-t-4", decls("border-top-width", "4px")))
	register(staticUtility("border-t-8", decls("border-top-width", "8px")))
	register(staticUtility("border-r", decls("border-right-width", "1px")))
	register(staticUtility("border-r-0", decls("border-right-width", "0px")))
	register(staticUtility("border-r-2", decls("border-right-width", "2px")))
	register(staticUtility("border-r-4", decls("border-right-width", "4px")))
	register(staticUtility("border-r-8", decls("border-right-width", "8px")))
	register(staticUtility("border-b", decls("border-bottom-width", "1px")))
	register(staticUtility("border-b-0", decls("border-bottom-width", "0px")))
	register(staticUtility("border-b-2", decls("border-bottom-width", "2px")))
	register(staticUtility("border-b-4", decls("border-bottom-width", "4px")))
	register(staticUtility("border-b-8", decls("border-bottom-width", "8px")))
	register(staticUtility("border-l", decls("border-left-width", "1px")))
	register(staticUtility("border-l-0", decls("border-left-width", "0px")))
	register(staticUtility("border-l-2", decls("border-left-width", "2px")))
	register(staticUtility("border-l-4", decls("border-left-width", "4px")))
	register(staticUtility("border-l-8", decls("border-left-width", "8px")))
	register(staticUtility("border-s", decls("border-inline-start-width", "1px")))
	register(staticUtility("border-s-0", decls("border-inline-start-width", "0px")))
	register(staticUtility("border-s-2", decls("border-inline-start-width", "2px")))
	register(staticUtility("border-s-4", decls("border-inline-start-width", "4px")))
	register(staticUtility("border-s-8", decls("border-inline-start-width", "8px")))
	register(staticUtility("border-e", decls("border-inline-end-width", "1px")))
	register(staticUtility("border-e-0", decls("border-inline-end-width", "0px")))
	register(staticUtility("border-e-2", decls("border-inline-end-width", "2px")))
	register(staticUtility("border-e-4", decls("border-inline-end-width", "4px")))
	register(staticUtility("border-e-8", decls("border-inline-end-width", "8px")))
	register(staticUtility("border-bs", decls("border-block-start-width", "1px")))
	register(staticUtility("border-bs-0", decls("border-block-start-width", "0px")))
	register(staticUtility("border-bs-2", decls("border-block-start-width", "2px")))
	register(staticUtility("border-bs-4", decls("border-block-start-width", "4px")))
	register(staticUtility("border-bs-8", decls("border-block-start-width", "8px")))
	register(staticUtility("border-be", decls("border-block-end-width", "1px")))
	register(staticUtility("border-be-0", decls("border-block-end-width", "0px")))
	register(staticUtility("border-be-2", decls("border-block-end-width", "2px")))
	register(staticUtility("border-be-4", decls("border-block-end-width", "4px")))
	register(staticUtility("border-be-8", decls("border-block-end-width", "8px")))

	// ===== Border Style =====
	register(staticUtility("border-solid", decls("border-style", "solid")))
	register(staticUtility("border-dashed", decls("border-style", "dashed")))
	register(staticUtility("border-dotted", decls("border-style", "dotted")))
	register(staticUtility("border-double", decls("border-style", "double")))
	register(staticUtility("border-hidden", decls("border-style", "hidden")))
	register(staticUtility("border-none", decls("border-style", "none")))

	// ===== Border Color =====
	register(staticUtility("border-inherit", decls("border-color", "inherit")))

	register(staticUtility("border-transparent", decls("border-color", "transparent")))

	// ===== Border Collapse =====
	register(staticUtility("border-collapse", decls("border-collapse", "collapse")))
	register(staticUtility("border-separate", decls("border-collapse", "separate")))

	// ===== Table Layout =====
	register(staticUtility("table-auto", decls("table-layout", "auto")))
	register(staticUtility("table-fixed", decls("table-layout", "fixed")))
	register(staticUtility("caption-top", decls("caption-side", "top")))
	register(staticUtility("caption-bottom", decls("caption-side", "bottom")))

	// ===== Border Radius =====
	register(staticUtility("rounded", decls("border-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-none", decls("border-radius", "0px")))
	register(staticUtility("rounded-sm", decls("border-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-md", decls("border-radius", "var(--radius-md)")))
	register(staticUtility("rounded-lg", decls("border-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-xl", decls("border-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-2xl", decls("border-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-3xl", decls("border-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-full", decls("border-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-t-none", decls("border-top-left-radius", "0px", "border-top-right-radius", "0px")))
	register(staticUtility("rounded-t-sm", decls("border-top-left-radius", "var(--radius-xs)", "border-top-right-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-t", decls("border-top-left-radius", "var(--radius-sm)", "border-top-right-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-t-md", decls("border-top-left-radius", "var(--radius-md)", "border-top-right-radius", "var(--radius-md)")))
	register(staticUtility("rounded-t-lg", decls("border-top-left-radius", "var(--radius-lg)", "border-top-right-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-t-xl", decls("border-top-left-radius", "var(--radius-xl)", "border-top-right-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-t-2xl", decls("border-top-left-radius", "var(--radius-2xl)", "border-top-right-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-t-3xl", decls("border-top-left-radius", "var(--radius-3xl)", "border-top-right-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-t-full", decls("border-top-left-radius", "calc(infinity * 1px)", "border-top-right-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-r-none", decls("border-top-right-radius", "0px", "border-bottom-right-radius", "0px")))
	register(staticUtility("rounded-r-sm", decls("border-top-right-radius", "var(--radius-xs)", "border-bottom-right-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-r", decls("border-top-right-radius", "var(--radius-sm)", "border-bottom-right-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-r-md", decls("border-top-right-radius", "var(--radius-md)", "border-bottom-right-radius", "var(--radius-md)")))
	register(staticUtility("rounded-r-lg", decls("border-top-right-radius", "var(--radius-lg)", "border-bottom-right-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-r-xl", decls("border-top-right-radius", "var(--radius-xl)", "border-bottom-right-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-r-2xl", decls("border-top-right-radius", "var(--radius-2xl)", "border-bottom-right-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-r-3xl", decls("border-top-right-radius", "var(--radius-3xl)", "border-bottom-right-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-r-full", decls("border-top-right-radius", "calc(infinity * 1px)", "border-bottom-right-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-b-none", decls("border-bottom-left-radius", "0px", "border-bottom-right-radius", "0px")))
	register(staticUtility("rounded-b-sm", decls("border-bottom-left-radius", "var(--radius-xs)", "border-bottom-right-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-b", decls("border-bottom-left-radius", "var(--radius-sm)", "border-bottom-right-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-b-md", decls("border-bottom-left-radius", "var(--radius-md)", "border-bottom-right-radius", "var(--radius-md)")))
	register(staticUtility("rounded-b-lg", decls("border-bottom-left-radius", "var(--radius-lg)", "border-bottom-right-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-b-xl", decls("border-bottom-left-radius", "var(--radius-xl)", "border-bottom-right-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-b-2xl", decls("border-bottom-left-radius", "var(--radius-2xl)", "border-bottom-right-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-b-3xl", decls("border-bottom-left-radius", "var(--radius-3xl)", "border-bottom-right-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-b-full", decls("border-bottom-left-radius", "calc(infinity * 1px)", "border-bottom-right-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-l-none", decls("border-top-left-radius", "0px", "border-bottom-left-radius", "0px")))
	register(staticUtility("rounded-l-sm", decls("border-top-left-radius", "var(--radius-xs)", "border-bottom-left-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-l", decls("border-top-left-radius", "var(--radius-sm)", "border-bottom-left-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-l-md", decls("border-top-left-radius", "var(--radius-md)", "border-bottom-left-radius", "var(--radius-md)")))
	register(staticUtility("rounded-l-lg", decls("border-top-left-radius", "var(--radius-lg)", "border-bottom-left-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-l-xl", decls("border-top-left-radius", "var(--radius-xl)", "border-bottom-left-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-l-2xl", decls("border-top-left-radius", "var(--radius-2xl)", "border-bottom-left-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-l-3xl", decls("border-top-left-radius", "var(--radius-3xl)", "border-bottom-left-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-l-full", decls("border-top-left-radius", "calc(infinity * 1px)", "border-bottom-left-radius", "calc(infinity * 1px)")))

	// Individual corner rounding (physical)
	register(staticUtility("rounded-tl-none", decls("border-top-left-radius", "0px")))
	register(staticUtility("rounded-tl-sm", decls("border-top-left-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-tl", decls("border-top-left-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-tl-md", decls("border-top-left-radius", "var(--radius-md)")))
	register(staticUtility("rounded-tl-lg", decls("border-top-left-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-tl-xl", decls("border-top-left-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-tl-2xl", decls("border-top-left-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-tl-3xl", decls("border-top-left-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-tl-full", decls("border-top-left-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-tr-none", decls("border-top-right-radius", "0px")))
	register(staticUtility("rounded-tr-sm", decls("border-top-right-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-tr", decls("border-top-right-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-tr-md", decls("border-top-right-radius", "var(--radius-md)")))
	register(staticUtility("rounded-tr-lg", decls("border-top-right-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-tr-xl", decls("border-top-right-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-tr-2xl", decls("border-top-right-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-tr-3xl", decls("border-top-right-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-tr-full", decls("border-top-right-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-br-none", decls("border-bottom-right-radius", "0px")))
	register(staticUtility("rounded-br-sm", decls("border-bottom-right-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-br", decls("border-bottom-right-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-br-md", decls("border-bottom-right-radius", "var(--radius-md)")))
	register(staticUtility("rounded-br-lg", decls("border-bottom-right-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-br-xl", decls("border-bottom-right-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-br-2xl", decls("border-bottom-right-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-br-3xl", decls("border-bottom-right-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-br-full", decls("border-bottom-right-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-bl-none", decls("border-bottom-left-radius", "0px")))
	register(staticUtility("rounded-bl-sm", decls("border-bottom-left-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-bl", decls("border-bottom-left-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-bl-md", decls("border-bottom-left-radius", "var(--radius-md)")))
	register(staticUtility("rounded-bl-lg", decls("border-bottom-left-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-bl-xl", decls("border-bottom-left-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-bl-2xl", decls("border-bottom-left-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-bl-3xl", decls("border-bottom-left-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-bl-full", decls("border-bottom-left-radius", "calc(infinity * 1px)")))

	// Logical side rounding
	register(staticUtility("rounded-s-none", decls("border-start-start-radius", "0px", "border-end-start-radius", "0px")))
	register(staticUtility("rounded-s-sm", decls("border-start-start-radius", "var(--radius-xs)", "border-end-start-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-s", decls("border-start-start-radius", "var(--radius-sm)", "border-end-start-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-s-md", decls("border-start-start-radius", "var(--radius-md)", "border-end-start-radius", "var(--radius-md)")))
	register(staticUtility("rounded-s-lg", decls("border-start-start-radius", "var(--radius-lg)", "border-end-start-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-s-xl", decls("border-start-start-radius", "var(--radius-xl)", "border-end-start-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-s-2xl", decls("border-start-start-radius", "var(--radius-2xl)", "border-end-start-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-s-3xl", decls("border-start-start-radius", "var(--radius-3xl)", "border-end-start-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-s-full", decls("border-start-start-radius", "calc(infinity * 1px)", "border-end-start-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-e-none", decls("border-start-end-radius", "0px", "border-end-end-radius", "0px")))
	register(staticUtility("rounded-e-sm", decls("border-start-end-radius", "var(--radius-xs)", "border-end-end-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-e", decls("border-start-end-radius", "var(--radius-sm)", "border-end-end-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-e-md", decls("border-start-end-radius", "var(--radius-md)", "border-end-end-radius", "var(--radius-md)")))
	register(staticUtility("rounded-e-lg", decls("border-start-end-radius", "var(--radius-lg)", "border-end-end-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-e-xl", decls("border-start-end-radius", "var(--radius-xl)", "border-end-end-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-e-2xl", decls("border-start-end-radius", "var(--radius-2xl)", "border-end-end-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-e-3xl", decls("border-start-end-radius", "var(--radius-3xl)", "border-end-end-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-e-full", decls("border-start-end-radius", "calc(infinity * 1px)", "border-end-end-radius", "calc(infinity * 1px)")))

	// Logical corner rounding
	register(staticUtility("rounded-ss-none", decls("border-start-start-radius", "0px")))
	register(staticUtility("rounded-ss-sm", decls("border-start-start-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-ss", decls("border-start-start-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-ss-md", decls("border-start-start-radius", "var(--radius-md)")))
	register(staticUtility("rounded-ss-lg", decls("border-start-start-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-ss-xl", decls("border-start-start-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-ss-2xl", decls("border-start-start-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-ss-3xl", decls("border-start-start-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-ss-full", decls("border-start-start-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-se-none", decls("border-start-end-radius", "0px")))
	register(staticUtility("rounded-se-sm", decls("border-start-end-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-se", decls("border-start-end-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-se-md", decls("border-start-end-radius", "var(--radius-md)")))
	register(staticUtility("rounded-se-lg", decls("border-start-end-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-se-xl", decls("border-start-end-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-se-2xl", decls("border-start-end-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-se-3xl", decls("border-start-end-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-se-full", decls("border-start-end-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-ee-none", decls("border-end-end-radius", "0px")))
	register(staticUtility("rounded-ee-sm", decls("border-end-end-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-ee", decls("border-end-end-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-ee-md", decls("border-end-end-radius", "var(--radius-md)")))
	register(staticUtility("rounded-ee-lg", decls("border-end-end-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-ee-xl", decls("border-end-end-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-ee-2xl", decls("border-end-end-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-ee-3xl", decls("border-end-end-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-ee-full", decls("border-end-end-radius", "calc(infinity * 1px)")))
	register(staticUtility("rounded-es-none", decls("border-end-start-radius", "0px")))
	register(staticUtility("rounded-es-sm", decls("border-end-start-radius", "var(--radius-xs)")))
	register(staticUtility("rounded-es", decls("border-end-start-radius", "var(--radius-sm)")))
	register(staticUtility("rounded-es-md", decls("border-end-start-radius", "var(--radius-md)")))
	register(staticUtility("rounded-es-lg", decls("border-end-start-radius", "var(--radius-lg)")))
	register(staticUtility("rounded-es-xl", decls("border-end-start-radius", "var(--radius-xl)")))
	register(staticUtility("rounded-es-2xl", decls("border-end-start-radius", "var(--radius-2xl)")))
	register(staticUtility("rounded-es-3xl", decls("border-end-start-radius", "var(--radius-3xl)")))
	register(staticUtility("rounded-es-full", decls("border-end-start-radius", "calc(infinity * 1px)")))

	// ===== Divide (child selector) =====
	register(staticUtilityWithSelector("divide-x", childSel, decls(
		"--tw-divide-x-reverse", "0",
		"border-inline-start-width", "calc(1px * calc(1 - var(--tw-divide-x-reverse)))",
		"border-inline-end-width", "calc(1px * var(--tw-divide-x-reverse))",
	)))
	register(staticUtilityWithSelector("divide-x-0", childSel, decls(
		"--tw-divide-x-reverse", "0",
		"border-inline-start-width", "calc(0px * calc(1 - var(--tw-divide-x-reverse)))",
		"border-inline-end-width", "calc(0px * var(--tw-divide-x-reverse))",
	)))
	register(staticUtilityWithSelector("divide-x-2", childSel, decls(
		"--tw-divide-x-reverse", "0",
		"border-inline-start-width", "calc(2px * calc(1 - var(--tw-divide-x-reverse)))",
		"border-inline-end-width", "calc(2px * var(--tw-divide-x-reverse))",
	)))
	register(staticUtilityWithSelector("divide-x-4", childSel, decls(
		"--tw-divide-x-reverse", "0",
		"border-inline-start-width", "calc(4px * calc(1 - var(--tw-divide-x-reverse)))",
		"border-inline-end-width", "calc(4px * var(--tw-divide-x-reverse))",
	)))
	register(staticUtilityWithSelector("divide-x-8", childSel, decls(
		"--tw-divide-x-reverse", "0",
		"border-inline-start-width", "calc(8px * calc(1 - var(--tw-divide-x-reverse)))",
		"border-inline-end-width", "calc(8px * var(--tw-divide-x-reverse))",
	)))
	register(staticUtilityWithSelector("divide-y", childSel, decls(
		"--tw-divide-y-reverse", "0",
		"border-top-width", "calc(1px * calc(1 - var(--tw-divide-y-reverse)))",
		"border-bottom-width", "calc(1px * var(--tw-divide-y-reverse))",
	)))
	register(staticUtilityWithSelector("divide-y-0", childSel, decls(
		"--tw-divide-y-reverse", "0",
		"border-top-width", "calc(0px * calc(1 - var(--tw-divide-y-reverse)))",
		"border-bottom-width", "calc(0px * var(--tw-divide-y-reverse))",
	)))
	register(staticUtilityWithSelector("divide-y-2", childSel, decls(
		"--tw-divide-y-reverse", "0",
		"border-top-width", "calc(2px * calc(1 - var(--tw-divide-y-reverse)))",
		"border-bottom-width", "calc(2px * var(--tw-divide-y-reverse))",
	)))
	register(staticUtilityWithSelector("divide-y-4", childSel, decls(
		"--tw-divide-y-reverse", "0",
		"border-top-width", "calc(4px * calc(1 - var(--tw-divide-y-reverse)))",
		"border-bottom-width", "calc(4px * var(--tw-divide-y-reverse))",
	)))
	register(staticUtilityWithSelector("divide-y-8", childSel, decls(
		"--tw-divide-y-reverse", "0",
		"border-top-width", "calc(8px * calc(1 - var(--tw-divide-y-reverse)))",
		"border-bottom-width", "calc(8px * var(--tw-divide-y-reverse))",
	)))
	register(staticUtilityWithSelector("divide-x-reverse", childSel, decls("--tw-divide-x-reverse", "1")))
	register(staticUtilityWithSelector("divide-y-reverse", childSel, decls("--tw-divide-y-reverse", "1")))
	register(staticUtilityWithSelector("divide-solid", childSel, decls("border-style", "solid")))
	register(staticUtilityWithSelector("divide-dashed", childSel, decls("border-style", "dashed")))
	register(staticUtilityWithSelector("divide-dotted", childSel, decls("border-style", "dotted")))
	register(staticUtilityWithSelector("divide-double", childSel, decls("border-style", "double")))
	register(staticUtilityWithSelector("divide-none", childSel, decls("border-style", "none")))

	// ===== Shadow =====
	register(staticUtility("shadow", decls(
		"--tw-shadow", "var(--shadow-sm, 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-none", decls(
		"--tw-shadow", "0 0 #0000",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-2xs", decls(
		"--tw-shadow", "var(--shadow-2xs, 0 1px rgb(0 0 0 / 0.05))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-xs", decls(
		"--tw-shadow", "var(--shadow-xs, 0 1px 2px 0 rgb(0 0 0 / 0.05))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-sm", decls(
		"--tw-shadow", "var(--shadow-sm, 0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-md", decls(
		"--tw-shadow", "var(--shadow-md, 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-lg", decls(
		"--tw-shadow", "var(--shadow-lg, 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-xl", decls(
		"--tw-shadow", "var(--shadow-xl, 0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-2xl", decls(
		"--tw-shadow", "var(--shadow-2xl, 0 25px 50px -12px rgb(0 0 0 / 0.25))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("shadow-inner", decls(
		"--tw-shadow", "inset 0 2px 4px 0 rgb(0 0 0 / 0.05)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))

	// ===== Inset Shadow =====
	register(staticUtility("inset-shadow-2xs", decls(
		"--tw-inset-shadow", "var(--inset-shadow-2xs, inset 0 1px rgb(0 0 0 / 0.05))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-shadow-xs", decls(
		"--tw-inset-shadow", "var(--inset-shadow-xs, inset 0 1px 1px rgb(0 0 0 / 0.05))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-shadow-sm", decls(
		"--tw-inset-shadow", "var(--inset-shadow-sm, inset 0 2px 4px rgb(0 0 0 / 0.05))",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))

	// ===== Ring =====
	register(staticUtility("ring", decls(
		"--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(3px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("ring-0", decls(
		"--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(0px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("ring-1", decls(
		"--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(1px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("ring-2", decls(
		"--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(2px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("ring-4", decls(
		"--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(4px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("ring-8", decls(
		"--tw-ring-shadow", "var(--tw-ring-inset,) 0 0 0 calc(8px + var(--tw-ring-offset-width)) var(--tw-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("ring-inset", decls("--tw-ring-inset", "inset")))

	// ===== Inset Ring =====
	register(staticUtility("inset-ring", decls(
		"--tw-inset-ring-shadow", "inset 0 0 0 1px var(--tw-inset-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-ring-0", decls(
		"--tw-inset-ring-shadow", "inset 0 0 0 0px var(--tw-inset-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-ring-1", decls(
		"--tw-inset-ring-shadow", "inset 0 0 0 1px var(--tw-inset-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-ring-2", decls(
		"--tw-inset-ring-shadow", "inset 0 0 0 2px var(--tw-inset-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-ring-4", decls(
		"--tw-inset-ring-shadow", "inset 0 0 0 4px var(--tw-inset-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))
	register(staticUtility("inset-ring-8", decls(
		"--tw-inset-ring-shadow", "inset 0 0 0 8px var(--tw-inset-ring-color, currentcolor)",
		"box-shadow", "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)",
	)))

	// ===== Ring Offset =====
	register(staticUtility("ring-offset-0", decls("--tw-ring-offset-width", "0px")))
	register(staticUtility("ring-offset-1", decls("--tw-ring-offset-width", "1px")))
	register(staticUtility("ring-offset-2", decls("--tw-ring-offset-width", "2px")))
	register(staticUtility("ring-offset-4", decls("--tw-ring-offset-width", "4px")))
	register(staticUtility("ring-offset-8", decls("--tw-ring-offset-width", "8px")))

	// ===== Outline =====
	register(staticUtility("outline-none", decls("outline", "2px solid transparent", "outline-offset", "2px")))
	register(staticUtility("outline", decls("outline-style", "solid")))
	register(staticUtility("outline-dashed", decls("outline-style", "dashed")))
	register(staticUtility("outline-dotted", decls("outline-style", "dotted")))
	register(staticUtility("outline-double", decls("outline-style", "double")))
	register(staticUtility("outline-0", decls("outline-width", "0px")))
	register(staticUtility("outline-1", decls("outline-width", "1px")))
	register(staticUtility("outline-2", decls("outline-width", "2px")))
	register(staticUtility("outline-4", decls("outline-width", "4px")))
	register(staticUtility("outline-8", decls("outline-width", "8px")))
	register(staticUtility("outline-offset-0", decls("outline-offset", "0px")))
	register(staticUtility("outline-offset-1", decls("outline-offset", "1px")))
	register(staticUtility("outline-offset-2", decls("outline-offset", "2px")))
	register(staticUtility("outline-offset-4", decls("outline-offset", "4px")))
	register(staticUtility("outline-offset-8", decls("outline-offset", "8px")))

	// ===== Cursor =====
	register(staticUtility("cursor-auto", decls("cursor", "auto")))
	register(staticUtility("cursor-default", decls("cursor", "default")))
	register(staticUtility("cursor-pointer", decls("cursor", "pointer")))
	register(staticUtility("cursor-wait", decls("cursor", "wait")))
	register(staticUtility("cursor-text", decls("cursor", "text")))
	register(staticUtility("cursor-move", decls("cursor", "move")))
	register(staticUtility("cursor-help", decls("cursor", "help")))
	register(staticUtility("cursor-not-allowed", decls("cursor", "not-allowed")))
	register(staticUtility("cursor-none", decls("cursor", "none")))
	register(staticUtility("cursor-context-menu", decls("cursor", "context-menu")))
	register(staticUtility("cursor-progress", decls("cursor", "progress")))
	register(staticUtility("cursor-cell", decls("cursor", "cell")))
	register(staticUtility("cursor-crosshair", decls("cursor", "crosshair")))
	register(staticUtility("cursor-vertical-text", decls("cursor", "vertical-text")))
	register(staticUtility("cursor-alias", decls("cursor", "alias")))
	register(staticUtility("cursor-copy", decls("cursor", "copy")))
	register(staticUtility("cursor-no-drop", decls("cursor", "no-drop")))
	register(staticUtility("cursor-grab", decls("cursor", "grab")))
	register(staticUtility("cursor-grabbing", decls("cursor", "grabbing")))
	register(staticUtility("cursor-all-scroll", decls("cursor", "all-scroll")))
	register(staticUtility("cursor-col-resize", decls("cursor", "col-resize")))
	register(staticUtility("cursor-row-resize", decls("cursor", "row-resize")))
	register(staticUtility("cursor-n-resize", decls("cursor", "n-resize")))
	register(staticUtility("cursor-e-resize", decls("cursor", "e-resize")))
	register(staticUtility("cursor-s-resize", decls("cursor", "s-resize")))
	register(staticUtility("cursor-w-resize", decls("cursor", "w-resize")))
	register(staticUtility("cursor-ne-resize", decls("cursor", "ne-resize")))
	register(staticUtility("cursor-nw-resize", decls("cursor", "nw-resize")))
	register(staticUtility("cursor-se-resize", decls("cursor", "se-resize")))
	register(staticUtility("cursor-sw-resize", decls("cursor", "sw-resize")))
	register(staticUtility("cursor-ew-resize", decls("cursor", "ew-resize")))
	register(staticUtility("cursor-ns-resize", decls("cursor", "ns-resize")))
	register(staticUtility("cursor-nesw-resize", decls("cursor", "nesw-resize")))
	register(staticUtility("cursor-nwse-resize", decls("cursor", "nwse-resize")))
	register(staticUtility("cursor-zoom-in", decls("cursor", "zoom-in")))
	register(staticUtility("cursor-zoom-out", decls("cursor", "zoom-out")))

	// ===== Caret Color =====
	register(staticUtility("caret-inherit", decls("caret-color", "inherit")))

	register(staticUtility("caret-transparent", decls("caret-color", "transparent")))

	// ===== Accent Color =====
	register(staticUtility("accent-auto", decls("accent-color", "auto")))
	register(staticUtility("accent-inherit", decls("accent-color", "inherit")))

	register(staticUtility("accent-transparent", decls("accent-color", "transparent")))

	// ===== Pointer Events =====
	register(staticUtility("pointer-events-none", decls("pointer-events", "none")))
	register(staticUtility("pointer-events-auto", decls("pointer-events", "auto")))

	// ===== Touch Action =====
	register(staticUtility("touch-auto", decls("touch-action", "auto")))
	register(staticUtility("touch-none", decls("touch-action", "none")))
	register(staticUtility("touch-pan-x", decls("touch-action", "pan-x")))
	register(staticUtility("touch-pan-left", decls("touch-action", "pan-left")))
	register(staticUtility("touch-pan-right", decls("touch-action", "pan-right")))
	register(staticUtility("touch-pan-y", decls("touch-action", "pan-y")))
	register(staticUtility("touch-pan-up", decls("touch-action", "pan-up")))
	register(staticUtility("touch-pan-down", decls("touch-action", "pan-down")))
	register(staticUtility("touch-pinch-zoom", decls("touch-action", "pinch-zoom")))
	register(staticUtility("touch-manipulation", decls("touch-action", "manipulation")))

	// ===== User Select =====
	register(staticUtility("select-none", decls("user-select", "none")))
	register(staticUtility("select-text", decls("user-select", "text")))
	register(staticUtility("select-all", decls("user-select", "all")))
	register(staticUtility("select-auto", decls("user-select", "auto")))

	// ===== Scroll Behavior =====
	register(staticUtility("scroll-auto", decls("scroll-behavior", "auto")))
	register(staticUtility("scroll-smooth", decls("scroll-behavior", "smooth")))

	// ===== Scroll Snap Align =====
	register(staticUtility("snap-start", decls("scroll-snap-align", "start")))
	register(staticUtility("snap-end", decls("scroll-snap-align", "end")))
	register(staticUtility("snap-center", decls("scroll-snap-align", "center")))
	register(staticUtility("snap-align-none", decls("scroll-snap-align", "none")))

	// ===== Scroll Snap Stop =====
	register(staticUtility("snap-normal", decls("scroll-snap-stop", "normal")))
	register(staticUtility("snap-always", decls("scroll-snap-stop", "always")))

	// ===== Scroll Snap Type =====
	register(staticUtility("snap-none", decls("scroll-snap-type", "none")))
	register(staticUtility("snap-x", decls("scroll-snap-type", "x var(--tw-scroll-snap-strictness, proximity)")))
	register(staticUtility("snap-y", decls("scroll-snap-type", "y var(--tw-scroll-snap-strictness, proximity)")))
	register(staticUtility("snap-both", decls("scroll-snap-type", "both var(--tw-scroll-snap-strictness, proximity)")))
	register(staticUtility("snap-mandatory", decls("--tw-scroll-snap-strictness", "mandatory")))
	register(staticUtility("snap-proximity", decls("--tw-scroll-snap-strictness", "proximity")))

	// ===== Resize =====
	register(staticUtility("resize-none", decls("resize", "none")))
	register(staticUtility("resize-y", decls("resize", "vertical")))
	register(staticUtility("resize-x", decls("resize", "horizontal")))
	register(staticUtility("resize", decls("resize", "both")))

	// ===== Object Fit =====
	register(staticUtility("object-contain", decls("object-fit", "contain")))
	register(staticUtility("object-cover", decls("object-fit", "cover")))
	register(staticUtility("object-fill", decls("object-fit", "fill")))
	register(staticUtility("object-none", decls("object-fit", "none")))
	register(staticUtility("object-scale-down", decls("object-fit", "scale-down")))

	// ===== Object Position =====
	register(staticUtility("object-bottom", decls("object-position", "bottom")))
	register(staticUtility("object-center", decls("object-position", "center")))
	register(staticUtility("object-left", decls("object-position", "left")))
	register(staticUtility("object-left-bottom", decls("object-position", "left bottom")))
	register(staticUtility("object-left-top", decls("object-position", "left top")))
	register(staticUtility("object-right", decls("object-position", "right")))
	register(staticUtility("object-right-bottom", decls("object-position", "right bottom")))
	register(staticUtility("object-right-top", decls("object-position", "right top")))
	register(staticUtility("object-top", decls("object-position", "top")))

	// ===== White Space =====
	register(staticUtility("whitespace-normal", decls("white-space", "normal")))
	register(staticUtility("whitespace-nowrap", decls("white-space", "nowrap")))
	register(staticUtility("whitespace-pre", decls("white-space", "pre")))
	register(staticUtility("whitespace-pre-line", decls("white-space", "pre-line")))
	register(staticUtility("whitespace-pre-wrap", decls("white-space", "pre-wrap")))
	register(staticUtility("whitespace-break-spaces", decls("white-space", "break-spaces")))

	// ===== Word Break =====
	register(staticUtility("break-normal", decls("overflow-wrap", "normal", "word-break", "normal")))
	register(staticUtility("break-words", decls("overflow-wrap", "break-word")))
	register(staticUtility("break-all", decls("word-break", "break-all")))
	register(staticUtility("break-keep", decls("word-break", "keep-all")))

	// ===== Vertical Align =====
	register(staticUtility("align-baseline", decls("vertical-align", "baseline")))
	register(staticUtility("align-top", decls("vertical-align", "top")))
	register(staticUtility("align-middle", decls("vertical-align", "middle")))
	register(staticUtility("align-bottom", decls("vertical-align", "bottom")))
	register(staticUtility("align-text-top", decls("vertical-align", "text-top")))
	register(staticUtility("align-text-bottom", decls("vertical-align", "text-bottom")))
	register(staticUtility("align-sub", decls("vertical-align", "sub")))
	register(staticUtility("align-super", decls("vertical-align", "super")))

	// ===== Content =====
	register(staticUtility("content-none", decls("content", "none")))

	// ===== List Style =====
	register(staticUtility("list-none", decls("list-style-type", "none")))
	register(staticUtility("list-disc", decls("list-style-type", "disc")))
	register(staticUtility("list-decimal", decls("list-style-type", "decimal")))
	register(staticUtility("list-inside", decls("list-style-position", "inside")))
	register(staticUtility("list-outside", decls("list-style-position", "outside")))
	register(staticUtility("list-image-none", decls("list-style-image", "none")))

	// ===== Appearance =====
	register(staticUtility("appearance-none", decls("appearance", "none")))
	register(staticUtility("appearance-auto", decls("appearance", "auto")))

	// ===== Fill =====
	register(staticUtility("fill-none", decls("fill", "none")))
	register(staticUtility("fill-inherit", decls("fill", "inherit")))

	register(staticUtility("fill-transparent", decls("fill", "transparent")))

	// ===== Stroke =====
	register(staticUtility("stroke-none", decls("stroke", "none")))
	register(staticUtility("stroke-inherit", decls("stroke", "inherit")))

	register(staticUtility("stroke-transparent", decls("stroke", "transparent")))
	register(staticUtility("stroke-0", decls("stroke-width", "0")))
	register(staticUtility("stroke-1", decls("stroke-width", "1")))
	register(staticUtility("stroke-2", decls("stroke-width", "2")))

	// ===== Filter =====
	filterChain := "var(--tw-blur,) var(--tw-brightness,) var(--tw-contrast,) var(--tw-grayscale,) var(--tw-hue-rotate,) var(--tw-invert,) var(--tw-saturate,) var(--tw-sepia,) var(--tw-drop-shadow,)"
	register(staticUtility("blur-none", decls("--tw-blur", "blur(0)", "filter", filterChain)))
	register(staticUtility("blur-sm", decls("--tw-blur", "blur(var(--blur-sm))", "filter", filterChain)))
	register(staticUtility("blur", decls("--tw-blur", "blur(var(--blur-md))", "filter", filterChain)))
	register(staticUtility("blur-md", decls("--tw-blur", "blur(var(--blur-md))", "filter", filterChain)))
	register(staticUtility("blur-lg", decls("--tw-blur", "blur(var(--blur-lg))", "filter", filterChain)))
	register(staticUtility("blur-xl", decls("--tw-blur", "blur(var(--blur-xl))", "filter", filterChain)))
	register(staticUtility("blur-2xl", decls("--tw-blur", "blur(var(--blur-2xl))", "filter", filterChain)))
	register(staticUtility("blur-3xl", decls("--tw-blur", "blur(var(--blur-3xl))", "filter", filterChain)))
	register(staticUtility("grayscale", decls("--tw-grayscale", "grayscale(100%)", "filter", filterChain)))
	register(staticUtility("grayscale-0", decls("--tw-grayscale", "grayscale(0)", "filter", filterChain)))
	register(staticUtility("invert", decls("--tw-invert", "invert(100%)", "filter", filterChain)))
	register(staticUtility("invert-0", decls("--tw-invert", "invert(0)", "filter", filterChain)))
	register(staticUtility("sepia", decls("--tw-sepia", "sepia(100%)", "filter", filterChain)))
	register(staticUtility("sepia-0", decls("--tw-sepia", "sepia(0)", "filter", filterChain)))
	register(staticUtility("drop-shadow-none", decls("--tw-drop-shadow", "drop-shadow(0 0 #0000)", "filter", filterChain)))
	register(staticUtility("drop-shadow-xs", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-xs, 0 1px 1px rgb(0 0 0 / 0.05)))", "filter", filterChain)))
	register(staticUtility("drop-shadow-sm", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-sm, 0 1px 2px rgb(0 0 0 / 0.15)))", "filter", filterChain)))
	register(staticUtility("drop-shadow", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-sm, 0 1px 2px rgb(0 0 0 / 0.15)))", "filter", filterChain)))
	register(staticUtility("drop-shadow-md", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-md, 0 3px 3px rgb(0 0 0 / 0.12)))", "filter", filterChain)))
	register(staticUtility("drop-shadow-lg", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-lg, 0 4px 4px rgb(0 0 0 / 0.15)))", "filter", filterChain)))
	register(staticUtility("drop-shadow-xl", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-xl, 0 9px 7px rgb(0 0 0 / 0.1)))", "filter", filterChain)))
	register(staticUtility("drop-shadow-2xl", decls("--tw-drop-shadow", "drop-shadow(var(--drop-shadow-2xl, 0 25px 25px rgb(0 0 0 / 0.15)))", "filter", filterChain)))

	// ===== Backdrop Filter =====
	backdropChain := "var(--tw-backdrop-blur,) var(--tw-backdrop-brightness,) var(--tw-backdrop-contrast,) var(--tw-backdrop-grayscale,) var(--tw-backdrop-hue-rotate,) var(--tw-backdrop-invert,) var(--tw-backdrop-saturate,) var(--tw-backdrop-sepia,) var(--tw-backdrop-opacity,)"
	register(staticUtility("backdrop-blur-none", decls("--tw-backdrop-blur", "blur(0)", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur-sm", decls("--tw-backdrop-blur", "blur(var(--blur-sm))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur", decls("--tw-backdrop-blur", "blur(var(--blur-md))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur-md", decls("--tw-backdrop-blur", "blur(var(--blur-md))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur-lg", decls("--tw-backdrop-blur", "blur(var(--blur-lg))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur-xl", decls("--tw-backdrop-blur", "blur(var(--blur-xl))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur-2xl", decls("--tw-backdrop-blur", "blur(var(--blur-2xl))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-blur-3xl", decls("--tw-backdrop-blur", "blur(var(--blur-3xl))", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-grayscale", decls("--tw-backdrop-grayscale", "grayscale(100%)", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-grayscale-0", decls("--tw-backdrop-grayscale", "grayscale(0)", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-invert", decls("--tw-backdrop-invert", "invert(100%)", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-invert-0", decls("--tw-backdrop-invert", "invert(0)", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-sepia", decls("--tw-backdrop-sepia", "sepia(100%)", "backdrop-filter", backdropChain)))
	register(staticUtility("backdrop-sepia-0", decls("--tw-backdrop-sepia", "sepia(0)", "backdrop-filter", backdropChain)))

	// ===== Mix Blend Mode =====
	register(staticUtility("mix-blend-normal", decls("mix-blend-mode", "normal")))
	register(staticUtility("mix-blend-multiply", decls("mix-blend-mode", "multiply")))
	register(staticUtility("mix-blend-screen", decls("mix-blend-mode", "screen")))
	register(staticUtility("mix-blend-overlay", decls("mix-blend-mode", "overlay")))
	register(staticUtility("mix-blend-darken", decls("mix-blend-mode", "darken")))
	register(staticUtility("mix-blend-lighten", decls("mix-blend-mode", "lighten")))
	register(staticUtility("mix-blend-color-dodge", decls("mix-blend-mode", "color-dodge")))
	register(staticUtility("mix-blend-color-burn", decls("mix-blend-mode", "color-burn")))
	register(staticUtility("mix-blend-hard-light", decls("mix-blend-mode", "hard-light")))
	register(staticUtility("mix-blend-soft-light", decls("mix-blend-mode", "soft-light")))
	register(staticUtility("mix-blend-difference", decls("mix-blend-mode", "difference")))
	register(staticUtility("mix-blend-exclusion", decls("mix-blend-mode", "exclusion")))
	register(staticUtility("mix-blend-hue", decls("mix-blend-mode", "hue")))
	register(staticUtility("mix-blend-saturation", decls("mix-blend-mode", "saturation")))
	register(staticUtility("mix-blend-color", decls("mix-blend-mode", "color")))
	register(staticUtility("mix-blend-luminosity", decls("mix-blend-mode", "luminosity")))
	register(staticUtility("mix-blend-plus-darker", decls("mix-blend-mode", "plus-darker")))
	register(staticUtility("mix-blend-plus-lighter", decls("mix-blend-mode", "plus-lighter")))

	// ===== Background Blend Mode =====
	register(staticUtility("bg-blend-normal", decls("background-blend-mode", "normal")))
	register(staticUtility("bg-blend-multiply", decls("background-blend-mode", "multiply")))
	register(staticUtility("bg-blend-screen", decls("background-blend-mode", "screen")))
	register(staticUtility("bg-blend-overlay", decls("background-blend-mode", "overlay")))
	register(staticUtility("bg-blend-darken", decls("background-blend-mode", "darken")))
	register(staticUtility("bg-blend-lighten", decls("background-blend-mode", "lighten")))
	register(staticUtility("bg-blend-color-dodge", decls("background-blend-mode", "color-dodge")))
	register(staticUtility("bg-blend-color-burn", decls("background-blend-mode", "color-burn")))
	register(staticUtility("bg-blend-hard-light", decls("background-blend-mode", "hard-light")))
	register(staticUtility("bg-blend-soft-light", decls("background-blend-mode", "soft-light")))
	register(staticUtility("bg-blend-difference", decls("background-blend-mode", "difference")))
	register(staticUtility("bg-blend-exclusion", decls("background-blend-mode", "exclusion")))
	register(staticUtility("bg-blend-hue", decls("background-blend-mode", "hue")))
	register(staticUtility("bg-blend-saturation", decls("background-blend-mode", "saturation")))
	register(staticUtility("bg-blend-color", decls("background-blend-mode", "color")))
	register(staticUtility("bg-blend-luminosity", decls("background-blend-mode", "luminosity")))

	// ===== Transition =====
	register(staticUtility("transition-none", decls("transition-property", "none")))
	register(staticUtility("transition-all", decls(
		"transition-property", "all",
		"transition-timing-function", "var(--tw-ease, var(--default-transition-timing-function))",
		"transition-duration", "var(--tw-duration, var(--default-transition-duration))",
	)))
	register(staticUtility("transition", decls(
		"transition-property", "color, background-color, border-color, outline-color, text-decoration-color, fill, stroke, --tw-gradient-from, --tw-gradient-via, --tw-gradient-to, opacity, box-shadow, transform, translate, scale, rotate, filter, -webkit-backdrop-filter, backdrop-filter, display, content-visibility, overlay, pointer-events",
		"transition-timing-function", "var(--tw-ease, var(--default-transition-timing-function))",
		"transition-duration", "var(--tw-duration, var(--default-transition-duration))",
	)))
	register(staticUtility("transition-colors", decls(
		"transition-property", "color, background-color, border-color, text-decoration-color, fill, stroke",
		"transition-timing-function", "var(--tw-ease, var(--default-transition-timing-function))",
		"transition-duration", "var(--tw-duration, var(--default-transition-duration))",
	)))
	register(staticUtility("transition-opacity", decls(
		"transition-property", "opacity",
		"transition-timing-function", "var(--tw-ease, var(--default-transition-timing-function))",
		"transition-duration", "var(--tw-duration, var(--default-transition-duration))",
	)))
	register(staticUtility("transition-shadow", decls(
		"transition-property", "box-shadow",
		"transition-timing-function", "var(--tw-ease, var(--default-transition-timing-function))",
		"transition-duration", "var(--tw-duration, var(--default-transition-duration))",
	)))
	register(staticUtility("transition-transform", decls(
		"transition-property", "transform, translate, scale, rotate",
		"transition-timing-function", "var(--tw-ease, var(--default-transition-timing-function))",
		"transition-duration", "var(--tw-duration, var(--default-transition-duration))",
	)))

	// ===== Duration =====
	register(staticUtility("duration-0", decls("--tw-duration", "0s", "transition-duration", "0s")))
	register(staticUtility("duration-75", decls("--tw-duration", "75ms", "transition-duration", "75ms")))
	register(staticUtility("duration-100", decls("--tw-duration", "100ms", "transition-duration", "100ms")))
	register(staticUtility("duration-150", decls("--tw-duration", "150ms", "transition-duration", "150ms")))
	register(staticUtility("duration-200", decls("--tw-duration", "200ms", "transition-duration", "200ms")))
	register(staticUtility("duration-300", decls("--tw-duration", "300ms", "transition-duration", "300ms")))
	register(staticUtility("duration-500", decls("--tw-duration", "500ms", "transition-duration", "500ms")))
	register(staticUtility("duration-700", decls("--tw-duration", "700ms", "transition-duration", "700ms")))
	register(staticUtility("duration-1000", decls("--tw-duration", "1000ms", "transition-duration", "1000ms")))

	// ===== Delay =====
	register(staticUtility("delay-0", decls("transition-delay", "0s")))
	register(staticUtility("delay-75", decls("transition-delay", "75ms")))
	register(staticUtility("delay-100", decls("transition-delay", "100ms")))
	register(staticUtility("delay-150", decls("transition-delay", "150ms")))
	register(staticUtility("delay-200", decls("transition-delay", "200ms")))
	register(staticUtility("delay-300", decls("transition-delay", "300ms")))
	register(staticUtility("delay-500", decls("transition-delay", "500ms")))
	register(staticUtility("delay-700", decls("transition-delay", "700ms")))
	register(staticUtility("delay-1000", decls("transition-delay", "1000ms")))

	// ===== Ease =====
	register(staticUtility("ease-linear", decls("--tw-ease", "linear", "transition-timing-function", "linear")))
	register(staticUtility("ease-in", decls("--tw-ease", "var(--ease-in)", "transition-timing-function", "var(--ease-in)")))
	register(staticUtility("ease-out", decls("--tw-ease", "var(--ease-out)", "transition-timing-function", "var(--ease-out)")))
	register(staticUtility("ease-in-out", decls("--tw-ease", "var(--ease-in-out)", "transition-timing-function", "var(--ease-in-out)")))

	// ===== Animation =====
	register(staticUtility("animate-none", decls("animation", "none")))
	register(staticUtility("animate-spin", decls("animation", "var(--animate-spin, spin 1s linear infinite)")))
	register(staticUtility("animate-ping", decls("animation", "var(--animate-ping, ping 1s cubic-bezier(0, 0, 0.2, 1) infinite)")))
	register(staticUtility("animate-pulse", decls("animation", "var(--animate-pulse, pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite)")))
	register(staticUtility("animate-bounce", decls("animation", "var(--animate-bounce, bounce 1s infinite)")))

	// ===== Transform =====
	register(staticUtility("scale-none", decls("scale", "none")))
	register(staticUtility("rotate-none", decls("rotate", "none")))
	register(staticUtility("translate-none", decls("translate", "none")))
	register(staticUtility("translate-3d", decls(
		"--tw-translate-z", "0px",
		"translate", "var(--tw-translate-x) var(--tw-translate-y) var(--tw-translate-z)",
	)))
	register(staticUtility("transform-none", decls("transform", "none")))

	// ===== Transform Origin =====
	register(staticUtility("origin-center", decls("transform-origin", "center")))
	register(staticUtility("origin-top", decls("transform-origin", "top")))
	register(staticUtility("origin-top-right", decls("transform-origin", "top right")))
	register(staticUtility("origin-right", decls("transform-origin", "right")))
	register(staticUtility("origin-bottom-right", decls("transform-origin", "bottom right")))
	register(staticUtility("origin-bottom", decls("transform-origin", "bottom")))
	register(staticUtility("origin-bottom-left", decls("transform-origin", "bottom left")))
	register(staticUtility("origin-left", decls("transform-origin", "left")))
	register(staticUtility("origin-top-left", decls("transform-origin", "top left")))

	// ===== Perspective =====
	register(staticUtility("perspective-none", decls("perspective", "none")))

	// ===== Will Change =====
	register(staticUtility("will-change-auto", decls("will-change", "auto")))
	register(staticUtility("will-change-scroll", decls("will-change", "scroll-position")))
	register(staticUtility("will-change-contents", decls("will-change", "contents")))
	register(staticUtility("will-change-transform", decls("will-change", "transform")))

	// ===== Text Shadow =====
	register(staticUtility("text-shadow-2xs", decls("text-shadow", "0px 1px 0px rgb(0 0 0 / 0.15)")))
	register(staticUtility("text-shadow-xs", decls("text-shadow", "0px 1px 1px rgb(0 0 0 / 0.2)")))
	register(staticUtility("text-shadow-sm", decls("text-shadow", "0px 1px 0px rgb(0 0 0 / 0.075), 0px 1px 1px rgb(0 0 0 / 0.075), 0px 2px 2px rgb(0 0 0 / 0.075)")))
	register(staticUtility("text-shadow-md", decls("text-shadow", "0px 1px 1px rgb(0 0 0 / 0.1), 0px 1px 2px rgb(0 0 0 / 0.1), 0px 2px 4px rgb(0 0 0 / 0.1)")))
	register(staticUtility("text-shadow-lg", decls("text-shadow", "0px 1px 2px rgb(0 0 0 / 0.1), 0px 3px 2px rgb(0 0 0 / 0.1), 0px 4px 8px rgb(0 0 0 / 0.1)")))
	register(staticUtility("text-shadow-none", decls("text-shadow", "none")))

	// ===== Font Stretch =====
	register(staticUtility("font-stretch-ultra-condensed", decls("font-stretch", "ultra-condensed")))
	register(staticUtility("font-stretch-extra-condensed", decls("font-stretch", "extra-condensed")))
	register(staticUtility("font-stretch-condensed", decls("font-stretch", "condensed")))
	register(staticUtility("font-stretch-semi-condensed", decls("font-stretch", "semi-condensed")))
	register(staticUtility("font-stretch-normal", decls("font-stretch", "normal")))
	register(staticUtility("font-stretch-semi-expanded", decls("font-stretch", "semi-expanded")))
	register(staticUtility("font-stretch-expanded", decls("font-stretch", "expanded")))
	register(staticUtility("font-stretch-extra-expanded", decls("font-stretch", "extra-expanded")))
	register(staticUtility("font-stretch-ultra-expanded", decls("font-stretch", "ultra-expanded")))

	// ===== Color Scheme =====
	register(staticUtility("scheme-normal", decls("color-scheme", "normal")))
	register(staticUtility("scheme-dark", decls("color-scheme", "dark")))
	register(staticUtility("scheme-light", decls("color-scheme", "light")))
	register(staticUtility("scheme-light-dark", decls("color-scheme", "light dark")))
	register(staticUtility("scheme-only-dark", decls("color-scheme", "only dark")))
	register(staticUtility("scheme-only-light", decls("color-scheme", "only light")))

	// ===== Field Sizing =====
	register(staticUtility("field-sizing-fixed", decls("field-sizing", "fixed")))
	register(staticUtility("field-sizing-content", decls("field-sizing", "content")))

	// ===== Backface Visibility =====
	register(staticUtility("backface-visible", decls("backface-visibility", "visible")))
	register(staticUtility("backface-hidden", decls("backface-visibility", "hidden")))

	// ===== Perspective Origin =====
	register(staticUtility("perspective-origin-center", decls("perspective-origin", "center")))
	register(staticUtility("perspective-origin-top", decls("perspective-origin", "top")))
	register(staticUtility("perspective-origin-top-right", decls("perspective-origin", "top right")))
	register(staticUtility("perspective-origin-right", decls("perspective-origin", "right")))
	register(staticUtility("perspective-origin-bottom-right", decls("perspective-origin", "bottom right")))
	register(staticUtility("perspective-origin-bottom", decls("perspective-origin", "bottom")))
	register(staticUtility("perspective-origin-bottom-left", decls("perspective-origin", "bottom left")))
	register(staticUtility("perspective-origin-left", decls("perspective-origin", "left")))
	register(staticUtility("perspective-origin-top-left", decls("perspective-origin", "top left")))

	// ===== Transform Style =====
	register(staticUtility("transform-3d", decls("transform-style", "preserve-3d")))
	register(staticUtility("transform-flat", decls("transform-style", "flat")))

	// ===== Font Smoothing =====
	register(staticUtility("antialiased", decls(
		"-webkit-font-smoothing", "antialiased",
		"-moz-osx-font-smoothing", "grayscale",
	)))
	register(staticUtility("subpixel-antialiased", decls(
		"-webkit-font-smoothing", "auto",
		"-moz-osx-font-smoothing", "auto",
	)))

	// ===== Wrap =====
	register(staticUtility("wrap-anywhere", decls("overflow-wrap", "anywhere")))
	register(staticUtility("wrap-break-word", decls("overflow-wrap", "break-word")))
	register(staticUtility("wrap-normal", decls("overflow-wrap", "normal")))

	// ===== Safe Alignment =====
	register(staticUtility("items-center-safe", decls("align-items", "safe center")))
	register(staticUtility("items-end-safe", decls("align-items", "safe end")))
	register(staticUtility("justify-center-safe", decls("justify-content", "safe center")))
	register(staticUtility("justify-end-safe", decls("justify-content", "safe end")))
	register(staticUtility("place-items-center-safe", decls("place-items", "safe center")))
	register(staticUtility("place-items-end-safe", decls("place-items", "safe end")))
	register(staticUtility("place-content-center-safe", decls("place-content", "safe center")))
	register(staticUtility("place-content-end-safe", decls("place-content", "safe end")))
	register(staticUtility("content-center-safe", decls("align-content", "safe center")))
	register(staticUtility("content-end-safe", decls("align-content", "safe end")))
	register(staticUtility("self-center-safe", decls("align-self", "safe center")))
	register(staticUtility("self-end-safe", decls("align-self", "safe end")))
	register(staticUtility("justify-self-center-safe", decls("justify-self", "safe center")))
	register(staticUtility("justify-self-end-safe", decls("justify-self", "safe end")))
	register(staticUtility("place-self-center-safe", decls("place-self", "safe center")))
	register(staticUtility("place-self-end-safe", decls("place-self", "safe end")))

	// ===== Baseline Last =====
	register(staticUtility("items-baseline-last", decls("align-items", "baseline last")))
	register(staticUtility("self-baseline-last", decls("align-self", "baseline last")))

	// ===== Transform Box =====
	register(staticUtility("transform-content", decls("transform-box", "content-box")))
	register(staticUtility("transform-border", decls("transform-box", "border-box")))
	register(staticUtility("transform-fill", decls("transform-box", "fill-box")))
	register(staticUtility("transform-stroke", decls("transform-box", "stroke-box")))
	register(staticUtility("transform-view", decls("transform-box", "view-box")))

	// ===== Contain =====
	register(staticUtility("contain-none", decls("contain", "none")))
	register(staticUtility("contain-content", decls("contain", "content")))
	register(staticUtility("contain-strict", decls("contain", "strict")))
	register(staticUtility("contain-size", decls("contain", "size")))
	register(staticUtility("contain-inline-size", decls("contain", "inline-size")))
	register(staticUtility("contain-layout", decls("contain", "layout")))
	register(staticUtility("contain-paint", decls("contain", "paint")))
	register(staticUtility("contain-style", decls("contain", "style")))

	// ===== @container (Container Query Context) =====
	// @container makes an element a container query context.
	// Named containers use /name modifier: @container/sidebar
	register(&UtilityRegistration{
		Name: "@container",
		Kind: "static",
		CompileFn: func(c ResolvedCandidate) []Declaration {
			if c.Modifier != "" {
				return decls(
					"container-type", "inline-size",
					"container-name", c.Modifier,
				)
			}
			return decls("container-type", "inline-size")
		},
	})
	register(staticUtility("@container-normal", decls("container-type", "normal")))
	register(staticUtility("@container-size", decls("container-type", "size")))
}
