package tailwind

import "testing"

func TestIsTextFile_ASCII(t *testing.T) {
	if !isTextFile([]byte("hello world\nfoo bar")) {
		t.Fatal("expected ASCII text to be detected as text")
	}
}

func TestIsTextFile_UTF8(t *testing.T) {
	if !isTextFile([]byte("héllo wörld 你好世界")) {
		t.Fatal("expected UTF-8 text to be detected as text")
	}
}

func TestIsTextFile_NullBytes(t *testing.T) {
	if isTextFile([]byte("hello\x00world")) {
		t.Fatal("expected content with null bytes to be detected as binary")
	}
}

func TestIsTextFile_PNGHeader(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	if isTextFile(png) {
		t.Fatal("expected PNG header to be detected as binary")
	}
}

func TestIsTextFile_Empty(t *testing.T) {
	if !isTextFile([]byte{}) {
		t.Fatal("expected empty input to be detected as text")
	}
}

func TestIsTextFile_HighControlRatio(t *testing.T) {
	// 20 bytes, 4 of which are non-text control chars (20%)
	data := []byte{0x01, 0x02, 0x03, 0x04, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p'}
	if isTextFile(data) {
		t.Fatal("expected high control-char ratio to be detected as binary")
	}
}

func TestHasTextExtension_TextFiles(t *testing.T) {
	textPaths := []string{
		"index.html", "main.go", "app.tsx", "page.jsx",
		"App.vue", "Header.svelte", "page.templ", "README.md",
		"image.svg",
	}
	for _, p := range textPaths {
		if !hasTextExtension(p) {
			t.Errorf("expected %q to be classified as text", p)
		}
	}
}

func TestHasTextExtension_BinaryFiles(t *testing.T) {
	binaryPaths := []string{
		"photo.png", "font.woff2", "program.exe",
		"archive.zip", "lib.so", "style.pdf",
	}
	for _, p := range binaryPaths {
		if hasTextExtension(p) {
			t.Errorf("expected %q to be classified as binary", p)
		}
	}
}

func TestHasTextExtension_Unknown(t *testing.T) {
	if !hasTextExtension("data.xyz") {
		t.Fatal("expected unknown extension to default to text")
	}
}

func TestHasTextExtension_NoExtension(t *testing.T) {
	if !hasTextExtension("Makefile") {
		t.Fatal("expected file with no extension to default to text")
	}
}
