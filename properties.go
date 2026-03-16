package tailwind

// twPropertyDeclarations contains the @property rules for Tailwind's
// internal --tw-* CSS custom properties. These use the CSS @property
// syntax to register properties with initial values, which modern
// browsers use to provide defaults.
const twPropertyDeclarations = `@property --tw-translate-x {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-translate-y {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-translate-z {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-scale-x {
  syntax: "*";
  inherits: false;
  initial-value: 1;
}
@property --tw-scale-y {
  syntax: "*";
  inherits: false;
  initial-value: 1;
}
@property --tw-scale-z {
  syntax: "*";
  inherits: false;
  initial-value: 1;
}
@property --tw-rotate-x {
  syntax: "*";
  inherits: false;
}
@property --tw-rotate-y {
  syntax: "*";
  inherits: false;
}
@property --tw-rotate-z {
  syntax: "*";
  inherits: false;
}
@property --tw-skew-x {
  syntax: "*";
  inherits: false;
}
@property --tw-skew-y {
  syntax: "*";
  inherits: false;
}
@property --tw-space-y-reverse {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-space-x-reverse {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-border-style {
  syntax: "*";
  inherits: false;
  initial-value: solid;
}
@property --tw-leading {
  syntax: "*";
  inherits: false;
}
@property --tw-font-weight {
  syntax: "*";
  inherits: false;
}
@property --tw-tracking {
  syntax: "*";
  inherits: false;
}
@property --tw-shadow {
  syntax: "*";
  inherits: false;
  initial-value: 0 0 #0000;
}
@property --tw-shadow-color {
  syntax: "*";
  inherits: false;
}
@property --tw-shadow-alpha {
  syntax: "<percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-inset-shadow {
  syntax: "*";
  inherits: false;
  initial-value: 0 0 #0000;
}
@property --tw-inset-shadow-color {
  syntax: "*";
  inherits: false;
}
@property --tw-inset-shadow-alpha {
  syntax: "<percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-ring-color {
  syntax: "*";
  inherits: false;
}
@property --tw-ring-shadow {
  syntax: "*";
  inherits: false;
  initial-value: 0 0 #0000;
}
@property --tw-inset-ring-color {
  syntax: "*";
  inherits: false;
}
@property --tw-inset-ring-shadow {
  syntax: "*";
  inherits: false;
  initial-value: 0 0 #0000;
}
@property --tw-ring-inset {
  syntax: "*";
  inherits: false;
}
@property --tw-ring-offset-width {
  syntax: "<length>";
  inherits: false;
  initial-value: 0px;
}
@property --tw-ring-offset-color {
  syntax: "*";
  inherits: false;
  initial-value: #fff;
}
@property --tw-ring-offset-shadow {
  syntax: "*";
  inherits: false;
  initial-value: 0 0 #0000;
}
@property --tw-blur {
  syntax: "*";
  inherits: false;
}
@property --tw-brightness {
  syntax: "*";
  inherits: false;
}
@property --tw-contrast {
  syntax: "*";
  inherits: false;
}
@property --tw-grayscale {
  syntax: "*";
  inherits: false;
}
@property --tw-hue-rotate {
  syntax: "*";
  inherits: false;
}
@property --tw-invert {
  syntax: "*";
  inherits: false;
}
@property --tw-opacity {
  syntax: "*";
  inherits: false;
}
@property --tw-saturate {
  syntax: "*";
  inherits: false;
}
@property --tw-sepia {
  syntax: "*";
  inherits: false;
}
@property --tw-drop-shadow {
  syntax: "*";
  inherits: false;
}
@property --tw-drop-shadow-color {
  syntax: "*";
  inherits: false;
}
@property --tw-drop-shadow-alpha {
  syntax: "<percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-text-shadow-color {
  syntax: "*";
  inherits: false;
}
@property --tw-text-shadow-alpha {
  syntax: "<percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-drop-shadow-size {
  syntax: "*";
  inherits: false;
}
@property --tw-duration {
  syntax: "*";
  inherits: false;
}
@property --tw-ease {
  syntax: "*";
  inherits: false;
}
@property --tw-outline-style {
  syntax: "*";
  inherits: false;
}
@property --tw-border-spacing-x {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-border-spacing-y {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-gradient-from {
  syntax: "<color>";
  inherits: false;
  initial-value: #0000;
}
@property --tw-gradient-via {
  syntax: "<color>";
  inherits: false;
  initial-value: #0000;
}
@property --tw-gradient-to {
  syntax: "<color>";
  inherits: false;
  initial-value: #0000;
}
@property --tw-gradient-stops {
  syntax: "*";
  inherits: false;
}
@property --tw-gradient-via-stops {
  syntax: "*";
  inherits: false;
}
@property --tw-gradient-position {
  syntax: "*";
  inherits: false;
}
@property --tw-gradient-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-gradient-via-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 50%;
}
@property --tw-gradient-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-ordinal {
  syntax: "*";
  inherits: false;
}
@property --tw-slashed-zero {
  syntax: "*";
  inherits: false;
}
@property --tw-numeric-figure {
  syntax: "*";
  inherits: false;
}
@property --tw-numeric-spacing {
  syntax: "*";
  inherits: false;
}
@property --tw-numeric-fraction {
  syntax: "*";
  inherits: false;
}
@property --tw-contain-size {
  syntax: "*";
  inherits: false;
}
@property --tw-contain-layout {
  syntax: "*";
  inherits: false;
}
@property --tw-contain-paint {
  syntax: "*";
  inherits: false;
}
@property --tw-contain-style {
  syntax: "*";
  inherits: false;
}
@property --tw-pan-x {
  syntax: "*";
  inherits: false;
}
@property --tw-pan-y {
  syntax: "*";
  inherits: false;
}
@property --tw-pinch-zoom {
  syntax: "*";
  inherits: false;
}
@property --tw-content {
  syntax: "*";
  inherits: false;
  initial-value: "";
}
@property --tw-scroll-snap-strictness {
  syntax: "*";
  inherits: false;
  initial-value: proximity;
}
@property --tw-divide-x-reverse {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-divide-y-reverse {
  syntax: "*";
  inherits: false;
  initial-value: 0;
}
@property --tw-mask-linear {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-radial {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-conic {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-top {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-bottom {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-left {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-right {
  syntax: "*";
  inherits: false;
  initial-value: linear-gradient(#fff, #fff);
}
@property --tw-mask-linear-position {
  syntax: "*";
  inherits: false;
  initial-value: 0deg;
}
@property --tw-mask-linear-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-linear-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-linear-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-linear-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-mask-linear-stops {
  syntax: "*";
  inherits: false;
}
@property --tw-mask-radial-shape {
  syntax: "*";
  inherits: false;
  initial-value: ellipse;
}
@property --tw-mask-radial-size {
  syntax: "*";
  inherits: false;
  initial-value: farthest-corner;
}
@property --tw-mask-radial-position {
  syntax: "*";
  inherits: false;
  initial-value: center;
}
@property --tw-mask-radial-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-radial-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-radial-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-radial-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-mask-radial-stops {
  syntax: "*";
  inherits: false;
}
@property --tw-mask-conic-position {
  syntax: "*";
  inherits: false;
  initial-value: 0deg;
}
@property --tw-mask-conic-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-conic-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-conic-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-conic-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-mask-conic-stops {
  syntax: "*";
  inherits: false;
}
@property --tw-mask-top-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-top-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-top-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-top-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-mask-bottom-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-bottom-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-bottom-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-bottom-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-mask-left-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-left-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-left-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-left-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
@property --tw-mask-right-from-color {
  syntax: "*";
  inherits: false;
  initial-value: black;
}
@property --tw-mask-right-to-color {
  syntax: "*";
  inherits: false;
  initial-value: transparent;
}
@property --tw-mask-right-from-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 0%;
}
@property --tw-mask-right-to-position {
  syntax: "<length-percentage>";
  inherits: false;
  initial-value: 100%;
}
`

