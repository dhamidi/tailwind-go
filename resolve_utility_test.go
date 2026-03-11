package tailwind

import (
	"strings"
	"testing"
)

func TestResolveUtilityColorClass(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; --color-red-400: #f87171; }
@utility bg-* { background-color: --value(--color); }
@utility text-* { color: --value(--color); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-blue-500 text-red-400"`))
	result := e.CSS()
	if !strings.Contains(result, "#3b82f6") {
		t.Errorf("bg-blue-500 not resolved: %s", result)
	}
	if !strings.Contains(result, "#f87171") {
		t.Errorf("text-red-400 not resolved: %s", result)
	}
}

func TestResolveUtilityBorderVariants(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility border-* { border-width: --value(--spacing); }
@utility border-t-* { border-top-width: --value(--spacing); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="border-2 border-t-4"`))
	result := e.CSS()
	if !strings.Contains(result, "border-width") {
		t.Errorf("border-2 not resolved: %s", result)
	}
	if !strings.Contains(result, "border-top-width") {
		t.Errorf("border-t-4 not resolved: %s", result)
	}
}
