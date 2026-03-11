package tailwind

import "strings"

// isTextFile reports whether the content appears to be a text file
// suitable for scanning for Tailwind class candidates.
// It inspects up to the first 512 bytes for binary indicators.
func isTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}

	n := len(content)
	if n > 512 {
		n = 512
	}
	sample := content[:n]

	controlCount := 0
	for _, b := range sample {
		if b == 0x00 {
			return false
		}
		if (b >= 0x01 && b <= 0x08) || (b >= 0x0E && b <= 0x1F) {
			controlCount++
		}
	}

	if float64(controlCount)/float64(n) > 0.10 {
		return false
	}

	return true
}

// hasTextExtension reports whether the file path has an extension
// commonly associated with text files that may contain Tailwind classes.
// Returns false for known binary extensions (images, fonts, audio, video,
// compiled artifacts). Returns true for unknown extensions (safe default:
// fall through to content-based detection).
func hasTextExtension(path string) bool {
	dot := strings.LastIndex(path, ".")
	if dot < 0 {
		return true
	}
	ext := strings.ToLower(path[dot:])

	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".avif", ".ico",
		".woff", ".woff2", ".ttf", ".otf", ".eot",
		".mp3", ".mp4", ".webm", ".ogg", ".wav",
		".zip", ".tar", ".gz", ".br", ".zst",
		".pdf", ".wasm",
		".exe", ".dll", ".so", ".dylib", ".o", ".a":
		return false
	}

	return true
}
