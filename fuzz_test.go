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
	f.Add("bg-[rgb(255,0,0)]")
	f.Add("grid-cols-[1fr_auto_2fr]")
	f.Add("text-[length:--my-size]")
	f.Add("bg-(--my-color)")
	f.Add("[&:nth-child(3)]:bg-red-500")
	f.Add("w-[calc(min(100%,_50rem)_-_2rem)]")
	f.Add("p-[var(--x,_var(--y,_1rem))]")
	f.Add("w-[")
	f.Add("]text-red")
	f.Add("w-[]")
	f.Add("w-[[[")
	f.Add("bg-blue-500/[.5]")
	f.Add("text-red-500/[0.75]")
	f.Add("border-green-300/[var(--opacity)]")
	f.Add("[@media(min-width:900px)]:p-4")
	f.Add("[@supports(display:grid)]:grid")
	f.Add("[&>svg]:fill-current")
	f.Add("[&:not(:first-child)]:mt-4")

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

// FuzzEngineMultiWrite ensures the engine produces identical output regardless
// of how input bytes are chunked across multiple Write calls.
func FuzzEngineMultiWrite(f *testing.F) {
	f.Add([]byte(`<div class="flex items-center p-4">`), uint8(5))
	f.Add([]byte(`<div class="bg-blue-500 hover:bg-blue-700 text-white">`), uint8(1))
	f.Add([]byte(`<div class="w-[calc(100%-2rem)] md:grid-cols-3">`), uint8(3))
	f.Add([]byte(`<div class="dark:md:hover:!-translate-x-1/2">`), uint8(7))
	f.Add([]byte{}, uint8(0))

	f.Fuzz(func(t *testing.T, data []byte, splitSeed uint8) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic: %v", r)
			}
		}()

		// Single-write reference.
		ref := New()
		ref.Write(data)
		refCSS := ref.CSS()

		// Multi-write: split data at various points based on seed.
		multi := New()
		if len(data) > 0 {
			// Generate split points from the seed.
			numSplits := int(splitSeed)%5 + 1
			step := len(data) / (numSplits + 1)
			if step == 0 {
				step = 1
			}
			pos := 0
			for i := 0; i < numSplits && pos < len(data); i++ {
				end := pos + step
				if end > len(data) {
					end = len(data)
				}
				multi.Write(data[pos:end])
				pos = end
			}
			if pos < len(data) {
				multi.Write(data[pos:])
			}
		}
		multiCSS := multi.CSS()

		if refCSS != multiCSS {
			t.Errorf("output differs between single-write and multi-write:\nsingle: %q\nmulti:  %q", refCSS, multiCSS)
		}
	})
}

// TestChunkBoundaryConsistency splits known tricky tokens at every possible
// byte boundary and verifies the output matches a single Write call.
func TestChunkBoundaryConsistency(t *testing.T) {
	inputs := []string{
		`<div class="bg-blue-500">`,
		`<div class="hover:bg-blue-500">`,
		`<div class="w-[calc(100%-2rem)]">`,
		`<div class="dark:md:hover:!-translate-x-1/2">`,
		`<div class="[mask-type:alpha]">`,
		`<div class="bg-(--my-color)">`,
		`<div class="text-[length:--my-size]">`,
		`<div class="group-hover/sidebar:flex">`,
	}

	for _, input := range inputs {
		data := []byte(input)
		ref := New()
		ref.Write(data)
		refCSS := ref.CSS()

		// Split at every byte position.
		for i := 0; i <= len(data); i++ {
			multi := New()
			multi.Write(data[:i])
			multi.Write(data[i:])
			multiCSS := multi.CSS()

			if refCSS != multiCSS {
				t.Errorf("split at %d for %q: single=%q multi=%q", i, input, refCSS, multiCSS)
			}
		}
	}
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
