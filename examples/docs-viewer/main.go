package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	goformat "go/format"
	"go/parser"
	"go/token"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	tailwind "github.com/dhamidi/tailwind-go"
)

// goListResult holds relevant fields from `go list -json`.
type goListResult struct {
	Dir        string
	Name       string
	Doc        string
	ImportPath string
	GoFiles    []string
}

// PackageData holds all doc data for a package.
type PackageData struct {
	ImportPath  string
	Name        string
	Doc         template.HTML
	Synopsis    string
	Files       []FileEntry
	TailwindURL string
	ActiveFile  string // currently selected file, empty for overview
}

// FileEntry groups declarations by source file.
type FileEntry struct {
	Name   string
	Funcs  []DeclEntry
	Types  []TypeEntry
	Vars   []DeclEntry
	Consts []DeclEntry
}

// DeclEntry is a documented declaration.
type DeclEntry struct {
	Name      string
	Doc       template.HTML
	Signature string
}

// TypeEntry is a documented type with its methods.
type TypeEntry struct {
	Name      string
	Doc       template.HTML
	Signature string
	Methods   []DeclEntry
	Funcs     []DeclEntry
}

// renderDocHTML renders a doc comment string into HTML with Tailwind classes.
func renderDocHTML(docStr string) template.HTML {
	if docStr == "" {
		return ""
	}
	p := &comment.Parser{}
	d := p.Parse(docStr)
	return template.HTML(renderDoc(d))
}

// renderDoc walks a *comment.Doc and produces HTML with Tailwind classes.
func renderDoc(d *comment.Doc) string {
	var buf strings.Builder
	for _, b := range d.Content {
		switch v := b.(type) {
		case *comment.Paragraph:
			buf.WriteString(`<p class="mb-3 leading-7 text-gray-700">`)
			renderTexts(&buf, v.Text)
			buf.WriteString("</p>\n")
		case *comment.Heading:
			id := headingID(v.Text)
			buf.WriteString(fmt.Sprintf(`<h3 id="hdr-%s" class="text-lg font-semibold text-gray-900 mt-6 mb-3">`, html.EscapeString(id)))
			renderTexts(&buf, v.Text)
			buf.WriteString("</h3>\n")
		case *comment.Code:
			buf.WriteString(`<pre class="bg-gray-900 text-gray-100 font-mono text-sm rounded-lg p-4 mb-4 overflow-x-auto leading-relaxed"><code>`)
			buf.WriteString(html.EscapeString(v.Text))
			buf.WriteString("</code></pre>\n")
		case *comment.List:
			tag := "ul"
			listClass := "list-disc"
			if v.Items[0].Number != "" {
				tag = "ol"
				listClass = "list-decimal"
			}
			buf.WriteString(fmt.Sprintf(`<%s class="%s pl-6 mb-4 space-y-1 text-gray-700">`, tag, listClass))
			for _, item := range v.Items {
				buf.WriteString("<li>")
				for _, content := range item.Content {
					if p, ok := content.(*comment.Paragraph); ok {
						renderTexts(&buf, p.Text)
					}
				}
				buf.WriteString("</li>\n")
			}
			buf.WriteString(fmt.Sprintf("</%s>\n", tag))
		}
	}
	return buf.String()
}

// headingID generates an ID string from heading text.
func headingID(texts []comment.Text) string {
	var parts []string
	for _, t := range texts {
		switch v := t.(type) {
		case comment.Plain:
			parts = append(parts, string(v))
		case comment.Italic:
			parts = append(parts, string(v))
		}
	}
	return strings.Join(parts, "")
}

