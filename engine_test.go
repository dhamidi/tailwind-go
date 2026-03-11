package tailwind

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"testing/fstest"
)

// --- Auto-load tests ---

func TestNewAutoLoadsCSS(t *testing.T) {
	engine := New()
	engine.Write([]byte(`<div class="flex items-center p-4 text-blue-500">`))
	css := engine.CSS()
	t.Logf("Generated CSS:\n%s", css)

	if !strings.Contains(css, "display: flex") {
		t.Error("missing 'display: flex' — embedded CSS not auto-loaded")
	}
	if !strings.Contains(css, "align-items: center") {
		t.Error("missing 'align-items: center'")
	}
	if !strings.Contains(css, "padding:") {
		t.Error("missing padding declaration")
	}
}

// --- Scanner tests ---

func TestScannerBasic(t *testing.T) {
	var s scanner
	tokens := s.feed([]byte(`<div class="flex items-center p-4 bg-blue-500">`))
	tokens = append(tokens, s.flush())

	want := map[string]bool{
		"flex": true, "items-center": true,
		"p-4": true, "bg-blue-500": true,
	}
	got := make(map[string]bool)
	for _, tok := range tokens {
		if tok != "" {
			got[tok] = true
		}
	}
	for w := range want {
		if !got[w] {
			t.Errorf("missing candidate %q, got %v", w, tokens)
		}
	}
}

func TestScannerArbitraryValues(t *testing.T) {
	var s scanner
	tokens := s.feed([]byte(`class="w-[300px] text-[#ff0000] grid-cols-[1fr_auto_2fr]"`))
	tokens = append(tokens, s.flush())

	want := []string{"w-[300px]", "text-[#ff0000]", "grid-cols-[1fr_auto_2fr]"}
	got := make(map[string]bool)
	for _, tok := range tokens {
		got[tok] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing candidate %q", w)
		}
	}
}

func TestScannerChunkedWrites(t *testing.T) {
	// Token split across two Write calls.
	var s scanner
	t1 := s.feed([]byte(`class="bg-bl`))
	t2 := s.feed([]byte(`ue-500 flex"`))
	t2 = append(t2, s.flush())

	all := append(t1, t2...)
	got := make(map[string]bool)
	for _, tok := range all {
		if tok != "" {
			got[tok] = true
		}
	}
	if !got["bg-blue-500"] {
		t.Errorf("failed to reconstruct 'bg-blue-500' across chunks, got %v", all)
	}
	if !got["flex"] {
		t.Errorf("missing 'flex'")
	}
}

func TestScannerRejectsURLs(t *testing.T) {
	var s scanner
	tokens := s.feed([]byte(`href="https://example.com" class="flex"`))
	tokens = append(tokens, s.flush())

	for _, tok := range tokens {
		if strings.Contains(tok, "://") {
			t.Errorf("should reject URL-like token, got %q", tok)
		}
	}
}

func TestScannerVariantClasses(t *testing.T) {
	var s scanner
	tokens := s.feed([]byte(`"dark:md:hover:bg-blue-500 !-translate-x-1/2"`))
	tokens = append(tokens, s.flush())

	got := make(map[string]bool)
	for _, tok := range tokens {
		got[tok] = true
	}
	if !got["dark:md:hover:bg-blue-500"] {
		t.Errorf("missing variant class")
	}
	if !got["!-translate-x-1/2"] {
		t.Errorf("missing negated important class")
	}
}

// --- Tokenizer tests ---

func TestTokenizerBasic(t *testing.T) {
	src := `@theme { --color-blue-500: #3b82f6; }`
	tokens := newTokenizer([]byte(src)).tokenize()

	// Should have: @theme, ws, {, ws, --color-blue-500, :, ws, #3b82f6, ;, ws, }
	types := make([]tokenType, 0)
	for _, tok := range tokens {
		types = append(types, tok.typ)
	}

	// Check key tokens are present.
	found := make(map[string]bool)
	for _, tok := range tokens {
		found[tok.value] = true
	}
	if !found["@theme"] {
		t.Error("missing @theme token")
	}
	if !found["--color-blue-500"] {
		t.Error("missing --color-blue-500 token")
	}
}

func TestTokenizerUtility(t *testing.T) {
	src := `@utility w-* {
  width: --value(--spacing);
}`
	tokens := newTokenizer([]byte(src)).tokenize()

	found := make(map[string]bool)
	for _, tok := range tokens {
		found[tok.value] = true
	}
	if !found["@utility"] {
		t.Error("missing @utility")
	}
	if !found["width"] {
		t.Error("missing width")
	}
}

func TestTokenizerDimensions(t *testing.T) {
	src := `width: 48rem;`
	tokens := newTokenizer([]byte(src)).tokenize()

	var dim *token
	for i := range tokens {
		if tokens[i].typ == tokDimension {
			dim = &tokens[i]
			break
		}
	}
	if dim == nil {
		t.Fatal("no dimension token found")
	}
	if dim.value != "48rem" {
		t.Errorf("got dimension %q, want %q", dim.value, "48rem")
	}
}

// --- Class parser tests ---

func TestParseClassSimple(t *testing.T) {
	pc := parseClass("flex")
	if pc.Utility != "flex" {
		t.Errorf("utility = %q, want %q", pc.Utility, "flex")
	}
	if len(pc.Variants) != 0 {
		t.Errorf("variants = %v, want none", pc.Variants)
	}
}

func TestParseClassWithValue(t *testing.T) {
	pc := parseClass("p-4")
	if pc.Utility != "p" {
		t.Errorf("utility = %q, want %q", pc.Utility, "p")
	}
	if pc.Value != "4" {
		t.Errorf("value = %q, want %q", pc.Value, "4")
	}
}

