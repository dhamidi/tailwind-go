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

func TestSpaceXChildSelector(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility space-x-* > :not(:last-child) {
  --tw-space-x-reverse: 0;
  margin-inline-end: calc(--value(--spacing) * var(--tw-space-x-reverse));
  margin-inline-start: calc(--value(--spacing) * calc(1 - var(--tw-space-x-reverse)));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="space-x-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	// Must target children, not the parent element.
	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-x-4 should use child selector > :not(:last-child): %s", result)
	}
	// Must use logical properties.
	if !strings.Contains(result, "margin-inline-start") {
		t.Errorf("space-x-4 should use margin-inline-start: %s", result)
	}
	if !strings.Contains(result, "margin-inline-end") {
		t.Errorf("space-x-4 should use margin-inline-end: %s", result)
	}
	// Must NOT use margin-left (old behavior).
	if strings.Contains(result, "margin-left") {
		t.Errorf("space-x-4 should NOT use margin-left: %s", result)
	}
	// Must include the reverse variable declaration.
	if !strings.Contains(result, "--tw-space-x-reverse: 0") {
		t.Errorf("space-x-4 should include --tw-space-x-reverse: 0: %s", result)
	}
}

func TestSpaceYChildSelector(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility space-y-* > :not(:last-child) {
  --tw-space-y-reverse: 0;
  margin-block-end: calc(--value(--spacing) * var(--tw-space-y-reverse));
  margin-block-start: calc(--value(--spacing) * calc(1 - var(--tw-space-y-reverse)));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="space-y-2"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-y-2 should use child selector: %s", result)
	}
	if !strings.Contains(result, "margin-block-start") {
		t.Errorf("space-y-2 should use margin-block-start: %s", result)
	}
	if !strings.Contains(result, "margin-block-end") {
		t.Errorf("space-y-2 should use margin-block-end: %s", result)
	}
	if strings.Contains(result, "margin-top") {
		t.Errorf("space-y-2 should NOT use margin-top: %s", result)
	}
}

func TestSpaceXReverseChildSelector(t *testing.T) {
	css := []byte(`
@utility space-x-reverse > :not(:last-child) { --tw-space-x-reverse: 1; }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="space-x-reverse"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-x-reverse should use child selector: %s", result)
	}
	if !strings.Contains(result, "--tw-space-x-reverse: 1") {
		t.Errorf("space-x-reverse should set --tw-space-x-reverse: 1: %s", result)
	}
}

func TestSpaceXWithEmbeddedCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="space-x-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("space-x-4 should use child selector: %s", result)
	}
	if !strings.Contains(result, "margin-inline-start") {
		t.Errorf("space-x-4 should use margin-inline-start: %s", result)
	}
	if strings.Contains(result, "margin-left") {
		t.Errorf("space-x-4 should NOT use margin-left: %s", result)
	}
}

func TestDivideXChildSelector(t *testing.T) {
	css := []byte(`
@utility divide-x > :not(:last-child) {
  --tw-divide-x-reverse: 0;
  border-inline-start-width: calc(1px * calc(1 - var(--tw-divide-x-reverse)));
  border-inline-end-width: calc(1px * var(--tw-divide-x-reverse));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="divide-x"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-x should use child selector > :not(:last-child): %s", result)
	}
	if !strings.Contains(result, "border-inline-start-width") {
		t.Errorf("divide-x should use border-inline-start-width: %s", result)
	}
	if !strings.Contains(result, "border-inline-end-width") {
		t.Errorf("divide-x should use border-inline-end-width: %s", result)
	}
	if strings.Contains(result, "border-left-width") {
		t.Errorf("divide-x should NOT use border-left-width: %s", result)
	}
	if !strings.Contains(result, "--tw-divide-x-reverse: 0") {
		t.Errorf("divide-x should include --tw-divide-x-reverse: 0: %s", result)
	}
}

func TestDivideYChildSelector(t *testing.T) {
	css := []byte(`
@utility divide-y > :not(:last-child) {
  --tw-divide-y-reverse: 0;
  border-top-width: calc(1px * calc(1 - var(--tw-divide-y-reverse)));
  border-bottom-width: calc(1px * var(--tw-divide-y-reverse));
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="divide-y"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-y should use child selector: %s", result)
	}
	if !strings.Contains(result, "border-top-width") {
		t.Errorf("divide-y should use border-top-width: %s", result)
	}
	if !strings.Contains(result, "border-bottom-width") {
		t.Errorf("divide-y should use border-bottom-width: %s", result)
	}
	if !strings.Contains(result, "--tw-divide-y-reverse: 0") {
		t.Errorf("divide-y should include --tw-divide-y-reverse: 0: %s", result)
	}
}

func TestDivideXReverseChildSelector(t *testing.T) {
	css := []byte(`
@utility divide-x-reverse > :not(:last-child) { --tw-divide-x-reverse: 1; }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="divide-x-reverse"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-x-reverse should use child selector: %s", result)
	}
	if !strings.Contains(result, "--tw-divide-x-reverse: 1") {
		t.Errorf("divide-x-reverse should set --tw-divide-x-reverse: 1: %s", result)
	}
}

func TestDivideXWithEmbeddedCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="divide-x"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-x should use child selector: %s", result)
	}
	if !strings.Contains(result, "border-inline-start-width") {
		t.Errorf("divide-x should use border-inline-start-width: %s", result)
	}
	if strings.Contains(result, "border-left-width") {
		t.Errorf("divide-x should NOT use border-left-width: %s", result)
	}
}

func TestDivideYWithEmbeddedCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`class="divide-y-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "> :not(:last-child)") {
		t.Errorf("divide-y-4 should use child selector: %s", result)
	}
	if !strings.Contains(result, "border-top-width") {
		t.Errorf("divide-y-4 should use border-top-width: %s", result)
	}
	if !strings.Contains(result, "border-bottom-width") {
		t.Errorf("divide-y-4 should use border-bottom-width: %s", result)
	}
}

func TestGridColsRepeat(t *testing.T) {
	e := New()
	tests := []struct {
		class string
		want  string
	}{
		{"grid-cols-1", "repeat(1, minmax(0, 1fr))"},
		{"grid-cols-3", "repeat(3, minmax(0, 1fr))"},
		{"grid-cols-12", "repeat(12, minmax(0, 1fr))"},
		{"grid-cols-none", "none"},
		{"grid-cols-subgrid", "subgrid"},
		{"grid-cols-[1fr_auto_2fr]", "1fr auto 2fr"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			if !strings.Contains(result, tt.want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tt.class, tt.want, result)
			}
			if strings.Contains(result, "grid-template-columns") != true {
				t.Errorf("%s: expected grid-template-columns property in CSS output:\n%s", tt.class, result)
			}
		})
	}
}

func TestGridRowsRepeat(t *testing.T) {
	e := New()
	tests := []struct {
		class string
		want  string
	}{
		{"grid-rows-2", "repeat(2, minmax(0, 1fr))"},
		{"grid-rows-4", "repeat(4, minmax(0, 1fr))"},
		{"grid-rows-none", "none"},
		{"grid-rows-subgrid", "subgrid"},
	}
	for _, tt := range tests {
		t.Run(tt.class, func(t *testing.T) {
			e.Write([]byte(`class="` + tt.class + `"`))
			result := e.CSS()
			if !strings.Contains(result, tt.want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tt.class, tt.want, result)
			}
			if !strings.Contains(result, "grid-template-rows") {
				t.Errorf("%s: expected grid-template-rows property in CSS output:\n%s", tt.class, result)
			}
		})
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
