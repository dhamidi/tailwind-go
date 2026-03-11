package tailwind

import (
	"strings"
)

// parseStylesheet parses a Tailwind v4 CSS source into a Stylesheet.
func parseStylesheet(css []byte) (*Stylesheet, error) {
	tokens := newTokenizer(css).tokenize()
	p := &parser{tokens: tokens}
	return p.parse()
}

type parser struct {
	tokens []token
	pos    int
	order  int // monotonic counter for definition ordering
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) advance() token {
	tok := p.peek()
	if tok.typ != tokEOF {
		p.pos++
	}
	return tok
}

func (p *parser) skipWhitespace() {
	for p.peek().typ == tokWhitespace {
		p.advance()
	}
}

func (p *parser) expect(typ tokenType) token {
	p.skipWhitespace()
	tok := p.advance()
	return tok // caller checks typ if needed
}

func (p *parser) parse() (*Stylesheet, error) {
	ss := &Stylesheet{
		Theme: ThemeConfig{Tokens: make(map[string]string)},
	}

	for p.peek().typ != tokEOF {
		p.skipWhitespace()
		tok := p.peek()

		if tok.typ == tokAtKeyword {
			switch tok.value {
			case "@theme":
				p.parseTheme(ss)
			case "@utility":
				p.parseUtility(ss)
			case "@variant":
				p.parseVariant(ss)
			case "@keyframes":
				p.parseKeyframes(ss)
			default:
				// Skip unknown at-rules (e.g., @layer, @import, @tailwind).
				p.skipAtRule()
			}
		} else if tok.typ == tokEOF {
			break
		} else {
			// Skip qualified rules (regular CSS rules).
			p.skipRule()
		}
	}

	return ss, nil
}

// parseTheme parses: @theme { --prop: value; ... }
func (p *parser) parseTheme(ss *Stylesheet) {
	p.advance() // consume @theme
	p.skipWhitespace()

	// Optional "inline" or other keywords before the block.
	for p.peek().typ == tokIdent {
		p.advance()
		p.skipWhitespace()
	}

	if p.peek().typ != tokBraceOpen {
		return
	}
	p.advance() // consume {

	for p.peek().typ != tokBraceClose && p.peek().typ != tokEOF {
		p.skipWhitespace()

		if p.peek().typ == tokBraceClose {
			break
		}

		// Expect: --property-name: value;
		prop, val := p.parseDeclaration()
		if prop != "" {
			ss.Theme.Tokens[prop] = val
		}
	}

	if p.peek().typ == tokBraceClose {
		p.advance() // consume }
	}
}

// parseUtility parses: @utility name-* { declarations }
// or: @utility name { declarations } (static utility)
func (p *parser) parseUtility(ss *Stylesheet) {
	p.advance() // consume @utility
	p.skipWhitespace()

	// Consume the pattern name (may include hyphens and *).
	pattern := p.consumeUtilityPattern()
	p.skipWhitespace()

	static := true
	if strings.HasSuffix(pattern, "-*") {
		pattern = strings.TrimSuffix(pattern, "-*")
		static = false
	} else if pattern == "*" {
		// Universal utility — not common, skip.
		p.skipBlock()
		return
	}

	if p.peek().typ != tokBraceOpen {
		return
	}
	p.advance() // consume {

	var decls []Declaration
	for p.peek().typ != tokBraceClose && p.peek().typ != tokEOF {
		p.skipWhitespace()
		if p.peek().typ == tokBraceClose {
			break
		}
		prop, val := p.parseDeclaration()
		if prop != "" {
			decls = append(decls, Declaration{Property: prop, Value: val})
		}
	}

	if p.peek().typ == tokBraceClose {
		p.advance() // consume }
	}

	p.order++
	ss.Utilities = append(ss.Utilities, &UtilityDef{
		Pattern:      pattern,
		Static:       static,
		Declarations: decls,
		Order:        p.order,
	})
}

// parseVariant parses: @variant name (&:selector);
// or: @variant name (@media ...);
func (p *parser) parseVariant(ss *Stylesheet) {
	p.advance() // consume @variant
	p.skipWhitespace()

	name := p.consumeIdentValue()
	p.skipWhitespace()

	v := &VariantDef{Name: name}
	p.order++
	v.Order = p.order

	// The rest until ; or { is the variant definition.
	// Could be (&:hover), (@media ...), or a block.
	if p.peek().typ == tokParenOpen {
		content := p.consumeParenContent()
		content = strings.TrimSpace(content)

		if strings.HasPrefix(content, "@media") {
			v.Media = strings.TrimPrefix(content, "@media ")
		} else if strings.HasPrefix(content, "@supports") {
			v.AtRule = "supports"
			v.Media = strings.TrimPrefix(content, "@supports ")
		} else if strings.HasPrefix(content, "@container") {
			v.AtRule = "container"
			v.Media = strings.TrimPrefix(content, "@container ")
		} else {
			v.Selector = content
		}
	}

	// Consume trailing semicolon if present.
	p.skipWhitespace()
	if p.peek().typ == tokSemicolon {
		p.advance()
	}

	// Or it might have a block body.
	if p.peek().typ == tokBraceOpen {
		p.skipBlock()
	}

	ss.Variants = append(ss.Variants, v)
}

