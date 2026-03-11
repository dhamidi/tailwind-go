package cssdata

import _ "embed"

//go:embed theme.css
var Theme []byte

//go:embed utilities.css
var Utilities []byte

//go:embed preflight.css
var Preflight []byte
