package main

import (
	"bytes"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"

	tailwind "github.com/dhamidi/tailwind-go"
)

// PageData holds template data for the doc viewer page.
type PageData struct {
	Package     string       // package import path
	Synopsis    string       // one-line synopsis
	Sections    []DocSection // parsed doc sections
	TailwindURL string       // cache-busted CSS URL
}

// DocSection represents a logical block of documentation.
type DocSection struct {
	Title   string // e.g. "FUNCTIONS", "TYPES", ""
	Content string // raw text content
	IsCode  bool   // true for indented/code blocks
}

// paragraphs joins consecutive non-blank lines into paragraphs,
// splitting on blank lines. Returns a slice of paragraph strings.
func paragraphs(s string) []string {
	lines := strings.Split(s, "\n")
	var result []string
	var current []string

	flush := func() {
		if len(current) > 0 {
			result = append(result, strings.Join(current, " "))
			current = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			flush()
		} else {
			current = append(current, trimmed)
		}
	}
	flush()
	return result
}

// sectionID returns a URL-friendly ID for a section title.
func sectionID(title string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(title)), " ", "-")
}

// titleCase converts an ALL-CAPS or "# Heading" title to title case for display.
func titleCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	pkg := flag.String("pkg", "github.com/dhamidi/tailwind-go", "package to document")
	flag.Parse()

	funcMap := template.FuncMap{
		"paragraphs": paragraphs,
		"sectionID":  sectionID,
		"titleCase":  titleCase,
	}
	tmpl := template.Must(template.New("page").Funcs(funcMap).Parse(pageTemplate))

	tw := tailwind.New()
	tw.Write([]byte(pageTemplate))
	handler := tailwind.NewHandler(tw)
	handler.Build()

	mux := http.NewServeMux()
	mux.Handle(handler.URL(), handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := *pkg
		if q := r.URL.Query().Get("pkg"); q != "" {
			target = q
		}

		data, err := buildPageData(target, handler.URL())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("template execute: %v", err)
		}
	})

	log.Printf("docs-viewer listening on %s (showing %s)", *addr, *pkg)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

// buildPageData runs go doc and builds the template data.
func buildPageData(pkg, tailwindURL string) (*PageData, error) {
	cmd := exec.Command("go", "doc", "-all", pkg)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return &PageData{
			Package:     pkg,
			Synopsis:    "Error fetching documentation: " + err.Error(),
			Sections:    []DocSection{{Content: out.String(), IsCode: true}},
			TailwindURL: tailwindURL,
		}, nil
	}

	raw := out.String()
	sections := parseGoDoc(raw)

	synopsis := ""
	for _, s := range sections {
		if !s.IsCode && s.Title == "" && strings.TrimSpace(s.Content) != "" {
			synopsis = strings.TrimSpace(strings.SplitN(s.Content, "\n", 2)[0])
			break
		}
	}

	return &PageData{
		Package:     pkg,
		Synopsis:    synopsis,
		Sections:    sections,
		TailwindURL: tailwindURL,
	}, nil
}

