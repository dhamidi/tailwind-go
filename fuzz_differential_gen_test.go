package tailwind

import (
	"math/rand"
	"strings"
	"testing"
)

// Building-block data for the differential fuzz class generator.

var staticUtilities = []string{
	"flex", "inline-flex", "block", "inline-block", "inline", "grid", "inline-grid",
	"table", "table-row", "table-cell", "hidden", "contents", "flow-root",
	"sr-only", "not-sr-only", "truncate", "italic", "not-italic",
	"underline", "overline", "line-through", "no-underline",
	"uppercase", "lowercase", "capitalize", "normal-case",
	"antialiased", "subpixel-antialiased",
	"break-words", "break-all", "break-normal", "break-keep",
	"relative", "absolute", "fixed", "sticky", "static",
	"isolate", "isolation-auto",
	"invisible", "visible", "collapse",
	"resize", "resize-x", "resize-y", "resize-none",
	"snap-start", "snap-end", "snap-center", "snap-align-none",
	"snap-normal", "snap-always",
	"grow", "grow-0", "shrink", "shrink-0",
	"border-collapse", "border-separate",
	"table-auto", "table-fixed",
	"overflow-hidden", "overflow-auto", "overflow-scroll", "overflow-visible", "overflow-clip",
	"overflow-x-hidden", "overflow-x-auto", "overflow-x-scroll", "overflow-x-visible", "overflow-x-clip",
	"overflow-y-hidden", "overflow-y-auto", "overflow-y-scroll", "overflow-y-visible", "overflow-y-clip",
	"overscroll-auto", "overscroll-contain", "overscroll-none",
	"object-contain", "object-cover", "object-fill", "object-none", "object-scale-down",
	"object-center", "object-top", "object-bottom", "object-left", "object-right",
	"whitespace-normal", "whitespace-nowrap", "whitespace-pre", "whitespace-pre-line", "whitespace-pre-wrap",
	"cursor-pointer", "cursor-default", "cursor-wait", "cursor-text", "cursor-move", "cursor-not-allowed",
	"cursor-auto", "cursor-cell", "cursor-crosshair", "cursor-grab", "cursor-grabbing",
	"cursor-help", "cursor-no-drop", "cursor-context-menu", "cursor-col-resize",
	"cursor-row-resize", "cursor-all-scroll", "cursor-zoom-in", "cursor-zoom-out",
	"cursor-copy", "cursor-alias", "cursor-progress", "cursor-none",
	"select-none", "select-text", "select-all", "select-auto",
	"pointer-events-none", "pointer-events-auto",
	"list-inside", "list-outside", "list-none", "list-disc", "list-decimal",
	"float-left", "float-right", "float-none", "float-start", "float-end",
	"clear-left", "clear-right", "clear-both", "clear-none",
	"box-border", "box-content",
	"appearance-none", "appearance-auto",
	"columns-1", "columns-2", "columns-3",
	"will-change-auto", "will-change-scroll", "will-change-contents", "will-change-transform",
	"transition", "transition-all", "transition-colors", "transition-opacity", "transition-shadow", "transition-transform", "transition-none",
	"ease-linear", "ease-in", "ease-out", "ease-in-out",
	// flexbox
	"flex-row", "flex-row-reverse", "flex-col", "flex-col-reverse",
	"flex-wrap", "flex-wrap-reverse", "flex-nowrap",
	"flex-1", "flex-auto", "flex-initial", "flex-none",
	// grid
	"grid-flow-row", "grid-flow-col", "grid-flow-dense", "grid-flow-row-dense", "grid-flow-col-dense",
	"col-auto", "row-auto",
	// alignment
	"justify-start", "justify-end", "justify-center", "justify-between", "justify-around", "justify-evenly", "justify-stretch", "justify-normal",
	"justify-items-start", "justify-items-end", "justify-items-center", "justify-items-stretch", "justify-items-normal",
	"justify-self-auto", "justify-self-start", "justify-self-end", "justify-self-center", "justify-self-stretch",
	"items-start", "items-end", "items-center", "items-baseline", "items-stretch",
	"self-auto", "self-start", "self-end", "self-center", "self-baseline", "self-stretch",
	"content-start", "content-end", "content-center", "content-between", "content-around", "content-evenly", "content-stretch", "content-baseline", "content-normal",
	"place-content-start", "place-content-end", "place-content-center", "place-content-between", "place-content-around", "place-content-evenly", "place-content-stretch", "place-content-baseline",
	"place-items-start", "place-items-end", "place-items-center", "place-items-baseline", "place-items-stretch",
	"place-self-auto", "place-self-start", "place-self-end", "place-self-center", "place-self-stretch",
	// aspect ratio
	"aspect-auto", "aspect-square", "aspect-video",
	// font weight
	"font-thin", "font-extralight", "font-light", "font-normal", "font-medium", "font-semibold", "font-bold", "font-extrabold", "font-black",
	// font family
	"font-sans", "font-serif", "font-mono",
	// leading (line-height) named values
	"leading-none", "leading-tight", "leading-snug", "leading-normal", "leading-relaxed", "leading-loose",
	// tracking (letter-spacing)
	"tracking-tighter", "tracking-tight", "tracking-normal", "tracking-wide", "tracking-wider", "tracking-widest",
	// text decoration style
	"decoration-solid", "decoration-double", "decoration-dotted", "decoration-dashed", "decoration-wavy",
	// text alignment
	"text-left", "text-center", "text-right", "text-justify", "text-start", "text-end",
	// text wrap
	"text-wrap", "text-nowrap", "text-balance", "text-pretty", "text-ellipsis", "text-clip",
	// font variant numeric
	"normal-nums", "ordinal", "slashed-zero", "lining-nums", "oldstyle-nums", "proportional-nums", "tabular-nums", "diagonal-fractions", "stacked-fractions",
	// vertical align
	"align-baseline", "align-top", "align-middle", "align-bottom", "align-text-top", "align-text-bottom", "align-sub", "align-super",
	// line clamp
	"line-clamp-1", "line-clamp-2", "line-clamp-3", "line-clamp-4", "line-clamp-5", "line-clamp-6", "line-clamp-none",
	// hyphens and overflow wrap
	"hyphens-none", "hyphens-manual", "hyphens-auto", "wrap-normal", "wrap-break-word", "wrap-anywhere",
	"grayscale", "grayscale-0", "invert", "invert-0", "sepia", "sepia-0",
	"backdrop-grayscale", "backdrop-grayscale-0", "backdrop-invert", "backdrop-invert-0", "backdrop-sepia", "backdrop-sepia-0",
	"mix-blend-normal", "mix-blend-multiply", "mix-blend-screen", "mix-blend-overlay",
	"mix-blend-darken", "mix-blend-lighten", "mix-blend-color-dodge", "mix-blend-color-burn",
	"mix-blend-hard-light", "mix-blend-soft-light", "mix-blend-difference", "mix-blend-exclusion",
	"mix-blend-hue", "mix-blend-saturation", "mix-blend-color", "mix-blend-luminosity",
	"mix-blend-plus-darker", "mix-blend-plus-lighter",
	// transform
	"transform-3d", "transform-flat", "transform-none",
	"translate-none", "rotate-none",
	"backface-visible", "backface-hidden",
	// text-shadow
	"text-shadow-sm", "text-shadow", "text-shadow-lg", "text-shadow-none",
	"text-shadow-initial", "text-shadow-inherit",
	"text-shadow-red-500", "text-shadow-blue-300/50",
	"bg-fixed", "bg-local", "bg-scroll",
	"bg-center", "bg-top", "bg-bottom", "bg-left", "bg-right",
	"bg-repeat", "bg-no-repeat", "bg-repeat-x", "bg-repeat-y",
	"bg-cover", "bg-contain", "bg-auto",
	// gradient directions
	"bg-linear-to-t", "bg-linear-to-tr", "bg-linear-to-r", "bg-linear-to-br",
	"bg-linear-to-b", "bg-linear-to-bl", "bg-linear-to-l", "bg-linear-to-tl",
	// gradient via reset
	"via-none",
	// background clip, origin, blend
	"bg-clip-border", "bg-clip-padding", "bg-clip-content", "bg-clip-text",
	"bg-origin-border", "bg-origin-padding", "bg-origin-content",
	"bg-blend-normal", "bg-blend-multiply", "bg-blend-screen", "bg-blend-overlay",
	"bg-blend-darken", "bg-blend-lighten", "bg-blend-color-dodge", "bg-blend-color-burn",
	"bg-blend-hard-light", "bg-blend-soft-light", "bg-blend-difference", "bg-blend-exclusion",
	"border-solid", "border-dashed", "border-dotted", "border-double", "border-none",
	"outline-none", "outline", "outline-dashed", "outline-dotted", "outline-double",
	"ring-inset",
	// outline width
	"outline-0", "outline-1", "outline-2", "outline-4", "outline-8",
	// ring width
	"ring-0", "ring-1", "ring-2", "ring-4", "ring-8", "ring",
	// inset ring width
	"inset-ring-0", "inset-ring-1", "inset-ring-2", "inset-ring-4", "inset-ring-8",
	// ring offset
	"ring-offset-0", "ring-offset-1", "ring-offset-2", "ring-offset-4", "ring-offset-8",
	// divide style
	"divide-solid", "divide-dashed", "divide-dotted", "divide-double", "divide-none",
	"accent-auto",
	// box decoration break
	"box-decoration-clone", "box-decoration-slice",
	// break before/after/inside
	"break-before-auto", "break-before-avoid", "break-before-all", "break-before-page",
	"break-before-left", "break-before-right", "break-before-column",
	"break-after-auto", "break-after-avoid", "break-after-all", "break-after-page",
	"break-after-left", "break-after-right", "break-after-column",
	"break-inside-auto", "break-inside-avoid", "break-inside-avoid-page", "break-inside-avoid-column",
	// color scheme
	"color-scheme-normal", "color-scheme-light", "color-scheme-dark", "color-scheme-light-dark",
	// field sizing
	"field-sizing-content", "field-sizing-fixed",
	// contain
	"contain-none", "contain-content", "contain-strict",
	"contain-size", "contain-inline-size", "contain-layout", "contain-paint", "contain-style",
	// caption side
	"caption-top", "caption-bottom",
	"touch-auto", "touch-none", "touch-manipulation",
	"touch-pan-x", "touch-pan-y", "touch-pan-up", "touch-pan-down",
	"touch-pan-left", "touch-pan-right", "touch-pinch-zoom",
	"scroll-auto", "scroll-smooth",
	"snap-none", "snap-x", "snap-y", "snap-both", "snap-mandatory", "snap-proximity",
	"forced-color-adjust-auto", "forced-color-adjust-none",
}