// parseDeclaration parses: property: value;
func (p *parser) parseDeclaration() (string, string) {
	p.skipWhitespace()

	// Consume property (may be a custom property like --spacing).
	var propParts []string
	for p.peek().typ != tokColon && p.peek().typ != tokBraceClose &&
		p.peek().typ != tokEOF && p.peek().typ != tokSemicolon {
		tok := p.advance()
		propParts = append(propParts, tok.value)
	}

	if p.peek().typ != tokColon {
		return "", ""
	}
	p.advance() // consume :

	// Consume value until ; or }
	var valParts []string
	depth := 0
	for {
		tok := p.peek()
		if tok.typ == tokEOF {
			break
		}
		if depth == 0 && (tok.typ == tokSemicolon || tok.typ == tokBraceClose) {
			break
		}
		if tok.typ == tokParenOpen || tok.typ == tokFunction {
			depth++
		} else if tok.typ == tokParenClose {
			depth--
		}
		valParts = append(valParts, tok.value)
		p.advance()
	}

	if p.peek().typ == tokSemicolon {
		p.advance() // consume ;
	}

	prop := strings.TrimSpace(strings.Join(propParts, ""))
	val := strings.TrimSpace(strings.Join(valParts, ""))

	return prop, val
}

// consumeUtilityPattern reads the utility pattern after @utility.
// e.g., "w-*", "bg-*", "translate-x-*", "flex", "block"
func (p *parser) consumeUtilityPattern() string {
	var parts []string
	for {
		tok := p.peek()
		if tok.typ == tokWhitespace || tok.typ == tokBraceOpen || tok.typ == tokEOF {
			break
		}
		parts = append(parts, tok.value)
		p.advance()
	}
	return strings.Join(parts, "")
}

// consumeIdentValue reads a single ident-like value (may contain hyphens).
func (p *parser) consumeIdentValue() string {
	var parts []string
	for {
		tok := p.peek()
		if tok.typ != tokIdent && tok.typ != tokDelim && tok.typ != tokNumber {
			break
		}
		// Stop on whitespace-like delimiters
		if tok.typ == tokDelim && tok.value != "-" && tok.value != "_" {
			break
		}
		parts = append(parts, tok.value)
		p.advance()
	}
	return strings.Join(parts, "")
}

// consumeParenContent reads everything inside (...), including nested parens.
func (p *parser) consumeParenContent() string {
	if p.peek().typ != tokParenOpen {
		return ""
	}
	p.advance() // consume (

	var parts []string
	depth := 1
	for depth > 0 && p.peek().typ != tokEOF {
		tok := p.advance()
		if tok.typ == tokParenOpen || tok.typ == tokFunction {
			depth++
			parts = append(parts, tok.value)
		} else if tok.typ == tokParenClose {
			depth--
			if depth > 0 {
				parts = append(parts, tok.value)
			}
		} else {
			parts = append(parts, tok.value)
		}
	}

	return strings.Join(parts, "")
}

// skipBlock skips a { ... } block, handling nesting.
func (p *parser) skipBlock() {
	if p.peek().typ != tokBraceOpen {
		return
	}
	p.advance()
	depth := 1
	for depth > 0 && p.peek().typ != tokEOF {
		tok := p.advance()
		if tok.typ == tokBraceOpen {
			depth++
		} else if tok.typ == tokBraceClose {
			depth--
		}
	}
}

// skipAtRule skips an at-rule (either with a block or semicolon-terminated).
func (p *parser) skipAtRule() {
	p.advance() // consume @keyword
	for p.peek().typ != tokEOF {
		tok := p.peek()
		if tok.typ == tokSemicolon {
			p.advance()
			return
		}
		if tok.typ == tokBraceOpen {
			p.skipBlock()
			return
		}
		p.advance()
	}
}

// skipRule skips a qualified rule (selector + block).
func (p *parser) skipRule() {
	for p.peek().typ != tokEOF {
		tok := p.peek()
		if tok.typ == tokBraceOpen {
			p.skipBlock()
			return
		}
		if tok.typ == tokSemicolon {
			p.advance()
			return
		}
		p.advance()
	}
}

// parseKeyframes parses: @keyframes name { ... }
func (p *parser) parseKeyframes(ss *Stylesheet) {
	p.advance() // consume @keyframes
	p.skipWhitespace()

	name := p.consumeIdentValue()
	if name == "" {
		p.skipAtRule()
		return
	}
	p.skipWhitespace()

	if p.peek().typ != tokBraceOpen {
		return
	}

	// Capture the block by serializing tokens.
	start := p.pos
	p.skipBlock()
	body := p.serializeTokens(start, p.pos)

	ss.Keyframes = append(ss.Keyframes, &KeyframesRule{
		Name: name,
		Body: "@keyframes " + name + " " + body,
	})
}

// serializeTokens reconstructs CSS text from a range of tokens.
func (p *parser) serializeTokens(start, end int) string {
	var sb strings.Builder
	for i := start; i < end && i < len(p.tokens); i++ {
		sb.WriteString(p.tokens[i].value)
	}
	return sb.String()
}
