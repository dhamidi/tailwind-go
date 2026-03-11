package tailwind

import (
	"fmt"
	"strings"
)

// CSS token types, loosely following the CSS Tokenization specification.
// We only implement what's needed to parse Tailwind's v4 CSS dialect.
type tokenType int

const (
	tokEOF tokenType = iota
	tokWhitespace
	tokIdent       // e.g., color, width, hover
	tokAtKeyword   // e.g., @theme, @utility, @variant
	tokHash        // e.g., #fff
	tokString      // "..." or '...'
	tokNumber      // 42, 3.14
	tokDimension   // 48rem, 0.25rem, 100%
	tokPercentage  // 50%
	tokFunction    // calc(, var(, --value(
	tokColon       // :
	tokSemicolon   // ;
	tokComma       // ,
	tokBraceOpen   // {
	tokBraceClose  // }
	tokBracketOpen // [
	tokBracketClose // ]
	tokParenOpen   // (
	tokParenClose  // )
	tokDelim       // any single character not matched above
	tokCDC         // -->
	tokCDO         // <!--
	tokComment     // /* ... */
)

type token struct {
	typ   tokenType
	value string
}

func (t token) String() string {
	return fmt.Sprintf("{%d %q}", t.typ, t.value)
}

// tokenizer converts raw CSS bytes into a stream of tokens.
type tokenizer struct {
	src []byte
	pos int
}

func newTokenizer(src []byte) *tokenizer {
	return &tokenizer{src: src}
}

func (t *tokenizer) peek() byte {
	if t.pos >= len(t.src) {
		return 0
	}
	return t.src[t.pos]
}

func (t *tokenizer) peekAt(offset int) byte {
	i := t.pos + offset
	if i >= len(t.src) {
		return 0
	}
	return t.src[i]
}

func (t *tokenizer) advance() byte {
	if t.pos >= len(t.src) {
		return 0
	}
	b := t.src[t.pos]
	t.pos++
	return b
}

func (t *tokenizer) remaining() int {
	return len(t.src) - t.pos
}

// tokenize returns all tokens from the source.
func (t *tokenizer) tokenize() []token {
	var tokens []token
	for {
		tok := t.next()
		if tok.typ == tokEOF {
			break
		}
		// Skip comments and whitespace for our purposes —
		// the parser doesn't need them.
		if tok.typ == tokComment {
			continue
		}
		tokens = append(tokens, tok)
	}
	return tokens
}

func (t *tokenizer) next() token {
	if t.pos >= len(t.src) {
		return token{typ: tokEOF}
	}

	b := t.peek()

	// Whitespace
	if isWSByte(b) {
		return t.consumeWhitespace()
	}

	// Comments
	if b == '/' && t.peekAt(1) == '*' {
		return t.consumeComment()
	}

	// Strings
	if b == '"' || b == '\'' {
		return t.consumeString()
	}

	// Numbers (and dimensions/percentages)
	if isDigit(b) || (b == '.' && isDigit(t.peekAt(1))) ||
		(b == '-' && (isDigit(t.peekAt(1)) || t.peekAt(1) == '.')) ||
		(b == '+' && (isDigit(t.peekAt(1)) || t.peekAt(1) == '.')) {
		// Check if this is actually the start of an ident (e.g., -webkit-)
		if b == '-' && !isDigit(t.peekAt(1)) && t.peekAt(1) != '.' {
			return t.consumeIdent()
		}
		return t.consumeNumeric()
	}

	// At-keyword
	if b == '@' {
		t.advance()
		if isIdentStart(t.peek()) {
			name := t.consumeIdentChars()
			return token{typ: tokAtKeyword, value: "@" + name}
		}
		return token{typ: tokDelim, value: "@"}
	}

	// Hash
	if b == '#' {
		t.advance()
		name := t.consumeIdentChars()
		if name == "" {
			// Also try to consume hex digits
			name = t.consumeHexChars()
		}
		return token{typ: tokHash, value: "#" + name}
	}

	// Ident or function
	if isIdentStart(b) || (b == '-' && isIdentStart(t.peekAt(1))) || (b == '-' && t.peekAt(1) == '-') {
		return t.consumeIdent()
	}

	// Simple single-character tokens
	t.advance()
	switch b {
	case ':':
		return token{typ: tokColon, value: ":"}
	case ';':
		return token{typ: tokSemicolon, value: ";"}
	case ',':
		return token{typ: tokComma, value: ","}
	case '{':
		return token{typ: tokBraceOpen, value: "{"}
	case '}':
		return token{typ: tokBraceClose, value: "}"}
	case '[':
		return token{typ: tokBracketOpen, value: "["}
	case ']':
		return token{typ: tokBracketClose, value: "]"}
	case '(':
		return token{typ: tokParenOpen, value: "("}
	case ')':
		return token{typ: tokParenClose, value: ")"}
	default:
		return token{typ: tokDelim, value: string(b)}
	}
}

