package main

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_Headings(t *testing.T) {
	tests := []struct {
		input    string
		contains string
	}{
		{"# Hello", `<h1 class="text-4xl font-black tracking-tight mb-6 text-gray-900 dark:text-white">Hello</h1>`},
		{"## World", `<h2 class="text-3xl font-extrabold tracking-tight mt-10 mb-4 text-gray-900 dark:text-white">World</h2>`},
		{"### Sub", `<h3 class="text-2xl font-bold tracking-tight mt-8 mb-4 text-gray-900 dark:text-white">Sub</h3>`},
	}
	for _, tt := range tests {
		got := renderMarkdown(tt.input)
		if !strings.Contains(got, tt.contains) {
			t.Errorf("renderMarkdown(%q)\ngot:  %s\nwant to contain: %s", tt.input, got, tt.contains)
		}
	}
}

func TestRenderMarkdown_Paragraph(t *testing.T) {
	got := renderMarkdown("Hello world")
	if !strings.Contains(got, `<p class="text-lg leading-relaxed text-gray-700 dark:text-gray-300 mb-4">Hello world</p>`) {
		t.Errorf("expected paragraph, got: %s", got)
	}
}

func TestRenderMarkdown_Bold(t *testing.T) {
	got := renderMarkdown("This is **bold** text")
	if !strings.Contains(got, `<strong class="font-bold text-gray-900 dark:text-white">bold</strong>`) {
		t.Errorf("expected bold, got: %s", got)
	}
}

func TestRenderMarkdown_Italic(t *testing.T) {
	got := renderMarkdown("This is *italic* text")
	if !strings.Contains(got, `<em class="italic">italic</em>`) {
		t.Errorf("expected italic, got: %s", got)
	}
}

func TestRenderMarkdown_Underline(t *testing.T) {
	got := renderMarkdown("This is __underlined__ text")
	if !strings.Contains(got, `<u class="underline`) {
		t.Errorf("expected underline, got: %s", got)
	}
}

func TestRenderMarkdown_Strikethrough(t *testing.T) {
	got := renderMarkdown("This is ~~deleted~~ text")
	if !strings.Contains(got, `<del class="line-through`) {
		t.Errorf("expected strikethrough, got: %s", got)
	}
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	got := renderMarkdown("Use `fmt.Println` here")
	if !strings.Contains(got, `<code class="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono`) {
		t.Errorf("expected inline code, got: %s", got)
	}
	if !strings.Contains(got, "fmt.Println") {
		t.Errorf("expected code content, got: %s", got)
	}
}

func TestRenderMarkdown_UnorderedList(t *testing.T) {
	input := "- First\n- Second\n- Third"
	got := renderMarkdown(input)
	if !strings.Contains(got, "<ul") {
		t.Errorf("expected ul, got: %s", got)
	}
	if strings.Count(got, "<li") != 3 {
		t.Errorf("expected 3 li elements, got: %s", got)
	}
}

func TestRenderMarkdown_OrderedList(t *testing.T) {
	input := "1. First\n2. Second\n3. Third"
	got := renderMarkdown(input)
	if !strings.Contains(got, "<ol") {
		t.Errorf("expected ol, got: %s", got)
	}
	if strings.Count(got, "<li") != 3 {
		t.Errorf("expected 3 li elements, got: %s", got)
	}
}

func TestRenderMarkdown_HorizontalRule(t *testing.T) {
	got := renderMarkdown("---")
	if !strings.Contains(got, "<hr") {
		t.Errorf("expected hr, got: %s", got)
	}
}

func TestRenderMarkdown_MultipleElements(t *testing.T) {
	input := `# Title

Some **bold** paragraph.

- Item one
- Item two

---

## Section Two

1. First
2. Second`

	got := renderMarkdown(input)
	if !strings.Contains(got, "<h1") {
		t.Errorf("missing h1")
	}
	if !strings.Contains(got, "<strong") {
		t.Errorf("missing strong")
	}
	if !strings.Contains(got, "<ul") {
		t.Errorf("missing ul")
	}
	if !strings.Contains(got, "<hr") {
		t.Errorf("missing hr")
	}
	if !strings.Contains(got, "<h2") {
		t.Errorf("missing h2")
	}
	if !strings.Contains(got, "<ol") {
		t.Errorf("missing ol")
	}
}

func TestRenderMarkdown_EmptyInput(t *testing.T) {
	got := renderMarkdown("")
	if got != "" {
		t.Errorf("expected empty output for empty input, got: %s", got)
	}
}

func TestRenderMarkdown_MultiLineParagraph(t *testing.T) {
	input := "Line one\nline two\nline three"
	got := renderMarkdown(input)
	if !strings.Contains(got, "Line one\nline two\nline three") {
		t.Errorf("expected multi-line paragraph content, got: %s", got)
	}
	if strings.Count(got, "<p") != 1 {
		t.Errorf("expected single paragraph, got: %s", got)
	}
}