// twPropertiesFallbackLayer provides the legacy fallback for browsers
// that do not support CSS @property. It uses a @supports query targeting
// older Safari/Firefox to set --tw-* defaults on all elements.
const twPropertiesFallbackLayer = `@layer properties {
  @supports ((-webkit-hyphens: none) and (not (margin-trim: inline))) or ((-moz-orient: inline) and (not (color:rgb(from red r g b)))) {
    *, ::before, ::after, ::backdrop {
      --tw-translate-x: 0;
      --tw-translate-y: 0;
      --tw-translate-z: 0;
      --tw-scale-x: 1;
      --tw-scale-y: 1;
      --tw-scale-z: 1;
      --tw-rotate-x: initial;
      --tw-rotate-y: initial;
      --tw-rotate-z: initial;
      --tw-skew-x: initial;
      --tw-skew-y: initial;
      --tw-space-x-reverse: 0;
      --tw-space-y-reverse: 0;
      --tw-border-style: solid;
      --tw-leading: initial;
      --tw-font-weight: initial;
      --tw-tracking: initial;
      --tw-shadow: 0 0 #0000;
      --tw-shadow-color: initial;
      --tw-shadow-alpha: 100%;
      --tw-inset-shadow: 0 0 #0000;
      --tw-inset-shadow-color: initial;
      --tw-inset-shadow-alpha: 100%;
      --tw-ring-color: initial;
      --tw-ring-shadow: 0 0 #0000;
      --tw-inset-ring-color: initial;
      --tw-inset-ring-shadow: 0 0 #0000;
      --tw-ring-inset: initial;
      --tw-ring-offset-width: 0px;
      --tw-ring-offset-color: #fff;
      --tw-ring-offset-shadow: 0 0 #0000;
      --tw-blur: initial;
      --tw-brightness: initial;
      --tw-contrast: initial;
      --tw-grayscale: initial;
      --tw-hue-rotate: initial;
      --tw-invert: initial;
      --tw-opacity: initial;
      --tw-saturate: initial;
      --tw-sepia: initial;
      --tw-drop-shadow: initial;
      --tw-drop-shadow-color: initial;
      --tw-drop-shadow-alpha: 100%;
      --tw-drop-shadow-size: initial;
      --tw-text-shadow-color: initial;
      --tw-text-shadow-alpha: 100%;
      --tw-duration: initial;
      --tw-ease: initial;
      --tw-outline-style: initial;
      --tw-border-spacing-x: 0;
      --tw-border-spacing-y: 0;
      --tw-gradient-from: #0000;
      --tw-gradient-via: #0000;
      --tw-gradient-to: #0000;
      --tw-gradient-stops: initial;
      --tw-gradient-via-stops: initial;
      --tw-gradient-position: initial;
      --tw-gradient-from-position: 0%;
      --tw-gradient-via-position: 50%;
      --tw-gradient-to-position: 100%;
      --tw-ordinal: initial;
      --tw-slashed-zero: initial;
      --tw-numeric-figure: initial;
      --tw-numeric-spacing: initial;
      --tw-numeric-fraction: initial;
      --tw-contain-size: initial;
      --tw-contain-layout: initial;
      --tw-contain-paint: initial;
      --tw-contain-style: initial;
      --tw-pan-x: initial;
      --tw-pan-y: initial;
      --tw-pinch-zoom: initial;
      --tw-content: "";
      --tw-scroll-snap-strictness: proximity;
      --tw-divide-x-reverse: 0;
      --tw-divide-y-reverse: 0;
      --tw-mask-linear: linear-gradient(#fff, #fff);
      --tw-mask-radial: linear-gradient(#fff, #fff);
      --tw-mask-conic: linear-gradient(#fff, #fff);
      --tw-mask-top: linear-gradient(#fff, #fff);
      --tw-mask-bottom: linear-gradient(#fff, #fff);
      --tw-mask-left: linear-gradient(#fff, #fff);
      --tw-mask-right: linear-gradient(#fff, #fff);
      --tw-mask-linear-position: 0deg;
      --tw-mask-linear-from-color: black;
      --tw-mask-linear-to-color: transparent;
      --tw-mask-linear-from-position: 0%;
      --tw-mask-linear-to-position: 100%;
      --tw-mask-linear-stops: initial;
      --tw-mask-radial-shape: ellipse;
      --tw-mask-radial-size: farthest-corner;
      --tw-mask-radial-position: center;
      --tw-mask-radial-from-color: black;
      --tw-mask-radial-to-color: transparent;
      --tw-mask-radial-from-position: 0%;
      --tw-mask-radial-to-position: 100%;
      --tw-mask-radial-stops: initial;
      --tw-mask-conic-position: 0deg;
      --tw-mask-conic-from-color: black;
      --tw-mask-conic-to-color: transparent;
      --tw-mask-conic-from-position: 0%;
      --tw-mask-conic-to-position: 100%;
      --tw-mask-conic-stops: initial;
      --tw-mask-top-from-color: black;
      --tw-mask-top-to-color: transparent;
      --tw-mask-top-from-position: 0%;
      --tw-mask-top-to-position: 100%;
      --tw-mask-bottom-from-color: black;
      --tw-mask-bottom-to-color: transparent;
      --tw-mask-bottom-from-position: 0%;
      --tw-mask-bottom-to-position: 100%;
      --tw-mask-left-from-color: black;
      --tw-mask-left-to-color: transparent;
      --tw-mask-left-from-position: 0%;
      --tw-mask-left-to-position: 100%;
      --tw-mask-right-from-color: black;
      --tw-mask-right-to-color: transparent;
      --tw-mask-right-from-position: 0%;
      --tw-mask-right-to-position: 100%;
    }
  }
}
`
