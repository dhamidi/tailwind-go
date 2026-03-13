package tailwind

import (
	"strings"
	"testing"
)

func TestContentArbitrary(t *testing.T) {
	tests := []struct {
		class    string
		expected string
	}{
		{"content-['']", "''"},
		{"content-[attr(data-label)]", "attr(data-label)"},
		{"content-['hello_world']", "'hello world'"},
		{"content-none", "none"},
	}

	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(tt.class))
			css := e.CSS()
			if !strings.Contains(css, tt.expected) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tt.class, tt.expected, css)
			}
		})
	}
}