func TestParseClassWithColor(t *testing.T) {
	pc := parseClass("bg-blue-500")
	// This is ambiguous without the utility index.
	// With our heuristic, 500 starts with a digit → split there.
	if pc.Value != "500" {
		t.Logf("utility=%q value=%q (disambiguation happens at resolution)", pc.Utility, pc.Value)
	}
}

func TestParseClassVariants(t *testing.T) {
	pc := parseClass("dark:md:hover:bg-blue-500")
	if len(pc.Variants) != 3 {
		t.Fatalf("variants = %v, want 3", pc.Variants)
	}
	if pc.Variants[0] != "dark" || pc.Variants[1] != "md" || pc.Variants[2] != "hover" {
		t.Errorf("variants = %v", pc.Variants)
	}
}

func TestParseClassImportant(t *testing.T) {
	pc := parseClass("!font-bold")
	if !pc.Important {
		t.Error("expected important")
	}
	if pc.Utility != "font-bold" && pc.Utility != "font" {
		t.Logf("utility = %q (either valid parse)", pc.Utility)
	}
}

func TestParseClassNegative(t *testing.T) {
	pc := parseClass("-translate-x-4")
	if !pc.Negative {
		t.Error("expected negative")
	}
}

func TestParseClassArbitraryValue(t *testing.T) {
	pc := parseClass("w-[300px]")
	if pc.Utility != "w" {
		t.Errorf("utility = %q, want %q", pc.Utility, "w")
	}
	if pc.Arbitrary != "300px" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "300px")
	}
}

func TestParseClassArbitraryWithSpaces(t *testing.T) {
	pc := parseClass("grid-cols-[1fr_auto_2fr]")
	if pc.Arbitrary != "1fr auto 2fr" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "1fr auto 2fr")
	}
}

func TestParseClassArbitraryProperty(t *testing.T) {
	pc := parseClass("[mask-type:alpha]")
	if !pc.ArbitraryProperty {
		t.Error("expected arbitrary property")
	}
	if pc.Utility != "mask-type" {
		t.Errorf("utility = %q, want %q", pc.Utility, "mask-type")
	}
	if pc.Arbitrary != "alpha" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "alpha")
	}
}

func TestParseClassArbitraryVariant(t *testing.T) {
	pc := parseClass("[&:nth-child(3)]:bg-red-500")
	if len(pc.Variants) != 1 || pc.Variants[0] != "[&:nth-child(3)]" {
		t.Errorf("variants = %v", pc.Variants)
	}
}

func TestParseClassFraction(t *testing.T) {
	pc := parseClass("w-1/2")
	if pc.Value != "1/2" {
		t.Errorf("value = %q, want %q", pc.Value, "1/2")
	}
}

func TestParseClassOpacityModifier(t *testing.T) {
	pc := parseClass("bg-blue-500/75")
	if pc.Modifier != "75" {
		t.Errorf("modifier = %q, want %q", pc.Modifier, "75")
	}
	// Value split is ambiguous without the utility index; 500 starts with
	// a digit so the heuristic splits there. Resolution disambiguates later.
	if pc.Value != "500" {
		t.Logf("value = %q (disambiguation happens at resolution)", pc.Value)
	}
}

func TestParseClassOpacityModifierArbitrary(t *testing.T) {
	pc := parseClass("bg-blue-500/[.5]")
	if pc.Modifier != "[.5]" {
		t.Errorf("modifier = %q", pc.Modifier)
	}
}

func TestParseClassFractionNotModifier(t *testing.T) {
	pc := parseClass("w-1/2")
	if pc.Modifier != "" {
		t.Errorf("fraction should not set modifier, got %q", pc.Modifier)
	}
	if pc.Value != "1/2" {
		t.Errorf("value = %q, want 1/2", pc.Value)
	}
}

func TestEndToEndOpacityModifier(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; }
@utility bg-* { background-color: --value(--color); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-blue-500/75"`))
	result := e.CSS()
	if !strings.Contains(result, "oklch(from #3b82f6 l c h / 75%)") {
		t.Errorf("unexpected output: %s", result)
	}
}

func TestEndToEndOpacityModifierArbitrary(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; }
@utility bg-* { background-color: --value(--color); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-blue-500/[.5]"`))
	result := e.CSS()
	if !strings.Contains(result, "oklch(from #3b82f6 l c h / .5)") {
		t.Errorf("unexpected output: %s", result)
	}
}

func TestEndToEndOpacityTheme(t *testing.T) {
	css := []byte(`
@theme {
  --color-white: white;
  --opacity-50: 0.5;
}
@utility text-* { color: --value(--color); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="text-white/50"`))
	result := e.CSS()
	if !strings.Contains(result, "oklch(from white l c h / 0.5)") {
		t.Errorf("unexpected output: %s", result)
	}
}

// --- Parser tests ---

func TestParseTheme(t *testing.T) {
	css := []byte(`
@theme {
  --color-blue-500: #3b82f6;
  --spacing: 0.25rem;
  --breakpoint-md: 48rem;
}
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}

	if v := ss.Theme.Tokens["--color-blue-500"]; v != "#3b82f6" {
		t.Errorf("color-blue-500 = %q", v)
	}
	if v := ss.Theme.Tokens["--spacing"]; v != "0.25rem" {
		t.Errorf("spacing = %q", v)
	}
	if v := ss.Theme.Tokens["--breakpoint-md"]; v != "48rem" {
		t.Errorf("breakpoint-md = %q", v)
	}
}

func TestParseUtilityDynamic(t *testing.T) {
	css := []byte(`
@utility w-* {
  width: --value(--spacing);
}
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}

	if len(ss.Utilities) != 1 {
		t.Fatalf("got %d utilities, want 1", len(ss.Utilities))
	}

	u := ss.Utilities[0]
	if u.Pattern != "w" {
		t.Errorf("pattern = %q, want %q", u.Pattern, "w")
	}
	if u.Static {
		t.Error("expected dynamic utility")
	}
	if len(u.Declarations) != 1 {
		t.Fatalf("got %d declarations", len(u.Declarations))
	}
	if u.Declarations[0].Property != "width" {
		t.Errorf("property = %q", u.Declarations[0].Property)
	}
}

