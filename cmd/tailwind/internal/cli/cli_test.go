package cli

import (
	"bytes"
	"strings"
	"testing"
)

func setupCommand() (*Command, *string, *string, *string, *bool, *bool, *string, *string) {
	c := NewCommand("tailwindcss", "TailwindCSS CLI")
	input := c.StringFlag("i", "input", "", "Input file")
	output := c.StringFlag("o", "output", "-", "Output file")
	watch := c.OptionalFlag("w", "watch", "", "Watch for changes and rebuild as needed")
	minify := c.BoolFlag("m", "minify", false, "Optimize and minify the output")
	optimize := c.BoolFlag("", "optimize", false, "Optimize the output without minifying")
	cwd := c.StringFlag("", "cwd", ".", "The current working directory")
	mapFlag := c.StringFlag("", "map", "false", "Generate a source map")
	return c, input, output, watch, minify, optimize, cwd, mapFlag
}

func TestParseLongFlags(t *testing.T) {
	c, input, output, _, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"--input", "foo.css", "--output", "bar.css"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *input != "foo.css" {
		t.Errorf("input = %q, want %q", *input, "foo.css")
	}
	if *output != "bar.css" {
		t.Errorf("output = %q, want %q", *output, "bar.css")
	}
}

func TestParseShortFlags(t *testing.T) {
	c, input, output, _, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"-i", "foo.css", "-o", "bar.css"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *input != "foo.css" {
		t.Errorf("input = %q, want %q", *input, "foo.css")
	}
	if *output != "bar.css" {
		t.Errorf("output = %q, want %q", *output, "bar.css")
	}
}

func TestParseWatchNoValue(t *testing.T) {
	c, _, _, watch, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"--watch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *watch != "" {
		t.Errorf("watch = %q, want %q", *watch, "")
	}
}

func TestParseWatchWithValue(t *testing.T) {
	c, _, _, watch, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"--watch=always"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *watch != "always" {
		t.Errorf("watch = %q, want %q", *watch, "always")
	}
}

func TestParseMinifyShort(t *testing.T) {
	c, _, _, _, minify, _, _, _ := setupCommand()
	err := c.Parse([]string{"-m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !*minify {
		t.Error("minify = false, want true")
	}
}

func TestParseDefaults(t *testing.T) {
	c, input, output, _, minify, _, cwd, _ := setupCommand()
	err := c.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *input != "" {
		t.Errorf("input = %q, want %q", *input, "")
	}
	if *output != "-" {
		t.Errorf("output = %q, want %q", *output, "-")
	}
	if *minify {
		t.Error("minify = true, want false")
	}
	if *cwd != "." {
		t.Errorf("cwd = %q, want %q", *cwd, ".")
	}
}

func TestParseHelp(t *testing.T) {
	c, _, _, _, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"--help"})
	if err != ErrHelp {
		t.Errorf("err = %v, want ErrHelp", err)
	}
}

func TestParseHelpShort(t *testing.T) {
	c, _, _, _, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"-h"})
	if err != ErrHelp {
		t.Errorf("err = %v, want ErrHelp", err)
	}
}

func TestParseUnknownFlag(t *testing.T) {
	c, _, _, _, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"--foo"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("error = %q, want it to contain 'unknown flag'", err.Error())
	}
}

func TestHelpOutputFormatting(t *testing.T) {
	c, _, _, _, _, _, _, _ := setupCommand()
	var buf bytes.Buffer
	c.PrintHelp(&buf)
	output := buf.String()

	if !strings.Contains(output, "Usage:") {
		t.Error("help output missing 'Usage:'")
	}
	if !strings.Contains(output, "Options:") {
		t.Error("help output missing 'Options:'")
	}
	if !strings.Contains(output, "-i, --input") {
		t.Error("help output missing '-i, --input'")
	}
	if !strings.Contains(output, "-o, --output") {
		t.Error("help output missing '-o, --output'")
	}

	// Check alignment: all description dots should start at the same column
	lines := strings.Split(output, "\n")
	var dotPositions []int
	for _, line := range lines {
		idx := strings.Index(line, "··")
		if idx >= 0 {
			dotPositions = append(dotPositions, idx)
		}
	}
	if len(dotPositions) > 1 {
		for i := 1; i < len(dotPositions); i++ {
			if dotPositions[i] != dotPositions[0] {
				t.Errorf("help columns not aligned: position %d vs %d", dotPositions[i], dotPositions[0])
				break
			}
		}
	}
}

func TestParseMixedShortLong(t *testing.T) {
	c, input, output, _, minify, _, _, _ := setupCommand()
	err := c.Parse([]string{"-i", "foo.css", "--output", "bar.css", "-m"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *input != "foo.css" {
		t.Errorf("input = %q, want %q", *input, "foo.css")
	}
	if *output != "bar.css" {
		t.Errorf("output = %q, want %q", *output, "bar.css")
	}
	if !*minify {
		t.Error("minify = false, want true")
	}
}

func TestParseStringFlagMissingValue(t *testing.T) {
	c, _, _, _, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"--input"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
	if !strings.Contains(err.Error(), "requires a value") {
		t.Errorf("error = %q, want it to contain 'requires a value'", err.Error())
	}
}

func TestWatchShortFlag(t *testing.T) {
	c, _, _, watch, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"-w"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *watch != "" {
		t.Errorf("watch = %q, want %q (empty string = truthy)", *watch, "")
	}
}

func TestWatchShortFlagWithEquals(t *testing.T) {
	c, _, _, watch, _, _, _, _ := setupCommand()
	err := c.Parse([]string{"-w=always"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *watch != "always" {
		t.Errorf("watch = %q, want %q", *watch, "always")
	}
}