var fuzzSpacingPrefixes = []string{
	"p", "m", "px", "py", "pt", "pr", "pb", "pl",
	"mx", "my", "mt", "mr", "mb", "ml",
	// logical margin/padding
	"ms", "me", "mbs", "mbe",
	"ps", "pe", "pbs", "pbe",
	"gap", "gap-x", "gap-y",
	"space-x", "space-y",
	"scroll-m", "scroll-p",
	"scroll-mx", "scroll-my", "scroll-mt", "scroll-mr", "scroll-mb", "scroll-ml",
	"scroll-px", "scroll-py", "scroll-pt", "scroll-pr", "scroll-pb", "scroll-pl",
	// logical scroll margin/padding
	"scroll-ms", "scroll-me", "scroll-mbs", "scroll-mbe",
	"scroll-ps", "scroll-pe", "scroll-pbs", "scroll-pbe",
	"leading",
	"border-spacing", "border-spacing-x", "border-spacing-y",
}

var fuzzSpacingValues = []string{
	"0", "0.5", "1", "1.5", "2", "2.5", "3", "3.5", "4", "5", "6", "7", "8",
	"9", "10", "11", "12", "14", "16", "20", "24", "28", "32", "36", "40",
	"44", "48", "52", "56", "60", "64", "72", "80", "96", "px", "auto",
}

var fuzzSizingPrefixes = []string{"w", "h", "min-w", "max-w", "min-h", "max-h", "size", "size-inline", "min-inline", "max-inline"}

var fuzzSizingValues = []string{
	"0", "0.5", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20",
	"24", "32", "40", "48", "56", "64", "72", "80", "96",
	"auto", "full", "screen", "min", "max", "fit", "px",
	"1/2", "1/3", "2/3", "1/4", "2/4", "3/4",
	"1/5", "2/5", "3/5", "4/5",
	"1/6", "2/6", "3/6", "4/6", "5/6",
	"1/12", "2/12", "3/12", "4/12", "5/12", "6/12",
	"7/12", "8/12", "9/12", "10/12", "11/12",
}

var fuzzColorPrefixes = []string{
	"bg", "text", "border", "outline", "ring",
	"accent", "caret", "fill", "stroke",
	"shadow", "inset-shadow", "decoration", "divide", "placeholder",
	"from", "via", "to",
	// border color per-side
	"border-t", "border-r", "border-b", "border-l", "border-x", "border-y",
	"border-s", "border-e", "border-bs", "border-be",
}

var fuzzColorFamilies = []string{
	"red", "orange", "amber", "yellow", "lime", "green", "emerald", "teal",
	"cyan", "sky", "blue", "indigo", "violet", "purple", "fuchsia", "pink",
	"rose", "slate", "gray", "zinc", "neutral", "stone",
}

var fuzzColorShades = []string{
	"50", "100", "200", "300", "400", "500", "600", "700", "800", "900", "950",
}

var fuzzColorSpecial = []string{"transparent", "current", "inherit", "white", "black"}