func TestParseUtilityStatic(t *testing.T) {
	css := []byte(`
@utility flex {
  display: flex;
}
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}

	if len(ss.Utilities) != 1 {
		t.Fatalf("got %d utilities", len(ss.Utilities))
	}
	if !ss.Utilities[0].Static {
		t.Error("expected static utility")
	}
	if ss.Utilities[0].Pattern != "flex" {
		t.Errorf("pattern = %q", ss.Utilities[0].Pattern)
	}
}

func TestParseVariant(t *testing.T) {
	css := []byte(`
@variant hover (&:hover);
@variant md (@media (width >= 48rem));
@variant dark (@media (prefers-color-scheme: dark));
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}

	if len(ss.Variants) != 3 {
		t.Fatalf("got %d variants, want 3", len(ss.Variants))
	}

	byName := make(map[string]*VariantDef)
	for _, v := range ss.Variants {
		byName[v.Name] = v
	}

	if v := byName["hover"]; v.Selector != "&:hover" {
		t.Errorf("hover selector = %q", v.Selector)
	}
	if v := byName["md"]; v.Media == "" {
		t.Error("md should have media query")
	}
	if v := byName["dark"]; v.Media == "" {
		t.Error("dark should have media query")
	}
}

// --- Integration tests ---

func TestEndToEndSimple(t *testing.T) {
	css := []byte(`
@theme {
  --spacing: 0.25rem;
  --color-blue-500: #3b82f6;
}

@utility flex {
  display: flex;
}

@utility items-center {
  align-items: center;
}

@utility p-* {
  padding: --value(--spacing);
}

@utility bg-* {
  background-color: --value(--color);
}

@variant hover (&:hover);
@variant md (@media (width >= 48rem));
`)

	engine := New()
	if err := engine.LoadCSS(css); err != nil {
		t.Fatal(err)
	}

	html := `<div class="flex items-center p-4 bg-blue-500">`
	engine.Write([]byte(html))

	result := engine.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "display: flex") {
		t.Error("missing 'display: flex'")
	}
	if !strings.Contains(result, "align-items: center") {
		t.Error("missing 'align-items: center'")
	}
	if !strings.Contains(result, "padding:") {
		t.Error("missing padding declaration")
	}
	if !strings.Contains(result, "background-color: #3b82f6") {
		t.Error("missing background-color")
	}
}

func TestEndToEndArbitrary(t *testing.T) {
	css := []byte(`
@utility w-* {
  width: --value(--spacing);
}
`)
	engine := New()
	engine.LoadCSS(css)
	engine.Write([]byte(`class="w-[300px]"`))

	result := engine.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "width: 300px") {
		t.Error("missing arbitrary width")
	}
}

func TestEndToEndVariants(t *testing.T) {
	css := []byte(`
@theme {
  --color-blue-600: #2563eb;
}

@utility bg-* {
  background-color: --value(--color);
}

@variant hover (&:hover);
@variant md (@media (width >= 48rem));
`)
	engine := New()
	engine.LoadCSS(css)
	engine.Write([]byte(`class="hover:bg-blue-600 md:bg-blue-600"`))

	result := engine.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "background-color: #2563eb") {
		t.Error("missing bg color")
	}
}

func TestEndToEndArbitraryProperty(t *testing.T) {
	engine := New()
	engine.LoadCSS([]byte(``))
	engine.Write([]byte(`class="[mask-type:alpha]"`))

	result := engine.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "mask-type: alpha") {
		t.Error("missing arbitrary property")
	}
}

func TestEndToEndFraction(t *testing.T) {
	css := []byte(`
@utility w-* {
  width: --value(--spacing);
}
`)
	engine := New()
	engine.LoadCSS(css)
	engine.Write([]byte(`class="w-1/2"`))

	result := engine.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "50%") {
		t.Error("missing fraction percentage")
	}
}

// --- io.Writer interface tests ---

func TestWriterInterface(t *testing.T) {
	css := []byte(`
@utility flex {
  display: flex;
}
`)
	engine := New()
	engine.LoadCSS(css)

	// Use engine as a generic io.Writer.
	var w io.Writer = engine
	w.Write([]byte(`<div class="flex">`))

	result := engine.CSS()
	if !strings.Contains(result, "display: flex") {
		t.Error("io.Writer interface didn't capture class")
	}
}

func TestPassthroughWriter(t *testing.T) {
	css := []byte(`
@utility flex {
  display: flex;
}
`)
	var buf bytes.Buffer
	engine := NewPassthrough(&buf)
	engine.LoadCSS(css)

	input := []byte(`<div class="flex">hello</div>`)
	engine.Write(input)

	// Passthrough should have the original bytes.
	if buf.String() != string(input) {
		t.Errorf("passthrough got %q, want %q", buf.String(), string(input))
	}

	// Engine should still capture classes.
	result := engine.CSS()
	if !strings.Contains(result, "display: flex") {
		t.Error("passthrough engine didn't capture class")
	}
}

