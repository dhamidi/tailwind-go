package main

import (
	"fmt"
	"log"
	"strings"

	tailwind "github.com/dhamidi/tailwind-go"
)

func main() {
	// === Tutorial: Complete example ===
	fmt.Println("=== Tutorial: Custom theme ===")
	engine := tailwind.New()

	err := engine.LoadCSS([]byte(`
@theme {
  --color-brand-50:  oklch(97% 0.01 250);
  --color-brand-100: oklch(93% 0.03 250);
  --color-brand-200: oklch(87% 0.06 250);
  --color-brand-300: oklch(78% 0.10 250);
  --color-brand-400: oklch(68% 0.14 250);
  --color-brand-500: oklch(58% 0.18 250);
  --color-brand-600: oklch(50% 0.16 250);
  --color-brand-700: oklch(42% 0.13 250);
  --color-brand-800: oklch(34% 0.10 250);
  --color-brand-900: oklch(27% 0.07 250);
  --color-brand-950: oklch(20% 0.04 250);
  --font-display: "Cal Sans", "Inter", sans-serif;
}
`))
	if err != nil {
		log.Fatal("LoadCSS failed: ", err)
	}

	engine.Write([]byte(`
<header class="bg-brand-500 text-white font-display p-4">
  <h1 class="text-2xl">Welcome</h1>
</header>
<main class="bg-brand-50 p-8">
  <p class="text-brand-900">Hello, custom theme!</p>
</main>
`))

	css := engine.CSS()
	fmt.Println(css)

	for _, want := range []string{"bg-brand-500", "text-brand-900", "font-display", "bg-brand-50"} {
		class := "." + strings.ReplaceAll(want, "/", `\/`)
		if !strings.Contains(css, class) {
			fmt.Printf("MISSING: expected %s in output\n", want)
		} else {
			fmt.Printf("OK: found %s\n", want)
		}
	}

	// === How-To: Custom breakpoints ===
	fmt.Println("\n=== Custom breakpoints ===")
	e4 := tailwind.New()
	e4.LoadCSS([]byte(`
@theme {
  --breakpoint-xs: 30rem;
  --breakpoint-3xl: 120rem;
}
`))
	e4.Write([]byte(`<div class="xs:flex 3xl:grid">`))
	css4 := e4.CSS()
	if strings.Contains(css4, "30rem") {
		fmt.Println("OK: xs breakpoint present")
	} else {
		fmt.Println("FAIL: xs breakpoint NOT found")
		fmt.Println(css4)
	}

	// === Opacity modifiers ===
	fmt.Println("\n=== Opacity modifiers ===")
	e7 := tailwind.New()
	e7.LoadCSS([]byte(`
@theme {
  --color-brand-500: oklch(58% 0.18 250);
}
`))
	e7.Write([]byte(`<div class="bg-brand-500/50">`))
	css7 := e7.CSS()
	fmt.Println(css7)
}