var fuzzMaskStatic = []string{
	"mask-clip-border", "mask-clip-padding", "mask-clip-content", "mask-clip-fill",
	"mask-clip-stroke", "mask-clip-view", "mask-clip-none",
	"mask-composite-add", "mask-composite-subtract", "mask-composite-intersect", "mask-composite-exclude",
	"mask-image-none",
	"mask-alpha", "mask-luminance", "mask-match",
	"mask-origin-border", "mask-origin-padding", "mask-origin-content",
	"mask-origin-fill", "mask-origin-stroke", "mask-origin-view",
	"mask-position-center", "mask-position-top", "mask-position-bottom",
	"mask-position-left", "mask-position-right",
	"mask-repeat", "mask-no-repeat", "mask-repeat-x", "mask-repeat-y",
	"mask-repeat-round", "mask-repeat-space",
	"mask-size-auto", "mask-size-cover", "mask-size-contain",
	"mask-type-alpha", "mask-type-luminance",
	"mask-circle", "mask-ellipse",
	"mask-radial-closest-side", "mask-radial-closest-corner",
	"mask-radial-farthest-side", "mask-radial-farthest-corner",
	"mask-radial-at-center", "mask-radial-at-top", "mask-radial-at-bottom",
	"mask-radial-at-left", "mask-radial-at-right",
}

var fuzzMaskGradientColorPrefixes = []string{
	"mask-linear-from", "mask-linear-to", "mask-radial-from", "mask-radial-to",
	"mask-conic-from", "mask-conic-to",
}

var fuzzMaskGradientPositionPrefixes = []string{
	"mask-x-from", "mask-x-to", "mask-y-from", "mask-y-to",
}

var fuzzMaskGradientPositionValues = []string{
	"0%", "5%", "10%", "25%", "50%", "75%", "100%",
}

var fuzzGradientPercentagePositions = []string{
	"from-0%", "from-5%", "from-10%", "from-50%", "via-30%", "via-50%", "to-90%", "to-100%",
}

var fuzzVariants = []string{
	"hover", "focus", "active", "visited", "disabled", "checked", "required",
	"invalid", "empty", "first", "last", "odd", "even",
	"first-of-type", "last-of-type", "only",
	"focus-within", "focus-visible",
	"before", "after", "placeholder", "file", "marker", "selection",
	"first-line", "first-letter",
	"sm", "md", "lg", "xl", "2xl",
	"dark",
	"motion-safe", "motion-reduce",
	"print",
	"open",
	"aria-checked", "aria-disabled", "aria-expanded", "aria-hidden",
	"group-hover", "group-focus",
	"peer-hover", "peer-focus", "peer-checked",
	// container query variants
	"@sm", "@md", "@lg", "@xl", "@2xl",
	// missing pseudo-class variants
	"target", "enabled", "indeterminate", "default", "valid",
	"placeholder-shown", "autofill", "read-only",
	"user-valid", "user-invalid", "optional", "inert",
	"in-range", "out-of-range", "only-of-type",
	// media query variants
	"portrait", "landscape",
	"contrast-more", "contrast-less",
	"forced-colors", "not-forced-colors",
	"noscript", "inverted-colors",
	"pointer-fine", "pointer-coarse", "pointer-none",
	"any-pointer-fine", "any-pointer-coarse", "any-pointer-none",
	"hover-hover", "hover-none", "any-hover-hover", "any-hover-none",
	// max-* responsive variants
	"max-sm", "max-md", "max-lg", "max-xl", "max-2xl",
	// missing pseudo-element variants
	"backdrop", "details-content",
	// directional variants
	"rtl", "ltr",
	// extended container query variants
	"@3xs", "@2xs", "@xs", "@3xl", "@4xl", "@5xl", "@6xl", "@7xl",
	// @starting-style variant
	"starting",
}

var fuzzCompoundVariantPrefixes = []string{
	"not", "has", "in",
}

var fuzzCompoundVariantValues = []string{
	"hover", "focus", "checked", "disabled", "open", "empty",
}

var fuzzGroupPeerNames = []string{
	"", "/sidebar", "/modal", "/input", "/card",
}

var fuzzAriaAttributes = []string{
	"checked", "disabled", "expanded", "hidden", "pressed", "readonly", "required", "selected", "busy",
}

var fuzzArbitraryAriaValues = []string{
	"[sort=ascending]", "[labelledby=title]", "[controls=panel]",
}

var fuzzDataAttributes = []string{
	"active", "loading", "current", "open",
}

var fuzzArbitraryDataValues = []string{
	"[size=large]", "[loading]", "[state=active]",
}

var fuzzOpacityModifiers = []string{
	"0", "5", "10", "15", "20", "25", "30", "40", "50", "60", "70", "75", "80", "90", "95", "100",
}

var fuzzNegatablePrefixes = []string{
	"m", "mx", "my", "mt", "mr", "mb", "ml",
	"ms", "me", "mbs", "mbe",
	"translate-x", "translate-y", "rotate", "skew-x", "skew-y",
	"order", "z", "indent", "tracking",
	"space-x", "space-y",
	"scroll-m", "scroll-mx", "scroll-my", "scroll-mt", "scroll-mr", "scroll-mb", "scroll-ml",
	"scroll-ms", "scroll-me", "scroll-mbs", "scroll-mbe",
}

var fuzzNegatableValues = []string{
	"0", "0.5", "1", "1.5", "2", "2.5", "3", "4", "5", "6", "8", "10", "12", "16", "px",
}

var fuzzArbitraryValuePrefixes = []string{
	"w", "h", "p", "m", "px", "py", "pt", "mt",
	"bg", "text", "border", "gap",
	"top", "right", "bottom", "left", "inset",
	"rounded", "translate-x", "translate-y", "rotate", "scale",
	"opacity", "z", "order", "grid-cols", "grid-rows",
	"col-span", "row-span", "basis", "min-w", "max-w", "min-h", "max-h",
	"line-clamp", "indent", "aspect", "content",
}

var fuzzPositionLogicalPrefixes = []string{
	"start", "end", "inset-bs", "inset-be",
}

var fuzzBasisValues = []string{
	"0", "0.5", "1", "2", "3", "4", "5", "6", "8", "10", "12", "16", "20",
	"24", "32", "40", "48", "56", "64", "72", "80", "96",
	"auto", "full", "1/2", "1/3", "2/3", "1/4", "3/4", "px",
}

var fuzzGridPrefixes = []string{
	"grid-cols", "grid-rows",
	"col-span", "col-start", "col-end",
	"row-span", "row-start", "row-end",
	"auto-cols", "auto-rows",
}

var fuzzGridValues = []string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "none", "subgrid",
}

var fuzzAutoColRowValues = []string{
	"auto", "min", "max", "fr",
}

var fuzzOrderValues = []string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
	"first", "last", "none",
}

var fuzzColumnsValues = []string{
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12",
	"auto", "3xs", "2xs", "xs", "sm", "md", "lg", "xl", "2xl", "3xl", "4xl", "5xl", "6xl", "7xl",
}

var fuzzFontSizePrefixes = []string{
	"text-xs", "text-sm", "text-base", "text-lg", "text-xl",
	"text-2xl", "text-3xl", "text-4xl", "text-5xl", "text-6xl",
	"text-7xl", "text-8xl", "text-9xl",
}

var fuzzLeadingNumeric = []string{
	"leading-3", "leading-4", "leading-5", "leading-6", "leading-7",
	"leading-8", "leading-9", "leading-10",
}

var fuzzUnderlineOffsetValues = []string{
	"underline-offset-0", "underline-offset-1", "underline-offset-2",
	"underline-offset-4", "underline-offset-8", "underline-offset-auto",
}