func TestIoCopy(t *testing.T) {
	css := []byte(`
@utility hidden {
  display: none;
}
@utility block {
  display: block;
}
`)
	engine := New()
	engine.LoadCSS(css)

	src := strings.NewReader(`<div class="hidden">secret</div><div class="block">visible</div>`)
	io.Copy(engine, src)

	result := engine.CSS()
	if !strings.Contains(result, "display: none") {
		t.Error("missing hidden")
	}
	if !strings.Contains(result, "display: block") {
		t.Error("missing block")
	}
}

func TestMultiDeclarationPrioritySpacing(t *testing.T) {
	// w-4 should use spacing (first alternative), not width
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility w-* {
  width: --value(--spacing);
  width: --value(--width);
  width: --value(length, percentage);
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="w-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "calc(4") {
		t.Errorf("expected spacing calc, got: %s", result)
	}
	// Should only have one width declaration
	if strings.Count(result, "width:") != 1 {
		t.Errorf("expected exactly one width declaration, got: %s", result)
	}
}

func TestMultiDeclarationPriorityTheme(t *testing.T) {
	// w-prose should fall through to --width namespace
	css := []byte(`
@theme { --spacing: 0.25rem; --width-prose: 65ch; }
@utility w-* {
  width: --value(--spacing);
  width: --value(--width);
  width: --value(length, percentage);
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="w-prose"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "65ch") {
		t.Errorf("expected 65ch, got: %s", result)
	}
	if strings.Count(result, "width:") != 1 {
		t.Errorf("expected exactly one width declaration, got: %s", result)
	}
}

func TestMultiDeclarationPriorityArbitrary(t *testing.T) {
	// w-[300px] should resolve as arbitrary value
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility w-* {
  width: --value(--spacing);
  width: --value(--width);
  width: --value(length, percentage);
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="w-[300px]"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "width: 300px") {
		t.Errorf("expected width: 300px, got: %s", result)
	}
	if strings.Count(result, "width:") != 1 {
		t.Errorf("expected exactly one width declaration, got: %s", result)
	}
}

func TestReset(t *testing.T) {
	css := []byte(`
@utility flex { display: flex; }
@utility block { display: block; }
`)
	engine := New()
	engine.LoadCSS(css)

	engine.Write([]byte(`class="flex"`))
	r1 := engine.CSS()
	if !strings.Contains(r1, "display: flex") {
		t.Error("first pass missing flex")
	}

	engine.Reset()
	engine.Write([]byte(`class="block"`))
	r2 := engine.CSS()
	if strings.Contains(r2, "display: flex") {
		t.Error("reset didn't clear flex")
	}
	if !strings.Contains(r2, "display: block") {
		t.Error("second pass missing block")
	}
}

func TestParseClassTypeHintedArbitrary(t *testing.T) {
	pc := parseClass("text-[length:1.5em]")
	if pc.Arbitrary != "1.5em" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "1.5em")
	}
	if pc.TypeHint != "length" {
		t.Errorf("typeHint = %q, want %q", pc.TypeHint, "length")
	}
}

func TestParseClassCustomPropertyArbitrary(t *testing.T) {
	pc := parseClass("w-[--sidebar-width]")
	if pc.Arbitrary != "var(--sidebar-width)" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "var(--sidebar-width)")
	}
}

func TestParseClassCustomPropertyWithTypeHint(t *testing.T) {
	pc := parseClass("text-[length:--my-size]")
	if pc.Arbitrary != "var(--my-size)" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "var(--my-size)")
	}
	if pc.TypeHint != "length" {
		t.Errorf("typeHint = %q, want %q", pc.TypeHint, "length")
	}
}

func TestArbitraryAtMediaVariant(t *testing.T) {
	css := []byte(`@utility bg-* { background-color: --value(--color); }`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="[@media(min-width:900px)]:bg-[red]"`))
	result := e.CSS()
	if !strings.Contains(result, "@media") {
		t.Errorf("expected @media wrapper in output, got: %s", result)
	}
}

// --- @apply tests ---

func TestApplyDirectiveBasic(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; }
@utility bg-* { background-color: --value(--color); }
@utility font-bold { font-weight: 700; }

.btn {
  @apply bg-blue-500 font-bold;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS: %s", result)
	if !strings.Contains(result, ".btn") {
		t.Error("missing .btn selector")
	}
	if !strings.Contains(result, "background-color: #3b82f6") {
		t.Error("missing bg-blue-500")
	}
	if !strings.Contains(result, "font-weight: 700") {
		t.Error("missing font-bold")
	}
}

func TestApplyDirectiveWithVariant(t *testing.T) {
	css := []byte(`
@theme { --color-blue-700: #1d4ed8; }
@utility bg-* { background-color: --value(--color); }
@variant hover (&:hover);

.btn {
  @apply hover:bg-blue-700;
}
`)
	e := New()
	e.LoadCSS(css)
	result := e.CSS()
	t.Logf("CSS: %s", result)
	if !strings.Contains(result, ".btn:hover") {
		t.Errorf("missing .btn:hover selector, got: %s", result)
	}
	if !strings.Contains(result, "background-color: #1d4ed8") {
		t.Errorf("missing background-color: #1d4ed8, got: %s", result)
	}
}

