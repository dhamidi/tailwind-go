package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed posts/*.md
var embeddedPosts embed.FS

//go:embed drafts/*.md
var embeddedDrafts embed.FS

//go:embed media/*
var embeddedMedia embed.FS

// Post represents a blog post or draft.
type Post struct {
	Slug      string
	Title     string
	Summary   string
	Body      string
	HTML      template.HTML
	Date      time.Time
	Tags      []string
	Published bool
}

// Media represents an uploaded media file.
type Media struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// Store manages blog content on disk.
type Store struct {
	dir string
}

// NewStore creates a new Store rooted at dir.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// Init copies embedded content into the store directory.
func (s *Store) Init() error {
	for _, sub := range []struct {
		name string
		fsys fs.FS
	}{
		{"posts", embeddedPosts},
		{"drafts", embeddedDrafts},
		{"media", embeddedMedia},
	} {
		if err := fs.WalkDir(sub.fsys, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			target := filepath.Join(s.dir, path)
			if d.IsDir() {
				return os.MkdirAll(target, 0o755)
			}
			// Skip .gitkeep
			if d.Name() == ".gitkeep" {
				return nil
			}
			data, err := fs.ReadFile(sub.fsys, path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0o644)
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListPosts returns all published posts sorted by date descending.
func (s *Store) ListPosts() ([]Post, error) {
	return s.listDir("posts", true)
}

// ListDrafts returns all draft posts.
func (s *Store) ListDrafts() ([]Post, error) {
	return s.listDir("drafts", false)
}

func (s *Store) listDir(subdir string, published bool) ([]Post, error) {
	dir := filepath.Join(s.dir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var posts []Post
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		post := parsePost(e.Name(), string(data), published)
		posts = append(posts, post)
	}
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts, nil
}

// GetPost returns a single post by slug.
func (s *Store) GetPost(slug string) (Post, error) {
	posts, err := s.ListPosts()
	if err != nil {
		return Post{}, err
	}
	for _, p := range posts {
		if p.Slug == slug {
			return p, nil
		}
	}
	// Also check drafts
	drafts, err := s.ListDrafts()
	if err != nil {
		return Post{}, err
	}
	for _, d := range drafts {
		if d.Slug == slug {
			return d, nil
		}
	}
	return Post{}, fmt.Errorf("post not found: %s", slug)
}

// SavePost writes a post to the posts directory.
func (s *Store) SavePost(post Post) error {
	dir := filepath.Join(s.dir, "posts")
	if !post.Published {
		dir = filepath.Join(s.dir, "drafts")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	filename := fmt.Sprintf("%s-%s.md", post.Date.Format("2006-01-02"), post.Slug)
	content := formatPostContent(post)
	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644)
}

// DeletePost removes a post by slug.
func (s *Store) DeletePost(slug string) error {
	return s.deleteBySlug("posts", slug)
}

// PublishDraft moves a draft to the posts directory.
func (s *Store) PublishDraft(slug string) error {
	drafts, err := s.ListDrafts()
	if err != nil {
		return err
	}
	for _, d := range drafts {
		if d.Slug == slug {
			// Delete draft file
			if err := s.deleteBySlug("drafts", slug); err != nil {
				return err
			}
			// Save as published post
			d.Published = true
			d.Date = time.Now()
			return s.SavePost(d)
		}
	}
	return fmt.Errorf("draft not found: %s", slug)
}

// ListMedia returns all files in the media directory.
func (s *Store) ListMedia() ([]Media, error) {
	dir := filepath.Join(s.dir, "media")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var items []Media
	for _, e := range entries {
		if e.IsDir() || e.Name() == ".gitkeep" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		items = append(items, Media{
			Name:    e.Name(),
			Path:    "/media/" + e.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return items, nil
}

// SaveMedia writes a media file.
func (s *Store) SaveMedia(name string, data []byte) error {
	dir := filepath.Join(s.dir, "media")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), data, 0o644)
}

// DeleteMedia removes a media file.
func (s *Store) DeleteMedia(name string) error {
	return os.Remove(filepath.Join(s.dir, "media", name))
}

func (s *Store) deleteBySlug(subdir, slug string) error {
	dir := filepath.Join(s.dir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "-"+slug+".md") || e.Name() == slug+".md" {
			return os.Remove(filepath.Join(dir, e.Name()))
		}
	}
	return fmt.Errorf("file not found for slug: %s", slug)
}

// parsePost parses a markdown file into a Post.
func parsePost(filename, content string, published bool) Post {
	// Parse date and slug from filename: 2026-03-01-welcome.md -> date=2026-03-01, slug=welcome
	name := strings.TrimSuffix(filename, ".md")
	var date time.Time
	var slug string
	if len(name) >= 10 {
		if d, err := time.Parse("2006-01-02", name[:10]); err == nil {
			date = d
			slug = name[11:] // skip the date and hyphen
		} else {
			slug = name
		}
	} else {
		slug = name
	}

	lines := strings.Split(content, "\n")
	var title string
	var tags []string
	bodyStart := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && title == "" {
			title = strings.TrimPrefix(trimmed, "# ")
			bodyStart = i + 1
			continue
		}
		if strings.HasPrefix(trimmed, "tags:") && i <= bodyStart+1 {
			tagStr := strings.TrimPrefix(trimmed, "tags:")
			for _, t := range strings.Split(tagStr, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
			bodyStart = i + 1
			continue
		}
		if trimmed != "" && i >= bodyStart {
			bodyStart = i
			break
		}
	}

	body := strings.Join(lines[bodyStart:], "\n")
	body = strings.TrimSpace(body)

	// Extract summary (first paragraph)
	summary := extractSummary(body)

	html := template.HTML(renderMarkdown(body))

	return Post{
		Slug:      slug,
		Title:     title,
		Summary:   summary,
		Body:      content,
		HTML:      html,
		Date:      date,
		Tags:      tags,
		Published: published,
	}
}

// extractSummary returns the first paragraph of text.
func extractSummary(body string) string {
	lines := strings.Split(body, "\n")
	var para []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(para) > 0 {
				break
			}
			continue
		}
		// Skip headings and special lines
		if strings.HasPrefix(trimmed, "#") || trimmed == "---" {
			if len(para) > 0 {
				break
			}
			continue
		}
		para = append(para, trimmed)
	}
	return strings.Join(para, " ")
}

// formatPostContent produces the markdown file content from a Post.
func formatPostContent(post Post) string {
	var b strings.Builder
	b.WriteString("# ")
	b.WriteString(post.Title)
	b.WriteByte('\n')
	if len(post.Tags) > 0 {
		b.WriteString("tags: ")
		b.WriteString(strings.Join(post.Tags, ", "))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	// Write body without the title/tags header
	body := post.Body
	// If Body contains the full file content (with # Title and tags), strip that
	lines := strings.Split(body, "\n")
	start := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			start = i + 1
			continue
		}
		if strings.HasPrefix(trimmed, "tags:") && i <= start+1 {
			start = i + 1
			continue
		}
		if trimmed != "" {
			start = i
			break
		}
	}
	b.WriteString(strings.TrimSpace(strings.Join(lines[start:], "\n")))
	b.WriteByte('\n')
	return b.String()
}
