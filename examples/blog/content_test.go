package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_PublishDraft_WithoutDatePrefix(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	// Create drafts directory and a draft without date prefix (like embedded drafts)
	draftsDir := filepath.Join(dir, "drafts")
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	postsDir := filepath.Join(dir, "posts")
	if err := os.MkdirAll(postsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "# Upcoming Features\ntags: news\n\nSome content here.\n"
	if err := os.WriteFile(filepath.Join(draftsDir, "upcoming-features.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Publish should succeed even without date prefix
	if err := s.PublishDraft("upcoming-features"); err != nil {
		t.Fatalf("PublishDraft failed: %v", err)
	}

	// Draft file should be gone
	entries, _ := os.ReadDir(draftsDir)
	for _, e := range entries {
		if e.Name() == "upcoming-features.md" {
			t.Error("draft file should have been deleted")
		}
	}

	// Post file should exist
	postEntries, _ := os.ReadDir(postsDir)
	found := false
	for _, e := range postEntries {
		if filepath.Ext(e.Name()) == ".md" {
			found = true
		}
	}
	if !found {
		t.Error("published post file not found")
	}
}

func TestStore_DeleteBySlug_WithDatePrefix(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	draftsDir := filepath.Join(dir, "drafts")
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "# Test Post\n\nBody.\n"
	if err := os.WriteFile(filepath.Join(draftsDir, "2026-03-01-test-post.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.deleteBySlug("drafts", "test-post"); err != nil {
		t.Fatalf("deleteBySlug failed for date-prefixed file: %v", err)
	}

	entries, _ := os.ReadDir(draftsDir)
	if len(entries) != 0 {
		t.Error("file should have been deleted")
	}
}

func TestStore_DeleteBySlug_WithoutDatePrefix(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	draftsDir := filepath.Join(dir, "drafts")
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	content := "# Simple Draft\n\nBody.\n"
	if err := os.WriteFile(filepath.Join(draftsDir, "simple-draft.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := s.deleteBySlug("drafts", "simple-draft"); err != nil {
		t.Fatalf("deleteBySlug failed for non-date-prefixed file: %v", err)
	}

	entries, _ := os.ReadDir(draftsDir)
	if len(entries) != 0 {
		t.Error("file should have been deleted")
	}
}