func (t *tokenizer) consumeWhitespace() token {
	start := t.pos
	for t.pos < len(t.src) && isWSByte(t.src[t.pos]) {
		t.pos++
	}
	return token{typ: tokWhitespace, value: string(t.src[start:t.pos])}
}

func (t *tokenizer) consumeComment() token {
	start := t.pos
	t.pos += 2 // skip /*
	for t.pos < len(t.src)-1 {
		if t.src[t.pos] == '*' && t.src[t.pos+1] == '/' {
			t.pos += 2
			return token{typ: tokComment, value: string(t.src[start:t.pos])}
		}
		t.pos++
	}
	t.pos = len(t.src) // unterminated comment
	return token{typ: tokComment, value: string(t.src[start:t.pos])}
}

func (t *tokenizer) consumeString() token {
	quote := t.advance()
	var sb strings.Builder
	sb.WriteByte(quote)
	for t.pos < len(t.src) {
		b := t.advance()
		sb.WriteByte(b)
		if b == quote {
			break
		}
		if b == '\\' && t.pos < len(t.src) {
			sb.WriteByte(t.advance())
		}
	}
	return token{typ: tokString, value: sb.String()}
}

func (t *tokenizer) consumeNumeric() token {
	start := t.pos

	// Optional sign
	if t.peek() == '+' || t.peek() == '-' {
		t.advance()
	}

	// Integer part
	for t.pos < len(t.src) && isDigit(t.src[t.pos]) {
		t.pos++
	}

	// Fractional part
	if t.pos < len(t.src)-1 && t.src[t.pos] == '.' && isDigit(t.src[t.pos+1]) {
		t.pos++ // skip .
		for t.pos < len(t.src) && isDigit(t.src[t.pos]) {
			t.pos++
		}
	}

	numStr := string(t.src[start:t.pos])

	// Check for percentage
	if t.peek() == '%' {
		t.advance()
		return token{typ: tokPercentage, value: numStr + "%"}
	}

	// Check for dimension (unit suffix)
	if isIdentStart(t.peek()) {
		unit := t.consumeIdentChars()
		return token{typ: tokDimension, value: numStr + unit}
	}

	return token{typ: tokNumber, value: numStr}
}

func (t *tokenizer) consumeIdent() token {
	name := t.consumeIdentChars()

	// If followed by '(', it's a function token.
	if t.peek() == '(' {
		t.advance()
		return token{typ: tokFunction, value: name + "("}
	}

	return token{typ: tokIdent, value: name}
}

func (t *tokenizer) consumeIdentChars() string {
	start := t.pos
	for t.pos < len(t.src) && isIdentChar(t.src[t.pos]) {
		t.pos++
	}
	return string(t.src[start:t.pos])
}

func (t *tokenizer) consumeHexChars() string {
	start := t.pos
	for t.pos < len(t.src) && isHexDigit(t.src[t.pos]) {
		t.pos++
	}
	return string(t.src[start:t.pos])
}

func isWSByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

func isIdentStart(b byte) bool {
	return isLetter(b) || b == '_' || b == '-'
}

func isIdentChar(b byte) bool {
	return isLetter(b) || isDigit(b) || b == '_' || b == '-'
}

func isHexDigit(b byte) bool {
	return isDigit(b) || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
