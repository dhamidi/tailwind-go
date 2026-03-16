package main

import (
	"regexp"
	"strings"
)

// renderMarkdown converts a subset of markdown to HTML with Tailwind CSS classes.
// It supports headings, paragraphs, bold, italic, underline, strikethrough,
// inline code, unordered lists, ordered lists, and horizontal rules.
func renderMarkdown(input string) string {
	lines := strings.Split(input, "\n")
	var out strings.Builder
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Horizontal rule
		if strings.TrimSpace(line) == "---" {
			out.WriteString(`<hr class="my-8 border-t-2 border-gray-200 dark:border-gray-700">`)
			out.WriteByte('\n')
			i++
			continue
		}

		// Headings
		if strings.HasPrefix(line, "### ") {
			content := renderInline(strings.TrimPrefix(line, "### "))
			out.WriteString(`<h3 class="text-2xl font-bold tracking-tight mt-8 mb-4 text-gray-900 dark:text-white">`)
			out.WriteString(content)
			out.WriteString("</h3>\n")
			i++
			continue
		}
		if strings.HasPrefix(line, "## ") {
			content := renderInline(strings.TrimPrefix(line, "## "))
			out.WriteString(`<h2 class="text-3xl font-extrabold tracking-tight mt-10 mb-4 text-gray-900 dark:text-white">`)
			out.WriteString(content)
			out.WriteString("</h2>\n")
			i++
			continue
		}
		if strings.HasPrefix(line, "# ") {
			content := renderInline(strings.TrimPrefix(line, "# "))
			out.WriteString(`<h1 class="text-4xl font-black tracking-tight mb-6 text-gray-900 dark:text-white">`)
			out.WriteString(content)
			out.WriteString("</h1>\n")
			i++
			continue
		}

		// Unordered list
		if strings.HasPrefix(line, "- ") {
			out.WriteString(`<ul class="list-disc list-inside space-y-2 mb-6 text-lg text-gray-700 dark:text-gray-300">`)
			out.WriteByte('\n')
			for i < len(lines) && strings.HasPrefix(lines[i], "- ") {
				content := renderInline(strings.TrimPrefix(lines[i], "- "))
				out.WriteString(`<li class="leading-relaxed">`)
				out.WriteString(content)
				out.WriteString("</li>\n")
				i++
			}
			out.WriteString("</ul>\n")
			continue
		}

		// Ordered list
		if matched, _ := regexp.MatchString(`^\d+\. `, line); matched {
			out.WriteString(`<ol class="list-decimal list-inside space-y-2 mb-6 text-lg text-gray-700 dark:text-gray-300">`)
			out.WriteByte('\n')
			re := regexp.MustCompile(`^\d+\. `)
			for i < len(lines) && re.MatchString(lines[i]) {
				content := renderInline(re.ReplaceAllString(lines[i], ""))
				out.WriteString(`<li class="leading-relaxed">`)
				out.WriteString(content)
				out.WriteString("</li>\n")
				i++
			}
			out.WriteString("</ol>\n")
			continue
		}

		// Empty line
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Paragraph: collect consecutive non-empty, non-special lines
		var para []string
		for i < len(lines) {
			l := lines[i]
			trimmed := strings.TrimSpace(l)
			if trimmed == "" {
				break
			}
			if strings.HasPrefix(l, "# ") || strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "### ") {
				break
			}
			if strings.HasPrefix(l, "- ") {
				break
			}
			if matched, _ := regexp.MatchString(`^\d+\. `, l); matched {
				break
			}
			if trimmed == "---" {
				break
			}
			para = append(para, l)
			i++
		}
		if len(para) > 0 {
			content := renderInline(strings.Join(para, "\n"))
			out.WriteString(`<p class="text-lg leading-relaxed text-gray-700 dark:text-gray-300 mb-4">`)
			out.WriteString(content)
			out.WriteString("</p>\n")
		}
	}

	return out.String()
}

// renderInline processes inline markdown formatting.
func renderInline(text string) string {
	// Process inline code first (to avoid processing markdown inside code spans)
	text = processInlineCode(text)
	// Bold: **text**
	text = replacePattern(text, `\*\*(.+?)\*\*`, `<strong class="font-bold text-gray-900 dark:text-white">$1</strong>`)
	// Italic: *text*
	text = replacePattern(text, `\*(.+?)\*`, `<em class="italic">$1</em>`)
	// Underline: __text__
	text = replacePattern(text, `__(.+?)__`, `<u class="underline decoration-violet-500 decoration-2 underline-offset-2">$1</u>`)
	// Strikethrough: ~~text~~
	text = replacePattern(text, `~~(.+?)~~`, `<del class="line-through text-gray-400 dark:text-gray-500">$1</del>`)
	return text
}

// processInlineCode replaces `code` with <code> tags, avoiding further processing of content inside.
func processInlineCode(text string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllString(text, `<code class="bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-violet-600 dark:text-violet-400">$1</code>`)
}

func replacePattern(text, pattern, replacement string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, replacement)
}