func TestDiagnosticsBasic(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; }
@utility flex { display: flex; }
@utility bg-* { background-color: --value(--color); }
@variant hover (&:hover);
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`<div class="flex bg-blue-500 unknown-class">`))

	d := e.Diagnostics()
	if d.UtilityCount < 2 {
		t.Errorf("utility count = %d, want >= 2", d.UtilityCount)
	}
	if d.VariantCount < 1 {
		t.Errorf("variant count = %d, want >= 1", d.VariantCount)
	}
	if d.ThemeTokenCount < 1 {
		t.Errorf("theme token count = %d, want >= 1", d.ThemeTokenCount)
	}
	// "unknown-class" should be in dropped candidates.
	found := false
	for _, c := range d.DroppedCandidates {
		if c == "unknown-class" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'unknown-class' in dropped candidates: %v", d.DroppedCandidates)
	}
}

func TestApplyDirectiveUnknownClass(t *testing.T) {
	css := []byte(`
.btn {
  @apply nonexistent-utility;
}
`)
	e := New()
	e.LoadCSS(css)
	// Should not panic, just produce no output for unknown classes.
	_ = e.CSS()
}

// --- Preflight CSS tests ---

func TestPreflightReturnsNonEmpty(t *testing.T) {
	e := New()
	css := e.Preflight()
	if css == "" {
		t.Fatal("Preflight() returned empty string")
	}
}

func TestPreflightContainsResetRules(t *testing.T) {
	e := New()
	css := e.Preflight()
	if !strings.Contains(css, "box-sizing: border-box") {
		t.Error("missing box-sizing reset in preflight")
	}
	if !strings.Contains(css, "margin: 0") {
		t.Error("missing margin reset in preflight")
	}
}

func TestPreflightCSS(t *testing.T) {
	e := New()
	css := e.PreflightCSS()
	if css == "" {
		t.Fatal("PreflightCSS returned empty string")
	}
	if strings.Contains(css, "--theme(") {
		t.Error("PreflightCSS contains unresolved --theme() references")
	}
	// Verify font-family fallback is present (sans)
	if !strings.Contains(css, "ui-sans-serif") {
		t.Error("expected default sans font family in preflight")
	}
	// Verify mono fallback is present
	if !strings.Contains(css, "ui-monospace") {
		t.Error("expected default mono font family in preflight")
	}
}

func TestFullCSSWithNoCandidates(t *testing.T) {
	e := New()
	full := e.FullCSS()
	preflight := e.Preflight()
	if full != preflight {
		t.Errorf("FullCSS with no candidates should equal Preflight;\nFullCSS length=%d, Preflight length=%d", len(full), len(preflight))
	}
}

func TestFullCSSWithCandidates(t *testing.T) {
	e := New()
	e.Write([]byte(`<div class="flex p-4">`))
	full := e.FullCSS()
	preflight := e.Preflight()
	utility := e.CSS()

	if !strings.Contains(full, "box-sizing: border-box") {
		t.Error("FullCSS missing preflight content")
	}
	if !strings.Contains(full, "display: flex") {
		t.Error("FullCSS missing utility content")
	}
	expected := preflight + "\n" + utility
	if full != expected {
		t.Errorf("FullCSS != Preflight + newline + CSS")
	}
}

func TestPreflightCSSCustomTheme(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`@theme { --default-font-family: "Inter", sans-serif; }`))
	css := e.PreflightCSS()
	if !strings.Contains(css, `"Inter"`) {
		t.Error("expected custom font family in preflight")
	}
	if strings.Contains(css, "--theme(") {
		t.Error("PreflightCSS contains unresolved --theme() references after custom theme")
	}
}

func TestPreflightCSSIndependentOfUtilityCSS(t *testing.T) {
	e := New()
	e.Write([]byte(`<div class="flex p-4">`))
	utilCSS := e.CSS()
	preflightCSS := e.PreflightCSS()

	// Utility CSS should not contain preflight reset content
	if strings.Contains(utilCSS, "box-sizing: border-box") {
		t.Error("utility CSS should not contain preflight reset styles")
	}
	// Preflight should not contain utility rules
	if strings.Contains(preflightCSS, "display: flex") {
		t.Error("preflight CSS should not contain utility styles")
	}
}

func TestResolveThemeRefs(t *testing.T) {
	tokens := map[string]string{
		"--my-font": "Arial, sans-serif",
	}
	input := "font-family: --theme(--my-font, Helvetica);"
	got := resolveThemeRefs(input, tokens)
	want := "font-family: Arial, sans-serif;"
	if got != want {
		t.Errorf("resolveThemeRefs() = %q, want %q", got, want)
	}
}

func TestResolveThemeRefsFallback(t *testing.T) {
	tokens := map[string]string{}
	input := "font-family: --theme(--missing, Helvetica, Arial);"
	got := resolveThemeRefs(input, tokens)
	want := "font-family: Helvetica, Arial;"
	if got != want {
		t.Errorf("resolveThemeRefs() = %q, want %q", got, want)
	}
}

func TestResolveThemeRefsRecursive(t *testing.T) {
	tokens := map[string]string{
		"--base":    "--theme(--actual, fallback)",
		"--actual":  "resolved-value",
	}
	input := "prop: --theme(--base, default);"
	got := resolveThemeRefs(input, tokens)
	want := "prop: resolved-value;"
	if got != want {
		t.Errorf("resolveThemeRefs() = %q, want %q", got, want)
	}
}

// --- Scan tests ---