var fuzzNegativeArbitraryValues = []string{
	"[2rem]", "[10px]", "[0.5em]", "[calc(1rem+4px)]", "[var(--offset)]",
}

var fuzzImportantArbitraryPrefixes = []string{
	"w", "h", "p", "m", "px", "py", "pt", "mt",
	"bg", "text", "border", "gap",
	"top", "right", "bottom", "left", "inset",
	"rounded", "opacity", "z", "order",
}

var fuzzImportantArbitraryValues = []string{
	"[300px]", "[1.5rem]", "[#ff0000]", "[50%]", "[2em]", "[var(--custom)]",
}

var fuzzArbitraryValues = []string{
	"300px", "1.5rem", "2em", "50%", "100vh",
	"calc(100%-2rem)", "var(--custom)", "#ff0000", "#3b82f6",
	"10px", "0.5", "200ms",
}

var fuzzTypeHints = []string{
	"length", "color", "percentage", "number", "integer",
	"ratio", "url", "shadow", "angle", "time", "any",
}

var fuzzTypeHintedValues = []string{
	"length:--my-size", "color:--my-bg", "percentage:50%",
	"number:1.5", "angle:45deg", "time:200ms",
	"length:1.5rem", "color:red", "shadow:0_1px_2px_black",
}

var fuzzCustomProperties = []string{
	"--my-color", "--sidebar-width", "--spacing-lg",
	"--tw-ring-color", "--custom-bg", "--header-height",
}

var fuzzComplexArbitraryValues = []string{
	"calc(100%_-_2rem)", "calc(100vh_-_4rem)",
	"min(100%,_50rem)", "max(10rem,_50%)",
	"clamp(1rem,_5vw,_3rem)",
	"rgb(255,0,0)", "rgb(255_0_0)", "hsl(200,100%,50%)",
	"oklch(0.7_0.15_200)", "var(--custom)",
	"var(--x,_fallback)", "env(safe-area-inset-top)",
	"url(data:image/svg+xml,...)",
	"100cqw", "50cqh", "100dvh", "100svh", "100lvh",
}

var fuzzArbitraryProperties = []string{
	"[mask-type:alpha]",
	"[content-visibility:auto]",
	"[contain:paint]",
	"[text-wrap:balance]",
	"[writing-mode:vertical-rl]",
	"[clip-path:circle(50%)]",
	"[scroll-timeline:--my-timeline]",
	"[view-transition-name:hero]",
	"[container-type:inline-size]",
	"[overflow-anchor:none]",
	"[overscroll-behavior:contain]",
	"[paint-order:stroke_fill]",
	"[text-rendering:optimizeLegibility]",
	"[word-spacing:0.1em]",
	"[tab-size:4]",
}

var fuzzArbitraryVariants = []string{
	"[&:nth-child(3)]", "[&>svg]", "[&.active]",
	"[&:not(:first-child)]", "[&_p]",
	"[@media(min-width:900px)]",
	"[@supports(display:grid)]",
	"[@container(width>=40rem)]",
}

var fuzzArbitraryOpacityModifiers = []string{
	"[.5]", "[0.75]", "[0.1]", "[var(--opacity)]", "[50%]",
}

var fuzzBorderWidthPrefixes = []string{
	"border", "border-t", "border-r", "border-b", "border-l",
	"border-x", "border-y",
	"border-s", "border-e", "border-bs", "border-be",
}

var fuzzBorderWidthValues = []string{"0", "2", "4", "8"}

var fuzzRoundedPrefixes = []string{
	"rounded", "rounded-t", "rounded-r", "rounded-b", "rounded-l",
	"rounded-tl", "rounded-tr", "rounded-br", "rounded-bl",
	"rounded-s", "rounded-e", "rounded-ss", "rounded-se", "rounded-ee", "rounded-es",
}

var fuzzRoundedValues = []string{
	"none", "sm", "md", "lg", "xl", "2xl", "3xl", "full",
}

var fuzzOutlineOffsetPrefixes = []string{"outline-offset"}

var fuzzOutlineOffsetValues = []string{"0", "1", "2", "4", "8"}

var fuzzDivideWidthPrefixes = []string{"divide-x", "divide-y"}

var fuzzDivideWidthValues = []string{"0", "2", "4", "8", "reverse"}

var fuzzFilterPrefixes = []string{
	"blur", "brightness", "contrast", "saturate", "hue-rotate", "drop-shadow",
}

var fuzzFilterValues = []string{
	"none", "sm", "md", "lg", "xl", "2xl", "3xl",
}

var fuzzBrightnessContrastValues = []string{
	"0", "50", "75", "90", "95", "100", "105", "110", "125", "150", "200",
}

var fuzzBackdropFilterPrefixes = []string{
	"backdrop-blur", "backdrop-brightness", "backdrop-contrast",
	"backdrop-saturate", "backdrop-hue-rotate", "backdrop-opacity",
}

var fuzzTransformPrefixes = []string{
	"translate-x", "translate-y", "translate-z",
	"rotate", "rotate-x", "rotate-y", "rotate-z",
	"scale", "scale-x", "scale-y", "scale-z",
	"skew-x", "skew-y",
}

var fuzzTranslateValues = []string{
	"0", "1", "2", "4", "8", "12", "16", "px", "full", "1/2", "1/3", "1/4",
}

var fuzzRotateValues = []string{
	"0", "1", "2", "3", "6", "12", "45", "90", "180",
}

var fuzzScaleValues = []string{
	"0", "50", "75", "90", "95", "100", "105", "110", "125", "150", "200",
}

var fuzzSkewValues = []string{
	"0", "1", "2", "3", "6", "12",
}

var fuzzShadowSizes = []string{
	"2xs", "xs", "sm", "", "md", "lg", "xl", "2xl", "none",
}

var fuzzInsetShadowSizes = []string{
	"2xs", "xs", "sm", "", "md", "lg", "xl", "2xl", "none",
}

var fuzzOpacityValues = []string{
	"0", "5", "10", "15", "20", "25", "30", "40", "50", "60", "70", "75", "80", "90", "95", "100",
}

var fuzzDurationDelayValues = []string{
	"0", "75", "100", "150", "200", "300", "500", "700", "1000",
}

var fuzzPerspectiveValues = []string{
	"dramatic", "near", "normal", "midrange", "far", "extreme", "none",
}

var fuzzPerspectiveOriginValues = []string{
	"center", "top", "top-right", "right", "bottom-right",
	"bottom", "bottom-left", "left", "top-left",
}

var fuzzOriginValues = []string{
	"center", "top", "top-right", "right", "bottom-right",
	"bottom", "bottom-left", "left", "top-left",
}

var fuzzRadialGradientValues = []string{
	"bg-radial", "bg-radial-at-t", "bg-radial-at-tl", "bg-radial-at-tr",
	"bg-radial-at-b", "bg-radial-at-bl", "bg-radial-at-br",
	"bg-radial-at-l", "bg-radial-at-r", "bg-radial-at-c",
}

var fuzzConicGradientValues = []string{
	"bg-conic",
}

var fuzzColorInterpolation = []string{
	"srgb", "oklab", "oklch", "lab", "lch", "hsl", "hwb", "srgb-linear",
}