// renderTexts renders inline text nodes to HTML.
func renderTexts(buf *strings.Builder, texts []comment.Text) {
	for _, t := range texts {
		switch v := t.(type) {
		case comment.Plain:
			buf.WriteString(html.EscapeString(string(v)))
		case comment.Italic:
			buf.WriteString("<em>")
			buf.WriteString(html.EscapeString(string(v)))
			buf.WriteString("</em>")
		case *comment.Link:
			buf.WriteString(fmt.Sprintf(`<a class="text-blue-600 hover:underline" href="%s">`, html.EscapeString(v.URL)))
			renderTexts(buf, v.Text)
			buf.WriteString("</a>")
		case *comment.DocLink:
			href := docLinkHref(v)
			buf.WriteString(fmt.Sprintf(`<a class="text-blue-600 hover:underline" href="%s">`, html.EscapeString(href)))
			renderTexts(buf, v.Text)
			buf.WriteString("</a>")
		}
	}
}

// docLinkHref builds an href for a doc link.
func docLinkHref(dl *comment.DocLink) string {
	if dl.ImportPath != "" {
		if dl.Name != "" {
			return fmt.Sprintf("/?pkg=%s#%s", dl.ImportPath, dl.Name)
		}
		return fmt.Sprintf("/?pkg=%s", dl.ImportPath)
	}
	// local symbol
	return "#" + dl.Name
}

// loadPackage uses go list and go/parser to load structured package documentation.
func loadPackage(pkg string) (*PackageData, error) {
	// Step 1: go list -json to find source dir
	cmd := exec.Command("go", "list", "-json", pkg)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list %s: %v\n%s", pkg, err, out.String())
	}

	var info goListResult
	if err := json.Unmarshal(out.Bytes(), &info); err != nil {
		return nil, fmt.Errorf("parse go list output: %v", err)
	}

	// Step 2: parse source files
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, info.Dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse dir %s: %v", info.Dir, err)
	}

	astPkg, ok := pkgs[info.Name]
	if !ok {
		// Try first available package
		for _, ap := range pkgs {
			astPkg = ap
			break
		}
		if astPkg == nil {
			return nil, fmt.Errorf("no Go package found in %s", info.Dir)
		}
	}

	var files []*ast.File
	for _, f := range astPkg.Files {
		files = append(files, f)
	}

	// Step 3: extract structured docs
	importPath := info.ImportPath
	if importPath == "" {
		importPath = pkg
	}
	dpkg, err := doc.NewFromFiles(fset, files, importPath)
	if err != nil {
		return nil, fmt.Errorf("doc.NewFromFiles: %v", err)
	}

	// Step 4: group declarations by file
	fileMap := make(map[string]*FileEntry)
	ensureFile := func(filename string) *FileEntry {
		// Extract just the basename
		parts := strings.Split(filename, "/")
		base := parts[len(parts)-1]
		if fe, ok := fileMap[base]; ok {
			return fe
		}
		fe := &FileEntry{Name: base}
		fileMap[base] = fe
		return fe
	}

	// Functions
	for _, fn := range dpkg.Funcs {
		filename := fset.Position(fn.Decl.Pos()).Filename
		fe := ensureFile(filename)
		fe.Funcs = append(fe.Funcs, DeclEntry{
			Name:      fn.Name,
			Doc:       renderDocHTML(fn.Doc),
			Signature: formatFuncSignature(fn),
		})
	}

	// Types
	for _, tp := range dpkg.Types {
		filename := fset.Position(tp.Decl.Pos()).Filename
		fe := ensureFile(filename)

		te := TypeEntry{
			Name:      tp.Name,
			Doc:       renderDocHTML(tp.Doc),
			Signature: formatTypeSignature(tp),
		}

		for _, m := range tp.Methods {
			te.Methods = append(te.Methods, DeclEntry{
				Name:      m.Name,
				Doc:       renderDocHTML(m.Doc),
				Signature: formatFuncSignature(m),
			})
		}
		for _, fn := range tp.Funcs {
			te.Funcs = append(te.Funcs, DeclEntry{
				Name:      fn.Name,
				Doc:       renderDocHTML(fn.Doc),
				Signature: formatFuncSignature(fn),
			})
		}

		fe.Types = append(fe.Types, te)
	}

	// Vars
	for _, v := range dpkg.Vars {
		filename := fset.Position(v.Decl.Pos()).Filename
		fe := ensureFile(filename)
		fe.Vars = append(fe.Vars, DeclEntry{
			Name:      strings.Join(v.Names, ", "),
			Doc:       renderDocHTML(v.Doc),
			Signature: formatValueSignature(v),
		})
	}

	// Consts
	for _, c := range dpkg.Consts {
		filename := fset.Position(c.Decl.Pos()).Filename
		fe := ensureFile(filename)
		fe.Consts = append(fe.Consts, DeclEntry{
			Name:      strings.Join(c.Names, ", "),
			Doc:       renderDocHTML(c.Doc),
			Signature: formatValueSignature(c),
		})
	}

	// Build sorted file list
	var fileEntries []FileEntry
	for _, fe := range fileMap {
		fileEntries = append(fileEntries, *fe)
	}
	sort.Slice(fileEntries, func(i, j int) bool {
		return fileEntries[i].Name < fileEntries[j].Name
	})

	return &PackageData{
		ImportPath: importPath,
		Name:       dpkg.Name,
		Doc:        renderDocHTML(dpkg.Doc),
		Synopsis:   doc.Synopsis(dpkg.Doc),
		Files:      fileEntries,
	}, nil
}

