package main

import (
	"fmt"
	"os"

	tailwind "github.com/dhamidi/tailwind-go"
	"github.com/dhamidi/tailwind-go/cmd/tailwind/internal/cli"
)

const version = "0.1.0"

func main() {
	cmd := cli.NewCommand("tailwind", "Tailwind CSS compiler (Go)")

	input := cmd.StringFlag("i", "input", "", "Input file")
	output := cmd.StringFlag("o", "output", "-", "Output file")
	cwd := cmd.StringFlag("", "cwd", ".", "The current working directory")
	cmd.BoolFlag("w", "watch", false, "Watch for changes")
	cmd.BoolFlag("m", "minify", false, "Minify output")
	cmd.BoolFlag("", "optimize", false, "Optimize output")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		if err == cli.ErrHelp {
			cmd.PrintHelp(os.Stdout)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "≈ tailwind v%s\n", version)

	if err := run(*input, *output, *cwd); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(input, output, cwd string) error {
	engine := tailwind.New()

	// Load custom CSS if input file provided
	if input != "" {
		css, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		if err := engine.LoadCSS(css); err != nil {
			return fmt.Errorf("parsing input CSS: %w", err)
		}
	}

	// Scan working directory
	fsys := os.DirFS(cwd)
	if err := engine.Scan(fsys); err != nil {
		return fmt.Errorf("scanning %s: %w", cwd, err)
	}

	// Generate CSS
	css := engine.FullCSS()

	// Write output
	if output == "-" {
		fmt.Print(css)
	} else {
		if err := os.WriteFile(output, []byte(css), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
	}

	return nil
}