var fuzzMaskGradientDirections = []string{
	"mask-linear-to-t", "mask-linear-to-r", "mask-linear-to-b", "mask-linear-to-l",
	"mask-radial", "mask-conic",
}

var fuzzLinearDirections = []string{
	"bg-linear-to-t", "bg-linear-to-tr", "bg-linear-to-r", "bg-linear-to-br",
	"bg-linear-to-b", "bg-linear-to-bl", "bg-linear-to-l", "bg-linear-to-tl",
}

var fuzzStrokeWidthValues = []string{"0", "1", "2"}

var fuzzInsetPrefixes = []string{
	"inset", "inset-x", "inset-y",
	"top", "right", "bottom", "left",
	"start", "end", "inset-bs", "inset-be",
}

var fuzzIndentValues = []string{
	"0", "0.5", "1", "2", "4", "8", "px",
}

var fuzzFontStretchValues = []string{
	"ultra-condensed", "extra-condensed", "condensed", "semi-condensed",
	"normal", "semi-expanded", "expanded", "extra-expanded", "ultra-expanded",
}

var fuzzAnimationValues = []string{
	"spin", "ping", "pulse", "bounce", "none",
}

var fuzzSafeAlignmentValues = []string{
	"items-center-safe", "items-end-safe",
	"justify-center-safe", "justify-end-safe",
	"content-center-safe", "content-end-safe",
	"self-center-safe", "self-end-safe",
}

// Complexity levels for class generation.
const (
	levelSimple = iota
	levelWithVariant
	levelWithModifier
	levelCompound
	levelMultiVariant
	levelNegative
	levelImportant
	levelArbitraryValue
	levelArbitraryProperty
	levelKitchenSink
	levelTypography
	levelBorderVariant
	levelFilterTransform
	levelCompoundVariant
	levelTypeHintedArbitrary
	levelParenCustomProperty
	levelComplexArbitrary
	levelArbitraryVariant
	levelArbitraryOpacityModifier
	levelNegativeImportant
	levelNegativeWithVariant
	levelNegativeArbitrary
	levelImportantArbitrary
	levelGradientChain
	levelGradientInterpolation
	levelMaskGradientDirection
)

// weightedChoice picks an index from a slice of weights using rng.
func weightedChoice(rng *rand.Rand, weights []int) int {
	total := 0
	for _, w := range weights {
		total += w
	}
	r := rng.Intn(total)
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(weights) - 1
}

// pick returns a random element from a string slice.
func pick(rng *rand.Rand, items []string) string {
	return items[rng.Intn(len(items))]
}

// generateFilterUtility generates a filter or backdrop-filter utility.
func generateFilterUtility(rng *rand.Rand, prefixes []string) string {
	prefix := pick(rng, prefixes)
	switch {
	case strings.HasSuffix(prefix, "brightness") || strings.HasSuffix(prefix, "contrast") ||
		strings.HasSuffix(prefix, "saturate") || strings.HasSuffix(prefix, "opacity"):
		return prefix + "-" + pick(rng, fuzzBrightnessContrastValues)
	case strings.HasSuffix(prefix, "hue-rotate"):
		return prefix + "-" + pick(rng, fuzzRotateValues)
	default: // blur, drop-shadow
		return prefix + "-" + pick(rng, fuzzFilterValues)
	}
}

// generateTransformUtility generates a transform utility.
func generateTransformUtility(rng *rand.Rand) string {
	prefix := pick(rng, fuzzTransformPrefixes)
	switch {
	case strings.HasPrefix(prefix, "translate"):
		return prefix + "-" + pick(rng, fuzzTranslateValues)
	case strings.HasPrefix(prefix, "rotate"):
		return prefix + "-" + pick(rng, fuzzRotateValues)
	case strings.HasPrefix(prefix, "scale"):
		return prefix + "-" + pick(rng, fuzzScaleValues)
	case strings.HasPrefix(prefix, "skew"):
		return prefix + "-" + pick(rng, fuzzSkewValues)
	}
	return prefix + "-" + pick(rng, fuzzScaleValues)
}

// generateShadowUtility generates a shadow, inset-shadow, or text-shadow utility.
func generateShadowUtility(rng *rand.Rand) string {
	sub := rng.Intn(3)
	switch sub {
	case 0: // shadow
		size := pick(rng, fuzzShadowSizes)
		if size == "" {
			return "shadow"
		}
		return "shadow-" + size
	case 1: // inset-shadow
		size := pick(rng, fuzzInsetShadowSizes)
		if size == "" {
			return "inset-shadow"
		}
		return "inset-shadow-" + size
	default: // text-shadow
		sizes := []string{"sm", "", "lg", "none"}
		size := pick(rng, sizes)
		if size == "" {
			return "text-shadow"
		}
		return "text-shadow-" + size
	}
}

// generateGradientChain generates a complete gradient specification (from + via + to).
func generateGradientChain(rng *rand.Rand) []string {
	allDirections := make([]string, 0, len(fuzzRadialGradientValues)+len(fuzzLinearDirections)+len(fuzzConicGradientValues))
	allDirections = append(allDirections, fuzzRadialGradientValues...)
	allDirections = append(allDirections, fuzzLinearDirections...)
	allDirections = append(allDirections, fuzzConicGradientValues...)

	direction := pick(rng, allDirections)
	fromColor := "from-" + pick(rng, fuzzColorFamilies) + "-" + pick(rng, fuzzColorShades)
	toColor := "to-" + pick(rng, fuzzColorFamilies) + "-" + pick(rng, fuzzColorShades)

	classes := []string{direction, fromColor}

	// Optionally add via
	if rng.Intn(2) == 0 {
		viaColor := "via-" + pick(rng, fuzzColorFamilies) + "-" + pick(rng, fuzzColorShades)
		classes = append(classes, viaColor)
	}

	// Optionally add positions
	if rng.Intn(3) == 0 {
		classes = append(classes, "from-"+pick(rng, []string{"5%", "10%", "25%", "50%"}))
	}

	classes = append(classes, toColor)
	return classes
}

// generateGridUtility generates a grid parametric utility.
func generateGridUtility(rng *rand.Rand) string {
	prefix := pick(rng, fuzzGridPrefixes)
	if prefix == "auto-cols" || prefix == "auto-rows" {
		return prefix + "-" + pick(rng, fuzzAutoColRowValues)
	}
	return prefix + "-" + pick(rng, fuzzGridValues)
}

