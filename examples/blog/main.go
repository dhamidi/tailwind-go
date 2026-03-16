package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tailwind "github.com/dhamidi/tailwind-go"
)

var (
	sessions   = map[string]bool{}
	sessionsMu sync.Mutex
)

func newSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isLoggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	return sessions[cookie.Value]
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isLoggedIn(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

var templates *template.Template
var store *Store

// PageData is the common data passed to all templates.
type PageData struct {
	TailwindURL string
	Posts       []Post
	Post        Post
	Prev        *Post
	Next        *Post
	Drafts      []Post
	Media       []Media
	Years       []YearGroup
	LoggedIn    bool
	LoginError  string
}

// YearGroup groups posts by year for the archive page.
type YearGroup struct {
	Year  int
	Posts []Post
}

var tailwindURL string

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	dataDir := flag.String("data", "", "data directory for content (default: temp dir)")
	flag.Parse()

	// Determine listen address: -addr flag > PORT env > default ":8080"
	listenAddr := *addr
	if port := os.Getenv("PORT"); port != "" && !isFlagSet("addr") {
		listenAddr = ":" + port
	}

	// Set up data directory
	dir := *dataDir
	if dir == "" {
		tmp, err := os.MkdirTemp("", "blog-*")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(tmp)
		dir = tmp
	}

	store = NewStore(dir)
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	// Parse templates
	templates = template.Must(template.ParseFiles(
		"layout.html",
		"home.html",
		"post.html",
		"archive.html",
		"editor.html",
		"drafts.html",
		"media.html",
		"login.html",
	))

	// Set up Tailwind
	tw := tailwind.New()

	// Scan template files
	for _, name := range []string{"layout.html", "home.html", "post.html", "archive.html", "editor.html", "drafts.html", "media.html", "login.html"} {
		data, err := os.ReadFile(name)
		if err != nil {
			log.Fatal(err)
		}
		tw.Write(data)
	}

	// Register classes used by the markdown renderer at runtime
	tw.Write([]byte(`
		text-4xl font-black tracking-tight mb-6 text-gray-900 dark:text-white
		text-3xl font-extrabold tracking-tight mt-10 mb-4
		text-2xl font-bold tracking-tight mt-8 mb-4
		text-lg leading-relaxed text-gray-700 dark:text-gray-300 mb-4
		font-bold text-gray-900 dark:text-white
		italic
		underline decoration-violet-500 decoration-2 underline-offset-2
		line-through text-gray-400 dark:text-gray-500
		bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded text-sm font-mono text-violet-600 dark:text-violet-400
		list-disc list-inside space-y-2 mb-6
		list-decimal list-inside space-y-2 mb-6
		leading-relaxed
		my-8 border-t-2 border-gray-200 dark:border-gray-700
		space-y-0 prose-custom
	`))

	handler := tailwind.NewHandler(tw)
	handler.Build()
	tailwindURL = handler.URL()

	mux := http.NewServeMux()
	mux.Handle(handler.URL(), handler)

	// Routes
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/posts", handleArchive)
	mux.HandleFunc("/posts/", handlePost)
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/logout", handleLogout)
	mux.HandleFunc("/admin/editor", requireAuth(handleEditor))
	mux.HandleFunc("/admin/editor/", requireAuth(handleEditor))
	mux.HandleFunc("/admin/posts", requireAuth(handleSavePost))
	mux.HandleFunc("/admin/posts/", requireAuth(handlePostAction))
	mux.HandleFunc("/admin/drafts", requireAuth(handleDrafts))
	mux.HandleFunc("/admin/drafts/", requireAuth(handleDraftAction))
	mux.HandleFunc("/admin/media", requireAuth(handleMediaRoute))
	mux.HandleFunc("/admin/media/", requireAuth(handleMediaAction))

	log.Printf("Starting blog server on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}

func renderPage(w http.ResponseWriter, name string, data PageData) {
	data.TailwindURL = tailwindURL
	// We need to execute layout with the content block
	err := templates.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("layout.html", "login.html"))
		data := PageData{TailwindURL: tailwindURL}
		if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			log.Printf("template error: %v", err)
		}
		return
	}

	user := r.FormValue("username")
	pass := r.FormValue("password")
	adminUser := os.Getenv("ADMIN_USER")
	adminPass := os.Getenv("ADMIN_PASSWORD")

	if adminUser == "" || adminPass == "" {
		http.Error(w, "Admin credentials not configured", http.StatusInternalServerError)
		return
	}

	if user == adminUser && pass == adminPass {
		token := newSession()
		sessionsMu.Lock()
		sessions[token] = true
		sessionsMu.Unlock()
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/admin/drafts", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFiles("layout.html", "login.html"))
	data := PageData{TailwindURL: tailwindURL, LoginError: "Invalid username or password"}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	posts, err := store.ListPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Show at most 3 recent posts
	if len(posts) > 3 {
		posts = posts[:3]
	}

	// Re-parse templates to pick up the home content block
	tmpl := template.Must(template.ParseFiles("layout.html", "home.html"))
	data := PageData{TailwindURL: tailwindURL, Posts: posts, LoggedIn: isLoggedIn(r)}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleArchive(w http.ResponseWriter, r *http.Request) {
	posts, err := store.ListPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	years := groupByYear(posts)

	tmpl := template.Must(template.ParseFiles("layout.html", "archive.html"))
	data := PageData{TailwindURL: tailwindURL, Years: years, LoggedIn: isLoggedIn(r)}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/posts/")
	if slug == "" {
		handleArchive(w, r)
		return
	}

	posts, err := store.ListPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var post *Post
	var idx int
	for i, p := range posts {
		if p.Slug == slug {
			post = &posts[i]
			idx = i
			break
		}
	}
	if post == nil {
		http.NotFound(w, r)
		return
	}

	data := PageData{TailwindURL: tailwindURL, Post: *post, LoggedIn: isLoggedIn(r)}
	if idx > 0 {
		data.Next = &posts[idx-1]
	}
	if idx < len(posts)-1 {
		data.Prev = &posts[idx+1]
	}

	tmpl := template.Must(template.ParseFiles("layout.html", "post.html"))
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleEditor(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/admin/editor/")
	slug = strings.TrimPrefix(slug, "/admin/editor")

	var post Post
	if slug != "" {
		p, err := store.GetPost(slug)
		if err == nil {
			post = p
		}
	}

	tmpl := template.Must(template.ParseFiles("layout.html", "editor.html"))
	data := PageData{TailwindURL: tailwindURL, Post: post, LoggedIn: isLoggedIn(r)}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleSavePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	body := r.FormValue("body")
	tagsStr := r.FormValue("tags")
	action := r.FormValue("action")
	originalSlug := r.FormValue("original_slug")

	var tags []string
	for _, t := range strings.Split(tagsStr, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	slug := slugify(title)
	published := action == "publish"

	// Build the full markdown content
	var content strings.Builder
	content.WriteString("# ")
	content.WriteString(title)
	content.WriteByte('\n')
	if len(tags) > 0 {
		content.WriteString("tags: ")
		content.WriteString(strings.Join(tags, ", "))
		content.WriteByte('\n')
	}
	content.WriteByte('\n')
	content.WriteString(body)

	postDate := time.Now()
	if originalSlug != "" {
		// Editing existing post — preserve original date and remove old file
		if orig, err := store.GetPost(originalSlug); err == nil {
			postDate = orig.Date
		}
		// Clean up old file(s); ignore errors since file may only exist in one location
		store.deleteBySlug("drafts", originalSlug)
		store.deleteBySlug("posts", originalSlug)
	}

	post := Post{
		Slug:      slug,
		Title:     title,
		Body:      content.String(),
		Date:      postDate,
		Tags:      tags,
		Published: published,
	}

	if err := store.SavePost(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if published {
		http.Redirect(w, r, "/posts/"+slug, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/admin/drafts", http.StatusSeeOther)
	}
}

func handlePostAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/posts/")
	if strings.HasSuffix(path, "/delete") {
		slug := strings.TrimSuffix(path, "/delete")
		if r.Method == http.MethodPost {
			store.DeletePost(slug)
			// Also try deleting from drafts
			store.deleteBySlug("drafts", slug)
		}
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
		return
	}
}

func handleDrafts(w http.ResponseWriter, r *http.Request) {
	drafts, err := store.ListDrafts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("layout.html", "drafts.html"))
	data := PageData{TailwindURL: tailwindURL, Drafts: drafts, LoggedIn: isLoggedIn(r)}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleDraftAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/drafts/")
	if strings.HasSuffix(path, "/publish") && r.Method == http.MethodPost {
		slug := strings.TrimSuffix(path, "/publish")
		if err := store.PublishDraft(slug); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
		return
	}
}

func handleMediaRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		handleUploadMedia(w, r)
		return
	}
	handleMediaList(w, r)
}

func handleMediaList(w http.ResponseWriter, r *http.Request) {
	media, err := store.ListMedia()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("layout.html", "media.html"))
	data := PageData{TailwindURL: tailwindURL, Media: media, LoggedIn: isLoggedIn(r)}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleUploadMedia(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20) // 32MB max
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	name := filepath.Base(header.Filename)
	if err := store.SaveMedia(name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}

func handleMediaAction(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/media/")
	if strings.HasSuffix(path, "/delete") && r.Method == http.MethodPost {
		name := strings.TrimSuffix(path, "/delete")
		store.DeleteMedia(name)
		http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
		return
	}
}

func groupByYear(posts []Post) []YearGroup {
	if len(posts) == 0 {
		return nil
	}
	var groups []YearGroup
	currentYear := -1
	for _, p := range posts {
		year := p.Date.Year()
		if year != currentYear {
			groups = append(groups, YearGroup{Year: year})
			currentYear = year
		}
		groups[len(groups)-1].Posts = append(groups[len(groups)-1].Posts, p)
	}
	return groups
}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		if r == ' ' || r == '-' {
			return '-'
		}
		return -1
	}, s)
	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

// isFlagSet reports whether the named flag was explicitly set on the command line.
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func init() {
	// Change to the directory containing this program if running via go run
	// This helps find template files
	if _, err := os.Stat("layout.html"); os.IsNotExist(err) {
		if exe, err := os.Executable(); err == nil {
			dir := filepath.Dir(exe)
			if _, err := os.Stat(filepath.Join(dir, "layout.html")); err == nil {
				os.Chdir(dir)
			}
		}
	}
}
