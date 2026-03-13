package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tailwind "github.com/dhamidi/tailwind-go"
	"github.com/fsnotify/fsnotify"
)

// watchLoop watches cwd for file changes using fsnotify and rebuilds CSS.
// If stdinAware is true, it exits when stdin is closed.
func watchLoop(engine *tailwind.Engine, input, output, cwd string, stdinAware bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	// Add all directories under cwd recursively.
	if err := addDirsRecursive(watcher, cwd); err != nil {
		return fmt.Errorf("watching directories: %w", err)
	}

	// Stdin closure channel.
	stdinClosed := make(chan struct{})
	if stdinAware {
		go func() {
			io.Copy(io.Discard, os.Stdin)
			close(stdinClosed)
		}()
	}

	// Debounce timer: fires after 100ms of quiet.
	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}
	pending := false

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Only care about writes, creates, removes, renames.
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
				continue
			}
			// Skip hidden files/dirs.
			base := filepath.Base(event.Name)
			if strings.HasPrefix(base, ".") {
				continue
			}
			// If a new directory is created, watch it too.
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					addDirsRecursive(watcher, event.Name)
				}
			}
			// Reset debounce timer.
			if !debounce.Stop() && pending {
				select {
				case <-debounce.C:
				default:
				}
			}
			debounce.Reset(100 * time.Millisecond)
			pending = true

		case <-debounce.C:
			pending = false
			start := time.Now()
			fmt.Fprintf(os.Stderr, "Rebuilding...\n")
			if err := rebuild(engine, input, output, cwd); err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Done in %dms\n", time.Since(start).Milliseconds())
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watch error: %s\n", err)

		case <-stdinClosed:
			return nil
		}
	}
}

// rebuild resets the engine, re-scans, and writes output.
func rebuild(engine *tailwind.Engine, input, output, cwd string) error {
	engine.Reset()

	if input != "" {
		css, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		if err := engine.LoadCSS(css); err != nil {
			return fmt.Errorf("parsing input CSS: %w", err)
		}
	}

	fsys := os.DirFS(cwd)
	if err := engine.Scan(fsys); err != nil {
		return fmt.Errorf("scanning %s: %w", cwd, err)
	}

	css := engine.FullCSS()

	if output == "-" {
		fmt.Print(css)
	} else {
		if err := os.WriteFile(output, []byte(css), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}

	return nil
}

// addDirsRecursive walks root and adds all directories to the watcher,
// skipping hidden directories and node_modules.
func addDirsRecursive(w *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable paths
		}
		if !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base != "." && strings.HasPrefix(base, ".") {
			return filepath.SkipDir
		}
		if base == "node_modules" {
			return filepath.SkipDir
		}
		return w.Add(path)
	})
}