// generateBaseUtility generates a random utility without variants or modifiers.
func generateBaseUtility(rng *rand.Rand) string {
	category := rng.Intn(37)
	switch category {
	case 0: // static
		return pick(rng, staticUtilities)
	case 1: // spacing
		return pick(rng, fuzzSpacingPrefixes) + "-" + pick(rng, fuzzSpacingValues)
	case 2: // sizing
		return pick(rng, fuzzSizingPrefixes) + "-" + pick(rng, fuzzSizingValues)
	case 3: // color
		prefix := pick(rng, fuzzColorPrefixes)
		if rng.Intn(5) == 0 {
			return prefix + "-" + pick(rng, fuzzColorSpecial)
		}
		return prefix + "-" + pick(rng, fuzzColorFamilies) + "-" + pick(rng, fuzzColorShades)
	case 4: // font size
		return pick(rng, fuzzFontSizePrefixes)
	case 5: // leading numeric
		return pick(rng, fuzzLeadingNumeric)
	case 6: // underline offset
		return pick(rng, fuzzUnderlineOffsetValues)
	case 7: // border width per-side
		return pick(rng, fuzzBorderWidthPrefixes) + "-" + pick(rng, fuzzBorderWidthValues)
	case 8: // rounded corner variants
		return pick(rng, fuzzRoundedPrefixes) + "-" + pick(rng, fuzzRoundedValues)
	case 9: // outline offset
		return pick(rng, fuzzOutlineOffsetPrefixes) + "-" + pick(rng, fuzzOutlineOffsetValues)
	case 10: // divide width
		return pick(rng, fuzzDivideWidthPrefixes) + "-" + pick(rng, fuzzDivideWidthValues)
	case 11: // border/rounded/ring combined (simple)
		sub := rng.Intn(3)
		switch sub {
		case 0:
			return pick(rng, fuzzBorderWidthPrefixes) + "-" + pick(rng, fuzzBorderWidthValues)
		case 1:
			return pick(rng, fuzzRoundedPrefixes) + "-" + pick(rng, fuzzRoundedValues)
		default:
			return pick(rng, fuzzDivideWidthPrefixes) + "-" + pick(rng, fuzzDivideWidthValues)
		}
	case 12: // filter
		return generateFilterUtility(rng, fuzzFilterPrefixes)
	case 13: // backdrop filter
		return generateFilterUtility(rng, fuzzBackdropFilterPrefixes)
	case 14: // transform
		return generateTransformUtility(rng)
	case 15: // shadow
		return generateShadowUtility(rng)
	case 16: // opacity
		return "opacity-" + pick(rng, fuzzOpacityValues)
	case 17: // duration/delay
		if rng.Intn(2) == 0 {
			return "duration-" + pick(rng, fuzzDurationDelayValues)
		}
		return "delay-" + pick(rng, fuzzDurationDelayValues)
	case 18: // logical positioning
		return pick(rng, fuzzPositionLogicalPrefixes) + "-" + pick(rng, fuzzSpacingValues)
	case 19: // basis
		return "basis-" + pick(rng, fuzzBasisValues)
	case 20: // grid parametric
		return generateGridUtility(rng)
	case 21: // order
		return "order-" + pick(rng, fuzzOrderValues)
	case 22: // columns
		return "columns-" + pick(rng, fuzzColumnsValues)
	case 23: // aspect ratio arbitrary
		arb := pick(rng, fuzzArbitraryValues)
		return "aspect-[" + arb + "]"
	case 24: // mask static
		return pick(rng, fuzzMaskStatic)
	case 25: // mask gradient color stops
		prefix := pick(rng, fuzzMaskGradientColorPrefixes)
		if rng.Intn(5) == 0 {
			return prefix + "-" + pick(rng, fuzzColorSpecial)
		}
		return prefix + "-" + pick(rng, fuzzColorFamilies) + "-" + pick(rng, fuzzColorShades)
	case 26: // mask gradient positions
		prefix := pick(rng, fuzzMaskGradientPositionPrefixes)
		return prefix + "-[" + pick(rng, fuzzMaskGradientPositionValues) + "]"
	case 27: // gradient percentage positions
		return pick(rng, fuzzGradientPercentagePositions)
	case 28: // perspective
		if rng.Intn(2) == 0 {
			return "perspective-" + pick(rng, fuzzPerspectiveValues)
		}
		return "perspective-origin-" + pick(rng, fuzzPerspectiveOriginValues)
	case 29: // transform origin
		return "origin-" + pick(rng, fuzzOriginValues)
	case 30: // advanced gradients (radial and conic)
		if rng.Intn(3) == 0 {
			return pick(rng, fuzzConicGradientValues)
		}
		return pick(rng, fuzzRadialGradientValues)
	case 31: // stroke width
		return "stroke-" + pick(rng, fuzzStrokeWidthValues)
	case 32: // inset positioning
		return pick(rng, fuzzInsetPrefixes) + "-" + pick(rng, fuzzSpacingValues)
	case 33: // indent
		val := pick(rng, fuzzIndentValues)
		if rng.Intn(3) == 0 {
			return "-indent-" + val
		}
		return "indent-" + val
	case 34: // font-stretch
		return "font-stretch-" + pick(rng, fuzzFontStretchValues)
	case 35: // content and animation
		if rng.Intn(2) == 0 {
			contentValues := []string{"content-none", "content-['hello']", "content-[attr(data-label)]"}
			return pick(rng, contentValues)
		}
		return "animate-" + pick(rng, fuzzAnimationValues)
	case 36: // safe alignment
		return pick(rng, fuzzSafeAlignmentValues)
	}
	return pick(rng, staticUtilities)
}

// generateColorUtility generates a color utility suitable for opacity modifiers.
func generateColorUtility(rng *rand.Rand) string {
	prefix := pick(rng, fuzzColorPrefixes)
	return prefix + "-" + pick(rng, fuzzColorFamilies) + "-" + pick(rng, fuzzColorShades)
}

