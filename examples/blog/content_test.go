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

func TestStore_EditDraft_PreservesDateAndRemovesOldFile(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	draftsDir := filepath.Join(dir, "drafts")
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create original draft with a specific date
	content := "# Original Title\ntags: test\n\nOriginal body.\n"
	if err := os.WriteFile(filepath.Join(draftsDir, "2026-03-01-original-title.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate editing: look up original to get date, delete old, save new
	orig, err := s.GetPost("original-title")
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}
	if orig.Date.Format("2006-01-02") != "2026-03-01" {
		t.Fatalf("expected date 2026-03-01, got %s", orig.Date.Format("2006-01-02"))
	}

	// Delete old file
	if err := s.deleteBySlug("drafts", "original-title"); err != nil {
		t.Fatalf("deleteBySlug failed: %v", err)
	}

	// Save updated draft with new title but preserved date
	updated := Post{
		Slug:      "updated-title",
		Title:     "Updated Title",
		Body:      "# Updated Title\ntags: test\n\nUpdated body.\n",
		Date:      orig.Date,
		Tags:      []string{"test"},
		Published: false,
	}
	if err := s.SavePost(updated); err != nil {
		t.Fatalf("SavePost failed: %v", err)
	}

	// Verify only one draft file exists
	entries, _ := os.ReadDir(draftsDir)
	mdFiles := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".md" {
			mdFiles++
		}
	}
	if mdFiles != 1 {
		t.Errorf("expected 1 draft file, got %d", mdFiles)
	}

	// Verify the new file uses the original date
	drafts, _ := s.ListDrafts()
	if len(drafts) != 1 {
		t.Fatalf("expected 1 draft, got %d", len(drafts))
	}
	if drafts[0].Slug != "updated-title" {
		t.Errorf("expected slug updated-title, got %s", drafts[0].Slug)
	}
	if drafts[0].Date.Format("2006-01-02") != "2026-03-01" {
		t.Errorf("expected preserved date 2026-03-01, got %s", drafts[0].Date.Format("2006-01-02"))
	}
}

func TestStore_EditDraft_SameTitle_NoDuplicates(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)

	draftsDir := filepath.Join(dir, "drafts")
	if err := os.MkdirAll(draftsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create original draft
	content := "# Test Draft\n\nHello world.\n"
	if err := os.WriteFile(filepath.Join(draftsDir, "2026-03-01-test-draft.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate editing with same title: get original, delete, save
	orig, err := s.GetPost("test-draft")
	if err != nil {
		t.Fatalf("GetPost failed: %v", err)
	}

	s.deleteBySlug("drafts", "test-draft")

	updated := Post{
		Slug:      "test-draft",
		Title:     "Test Draft",
		Body:      "# Test Draft\n\nUpdated content.\n",
		Date:      orig.Date,
		Published: false,
	}
	if err := s.SavePost(updated); err != nil {
		t.Fatalf("SavePost failed: %v", err)
	}

	// Verify exactly one draft
	drafts, _ := s.ListDrafts()
	if len(drafts) != 1 {
		t.Fatalf("expected 1 draft, got %d", len(drafts))
	}
	if drafts[0].Date.Format("2006-01-02") != "2026-03-01" {
		t.Errorf("expected preserved date 2026-03-01, got %s", drafts[0].Date.Format("2006-01-02"))
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
