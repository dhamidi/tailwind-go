package minify

// CSS minifies a CSS string by removing unnecessary whitespace and comments.
// It preserves content inside quoted strings and url() values.
func CSS(input string) string {
	if len(input) == 0 {
		return ""
	}

	out := make([]byte, 0, len(input))
	i := 0
	n := len(input)

	for i < n {
		// Check for comments: /* ... */
		if i+1 < n && input[i] == '/' && input[i+1] == '*' {
			i += 2
			for i+1 < n {
				if input[i] == '*' && input[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			// Skip any trailing whitespace/comments after this comment
			i = skipWhitespaceAndComments(input, i)
			// Emit a space if needed between non-punctuation tokens
			if len(out) > 0 && i < n {
				prev := out[len(out)-1]
				next := input[i]
				if !isPunctuation(prev) && !isPunctuation(next) {
					out = append(out, ' ')
				}
			}
			continue
		}

		// Check for quoted strings
		if input[i] == '"' || input[i] == '\'' {
			quote := input[i]
			out = append(out, quote)
			i++
			for i < n && input[i] != quote {
				if input[i] == '\\' && i+1 < n {
					out = append(out, input[i], input[i+1])
					i += 2
					continue
				}
				out = append(out, input[i])
				i++
			}
			if i < n {
				out = append(out, quote)
				i++
			}
			continue
		}

		// Check for url(...)
		if i+3 < n && input[i] == 'u' && input[i+1] == 'r' && input[i+2] == 'l' && input[i+3] == '(' {
			out = append(out, 'u', 'r', 'l', '(')
			i += 4
			// Skip whitespace after (
			for i < n && isWhitespace(input[i]) {
				i++
			}
			// Check if url content is quoted
			if i < n && (input[i] == '"' || input[i] == '\'') {
				quote := input[i]
				out = append(out, quote)
				i++
				for i < n && input[i] != quote {
					if input[i] == '\\' && i+1 < n {
						out = append(out, input[i], input[i+1])
						i += 2
						continue
					}
					out = append(out, input[i])
					i++
				}
				if i < n {
					out = append(out, quote)
					i++
				}
			} else {
				// Unquoted url content - collect until )
				for i < n && input[i] != ')' {
					if !isWhitespace(input[i]) {
						out = append(out, input[i])
					}
					i++
				}
			}
			// Skip whitespace before )
			// The ) will be handled below or we append it
			if i < n && input[i] == ')' {
				out = append(out, ')')
				i++
			}
			continue
		}

		// Whitespace handling
		if isWhitespace(input[i]) {
			// Skip whitespace and any comments that follow
			i = skipWhitespaceAndComments(input, i)
			// Decide whether to emit a space based on context
			if len(out) > 0 && i < n {
				prev := out[len(out)-1]
				next := input[i]
				if !isPunctuation(prev) && !isPunctuation(next) {
					out = append(out, ' ')
				}
			}
			continue
		}

		// Remove space before punctuation that was already emitted
		// (handled above by not emitting space)

		// Remove trailing semicolons before }
		if input[i] == ';' {
			// Look ahead past whitespace and comments for }
			j := i + 1
			for j < n {
				if isWhitespace(input[j]) {
					j++
				} else if j+1 < n && input[j] == '/' && input[j+1] == '*' {
					j += 2
					for j+1 < n {
						if input[j] == '*' && input[j+1] == '/' {
							j += 2
							break
						}
						j++
					}
				} else {
					break
				}
			}
			if j < n && input[j] == '}' {
				// Skip the semicolon
				i++
				continue
			}
		}

		out = append(out, input[i])
		i++
	}

	// Trim leading/trailing whitespace
	start := 0
	for start < len(out) && isWhitespace(out[start]) {
		start++
	}
	end := len(out)
	for end > start && isWhitespace(out[end-1]) {
		end--
	}

	return string(out[start:end])
}

// skipWhitespaceAndComments advances past any sequence of whitespace and comments.
func skipWhitespaceAndComments(input string, i int) int {
	n := len(input)
	for i < n {
		if isWhitespace(input[i]) {
			i++
		} else if i+1 < n && input[i] == '/' && input[i+1] == '*' {
			i += 2
			for i+1 < n {
				if input[i] == '*' && input[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
		} else {
			break
		}
	}
	return i
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isPunctuation(b byte) bool {
	return b == '{' || b == '}' || b == ':' || b == ';' || b == ',' || b == '(' || b == ')'
}