// generateClassAtLevel generates a class at the specified complexity level.
func generateClassAtLevel(rng *rand.Rand, level int) string {
	switch level {
	case levelSimple:
		return generateBaseUtility(rng)

	case levelWithVariant:
		return pick(rng, fuzzVariants) + ":" + generateBaseUtility(rng)

	case levelWithModifier:
		util := generateColorUtility(rng)
		return util + "/" + pick(rng, fuzzOpacityModifiers)

	case levelCompound:
		util := generateColorUtility(rng)
		return pick(rng, fuzzVariants) + ":" + util + "/" + pick(rng, fuzzOpacityModifiers)

	case levelMultiVariant:
		v1 := pick(rng, fuzzVariants)
		v2 := pick(rng, fuzzVariants)
		for v2 == v1 {
			v2 = pick(rng, fuzzVariants)
		}
		return v1 + ":" + v2 + ":" + generateBaseUtility(rng)

	case levelNegative:
		prefix := pick(rng, fuzzNegatablePrefixes)
		val := pick(rng, fuzzNegatableValues)
		return "-" + prefix + "-" + val

	case levelImportant:
		return "!" + generateBaseUtility(rng)

	case levelArbitraryValue:
		prefix := pick(rng, fuzzArbitraryValuePrefixes)
		val := pick(rng, fuzzArbitraryValues)
		return prefix + "-[" + val + "]"

	case levelArbitraryProperty:
		return pick(rng, fuzzArbitraryProperties)

	case levelKitchenSink:
		variant := pick(rng, fuzzVariants)
		util := generateColorUtility(rng)
		mod := pick(rng, fuzzOpacityModifiers)
		return variant + ":!" + util + "/" + mod

	case levelTypography:
		typoSets := [][]string{
			fuzzFontSizePrefixes,
			fuzzLeadingNumeric,
			fuzzUnderlineOffsetValues,
		}
		return pick(rng, typoSets[rng.Intn(len(typoSets))])

	case levelBorderVariant:
		variant := pick(rng, fuzzVariants)
		sub := rng.Intn(4)
		switch sub {
		case 0: // border width with variant
			return variant + ":" + pick(rng, fuzzBorderWidthPrefixes) + "-" + pick(rng, fuzzBorderWidthValues)
		case 1: // rounded with variant
			return variant + ":" + pick(rng, fuzzRoundedPrefixes) + "-" + pick(rng, fuzzRoundedValues)
		case 2: // divide width with variant
			return variant + ":" + pick(rng, fuzzDivideWidthPrefixes) + "-" + pick(rng, fuzzDivideWidthValues)
		default: // outline offset with variant
			return variant + ":" + pick(rng, fuzzOutlineOffsetPrefixes) + "-" + pick(rng, fuzzOutlineOffsetValues)
		}

	case levelFilterTransform:
		sub := rng.Intn(5)
		switch sub {
		case 0: // filter
			return generateFilterUtility(rng, fuzzFilterPrefixes)
		case 1: // backdrop filter
			return generateFilterUtility(rng, fuzzBackdropFilterPrefixes)
		case 2: // transform
			return generateTransformUtility(rng)
		case 3: // shadow
			return generateShadowUtility(rng)
		default: // opacity/duration/delay
			switch rng.Intn(3) {
			case 0:
				return "opacity-" + pick(rng, fuzzOpacityValues)
			case 1:
				return "duration-" + pick(rng, fuzzDurationDelayValues)
			default:
				return "delay-" + pick(rng, fuzzDurationDelayValues)
			}
		}

	case levelCompoundVariant:
		sub := rng.Intn(7)
		switch sub {
		case 0: // not-*/has-*/in-*
			prefix := pick(rng, fuzzCompoundVariantPrefixes)
			val := pick(rng, fuzzCompoundVariantValues)
			return prefix + "-" + val + ":" + generateBaseUtility(rng)
		case 1: // group-*/name
			val := pick(rng, fuzzCompoundVariantValues)
			name := pick(rng, fuzzGroupPeerNames)
			return "group-" + val + name + ":" + generateBaseUtility(rng)
		case 2: // peer-*/name
			val := pick(rng, fuzzCompoundVariantValues)
			name := pick(rng, fuzzGroupPeerNames)
			return "peer-" + val + name + ":" + generateBaseUtility(rng)
		case 3: // aria-*
			if rng.Intn(2) == 0 {
				return "aria-" + pick(rng, fuzzAriaAttributes) + ":" + generateBaseUtility(rng)
			}
			return "aria-" + pick(rng, fuzzArbitraryAriaValues) + ":" + generateBaseUtility(rng)
		case 4: // data-*
			if rng.Intn(2) == 0 {
				return "data-" + pick(rng, fuzzDataAttributes) + ":" + generateBaseUtility(rng)
			}
			return "data-" + pick(rng, fuzzArbitraryDataValues) + ":" + generateBaseUtility(rng)
		case 5: // nth-*
			nths := []string{"[2]", "[3]", "[3n+1]", "[odd]", "[even]"}
			return "nth-" + pick(rng, nths) + ":" + generateBaseUtility(rng)
		case 6: // *: and **: children/descendants
			if rng.Intn(2) == 0 {
				return "*:" + generateBaseUtility(rng)
			}
			return "**:" + generateBaseUtility(rng)
		}

	case levelTypeHintedArbitrary:
		prefix := pick(rng, fuzzArbitraryValuePrefixes)
		val := pick(rng, fuzzTypeHintedValues)
		return prefix + "-[" + val + "]"

	case levelParenCustomProperty:
		prefix := pick(rng, fuzzArbitraryValuePrefixes)
		prop := pick(rng, fuzzCustomProperties)
		return prefix + "-(" + prop + ")"

	case levelComplexArbitrary:
		prefix := pick(rng, fuzzArbitraryValuePrefixes)
		val := pick(rng, fuzzComplexArbitraryValues)
		return prefix + "-[" + val + "]"

	case levelArbitraryVariant:
		variant := pick(rng, fuzzArbitraryVariants)
		return variant + ":" + generateBaseUtility(rng)

	case levelArbitraryOpacityModifier:
		util := generateColorUtility(rng)
		return util + "/" + pick(rng, fuzzArbitraryOpacityModifiers)

	case levelNegativeImportant:
		prefix := pick(rng, fuzzNegatablePrefixes)
		val := pick(rng, fuzzNegatableValues)
		return "!-" + prefix + "-" + val

	case levelNegativeWithVariant:
		prefix := pick(rng, fuzzNegatablePrefixes)
		val := pick(rng, fuzzNegatableValues)
		variant := pick(rng, fuzzVariants)
		return variant + ":-" + prefix + "-" + val

	case levelNegativeArbitrary:
		prefix := pick(rng, fuzzNegatablePrefixes)
		val := pick(rng, fuzzNegativeArbitraryValues)
		return "-" + prefix + "-" + val

	case levelImportantArbitrary:
		prefix := pick(rng, fuzzImportantArbitraryPrefixes)
		val := pick(rng, fuzzImportantArbitraryValues)
		return "!" + prefix + "-" + val

	case levelGradientChain:
		chain := generateGradientChain(rng)
		return pick(rng, chain)

	case levelGradientInterpolation:
		allDirections := make([]string, 0, len(fuzzLinearDirections)+len(fuzzRadialGradientValues)+len(fuzzConicGradientValues))
		allDirections = append(allDirections, fuzzLinearDirections...)
		allDirections = append(allDirections, fuzzRadialGradientValues...)
		allDirections = append(allDirections, fuzzConicGradientValues...)
		return pick(rng, allDirections) + "/" + pick(rng, fuzzColorInterpolation)

	case levelMaskGradientDirection:
		return pick(rng, fuzzMaskGradientDirections)
	}
	return generateBaseUtility(rng)
}

// generateRandomClasses produces count pseudo-random Tailwind classes.
func generateRandomClasses(rng *rand.Rand, count int) []string {
	classes := make([]string, 0, count)
	weights := []int{
		25, // simple
		16, // with variant
		8,  // with modifier
		8,  // compound
		6,  // multi-variant
		5,  // negative
		4,  // important
		5,  // arbitrary value
		4,  // arbitrary property
		2,  // kitchen sink
		10, // typography
		8,  // border variant
		8,  // filter/transform
		12, // compound variant
		5,  // type-hinted arbitrary
		4,  // parenthesized custom property
		5,  // complex arbitrary
		5,  // arbitrary variant
		4,  // arbitrary opacity modifier
		4,  // negative + important
		4,  // negative with variant
		4,  // negative with arbitrary value
		4,  // important with arbitrary value
		6,  // gradient chain
		4,  // gradient with color interpolation
		3,  // mask gradient direction
	}

	for i := 0; i < count; i++ {
		level := weightedChoice(rng, weights)
		classes = append(classes, generateClassAtLevel(rng, level))
	}
	return classes
}

