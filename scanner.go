package tailwind

// scanner extracts candidate Tailwind class names from a raw byte stream.
//
// It is completely syntax-agnostic: it doesn't know about HTML, JSX,
// Go templates, or any other format. It simply tokenizes the byte
// stream into whitespace/delimiter-separated tokens and applies a
// lightweight filter to reject tokens that obviously can't be Tailwind
// classes (e.g., pure punctuation, HTML tags, etc.).
//
// This deliberately over-extracts. False positives are expected and
// harmless — they simply won't match any utility definition and will
// be discarded during generation. The alternative (parsing specific
// template syntaxes) would couple the engine to particular formats,
// which is exactly what we want to avoid.
type scanner struct {
	buf   []byte // current partial token carried across Write calls
	depth int    // bracket nesting depth (for arbitrary values)
}

// feed processes a chunk of bytes and returns completed candidate tokens.
// It maintains state across calls to handle tokens split across chunks.
func (s *scanner) feed(p []byte) []string {
	var tokens []string

	for _, b := range p {
		if s.depth > 0 {
			// Inside brackets — keep accumulating regardless of
			// what the byte is (arbitrary values can contain
			// almost anything: spaces map to _, but we might see
			// commas, parens, etc.)
			switch b {
			case '[':
				s.depth++
				s.buf = append(s.buf, b)
			case ']':
				s.depth--
				s.buf = append(s.buf, b)
			default:
				s.buf = append(s.buf, b)
			}
			continue
		}

		if isTokenByte(b) {
			if b == '[' {
				s.depth++
			}
			s.buf = append(s.buf, b)
			continue
		}

		// Delimiter byte — flush current token.
		if len(s.buf) > 0 {
			if tok := s.accept(); tok != "" {
				tokens = append(tokens, tok)
			}
			s.buf = s.buf[:0]
		}
	}

	return tokens
}

// flush returns the final buffered token (if any) and resets state.
func (s *scanner) flush() string {
	if len(s.buf) == 0 {
		return ""
	}
	tok := s.accept()
	s.buf = s.buf[:0]
	s.depth = 0
	return tok
}

// accept examines the buffered bytes and returns a candidate class
// name, or "" if the token doesn't look like a Tailwind class.
func (s *scanner) accept() string {
	tok := string(s.buf)

	// Quick rejections for obviously non-class tokens.
	if len(tok) == 0 {
		return ""
	}

	// Must start with a letter, !, -, [ (for arbitrary properties), @ (for container variants),
	// or * (for child/descendant selector variants).
	first := tok[0]
	if !isLetter(first) && first != '!' && first != '-' && first != '[' && first != '@' && first != '*' {
		return ""
	}

	// Reject if it's just punctuation / too short to be meaningful.
	if len(tok) == 1 && !isLetter(first) {
		return ""
	}

	// Reject common non-class patterns:
	// - Pure numbers after a letter prefix (e.g., "x0a2b3" — hex-like)
	// - Strings that contain '=' (attributes like href=...)
	// - Strings starting with '//' or '/*' (comments)
	// - Strings that look like URLs (contain "://")
	for i := 0; i < len(tok)-2; i++ {
		if tok[i] == ':' && tok[i+1] == '/' && tok[i+2] == '/' {
			return "" // URL
		}
	}

	// Reject tokens with unbalanced brackets.
	depth := 0
	for _, c := range tok {
		if c == '[' {
			depth++
		} else if c == ']' {
			depth--
		}
		if depth < 0 {
			return ""
		}
	}
	if depth != 0 {
		return ""
	}

	return tok
}

// isTokenByte returns true if b can appear in a Tailwind class name.
// This includes:
//   - Letters (a-z, A-Z)
//   - Digits (0-9)
//   - Hyphen (-), underscore (_)
//   - Colon (:) — variant separator
//   - Square brackets ([ ]) — arbitrary values
//   - Forward slash (/) — fractions and opacity modifiers
//   - Dot (.) — decimal values in arbitrary
//   - Hash (#) — hex colors in arbitrary
//   - Exclamation (!) — important modifier
//   - Percent (%) — percentage values
//   - Parentheses — inside arbitrary values (e.g., calc())
//   - Comma — inside arbitrary values
//   - Plus, asterisk — inside arbitrary values
//   - At (@) — for @-prefixed variants
func isTokenByte(b byte) bool {
	switch {
	case b >= 'a' && b <= 'z':
		return true
	case b >= 'A' && b <= 'Z':
		return true
	case b >= '0' && b <= '9':
		return true
	case b == '-', b == '_', b == ':', b == '[', b == ']',
		b == '/', b == '.', b == '#', b == '!', b == '%',
		b == '(', b == ')', b == ',', b == '+', b == '*',
		b == '@':
		return true
	}
	return false
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