// formatNode formats an AST node using go/format.
func formatNode(node ast.Node) string {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	if err := goformat.Node(&buf, fset, node); err != nil {
		return ""
	}
	return buf.String()
}

// formatFuncSignature formats a function declaration as a signature.
func formatFuncSignature(fn *doc.Func) string {
	if s := formatNode(fn.Decl); s != "" {
		return s
	}
	return "func " + fn.Name
}

// formatTypeSignature formats a type declaration.
func formatTypeSignature(tp *doc.Type) string {
	if s := formatNode(tp.Decl); s != "" {
		return s
	}
	return "type " + tp.Name
}

// formatValueSignature formats a var/const declaration.
func formatValueSignature(v *doc.Value) string {
	if s := formatNode(v.Decl); s != "" {
		return s
	}
	return strings.Join(v.Names, ", ")
}

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	pkg := flag.String("pkg", "github.com/dhamidi/tailwind-go", "package to document")
	templateFile := flag.String("template", "page.html", "path to HTML template file")
	flag.Parse()

	pageTemplate, err := os.ReadFile(*templateFile)
	if err != nil {
		log.Fatalf("read template %s: %v", *templateFile, err)
	}

	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"declCount": func(fe FileEntry) int {
			return len(fe.Funcs) + len(fe.Types) + len(fe.Vars) + len(fe.Consts)
		},
	}).Parse(string(pageTemplate)))

	tw := tailwind.New()
	tw.Write(pageTemplate)
	// Also register classes used by renderDoc() which generates HTML at runtime.
	tw.Write([]byte(`mb-3 leading-7 text-gray-700 text-lg font-semibold text-gray-900 mt-6 mb-3
		bg-gray-900 text-gray-100 font-mono text-sm rounded-lg p-4 mb-4 overflow-x-auto leading-relaxed
		list-disc list-decimal pl-6 mb-4 space-y-1 text-blue-600 hover:underline`))
	handler := tailwind.NewHandler(tw)
	handler.Build()

	mux := http.NewServeMux()
	mux.Handle(handler.URL(), handler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		target := *pkg
		if q := r.URL.Query().Get("pkg"); q != "" {
			target = q
		}
		file := r.URL.Query().Get("file")

		data, err := loadPackage(target)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data.TailwindURL = handler.URL()
		data.ActiveFile = file

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("template execute: %v", err)
		}
	})

	log.Printf("docs-viewer listening on %s (showing %s)", *addr, *pkg)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