// TestClassGenerator verifies the class generator produces correct output.
// This test does NOT require npm/node — it only tests the generator itself.
func TestClassGenerator(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	classes := generateRandomClasses(rng, 200)

	if len(classes) != 200 {
		t.Fatalf("expected 200 classes, got %d", len(classes))
	}

	// Verify all classes are non-empty strings.
	for i, c := range classes {
		if c == "" {
			t.Errorf("class %d is empty", i)
		}
	}

	// Verify determinism: same seed produces same output.
	rng2 := rand.New(rand.NewSource(42))
	classes2 := generateRandomClasses(rng2, 200)
	for i := range classes {
		if classes[i] != classes2[i] {
			t.Errorf("non-deterministic at index %d: %q vs %q", i, classes[i], classes2[i])
		}
	}

	// Verify diversity: check that multiple complexity levels are represented.
	hasVariant := false
	hasNegative := false
	hasArbitrary := false
	hasCompoundVariant := false
	hasTypeHinted := false
	hasParenCustomProp := false
	hasArbitraryVariant := false
	hasNegativeImportant := false
	hasNegativeArbitrary := false
	hasImportantArbitrary := false
	hasGradientInterpolation := false
	hasRadialOrConic := false
	hasMaskGradientDir := false
	for _, c := range classes {
		if len(c) > 0 && c[0] == '-' {
			hasNegative = true
		}
		if strings.Contains(c, ":") {
			hasVariant = true
		}
		if strings.Contains(c, "[") {
			hasArbitrary = true
		}
		// Detect compound variants: not-*, has-*, in-*, group-*/name, peer-*/name, nth-*, *:, **:
		for _, prefix := range []string{"not-", "has-", "in-", "group-", "peer-", "nth-", "*:", "**:"} {
			if strings.HasPrefix(c, prefix) && strings.Contains(c, ":") {
				hasCompoundVariant = true
			}
		}
		// Detect type-hinted arbitrary values like text-[length:--my-size]
		for _, hint := range fuzzTypeHints {
			if strings.Contains(c, "["+hint+":") {
				hasTypeHinted = true
			}
		}
		// Detect parenthesized custom properties like w-(--my-color)
		if strings.Contains(c, "(--") {
			hasParenCustomProp = true
		}
		// Detect arbitrary variants like [&:nth-child(3)]:
		if strings.HasPrefix(c, "[") && strings.Contains(c, "]:") {
			hasArbitraryVariant = true
		}
		// Detect negative+important like !-translate-x-4
		if strings.HasPrefix(c, "!-") {
			hasNegativeImportant = true
		}
		// Detect negative with arbitrary values like -m-[10px]
		if len(c) > 0 && c[0] == '-' && strings.Contains(c, "[") {
			hasNegativeArbitrary = true
		}
		// Detect important with arbitrary values like !w-[300px]
		if len(c) > 0 && c[0] == '!' && strings.Contains(c, "[") {
			hasImportantArbitrary = true
		}
		// Detect gradient with color interpolation like bg-linear-to-r/oklab
		for _, interp := range fuzzColorInterpolation {
			if strings.HasSuffix(c, "/"+interp) {
				hasGradientInterpolation = true
			}
		}
		// Detect radial or conic gradients
		if strings.HasPrefix(c, "bg-radial") || strings.HasPrefix(c, "bg-conic") {
			hasRadialOrConic = true
		}
		// Detect mask gradient directions
		for _, d := range fuzzMaskGradientDirections {
			if c == d {
				hasMaskGradientDir = true
			}
		}
	}
	if !hasVariant {
		t.Error("no variant classes generated")
	}
	if !hasNegative {
		t.Error("no negative classes generated")
	}
	if !hasArbitrary {
		t.Error("no arbitrary value classes generated")
	}
	if !hasCompoundVariant {
		t.Error("no compound variant classes generated")
	}
	if !hasTypeHinted {
		t.Error("no type-hinted arbitrary value classes generated")
	}
	if !hasParenCustomProp {
		t.Error("no parenthesized custom property classes generated")
	}
	if !hasArbitraryVariant {
		t.Error("no arbitrary variant classes generated")
	}
	if !hasNegativeImportant {
		t.Error("no negative+important classes generated")
	}
	if !hasNegativeArbitrary {
		t.Error("no negative with arbitrary value classes generated")
	}
	if !hasImportantArbitrary {
		t.Error("no important with arbitrary value classes generated")
	}
	if !hasGradientInterpolation {
		t.Error("no gradient with color interpolation classes generated")
	}
	if !hasRadialOrConic {
		t.Error("no radial or conic gradient classes generated")
	}
	if !hasMaskGradientDir {
		t.Error("no mask gradient direction classes generated")
	}

	// Verify 500 classes produces at least 400 unique (high diversity).
	rng3 := rand.New(rand.NewSource(42))
	large := generateRandomClasses(rng3, 500)
	seen := map[string]bool{}
	for _, c := range large {
		seen[c] = true
	}
	if len(seen) < 350 {
		t.Errorf("expected at least 350 unique classes from 500 generated, got %d", len(seen))
	}
	t.Logf("Generated %d unique classes from 500 total", len(seen))
}

// TestFuzzPropertyDeclarationCoverage verifies that specific class combinations
// involving --tw-content, --tw-scroll-snap-strictness, and --tw-divide-*-reverse
// produce valid CSS with the correct @property declarations.
func TestFuzzPropertyDeclarationCoverage(t *testing.T) {
	cases := []struct {
		name    string
		classes string
		check   func(t *testing.T, css string)
	}{
		{
			name:    "snap-x alone",
			classes: "snap-x",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "scroll-snap-type: x var(--tw-scroll-snap-strictness)") {
					t.Error("snap-x should emit scroll-snap-type with strictness variable")
				}
			},
		},
		{
			name:    "snap-y alone",
			classes: "snap-y",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "scroll-snap-type: y var(--tw-scroll-snap-strictness)") {
					t.Error("snap-y should emit scroll-snap-type with strictness variable")
				}
			},
		},
		{
			name:    "snap-both alone",
			classes: "snap-both",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "scroll-snap-type: both var(--tw-scroll-snap-strictness)") {
					t.Error("snap-both should emit scroll-snap-type with strictness variable")
				}
			},
		},
		{
			name:    "snap-x snap-mandatory combined",
			classes: "snap-x snap-mandatory",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "scroll-snap-type: x var(--tw-scroll-snap-strictness)") {
					t.Error("snap-x should emit scroll-snap-type")
				}
				if !strings.Contains(css, "--tw-scroll-snap-strictness: mandatory") {
					t.Error("snap-mandatory should set strictness to mandatory")
				}
			},
		},
		{
			name:    "divide-x",
			classes: "divide-x",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "--tw-divide-x-reverse") {
					t.Error("divide-x should reference --tw-divide-x-reverse")
				}
			},
		},
		{
			name:    "divide-y",
			classes: "divide-y",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "--tw-divide-y-reverse") {
					t.Error("divide-y should reference --tw-divide-y-reverse")
				}
			},
		},
		{
			name:    "before:content-['hello']",
			classes: "before:content-['hello']",
			check: func(t *testing.T, css string) {
				if !strings.Contains(css, "::before") {
					t.Error("before: variant should produce ::before pseudo-element")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := New()
			e.Write([]byte(`class="` + tc.classes + `"`))
			css := e.CSS()
			t.Logf("CSS for %q:\n%s", tc.classes, css)
			tc.check(t, css)
		})
	}
}