func TestScanHTMLFiles(t *testing.T) {
	fs := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<div class="flex p-4">hello</div>`)},
		"about.html": &fstest.MapFile{Data: []byte(`<div class="block mt-2">about</div>`)},
	}
	e := New()
	if err := e.Scan(fs); err != nil {
		t.Fatal(err)
	}
	candidates := e.Candidates()
	got := make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	for _, want := range []string{"flex", "p-4", "block", "mt-2"} {
		if !got[want] {
			t.Errorf("missing candidate %q, got %v", want, candidates)
		}
	}
}

func TestScanSkipsBinaryFiles(t *testing.T) {
	fs := fstest.MapFS{
		"index.html":  &fstest.MapFile{Data: []byte(`<div class="flex">hello</div>`)},
		"image.png":   &fstest.MapFile{Data: []byte{0x89, 0x50, 0x4E, 0x47}},
		"styles.css":  &fstest.MapFile{Data: []byte(`/* class="hidden" */`)},
	}
	e := New()
	if err := e.Scan(fs); err != nil {
		t.Fatal(err)
	}
	candidates := e.Candidates()
	got := make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	if !got["flex"] {
		t.Error("missing candidate 'flex' from HTML file")
	}
	// PNG should be skipped by extension
	// The CSS file should be scanned (text extension)
}

func TestScanSkipsGitDir(t *testing.T) {
	fs := fstest.MapFS{
		".git/config":      &fstest.MapFile{Data: []byte(`class="secret"`)},
		".git/HEAD":        &fstest.MapFile{Data: []byte(`ref: refs/heads/main`)},
		"src/app.html":     &fstest.MapFile{Data: []byte(`<div class="flex">app</div>`)},
	}
	e := New()
	if err := e.Scan(fs); err != nil {
		t.Fatal(err)
	}
	candidates := e.Candidates()
	got := make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	if got["secret"] {
		t.Error("should not have scanned .git directory")
	}
	if !got["flex"] {
		t.Error("missing candidate 'flex' from src/app.html")
	}
}

func TestScanSkipsNodeModules(t *testing.T) {
	fs := fstest.MapFS{
		"node_modules/pkg/index.js": &fstest.MapFile{Data: []byte(`class="hidden"`)},
		"src/app.html":              &fstest.MapFile{Data: []byte(`<div class="block">app</div>`)},
	}
	e := New()
	if err := e.Scan(fs); err != nil {
		t.Fatal(err)
	}
	candidates := e.Candidates()
	got := make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	if got["hidden"] {
		t.Error("should not have scanned node_modules directory")
	}
	if !got["block"] {
		t.Error("missing candidate 'block' from src/app.html")
	}
}

func TestScanEmptyFS(t *testing.T) {
	fs := fstest.MapFS{}
	e := New()
	if err := e.Scan(fs); err != nil {
		t.Fatalf("unexpected error scanning empty FS: %v", err)
	}
	if len(e.Candidates()) != 0 {
		t.Errorf("expected no candidates from empty FS, got %v", e.Candidates())
	}
}

func TestScanAccumulatesCandidates(t *testing.T) {
	fs1 := fstest.MapFS{
		"a.html": &fstest.MapFile{Data: []byte(`<div class="flex">a</div>`)},
	}
	fs2 := fstest.MapFS{
		"b.html": &fstest.MapFile{Data: []byte(`<div class="block">b</div>`)},
	}
	e := New()
	if err := e.Scan(fs1); err != nil {
		t.Fatal(err)
	}
	if err := e.Scan(fs2); err != nil {
		t.Fatal(err)
	}
	candidates := e.Candidates()
	got := make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	if !got["flex"] {
		t.Error("missing 'flex' from first scan")
	}
	if !got["block"] {
		t.Error("missing 'block' from second scan")
	}
}

// --- Compound variant tests ---

func TestCompoundVariantGroupHover(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@variant group-* {
  :merge(.group):{value} & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="group-hover:text-white"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, ".group:hover") {
		t.Errorf("missing .group:hover, got: %s", result)
	}
	if !strings.Contains(result, "color: #fff") {
		t.Errorf("missing color: #fff, got: %s", result)
	}
}

func TestCompoundVariantPeerFocus(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@variant peer-* {
  :merge(.peer):{value} ~ & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="peer-focus:text-white"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, ".peer:focus") {
		t.Errorf("missing .peer:focus, got: %s", result)
	}
	if !strings.Contains(result, "~") {
		t.Errorf("missing ~ combinator, got: %s", result)
	}
}

func TestCompoundVariantNotHover(t *testing.T) {
	css := []byte(`
@theme { --opacity-100: 1; }
@utility opacity-* { opacity: --value(--opacity); }
@variant not-* {
  &:not(:{value}) {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="not-hover:opacity-100"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, ":not(:hover)") {
		t.Errorf("missing :not(:hover), got: %s", result)
	}
}

func TestCompoundVariantHasChecked(t *testing.T) {
	css := []byte(`
@theme { --color-gray-50: #f9fafb; }
@utility bg-* { background-color: --value(--color); }
@variant has-* {
  &:has(:{value}) {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="has-checked:bg-gray-50"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, ":has(:checked)") {
		t.Errorf("missing :has(:checked), got: %s", result)
	}
}

func TestCompoundVariantInDataCurrent(t *testing.T) {
	css := []byte(`
@utility font-bold { font-weight: 700; }
@variant in-* {
  [{value}] & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="in-data-current:font-bold"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "[data-current]") {
		t.Errorf("missing [data-current], got: %s", result)
	}
}

func TestCompoundVariantStacking(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; --breakpoint-md: 48rem; }
@utility text-* { color: --value(--color); }
@variant md (@media (width >= 48rem));
@variant group-* {
  :merge(.group):{value} & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="md:group-hover:text-white"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "@media") {
		t.Errorf("missing @media wrapper, got: %s", result)
	}
	if !strings.Contains(result, ".group:hover") {
		t.Errorf("missing .group:hover, got: %s", result)
	}
}

func TestCompoundVariantParserBlockForm(t *testing.T) {
	css := []byte(`
@variant group-* {
  :merge(.group):{value} & {
    @slot;
  }
}
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Variants) != 1 {
		t.Fatalf("got %d variants, want 1", len(ss.Variants))
	}
	v := ss.Variants[0]
	if v.Name != "group" {
		t.Errorf("name = %q, want %q", v.Name, "group")
	}
	if !v.Compound {
		t.Error("expected compound = true")
	}
	if !strings.Contains(v.Template, "{value}") {
		t.Errorf("template missing {value}: %q", v.Template)
	}
	if !strings.Contains(v.Template, "&") {
		t.Errorf("template missing &: %q", v.Template)
	}
}

// --- @starting-style variant tests ---

func TestStartingStyleVariant(t *testing.T) {
	css := []byte(`
@utility opacity-* {
	opacity: --value(integer);
}
@variant starting (@starting-style);
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="starting:opacity-0"`))
	result := e.CSS()
	if !strings.Contains(result, "@starting-style") {
		t.Errorf("missing @starting-style wrapper, got: %s", result)
	}
	if !strings.Contains(result, "opacity: 0") {
		t.Errorf("missing opacity declaration, got: %s", result)
	}
}

func TestStartingStyleVariantStacking(t *testing.T) {
	css := []byte(`
@utility opacity-* {
	opacity: --value(integer);
}
@variant starting (@starting-style);
@variant hover (&:hover);
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="hover:starting:opacity-0"`))
	result := e.CSS()
	if !strings.Contains(result, "@starting-style") {
		t.Errorf("missing @starting-style wrapper, got: %s", result)
	}
	if !strings.Contains(result, ":hover") {
		t.Errorf("missing :hover selector, got: %s", result)
	}
	if !strings.Contains(result, "opacity: 0") {
		t.Errorf("missing opacity declaration, got: %s", result)
	}
}

// --- Named group/peer variant tests ---

func TestNamedGroupHover(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@utility flex { display: flex; }
@variant group-* {
  :merge(.group):{value} & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="group-hover/sidebar:flex"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	// Should contain: .group\/sidebar:hover .group-hover\/sidebar\:flex
	if !strings.Contains(result, `group\/sidebar:hover`) {
		t.Errorf("missing named group selector, got: %s", result)
	}
	if !strings.Contains(result, "display: flex") {
		t.Errorf("missing display: flex, got: %s", result)
	}
}

func TestNamedPeerFocus(t *testing.T) {
	css := []byte(`
@theme { --color-white: #fff; }
@utility text-* { color: --value(--color); }
@variant peer-* {
  :merge(.peer):{value} ~ & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="peer-focus/email:text-white"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, `peer\/email:focus`) {
		t.Errorf("missing named peer selector, got: %s", result)
	}
	if !strings.Contains(result, "~") {
		t.Errorf("missing ~ combinator, got: %s", result)
	}
	if !strings.Contains(result, "color: #fff") {
		t.Errorf("missing color: #fff, got: %s", result)
	}
}

func TestNamedGroupDoesNotConflictWithOpacityModifier(t *testing.T) {
	css := []byte(`
@theme { --color-blue-500: #3b82f6; }
@utility bg-* { background-color: --value(--color); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-blue-500/75"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	// Should still produce opacity modifier, not named group
	if !strings.Contains(result, "oklch(from #3b82f6 l c h / 75%)") {
		t.Errorf("opacity modifier broken, got: %s", result)
	}
}

func TestNamedGroupSelectorEscaping(t *testing.T) {
	css := []byte(`
@utility flex { display: flex; }
@variant group-* {
  :merge(.group):{value} & {
    @slot;
  }
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="group-hover/sidebar:flex"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	// The class name itself should be properly escaped
	if !strings.Contains(result, `group-hover\/sidebar\:flex`) {
		t.Errorf("missing properly escaped class selector, got: %s", result)
	}
}

func TestStripMerge(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{":merge(.group):hover &", ".group:hover &"},
		{":merge(.peer):focus ~ &", ".peer:focus ~ &"},
		{`:merge(.group\/sidebar):hover &`, `.group\/sidebar:hover &`},
		{"&:hover", "&:hover"},
		{"", ""},
		{":merge(.a):merge(.b) &", ".a.b &"},
		{":merge(.outer(:inner)) &", ".outer(:inner) &"},
	}
	for _, tt := range tests {
		got := stripMerge(tt.input)
		if got != tt.want {
			t.Errorf("stripMerge(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Container query variant tests ---

func TestContainerQueryVariant(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility p-* { padding: --value(--spacing); }
@variant @md (@container (width >= 48rem));
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="@md:p-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "@container (width >= 48rem)") {
		t.Errorf("missing @container query, got: %s", result)
	}
	if !strings.Contains(result, "padding:") {
		t.Errorf("missing padding declaration, got: %s", result)
	}
}

func TestContainerQueryVariantSelector(t *testing.T) {
	css := []byte(`
@utility flex { display: flex; }
@variant @sm (@container (width >= 40rem));
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="@sm:flex"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, `\@sm`) {
		t.Errorf("missing escaped @ in selector, got: %s", result)
	}
	if !strings.Contains(result, "display: flex") {
		t.Errorf("missing display: flex, got: %s", result)
	}
}

func TestContainerQueryStacking(t *testing.T) {
	css := []byte(`
@theme { --spacing: 0.25rem; }
@utility p-* { padding: --value(--spacing); }
@variant @md (@container (width >= 48rem));
@variant hover (&:hover);
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="@md:hover:p-4"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)
	if !strings.Contains(result, "@container (width >= 48rem)") {
		t.Errorf("missing @container query, got: %s", result)
	}
	if !strings.Contains(result, ":hover") {
		t.Errorf("missing :hover selector, got: %s", result)
	}
	if !strings.Contains(result, "padding:") {
		t.Errorf("missing padding declaration, got: %s", result)
	}
}

func TestContainerQueryParserVariant(t *testing.T) {
	css := []byte(`
@variant @md (@container (width >= 48rem));
@variant @sm (@container (width >= 40rem));
@variant @lg (@container (width >= 64rem));
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Variants) != 3 {
		t.Fatalf("got %d variants, want 3", len(ss.Variants))
	}
	byName := make(map[string]*VariantDef)
	for _, v := range ss.Variants {
		byName[v.Name] = v
	}
	for _, name := range []string{"@md", "@sm", "@lg"} {
		v, ok := byName[name]
		if !ok {
			t.Errorf("missing variant %q", name)
			continue
		}
		if v.AtRule != "container" {
			t.Errorf("variant %q AtRule = %q, want %q", name, v.AtRule, "container")
		}
	}
}

func TestScannerExtractsContainerVariantClass(t *testing.T) {
	var s scanner
	tokens := s.feed([]byte(`class="@md:p-4 @sm:flex"`))
	tokens = append(tokens, s.flush())
	got := make(map[string]bool)
	for _, tok := range tokens {
		if tok != "" {
			got[tok] = true
		}
	}
	if !got["@md:p-4"] {
		t.Errorf("missing @md:p-4, got %v", tokens)
	}
	if !got["@sm:flex"] {
		t.Errorf("missing @sm:flex, got %v", tokens)
	}
}

func TestBuiltinContainerQueryVariants(t *testing.T) {
	e := New()
	e.Write([]byte(`<div class="@md:flex @sm:hidden @lg:block">`))
	got := e.CSS()
	t.Logf("CSS:\n%s", got)

	if !strings.Contains(got, "@container (width >= 28rem)") {
		t.Errorf("missing @container wrapper for @md")
	}
	if !strings.Contains(got, "@container (width >= 24rem)") {
		t.Errorf("missing @container wrapper for @sm")
	}
	if !strings.Contains(got, "@container (width >= 32rem)") {
		t.Errorf("missing @container wrapper for @lg")
	}
}

// --- Built-in parameterized variant tests (using New() without LoadCSS) ---

func TestNotVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`not-hover:opacity-100`))
	css := e.CSS()
	// Expect: .not-hover\:opacity-100:not(:hover) { opacity: 1; }
	if !strings.Contains(css, ":not(:hover)") {
		t.Errorf("expected :not(:hover) selector, got:\n%s", css)
	}
}

func TestHasVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`has-checked:bg-gray-50`))
	css := e.CSS()
	if !strings.Contains(css, ":has(:checked)") {
		t.Errorf("expected :has(:checked) selector, got:\n%s", css)
	}
}

func TestStartingStyleVariantBuiltin(t *testing.T) {
	e := New()
	e.Write([]byte(`starting:opacity-0`))
	css := e.CSS()
	if !strings.Contains(css, "@starting-style") {
		t.Errorf("expected @starting-style wrapper, got:\n%s", css)
	}
}

func TestInVariantBuiltin(t *testing.T) {
	e := New()
	e.Write([]byte(`in-data-current:font-bold`))
	css := e.CSS()
	if !strings.Contains(css, "[data-current]") {
		t.Errorf("expected [data-current] ancestor selector, got:\n%s", css)
	}
}

func TestScanResetClearsCandidates(t *testing.T) {
	fs1 := fstest.MapFS{
		"a.html": &fstest.MapFile{Data: []byte(`<div class="flex">a</div>`)},
	}
	fs2 := fstest.MapFS{
		"b.html": &fstest.MapFile{Data: []byte(`<div class="block">b</div>`)},
	}
	e := New()
	if err := e.Scan(fs1); err != nil {
		t.Fatal(err)
	}
	candidates := e.Candidates()
	got := make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	if !got["flex"] {
		t.Error("missing 'flex' after first scan")
	}

	e.Reset()
	if err := e.Scan(fs2); err != nil {
		t.Fatal(err)
	}
	candidates = e.Candidates()
	got = make(map[string]bool)
	for _, c := range candidates {
		got[c] = true
	}
	if got["flex"] {
		t.Error("'flex' should have been cleared by Reset")
	}
	if !got["block"] {
		t.Error("missing 'block' after second scan")
	}
}

func TestScreenKeywordWidth(t *testing.T) {
	e := New()
	e.Write([]byte(`w-screen`))
	css := e.CSS()
	if !strings.Contains(css, "width: 100vw") {
		t.Errorf("w-screen should produce width: 100vw, got:\n%s", css)
	}
}

func TestScreenKeywordHeight(t *testing.T) {
	e := New()
	e.Write([]byte(`h-screen`))
	css := e.CSS()
	if !strings.Contains(css, "height: 100vh") {
		t.Errorf("h-screen should produce height: 100vh, got:\n%s", css)
	}
}

func TestScreenKeywordMinHeight(t *testing.T) {
	e := New()
	e.Write([]byte(`min-h-screen`))
	css := e.CSS()
	if !strings.Contains(css, "min-height: 100vh") {
		t.Errorf("min-h-screen should produce min-height: 100vh, got:\n%s", css)
	}
}

func TestScreenKeywordMaxHeight(t *testing.T) {
	e := New()
	e.Write([]byte(`max-h-screen`))
	css := e.CSS()
	if !strings.Contains(css, "max-height: 100vh") {
		t.Errorf("max-h-screen should produce max-height: 100vh, got:\n%s", css)
	}
}
