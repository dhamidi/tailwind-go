# tailwind CLI

A standalone build tool for [tailwind-go](../../README.md) — scans your project for Tailwind class candidates and emits CSS.

## Tutorials

### Getting Started

Install the CLI:

```sh
go install github.com/dhamidi/tailwind-go/cmd/tailwind@latest
```

Create a minimal input CSS file (`input.css`):

```css
@theme {
  --color-brand: #e11d48;
}
@utility brand-bg {
  background-color: var(--color-brand);
}
```

Create a small HTML file (`index.html`):

```html
<div class="flex items-center p-4 brand-bg">Hello</div>
```

Run the CLI:

```sh
tailwind -i input.css
```

The generated CSS prints to stdout, containing utility definitions for `flex`, `items-center`, `p-4`, and your custom `brand-bg` utility.

## How-To Guides

### Scan a project and write CSS to a file

```sh
tailwind -i input.css -o dist/styles.css
```

This scans the current directory for class candidates, resolves them against the utilities defined in `input.css`, and writes the result to `dist/styles.css`.

To scan a different directory, use `--cwd`:

```sh
tailwind -i input.css -o dist/styles.css --cwd ./src
```

### Use watch mode during development

```sh
tailwind -i input.css -o dist/styles.css --watch
```

The CLI watches all directories under `--cwd` (default `.`) for file changes and rebuilds automatically. Rebuilds are debounced by 100 ms to avoid redundant work during batch saves.

By default, watch mode exits when stdin is closed. This is useful for editor or tool integrations that launch the CLI as a subprocess — when the parent process exits and stdin closes, the watcher shuts down cleanly.

To keep watching regardless of stdin state, pass the `always` value:

```sh
tailwind -i input.css -o dist/styles.css --watch=always
```

### Minify output for production

```sh
tailwind -i input.css -o dist/styles.css --minify
```

or equivalently:

```sh
tailwind -i input.css -o dist/styles.css --optimize
```

Both flags strip whitespace, comments, and trailing semicolons from the output while preserving string literals and `url()` values.

### Pipe output to another tool

Output goes to stdout by default (`-o -`), so you can pipe it directly:

```sh
tailwind -i input.css | cat > styles.css
```

## Reference

### Flags

| Flag | Alias | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--input` | `-i` | string | *(none)* | Input CSS file defining custom utilities, themes, and variants |
| `--output` | `-o` | string | `-` (stdout) | Output file path; use `-` for stdout |
| `--cwd` | | string | `.` | Working directory to scan for class candidates |
| `--watch` | `-w` | optional | *(off)* | Watch for changes and rebuild; pass `always` to ignore stdin closure |
| `--minify` | `-m` | bool | `false` | Minify CSS output |
| `--optimize` | | bool | `false` | Alias for `--minify` |

`--help` / `-h` prints usage information and exits.

### Exit behaviour

- Exit code **0**: successful build (or clean shutdown of watch mode).
- Exit code **1**: error (unknown flag, missing required value, I/O failure).

### Watch mode and stdin

When `--watch` is used without a value, the CLI monitors stdin in a background goroutine. Once stdin reaches EOF (e.g., the parent process exits), the watcher terminates. Passing `--watch=always` disables this behaviour, keeping the process alive until it receives a signal.

### Directory walking

The file watcher recursively adds all directories under `--cwd`, skipping hidden directories (names starting with `.`) and `node_modules`.

## Explanation

### What the CLI does

The CLI is a thin build-tool wrapper around the [tailwind-go](../../README.md) library. It handles file I/O, directory scanning, watch mode, and minification so the core library can remain a pure `io.Writer`-based engine with no filesystem coupling.

### How watch mode works

Watch mode uses [fsnotify](https://github.com/fsnotify/fsnotify) to receive OS-level file-system events. On startup it recursively walks `--cwd` and adds every directory to the watcher (skipping hidden dirs and `node_modules`). When a new directory is created at runtime, it is automatically added as well.

Events are debounced with a 100 ms timer — rapid successive changes reset the timer so only one rebuild fires after edits settle. Each rebuild resets the engine, re-reads the input CSS, re-scans the directory tree, and writes the output.

### Relationship between `--input` CSS and scanned candidates

The `--input` file teaches the engine which utilities, themes, variants, and keyframes are available. The directory scan under `--cwd` extracts candidate class names from every file it encounters. The engine matches candidates against loaded definitions and emits CSS only for classes that are actually used. If no `--input` is provided, the engine uses only its built-in Tailwind v4 defaults.
