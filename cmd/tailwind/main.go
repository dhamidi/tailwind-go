package main

import (
	"fmt"
	"os"

	tailwind "github.com/dhamidi/tailwind-go"
	"github.com/dhamidi/tailwind-go/cmd/tailwind/internal/cli"
	"github.com/dhamidi/tailwind-go/cmd/tailwind/internal/minify"
)

const version = "0.2.0"

// watchOff is the sentinel default for the --watch flag when not provided.
const watchOff = "\x00"

func main() {
	cmd := cli.NewCommand("tailwind", "Tailwind CSS compiler (Go)")

	input := cmd.StringFlag("i", "input", "", "Input file")
	output := cmd.StringFlag("o", "output", "-", "Output file")
	cwd := cmd.StringFlag("", "cwd", ".", "The current working directory")
	watch := cmd.OptionalFlag("w", "watch", watchOff, "Watch for changes and rebuild")
	doMinify := cmd.BoolFlag("m", "minify", false, "Minify output")
	doOptimize := cmd.BoolFlag("", "optimize", false, "Optimize output")

	if err := cmd.Parse(os.Args[1:]); err != nil {
		if err == cli.ErrHelp {
			cmd.PrintHelp(os.Stdout)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "≈ tailwind v%s\n", version)

	if err := run(*input, *output, *cwd, *watch, *doMinify || *doOptimize); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(input, output, cwd, watch string, shouldMinify bool) error {
	engine := tailwind.New()

	// Initial build.
	if err := build(engine, input, output, cwd, shouldMinify); err != nil {
		return err
	}

	// If watch mode is active, enter the watch loop.
	if watch != watchOff {
		stdinAware := watch != "always"
		return watchLoop(engine, input, output, cwd, stdinAware)
	}

	return nil
}

func build(engine *tailwind.Engine, input, output, cwd string, shouldMinify bool) error {
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

	// Minify if requested
	if shouldMinify {
		css = minify.CSS(css)
	}

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
