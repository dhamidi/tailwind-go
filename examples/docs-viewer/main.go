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

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	pkg := flag.String("pkg", "github.com/dhamidi/tailwind-go", "package to document")
	flag.Parse()

	// Pre-parse the template so we can scan it for classes
	funcMap := template.FuncMap{
		"splitLines": func(s string) []string { return strings.Split(s, "\n") },
	}
	tmpl := template.Must(template.New("page").Funcs(funcMap).Parse(pageTemplate))

	// Build the tailwind handler by scanning the template source
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

	// Extract synopsis from first non-empty prose section
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
func parseGoDoc(raw string) []DocSection {
	lines := strings.Split(raw, "\n")
	var sections []DocSection
	var current *DocSection

	flush := func() {
		if current != nil && strings.TrimSpace(current.Content) != "" {
			sections = append(sections, *current)
		}
		current = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimRight(line, " \t")

		// Detect ALL-CAPS section headings (e.g. "FUNCTIONS", "TYPES", "CONSTANTS")
		if isHeading(trimmed) {
			flush()
			current = &DocSection{Title: trimmed}
			continue
		}

		// Detect code lines: indented with tab or 4+ spaces
		isCode := len(line) > 0 && (line[0] == '\t' || strings.HasPrefix(line, "    "))

		if current == nil {
			current = &DocSection{IsCode: isCode}
		} else if current.IsCode != isCode && strings.TrimSpace(line) != "" {
			// Transition between code and prose
			flush()
			current = &DocSection{Title: "", IsCode: isCode}
		}

		if current.Content != "" {
			current.Content += "\n"
		}
		current.Content += trimmed
	}
	flush()

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

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Package}} — Go Documentation</title>
  <link rel="stylesheet" href="{{.TailwindURL}}">
</head>
<body class="min-h-screen bg-slate-50 flex flex-col">
  <header class="bg-slate-900 text-white px-4 py-3 shadow-md">
    <div class="flex items-center gap-6">
      <h1 class="text-xl font-bold tracking-tight">Go Docs</h1>
      <span class="text-slate-400 text-sm font-mono hidden md:block">{{.Package}}</span>
      <form class="flex-1 flex justify-end" action="/" method="get">
        <input
          type="text"
          name="pkg"
          placeholder="Search package..."
          value="{{.Package}}"
          class="bg-slate-800 text-white rounded px-4 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-slate-400 w-full max-w-[400px]"
        >
      </form>
    </div>
    {{if .Synopsis}}<p class="text-slate-400 text-sm mt-2">{{.Synopsis}}</p>{{end}}
  </header>

  <main class="flex-1 grid md:grid-cols-[250px_1fr] gap-6 p-4">
    <nav class="hidden md:block overflow-y-auto max-h-screen space-y-2 p-4 bg-white rounded-xl shadow-sm border border-slate-200">
      <h2 class="text-sm font-semibold text-slate-500 mb-4">Sections</h2>
      {{range .Sections}}{{if .Title}}
      <a href="#{{.Title}}" class="block text-sm px-4 py-2 rounded text-slate-700 hover:bg-slate-700 hover:text-white">{{.Title}}</a>
      {{end}}{{end}}
    </nav>

    <div class="space-y-4 overflow-x-auto">
      {{range .Sections}}
      <section {{if .Title}}id="{{.Title}}"{{end}} class="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
        {{if .Title}}<h2 class="text-lg font-semibold text-slate-900 border-b border-slate-200 pb-2 mb-4">{{.Title}}</h2>{{end}}
        {{if .IsCode}}<pre class="bg-slate-800 text-green-400 font-mono text-sm rounded p-4 overflow-x-auto leading-relaxed"><code>{{.Content}}</code></pre>
        {{else}}<div class="text-slate-700 leading-relaxed text-sm space-y-2">{{range $line := .Content | splitLines}}<p>{{$line}}</p>{{end}}</div>
        {{end}}
      </section>
      {{end}}
    </div>
  </main>

  <footer class="bg-slate-100 text-slate-500 text-sm text-center py-3 border-t border-slate-200">
    Powered by <span class="font-semibold text-slate-700">tailwind-go</span>
  </footer>
</body>
</html>`
