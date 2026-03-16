# Blog Example

A fully featured, server-rendered blog built with [tailwind-go](https://github.com/dhamidi/tailwind-go). Demonstrates dark mode, animations, responsive design, gradients, and bold typography — all with zero JavaScript build tools.

## Quick Start

```bash
cd examples/blog
go run .
```

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Usage

```
go run . [flags]

Flags:
  -addr string    listen address (default ":8080")
  -data string    data directory for content (default: copies embedded content to temp dir)
```

The server also respects the `PORT` environment variable, which is set by
many deployment platforms (Heroku, Cloud Run, Railway, etc.). The precedence
order is:

1. `-addr` flag (highest)
2. `PORT` environment variable
3. Default `:8080`

```bash
# Listen on :9090 via PORT
PORT=9090 go run .

# -addr flag takes precedence over PORT
PORT=9090 go run . -addr :7070   # listens on :7070
```

## Features

- **Server-side rendering** with Go's `html/template`
- **Dark mode** with localStorage persistence and system preference detection
- **Responsive design** using Tailwind breakpoints (`sm:`, `md:`, `lg:`)
- **Gradient navigation** and accent elements
- **Markdown rendering** — a pure-Go subset renderer with Tailwind-styled output
- **CRUD operations** — create, edit, publish, and delete posts and drafts
- **Media uploads** — simple file management interface
- **Content-hashed CSS** — generated at startup, cached with immutable headers

## Architecture

| File | Purpose |
|------|---------|
| `main.go` | HTTP server, routing, request handlers |
| `content.go` | `Post`, `Draft`, and `Media` types; `Store` for filesystem CRUD |
| `markdown.go` | Pure-function markdown-to-HTML renderer with Tailwind classes |
| `markdown_test.go` | Unit tests for the markdown renderer (no I/O) |
| `layout.html` | Base template: nav, footer, dark mode toggle |
| `home.html` | Hero section and recent posts grid |
| `post.html` | Single post view with prev/next navigation |
| `archive.html` | All posts grouped by year |
| `editor.html` | Create/edit post form with live character count |
| `drafts.html` | Draft management with publish/delete actions |
| `media.html` | Media library with upload |

## Sample Content

Three sample posts ship embedded in the binary:

- **Welcome to the Blog** — introductory post about tailwind-go
- **Building with tailwind-go** — tutorial on getting started
- **The Complete Dark Mode Guide** — design principles for dark mode

One draft is also included for testing the drafts workflow.

## Tailwind Features Showcased

This example exercises a wide range of Tailwind utilities:

- Responsive breakpoints (`sm:`, `md:`, `lg:`, `xl:`)
- Dark mode variants (`dark:`)
- Hover and focus states (`hover:`, `focus:`, `focus-visible:`)
- Gradients (`bg-gradient-to-r`, `from-*`, `via-*`, `to-*`)
- Shadows and transitions (`shadow-lg`, `transition-all`, `duration-300`)
- Transforms (`hover:scale-105`, `hover:-translate-y-1`)
- Typography scale (`text-xs` through `text-6xl`)
- Flexbox and Grid layouts
- Borders, rings, and rounded corners
- Arbitrary values (`min-h-[400px]`)