// parseGoDoc splits raw go doc output into DocSections.
// It recognizes ALL-CAPS headings (FUNCTIONS, TYPES, etc.) and
// markdown-style headings (# Printing, # Scanning, etc.).
// Code blocks and prose are grouped under the most recent heading.
func parseGoDoc(raw string) []DocSection {
	lines := strings.Split(raw, "\n")
	var sections []DocSection

	// currentTitle tracks the most recent heading for grouping
	currentTitle := ""
	var codeLines []string
	var proseLines []string

	flushCode := func() {
		if len(codeLines) > 0 {
			content := strings.Join(codeLines, "\n")
			if strings.TrimSpace(content) != "" {
				sections = append(sections, DocSection{
					Title:   currentTitle,
					Content: content,
					IsCode:  true,
				})
				currentTitle = "" // title consumed
			}
			codeLines = nil
		}
	}

	flushProse := func() {
		if len(proseLines) > 0 {
			content := strings.Join(proseLines, "\n")
			if strings.TrimSpace(content) != "" {
				sections = append(sections, DocSection{
					Title:   currentTitle,
					Content: content,
					IsCode:  false,
				})
				currentTitle = "" // title consumed
			}
			proseLines = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		// Detect headings: ALL-CAPS or "# Heading"
		if isHeading(trimmed) {
			flushProse()
			flushCode()
			currentTitle = trimmed
			continue
		}
		if isMarkdownHeading(trimmed) {
			flushProse()
			flushCode()
			currentTitle = strings.TrimSpace(trimmed[1:]) // strip "# "
			continue
		}

		// Detect code lines: indented with tab or 4+ spaces
		isCode := len(line) > 0 && (line[0] == '\t' || strings.HasPrefix(line, "    "))

		if isCode {
			flushProse()
			codeLines = append(codeLines, trimmed)
		} else {
			flushCode()
			proseLines = append(proseLines, trimmed)
		}
	}
	flushProse()
	flushCode()

	return sections
}

// isHeading returns true if line is an ALL-CAPS heading like "FUNCTIONS" or "TYPES".
func isHeading(line string) bool {
	if len(line) == 0 || line[0] == ' ' || line[0] == '\t' {
		return false
	}
	hasLetter := false
	for _, r := range line {
		if r >= 'A' && r <= 'Z' {
			hasLetter = true
		} else if r == ' ' {
			continue
		} else {
			return false
		}
	}
	return hasLetter
}

// isMarkdownHeading returns true if line is a markdown-style heading like "# Printing".
func isMarkdownHeading(line string) bool {
	return len(line) >= 2 && line[0] == '#' && line[1] == ' '
}

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Package}} — Go Documentation</title>
  <link rel="stylesheet" href="{{.TailwindURL}}">
</head>
<body class="min-h-screen bg-gray-50 flex flex-col">

  <header class="bg-white border-b border-gray-200 sticky top-0 z-10">
    <div class="max-w-7xl mx-auto px-6 py-4">
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-3">
          <h1 class="text-xl font-bold text-gray-900">Go Docs</h1>
          <span class="text-gray-400 font-mono text-sm hidden lg:block">{{.Package}}</span>
        </div>
        <form class="flex-1 flex justify-end" action="/" method="get">
          <div class="relative w-full max-w-sm">
            <input
              type="text"
              name="pkg"
              placeholder="Search packages..."
              value="{{.Package}}"
              class="w-full bg-gray-100 text-gray-900 rounded-full px-5 py-2 text-sm font-mono border border-gray-200 focus:outline-none focus:bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
          </div>
        </form>
      </div>
      {{if .Synopsis}}<p class="text-gray-500 text-sm mt-2">{{.Synopsis}}</p>{{end}}
    </div>
  </header>

  <div class="flex-1 flex max-w-7xl mx-auto w-full">

    <nav class="hidden lg:block w-64 shrink-0 overflow-y-auto sticky top-16 self-start max-h-screen p-6">
      <div class="bg-white rounded-xl border border-gray-200 shadow-sm p-5">
        <h2 class="text-xs font-semibold uppercase tracking-wide text-gray-400 mb-4">On this page</h2>
        {{range .Sections}}{{if .Title}}
        <a href="#{{.Title | sectionID}}" class="block text-sm text-gray-600 hover:text-blue-600 hover:bg-blue-50 rounded-lg px-3 py-2 mb-1">{{.Title | titleCase}}</a>
        {{end}}{{end}}
      </div>
    </nav>

    <main class="flex-1 min-w-0 p-6 lg:p-8">
      <div class="space-y-6">
        {{range .Sections}}
        <section {{if .Title}}id="{{.Title | sectionID}}"{{end}} class="bg-white rounded-xl shadow-sm border border-gray-200">
          {{if .Title}}<div class="px-6 py-4 border-b border-gray-100">
            <h2 class="text-lg font-semibold text-gray-900">{{.Title | titleCase}}</h2>
          </div>{{end}}
          {{if .IsCode}}<div class="p-0">
            <pre class="bg-gray-900 text-gray-100 font-mono text-sm rounded-b-xl p-5 overflow-x-auto leading-relaxed"><code>{{.Content}}</code></pre>
          </div>
          {{else}}<div class="p-6 text-gray-700 text-base leading-7">
            {{range $p := .Content | paragraphs}}<p class="mb-3">{{$p}}</p>
            {{end}}
          </div>
          {{end}}
        </section>
        {{end}}
      </div>
    </main>

  </div>

  <footer class="py-6 text-center text-sm text-gray-400 border-t border-gray-200">
    Powered by <span class="font-semibold text-gray-600">tailwind-go</span>
  </footer>

</body>
</html>`
