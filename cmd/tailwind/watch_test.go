package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tailwind "github.com/dhamidi/tailwind-go"
	"github.com/fsnotify/fsnotify"
)

func TestAddDirsRecursive(t *testing.T) {
	dir := t.TempDir()

	// Create directory structure.
	os.MkdirAll(filepath.Join(dir, "src", "components"), 0755)
	os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0755)
	os.MkdirAll(filepath.Join(dir, "node_modules", "pkg"), 0755)
	os.MkdirAll(filepath.Join(dir, "public"), 0755)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := addDirsRecursive(w, dir); err != nil {
		t.Fatal(err)
	}

	watched := w.WatchList()
	watchedSet := make(map[string]bool)
	for _, p := range watched {
		watchedSet[p] = true
	}

	// Should watch root, src, src/components, public.
	for _, want := range []string{dir, filepath.Join(dir, "src"), filepath.Join(dir, "src", "components"), filepath.Join(dir, "public")} {
		if !watchedSet[want] {
			t.Errorf("expected %s to be watched", want)
		}
	}

	// Should NOT watch .git or node_modules.
	for _, skip := range []string{filepath.Join(dir, ".git"), filepath.Join(dir, "node_modules")} {
		if watchedSet[skip] {
			t.Errorf("expected %s to NOT be watched", skip)
		}
	}
}

func TestWatchLoopExitsOnStdinClose(t *testing.T) {
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	origStdin := os.Stdin
	os.Stdin = pr
	defer func() { os.Stdin = origStdin }()

	dir := t.TempDir()

	done := make(chan error, 1)
	go func() {
		done <- watchLoop(tailwind.New(), "", os.DevNull, dir, true)
	}()

	// Close the write end to signal stdin closure.
	pw.Close()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("watchLoop returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("watchLoop did not exit after stdin closed")
	}
}

func TestWatchLoopAlwaysModeIgnoresStdinClose(t *testing.T) {
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	origStdin := os.Stdin
	os.Stdin = pr
	defer func() { os.Stdin = origStdin }()

	dir := t.TempDir()

	done := make(chan error, 1)
	go func() {
		done <- watchLoop(tailwind.New(), "", os.DevNull, dir, false)
	}()

	// Close stdin - should NOT cause exit in always mode.
	pw.Close()

	select {
	case <-done:
		t.Fatal("watchLoop should not exit in always mode when stdin closes")
	case <-time.After(300 * time.Millisecond):
		// Good - still running.
	}
}

func TestDebounceCoalescesChanges(t *testing.T) {
	dir := t.TempDir()
	outFile := filepath.Join(dir, "out.css")

	// Write an initial HTML file.
	htmlFile := filepath.Join(dir, "index.html")
	os.WriteFile(htmlFile, []byte(`<div class="flex"></div>`), 0644)

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStdin := os.Stdin
	os.Stdin = pr
	defer func() { os.Stdin = origStdin }()

	done := make(chan error, 1)
	go func() {
		done <- watchLoop(tailwind.New(), "", outFile, dir, true)
	}()

	// Wait for watcher to set up.
	time.Sleep(200 * time.Millisecond)

	// Simulate rapid successive writes.
	for i := 0; i < 5; i++ {
		os.WriteFile(htmlFile, []byte(`<div class="flex block"></div>`), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait for debounce to fire (100ms + some margin).
	time.Sleep(300 * time.Millisecond)

	// Verify output was written.
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Error("expected output file to be written after debounced rebuild")
	}

	// Clean up.
	pw.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("watchLoop did not exit")
	}
}
