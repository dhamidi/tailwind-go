//go:build go1.18

package tailwind

import "testing"

// FuzzTokenizer ensures the CSS tokenizer never panics on arbitrary input.
func FuzzTokenizer(f *testing.F) {
	// Seed corpus from existing test cases.
	f.Add([]byte(`@theme { --color-blue-500: #3b82f6; }`))
	f.Add([]byte(`@utility w-* { width: --value(--spacing); }`))
	f.Add([]byte(`@variant hover (&:hover);`))
	f.Add([]byte(`/* comment */ .class { color: red; }`))
	f.Add([]byte{0x00, 0xFF, 0x80})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic: %v", r)
			}
		}()
		_ = newTokenizer(data).tokenize()
	})
}

// FuzzParser ensures the CSS parser never panics on arbitrary input.
func FuzzParser(f *testing.F) {
	f.Add([]byte(`@theme { --x: 1px; }`))
	f.Add([]byte(`@utility flex { display: flex; }`))
	f.Add([]byte(`@variant hover (&:hover);`))
	f.Add([]byte(`}}}{{{{`))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic: %v", r)
			}
		}()
		_, _ = parseStylesheet(data)
	})
}

// FuzzClassParser ensures the class parser never panics on arbitrary input.
func FuzzClassParser(f *testing.F) {
	f.Add("flex")
	f.Add("bg-blue-500")
	f.Add("dark:md:hover:!-translate-x-1/2")
	f.Add("[mask-type:alpha]")
	f.Add("w-[calc(100%_-_2rem)]")
	f.Add("::::")
	f.Add("")
	f.Add("--")

	f.Fuzz(func(t *testing.T, s string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on %q: %v", s, r)
			}
		}()
		_ = parseClass(s)
	})
}

// FuzzScanner ensures the byte scanner never panics on arbitrary input.
func FuzzScanner(f *testing.F) {
	f.Add([]byte(`<div class="flex items-center">`))
	f.Add([]byte(`class="w-[300px] text-[#ff0000]"`))
	f.Add([]byte{0x00, 0x01, 0xFF})
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic: %v", r)
			}
		}()
		var s scanner
		tokens := s.feed(data)
		_ = tokens
		_ = s.flush()
	})
}

// FuzzEngine ensures the full pipeline never panics.
func FuzzEngine(f *testing.F) {
	f.Add([]byte(`<div class="flex items-center p-4">`))
	f.Add([]byte(`random bytes that are not HTML`))
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic: %v", r)
			}
		}()
		e := New()
		_, _ = e.Write(data)
		_ = e.CSS()
	})
}
