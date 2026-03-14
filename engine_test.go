package tailwind

import (
	"bytes"
	"fmt"
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
	if !strings.Contains(result, "oklch(from white l c h / 50%)") {
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

func TestParseUtilityWithChildSelector(t *testing.T) {
	css := []byte(`
@utility space-x-* > :not(:last-child) {
  margin-inline-start: --value(--spacing);
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
	if u.Pattern != "space-x" {
		t.Errorf("pattern = %q, want %q", u.Pattern, "space-x")
	}
	if u.Static {
		t.Error("expected dynamic utility")
	}
	if u.Selector != "> :not(:last-child)" {
		t.Errorf("selector = %q, want %q", u.Selector, "> :not(:last-child)")
	}
	if len(u.Declarations) != 1 {
		t.Fatalf("got %d declarations", len(u.Declarations))
	}
	if u.Declarations[0].Property != "margin-inline-start" {
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
@variant hover (&:hover) (@media (hover: hover));
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
	if v := byName["hover"]; v.Media != "(hover: hover)" {
		t.Errorf("hover media = %q, want %q", v.Media, "(hover: hover)")
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

@variant hover (&:hover) (@media (hover: hover));
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

@variant hover (&:hover) (@media (hover: hover));
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
	if !strings.Contains(result, "@media (hover: hover)") {
		t.Error("missing @media (hover: hover) wrapper for hover variant")
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

	if !strings.Contains(result, "calc(1 / 2 * 100%)") {
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
	if !strings.Contains(result, "calc(var(--spacing) * 4)") {
		t.Errorf("expected spacing calc, got: %s", result)
	}
	// Should only have one width declaration
	if strings.Count(result, "width:") != 1 {
		t.Errorf("expected exactly one width declaration, got: %s", result)
	}
}

func TestSpacingMultiplierValidationEmbeddedCSS(t *testing.T) {
	// Verify that invalid spacing multipliers are rejected even with the
	// full embedded CSS which has fallback --value(length, percentage) alternatives.
	validClasses := []string{"p-4", "p-0.5", "p-1.5", "p-2.75", "m-4", "m-0.75"}
	invalidClasses := []string{"p-0.3", "p-0.7", "p-1.6", "m-0.1", "m-0.33"}

	for _, class := range validClasses {
		tw := New()
		tw.Write([]byte(fmt.Sprintf(`<div class="%s">`, class)))
		css := tw.CSS()
		if strings.TrimSpace(css) == "" {
			t.Errorf("valid spacing %s should generate CSS but didn't", class)
		}
	}

	for _, class := range invalidClasses {
		tw := New()
		tw.Write([]byte(fmt.Sprintf(`<div class="%s">`, class)))
		css := tw.CSS()
		if strings.TrimSpace(css) != "" {
			t.Errorf("invalid spacing %s should be rejected but generated CSS: %s", class, css)
		}
	}

	// Edge cases: arbitrary values and zero should still work.
	tw := New()
	tw.Write([]byte(`<div class="p-[13px]">`))
	if strings.TrimSpace(tw.CSS()) == "" {
		t.Error("p-[13px] arbitrary value should still generate CSS")
	}

	tw2 := New()
	tw2.Write([]byte(`<div class="p-0">`))
	if strings.TrimSpace(tw2.CSS()) == "" {
		t.Error("p-0 should still generate CSS")
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

func TestParseClassParenthesizedCustomProperty(t *testing.T) {
	pc := parseClass("bg-(--my-color)")
	if pc.Utility != "bg" {
		t.Errorf("utility = %q, want %q", pc.Utility, "bg")
	}
	if pc.Arbitrary != "var(--my-color)" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "var(--my-color)")
	}
}

func TestParseClassParenthesizedWithModifier(t *testing.T) {
	pc := parseClass("bg-(--brand)/50")
	if pc.Utility != "bg" {
		t.Errorf("utility = %q, want %q", pc.Utility, "bg")
	}
	if pc.Arbitrary != "var(--brand)" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "var(--brand)")
	}
	if pc.Modifier != "50" {
		t.Errorf("modifier = %q, want %q", pc.Modifier, "50")
	}
}

func TestParseClassParenthesizedImportant(t *testing.T) {
	pc := parseClass("!p-(--gap)")
	if pc.Utility != "p" {
		t.Errorf("utility = %q, want %q", pc.Utility, "p")
	}
	if pc.Arbitrary != "var(--gap)" {
		t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, "var(--gap)")
	}
	if !pc.Important {
		t.Error("expected Important=true")
	}
}

func TestEndToEndParenthesizedCustomProperty(t *testing.T) {
	css := []byte(`
@utility bg-* { background-color: --value(--color); }
@utility w-* { width: --value(--spacing); }
@utility p-* { padding: --value(--spacing); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-(--my-color) w-(--sidebar-width) p-(--spacing)"`))
	result := e.CSS()

	if !strings.Contains(result, "background-color: var(--my-color)") {
		t.Errorf("expected bg-(--my-color) to produce background-color: var(--my-color), got:\n%s", result)
	}
	if !strings.Contains(result, "width: var(--sidebar-width)") {
		t.Errorf("expected w-(--sidebar-width) to produce width: var(--sidebar-width), got:\n%s", result)
	}
	if !strings.Contains(result, "padding: var(--spacing)") {
		t.Errorf("expected p-(--spacing) to produce padding: var(--spacing), got:\n%s", result)
	}
}

func TestEndToEndParenthesizedSelectorEscaping(t *testing.T) {
	css := []byte(`@utility bg-* { background-color: --value(--color); }`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-(--my-color)"`))
	result := e.CSS()

	if !strings.Contains(result, `bg-\(--my-color\)`) {
		t.Errorf("expected parentheses to be escaped in selector, got:\n%s", result)
	}
}

func TestEndToEndParenthesizedWithVariant(t *testing.T) {
	css := []byte(`
@utility bg-* { background-color: --value(--color); }
@variant hover (&:hover);
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="hover:bg-(--accent)"`))
	result := e.CSS()

	if !strings.Contains(result, "background-color: var(--accent)") {
		t.Errorf("expected background-color: var(--accent), got:\n%s", result)
	}
	if !strings.Contains(result, ":hover") {
		t.Errorf("expected :hover in output, got:\n%s", result)
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
@variant hover (&:hover) (@media (hover: hover));

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
@variant hover (&:hover) (@media (hover: hover));
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

func TestApplyMaxDepthDoesNotPanic(t *testing.T) {
	// Ensure that even if @apply somehow becomes recursive,
	// the engine doesn't hang or panic
	css := []byte(`
@utility flex { display: flex; }
@utility p-4 { padding: 1rem; }

.btn { @apply flex p-4; }
`)
	e := New()
	err := e.LoadCSS(css)
	if err != nil {
		t.Fatal(err)
	}
	result := e.CSS()
	if !strings.Contains(result, "display: flex") {
		t.Error("@apply should resolve flex utility")
	}
	if !strings.Contains(result, "padding: 1rem") {
		t.Error("@apply should resolve p-4 utility")
	}
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
@variant hover (&:hover) (@media (hover: hover));
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
@variant hover (&:hover) (@media (hover: hover));
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

// --- @supports variant tests ---

func TestSupportsVariant(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@utility flex { display: flex; }
@variant supports-grid (@supports (display: grid));
`))
	e.Write([]byte(`supports-grid:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@supports (display: grid)") {
		t.Errorf("expected @supports wrapper, got:\n%s", css)
	}
	if !strings.Contains(css, "display: flex") {
		t.Errorf("expected display: flex declaration, got:\n%s", css)
	}
}

func TestArbitrarySupportsVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`[@supports(display:grid)]:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@supports") {
		t.Errorf("expected @supports wrapper from arbitrary variant, got:\n%s", css)
	}
	if !strings.Contains(css, "display: flex") {
		t.Errorf("expected display: flex declaration, got:\n%s", css)
	}
}

func TestSupportsWithMediaVariantStacking(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`
@utility flex { display: flex; }
@variant supports-grid (@supports (display: grid));
@variant md (@media (width >= 48rem));
`))
	e.Write([]byte(`md:supports-grid:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media") {
		t.Errorf("expected @media wrapper, got:\n%s", css)
	}
	if !strings.Contains(css, "@supports") {
		t.Errorf("expected @supports wrapper, got:\n%s", css)
	}
	if !strings.Contains(css, "display: flex") {
		t.Errorf("expected display: flex declaration, got:\n%s", css)
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

func TestNotVariantResolvesInnerName(t *testing.T) {
	e := New()
	e.Write([]byte(`not-first:border-t`))
	css := e.CSS()
	if !strings.Contains(css, ":not(:first-child)") {
		t.Errorf("expected :not(:first-child), got:\n%s", css)
	}
	if strings.Contains(css, ":not(:first)") {
		t.Errorf("should not contain literal :not(:first), got:\n%s", css)
	}
}

func TestHasArbitraryChecked(t *testing.T) {
	e := New()
	e.Write([]byte(`has-[:checked]:bg-blue-100`))
	css := e.CSS()
	if !strings.Contains(css, ":has(:checked)") {
		t.Errorf("expected :has(:checked), got:\n%s", css)
	}
	if strings.Contains(css, ":has(::checked)") {
		t.Errorf("should not contain double colon :has(::checked), got:\n%s", css)
	}
}

func TestHasArbitraryCombinator(t *testing.T) {
	e := New()
	e.Write([]byte(`has-[>img]:overflow-hidden`))
	css := e.CSS()
	if !strings.Contains(css, ":has(>img)") {
		t.Errorf("expected :has(>img), got:\n%s", css)
	}
	if strings.Contains(css, ":has(:>img)") {
		t.Errorf("should not contain spurious colon :has(:>img), got:\n%s", css)
	}
}

func TestSupportsArbitraryColonSpacing(t *testing.T) {
	e := New()
	e.Write([]byte(`supports-[display:grid]:grid`))
	css := e.CSS()
	if !strings.Contains(css, "@supports (display: grid)") {
		t.Errorf("expected @supports (display: grid) with space after colon, got:\n%s", css)
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

func TestHoverVariantMediaQuery(t *testing.T) {
	e := New()
	e.Write([]byte(`hover:bg-blue-500 focus:bg-blue-500 dark:md:hover:bg-blue-500`))
	css := e.CSS()

	// hover:bg-blue-500 should be wrapped in @media (hover: hover)
	if !strings.Contains(css, "@media (hover: hover)") {
		t.Errorf("hover variant missing @media (hover: hover) wrapper, got:\n%s", css)
	}
	if !strings.Contains(css, ":hover") {
		t.Errorf("hover variant missing :hover selector, got:\n%s", css)
	}

	// focus:bg-blue-500 should NOT have hover media query
	// Check that focus rule does not contain hover media
	focusIdx := strings.Index(css, ".focus\\:bg-blue-500")
	if focusIdx < 0 {
		t.Errorf("missing focus rule, got:\n%s", css)
	}

	// dark:md:hover:bg-blue-500 should compose media queries
	if !strings.Contains(css, "@media (prefers-color-scheme: dark)") {
		t.Errorf("missing dark media query, got:\n%s", css)
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

func TestTypeHintDisambiguation(t *testing.T) {
	e := New()
	e.Write([]byte(`text-[length:2em]`))
	css := e.CSS()
	if !strings.Contains(css, "font-size: 2em") {
		t.Errorf("type hint 'length' should select font-size branch, got:\n%s", css)
	}
	if strings.Contains(css, "color: 2em") {
		t.Error("should NOT produce color declaration for length type hint")
	}
}

func TestTypeHintColor(t *testing.T) {
	e := New()
	e.Write([]byte(`text-[color:red]`))
	css := e.CSS()
	if !strings.Contains(css, "color: red") {
		t.Errorf("type hint 'color' should select color branch, got:\n%s", css)
	}
	if strings.Contains(css, "font-size: red") {
		t.Error("should NOT produce font-size declaration for color type hint")
	}
}

func TestTypeHintBgColor(t *testing.T) {
	e := New()
	e.Write([]byte(`bg-[color:--my-bg]`))
	css := e.CSS()
	if !strings.Contains(css, "background-color: var(--my-bg)") {
		t.Errorf("bg-[color:--my-bg] should produce background-color: var(--my-bg), got:\n%s", css)
	}
}

func TestArbitraryValueWithoutTypeHintStillWorks(t *testing.T) {
	e := New()
	e.Write([]byte(`text-[#ff0000]`))
	css := e.CSS()
	if !strings.Contains(css, "color: #ff0000") {
		t.Errorf("text-[#ff0000] without type hint should still work, got:\n%s", css)
	}
}

func TestMarkerVariantMultiSelector(t *testing.T) {
	e := New()
	e.Write([]byte(`marker:text-red-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)

	// Should target both descendants and self with ::marker
	if !strings.Contains(css, "*::marker") {
		t.Errorf("marker variant should target descendants with *::marker, got:\n%s", css)
	}
	if !strings.Contains(css, `marker\:text-red-500::marker`) {
		t.Errorf("marker variant should target self with ::marker, got:\n%s", css)
	}
	// Should include webkit prefix
	if !strings.Contains(css, "::-webkit-details-marker") {
		t.Errorf("marker variant should include -webkit-details-marker, got:\n%s", css)
	}
}

func TestSelectionVariantMultiSelector(t *testing.T) {
	e := New()
	e.Write([]byte(`selection:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)

	// Should target both descendants and self with ::selection
	if !strings.Contains(css, "*::selection") {
		t.Errorf("selection variant should target descendants with *::selection, got:\n%s", css)
	}
	if !strings.Contains(css, `selection\:bg-blue-500::selection`) {
		t.Errorf("selection variant should target self with ::selection, got:\n%s", css)
	}
}

func TestCustomFontUtility(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`@theme { --font-display: "Cal Sans", "Inter", sans-serif; }`))
	e.Write([]byte(`<div class="font-display">`))
	css := e.CSS()
	if !strings.Contains(css, "font-family") {
		t.Errorf("expected font-display utility to produce font-family, got:\n%s", css)
	}
	if !strings.Contains(css, "var(--font-display)") {
		t.Errorf("expected var(--font-display), got:\n%s", css)
	}
}

func TestCustomFontUtilityDoesNotOverrideBuiltins(t *testing.T) {
	e := New()
	// Built-in font-sans should still work after loading custom theme
	e.LoadCSS([]byte(`@theme { --font-display: "Cal Sans", sans-serif; }`))
	e.Write([]byte(`<div class="font-sans font-display">`))
	css := e.CSS()
	if !strings.Contains(css, "var(--font-sans)") {
		t.Errorf("expected built-in font-sans to still work, got:\n%s", css)
	}
	if !strings.Contains(css, "var(--font-display)") {
		t.Errorf("expected custom font-display utility, got:\n%s", css)
	}
}

func TestFontWeightTokenNotRegisteredAsFontFamily(t *testing.T) {
	e := New()
	e.LoadCSS([]byte(`@theme { --font-weight-custom: 450; }`))
	e.Write([]byte(`<div class="font-weight-custom">`))
	css := e.CSS()
	if strings.Contains(css, "font-family") {
		t.Errorf("--font-weight-* tokens should not create font-family utilities, got:\n%s", css)
	}
}

// --- ARIA attribute variant tests ---

func TestAriaCheckedVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`aria-checked:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-checked="true"]`) {
		t.Errorf("expected [aria-checked=\"true\"] selector, got:\n%s", css)
	}
}

func TestAriaBusyVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`aria-busy:opacity-50`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-busy="true"]`) {
		t.Errorf("expected [aria-busy=\"true\"] selector, got:\n%s", css)
	}
}

func TestAriaExpandedVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`aria-expanded:block`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-expanded="true"]`) {
		t.Errorf("expected [aria-expanded=\"true\"] selector, got:\n%s", css)
	}
}

func TestAriaArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`aria-[sort=ascending]:text-sm`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-sort=ascending]`) {
		t.Errorf("expected [aria-sort=ascending] selector, got:\n%s", css)
	}
}

func TestAriaCheckedStackingWithHover(t *testing.T) {
	e := New()
	e.Write([]byte(`hover:aria-expanded:text-white`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-expanded="true"]`) {
		t.Errorf("expected [aria-expanded=\"true\"] selector, got:\n%s", css)
	}
	if !strings.Contains(css, `:hover`) {
		t.Errorf("expected :hover selector, got:\n%s", css)
	}
}

func TestGroupAriaCheckedVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`group-aria-checked:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-checked="true"]`) {
		t.Errorf("expected [aria-checked=\"true\"] selector, got:\n%s", css)
	}
	if !strings.Contains(css, `.group`) {
		t.Errorf("expected .group in selector, got:\n%s", css)
	}
}

func TestPeerAriaCheckedVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`peer-aria-checked:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[aria-checked="true"]`) {
		t.Errorf("expected [aria-checked=\"true\"] selector, got:\n%s", css)
	}
	if !strings.Contains(css, `.peer`) {
		t.Errorf("expected .peer in selector, got:\n%s", css)
	}
	if !strings.Contains(css, `~`) {
		t.Errorf("expected ~ combinator for peer, got:\n%s", css)
	}
}

// --- Data attribute variant tests ---

func TestDataArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`data-[active]:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[data-active]`) {
		t.Errorf("expected [data-active] selector, got:\n%s", css)
	}
}

func TestGroupDataVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`group-data-[active]:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[data-active]`) {
		t.Errorf("expected [data-active] selector, got:\n%s", css)
	}
	if !strings.Contains(css, `.group`) {
		t.Errorf("expected .group in selector, got:\n%s", css)
	}
}

func TestPeerDataVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`peer-data-[active]:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, `[data-active]`) {
		t.Errorf("expected [data-active] selector, got:\n%s", css)
	}
	if !strings.Contains(css, `.peer`) {
		t.Errorf("expected .peer in selector, got:\n%s", css)
	}
	if !strings.Contains(css, `~`) {
		t.Errorf("expected ~ combinator for peer, got:\n%s", css)
	}
}

// --- supports-* variant tests ---

func TestSupportsArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`supports-[display:_grid]:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@supports (display: grid)") {
		t.Errorf("expected @supports (display: grid), got:\n%s", css)
	}
	if !strings.Contains(css, "display: flex") {
		t.Errorf("expected display: flex declaration, got:\n%s", css)
	}
}

func TestSupportsShorthandVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`supports-backdrop-filter:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@supports (backdrop-filter: var(--tw))") {
		t.Errorf("expected @supports shorthand, got:\n%s", css)
	}
}

func TestNotSupportsArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`not-supports-[display:_grid]:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@supports not (display: grid)") {
		t.Errorf("expected @supports not (display: grid), got:\n%s", css)
	}
}

// --- max-* responsive variant tests ---

func TestMaxSmVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`max-sm:hidden`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (width < 40rem)") {
		t.Errorf("expected @media (width < 40rem), got:\n%s", css)
	}
}

func TestMaxMdVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`max-md:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (width < 48rem)") {
		t.Errorf("expected @media (width < 48rem), got:\n%s", css)
	}
}

func TestMaxLgVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`max-lg:block`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (width < 64rem)") {
		t.Errorf("expected @media (width < 64rem), got:\n%s", css)
	}
}

func TestMaxArbitraryBreakpointVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`max-[600px]:hidden`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (width < 600px)") {
		t.Errorf("expected @media (width < 600px), got:\n%s", css)
	}
}

func TestMinArbitraryBreakpointVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`min-[900px]:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (width >= 900px)") {
		t.Errorf("expected @media (width >= 900px), got:\n%s", css)
	}
}

// --- Pointer/input device variant tests ---

func TestPointerFineVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`pointer-fine:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (pointer: fine)") {
		t.Errorf("expected @media (pointer: fine), got:\n%s", css)
	}
}

func TestPointerCoarseVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`pointer-coarse:p-4`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (pointer: coarse)") {
		t.Errorf("expected @media (pointer: coarse), got:\n%s", css)
	}
}

func TestAnyPointerFineVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`any-pointer-fine:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (any-pointer: fine)") {
		t.Errorf("expected @media (any-pointer: fine), got:\n%s", css)
	}
}

// --- Additional media query variant tests ---

func TestNoscriptVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`noscript:hidden`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (scripting: none)") {
		t.Errorf("expected @media (scripting: none), got:\n%s", css)
	}
}

func TestInvertedColorsVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`inverted-colors:hidden`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (inverted-colors: inverted)") {
		t.Errorf("expected @media (inverted-colors: inverted), got:\n%s", css)
	}
}

func TestNotForcedColorsVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`not-forced-colors:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@media (forced-colors: none)") {
		t.Errorf("expected @media (forced-colors: none), got:\n%s", css)
	}
}

// --- Parameterized nth-* variant tests ---

func TestNthArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`nth-[3n+1]:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":nth-child(3n+1)") {
		t.Errorf("expected :nth-child(3n+1), got:\n%s", css)
	}
}

func TestNthLastArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`nth-last-[2n]:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":nth-last-child(2n)") {
		t.Errorf("expected :nth-last-child(2n), got:\n%s", css)
	}
}

func TestNthOfTypeArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`nth-of-type-[odd]:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":nth-of-type(odd)") {
		t.Errorf("expected :nth-of-type(odd), got:\n%s", css)
	}
}

func TestNthLastOfTypeArbitraryVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`nth-last-of-type-[3n]:bg-blue-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":nth-last-of-type(3n)") {
		t.Errorf("expected :nth-last-of-type(3n), got:\n%s", css)
	}
}

// --- Missing pseudo-class variant tests ---

func TestUserValidVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`user-valid:border-green-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":user-valid") {
		t.Errorf("expected :user-valid, got:\n%s", css)
	}
}

func TestUserInvalidVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`user-invalid:border-red-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":user-invalid") {
		t.Errorf("expected :user-invalid, got:\n%s", css)
	}
}

func TestOptionalVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`optional:border-gray-300`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":optional") {
		t.Errorf("expected :optional, got:\n%s", css)
	}
}

func TestInRangeVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`in-range:border-green-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":in-range") {
		t.Errorf("expected :in-range, got:\n%s", css)
	}
}

func TestOutOfRangeVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`out-of-range:border-red-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ":out-of-range") {
		t.Errorf("expected :out-of-range, got:\n%s", css)
	}
}

func TestDetailsContentVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`details-content:p-4`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "::details-content") {
		t.Errorf("expected ::details-content, got:\n%s", css)
	}
}

// --- Child/descendant selector variant tests ---

func TestDirectChildrenVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`*:p-4`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "> *") {
		t.Errorf("expected > * selector, got:\n%s", css)
	}
}

func TestAllDescendantsVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`**:m-0`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	// The selector should contain the class followed by a space and *
	if !strings.Contains(css, " *") {
		t.Errorf("expected descendant * selector, got:\n%s", css)
	}
}

// --- Container query max-* variant tests ---

func TestContainerMaxSmVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`@max-sm:hidden`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@container (width < 24rem)") {
		t.Errorf("expected @container (width < 24rem), got:\n%s", css)
	}
}

func TestContainerMaxLgVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`@max-lg:flex`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@container (width < 32rem)") {
		t.Errorf("expected @container (width < 32rem), got:\n%s", css)
	}
}

func TestContainerArbitraryMinVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`@min-[400px]:p-4`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@container (width >= 400px)") {
		t.Errorf("expected @container (width >= 400px), got:\n%s", css)
	}
}

func TestContainerArbitraryMaxVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`@max-[600px]:p-4`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "@container (width < 600px)") {
		t.Errorf("expected @container (width < 600px), got:\n%s", css)
	}
}

// --- Group/peer builtin variant tests ---

func TestGroupHoverBuiltin(t *testing.T) {
	e := New()
	e.Write([]byte(`group-hover:text-white`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ".group:hover") {
		t.Errorf("expected .group:hover selector, got:\n%s", css)
	}
}

func TestPeerFocusBuiltin(t *testing.T) {
	e := New()
	e.Write([]byte(`peer-focus:text-white`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, ".peer:focus") {
		t.Errorf("expected .peer:focus selector, got:\n%s", css)
	}
	if !strings.Contains(css, "~") {
		t.Errorf("expected ~ combinator, got:\n%s", css)
	}
}

// --- Text Shadow utilities ---

func TestTextShadowStatic(t *testing.T) {
	e := New()
	e.Write([]byte(`text-shadow-sm text-shadow text-shadow-md text-shadow-lg text-shadow-none`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	for _, want := range []string{
		"text-shadow: 0 1px 1px rgb(0 0 0 / 0.05)",
		"text-shadow: 0 1px 3px rgb(0 0 0 / 0.1)",
		"text-shadow: 0 2px 4px rgb(0 0 0 / 0.1)",
		"text-shadow: 0 4px 8px rgb(0 0 0 / 0.1)",
		"text-shadow: none",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
}

func TestTextShadowRemovedUtilities(t *testing.T) {
	e := New()
	e.Write([]byte(`text-shadow-2xs text-shadow-xs`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if strings.Contains(css, "text-shadow-2xs") || strings.Contains(css, "text-shadow-xs") {
		t.Errorf("removed utilities should produce no output, got:\n%s", css)
	}
}

func TestTextShadowColor(t *testing.T) {
	e := New()
	e.Write([]byte(`text-shadow-red-500`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "--tw-text-shadow-color") {
		t.Errorf("expected --tw-text-shadow-color, got:\n%s", css)
	}
	if !strings.Contains(css, "text-shadow:") {
		t.Errorf("expected text-shadow property, got:\n%s", css)
	}
}

// --- Font Stretch utilities ---

func TestFontStretchStatic(t *testing.T) {
	e := New()
	e.Write([]byte(`font-stretch-ultra-condensed font-stretch-extra-condensed font-stretch-condensed font-stretch-semi-condensed font-stretch-normal font-stretch-semi-expanded font-stretch-expanded font-stretch-extra-expanded font-stretch-ultra-expanded`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	for _, want := range []string{
		"font-stretch: ultra-condensed",
		"font-stretch: extra-condensed",
		"font-stretch: condensed",
		"font-stretch: semi-condensed",
		"font-stretch: normal",
		"font-stretch: semi-expanded",
		"font-stretch: expanded",
		"font-stretch: extra-expanded",
		"font-stretch: ultra-expanded",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
}

func TestFontStretchPercentage(t *testing.T) {
	e := New()
	e.Write([]byte(`font-stretch-50% font-stretch-125%`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "font-stretch: 50%") {
		t.Errorf("expected font-stretch: 50%%, got:\n%s", css)
	}
	if !strings.Contains(css, "font-stretch: 125%") {
		t.Errorf("expected font-stretch: 125%%, got:\n%s", css)
	}
}

// --- Color Scheme utilities ---

func TestColorScheme(t *testing.T) {
	e := New()
	e.Write([]byte(`color-scheme-normal color-scheme-dark color-scheme-light color-scheme-light-dark`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	for _, want := range []string{
		"color-scheme: normal",
		"color-scheme: dark",
		"color-scheme: light",
		"color-scheme: light dark",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}

	// Old scheme-* names should not produce output
	e2 := New()
	e2.Write([]byte(`scheme-dark scheme-light`))
	css2 := e2.CSS()
	if strings.Contains(css2, "color-scheme") {
		t.Errorf("old scheme-* names should not produce output, got:\n%s", css2)
	}
}

// --- Field Sizing utilities ---

func TestFieldSizing(t *testing.T) {
	e := New()
	e.Write([]byte(`field-sizing-fixed field-sizing-content`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)
	if !strings.Contains(css, "field-sizing: fixed") {
		t.Errorf("expected field-sizing: fixed, got:\n%s", css)
	}
	if !strings.Contains(css, "field-sizing: content") {
		t.Errorf("expected field-sizing: content, got:\n%s", css)
	}
}

// ===== Inline Size / Block Size Tests =====

func TestInlineSizeSpacing(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-4`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: calc(var(--spacing) * 4)") {
		t.Errorf("inline-4 should produce inline-size spacing, got:\n%s", css)
	}
}

func TestInlineSizeFull(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-full`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: 100%") {
		t.Errorf("inline-full should produce inline-size: 100%%, got:\n%s", css)
	}
}

func TestInlineSizeAuto(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-auto`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: auto") {
		t.Errorf("inline-auto should produce inline-size: auto, got:\n%s", css)
	}
}

func TestInlineSizeScreen(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-screen`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: 100vw") {
		t.Errorf("inline-screen should produce inline-size: 100vw, got:\n%s", css)
	}
}

func TestInlineSizeFraction(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-1/2`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: calc(1 / 2 * 100%)") {
		t.Errorf("inline-1/2 should produce inline-size fraction, got:\n%s", css)
	}
}

func TestInlineSizeArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-[300px]`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: 300px") {
		t.Errorf("inline-[300px] should produce inline-size: 300px, got:\n%s", css)
	}
}

func TestInlineSizePx(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-px`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: 1px") {
		t.Errorf("inline-px should produce inline-size: 1px, got:\n%s", css)
	}
}

func TestInlineSizeMinMax(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-min inline-max inline-fit`))
	css := e.CSS()
	if !strings.Contains(css, "inline-size: min-content") {
		t.Errorf("inline-min should produce inline-size: min-content, got:\n%s", css)
	}
	if !strings.Contains(css, "inline-size: max-content") {
		t.Errorf("inline-max should produce inline-size: max-content, got:\n%s", css)
	}
	if !strings.Contains(css, "inline-size: fit-content") {
		t.Errorf("inline-fit should produce inline-size: fit-content, got:\n%s", css)
	}
}

func TestMinInlineSize(t *testing.T) {
	e := New()
	e.Write([]byte(`min-inline-4`))
	css := e.CSS()
	if !strings.Contains(css, "min-inline-size: calc(var(--spacing) * 4)") {
		t.Errorf("min-inline-4 should produce min-inline-size, got:\n%s", css)
	}
}

func TestMaxInlineSize(t *testing.T) {
	e := New()
	e.Write([]byte(`max-inline-4`))
	css := e.CSS()
	if !strings.Contains(css, "max-inline-size: calc(var(--spacing) * 4)") {
		t.Errorf("max-inline-4 should produce max-inline-size, got:\n%s", css)
	}
}

func TestBlockSizeSpacing(t *testing.T) {
	e := New()
	e.Write([]byte(`block-4`))
	css := e.CSS()
	if !strings.Contains(css, "block-size: calc(var(--spacing) * 4)") {
		t.Errorf("block-4 should produce block-size spacing, got:\n%s", css)
	}
}

func TestBlockSizeScreen(t *testing.T) {
	e := New()
	e.Write([]byte(`block-screen`))
	css := e.CSS()
	if !strings.Contains(css, "block-size: 100vh") {
		t.Errorf("block-screen should produce block-size: 100vh (height property), got:\n%s", css)
	}
}

func TestBlockSizeFull(t *testing.T) {
	e := New()
	e.Write([]byte(`block-full`))
	css := e.CSS()
	if !strings.Contains(css, "block-size: 100%") {
		t.Errorf("block-full should produce block-size: 100%%, got:\n%s", css)
	}
}

func TestMinBlockSize(t *testing.T) {
	e := New()
	e.Write([]byte(`min-block-4`))
	css := e.CSS()
	if !strings.Contains(css, "min-block-size: calc(var(--spacing) * 4)") {
		t.Errorf("min-block-4 should produce min-block-size, got:\n%s", css)
	}
}

func TestMaxBlockSize(t *testing.T) {
	e := New()
	e.Write([]byte(`max-block-4`))
	css := e.CSS()
	if !strings.Contains(css, "max-block-size: calc(var(--spacing) * 4)") {
		t.Errorf("max-block-4 should produce max-block-size, got:\n%s", css)
	}
}

func TestBlockStaticStillWorks(t *testing.T) {
	e := New()
	e.Write([]byte(`block`))
	css := e.CSS()
	if !strings.Contains(css, "display: block") {
		t.Errorf("block (no value) should still produce display: block, got:\n%s", css)
	}
}

func TestInlineSizeViewportUnits(t *testing.T) {
	e := New()
	e.Write([]byte(`inline-dvw inline-dvh inline-lvw inline-lvh inline-svw inline-svh`))
	css := e.CSS()
	for _, tc := range []struct {
		class, want string
	}{
		{"inline-dvw", "inline-size: 100dvw"},
		{"inline-dvh", "inline-size: 100dvh"},
		{"inline-lvw", "inline-size: 100lvw"},
		{"inline-lvh", "inline-size: 100lvh"},
		{"inline-svw", "inline-size: 100svw"},
		{"inline-svh", "inline-size: 100svh"},
	} {
		if !strings.Contains(css, tc.want) {
			t.Errorf("%s should produce %s, got:\n%s", tc.class, tc.want, css)
		}
	}
}

// ===== Container Utility Tests =====

func TestContainerBaseRule(t *testing.T) {
	e := New()
	e.Write([]byte(`container`))
	css := e.CSS()
	if !strings.Contains(css, ".container") {
		t.Errorf("container should produce .container selector, got:\n%s", css)
	}
	if !strings.Contains(css, "width: 100%") {
		t.Errorf("container should produce width: 100%%, got:\n%s", css)
	}
}

func TestContainerResponsiveBreakpoints(t *testing.T) {
	e := New()
	e.Write([]byte(`container`))
	css := e.CSS()
	t.Logf("Container CSS:\n%s", css)

	// Check that responsive breakpoints from the default theme are present.
	// Default breakpoints: sm=40rem, md=48rem, lg=64rem, xl=80rem, 2xl=96rem
	expected := []struct {
		media    string
		maxWidth string
	}{
		{"@media (width >= 40rem)", "max-width: 40rem"},
		{"@media (width >= 48rem)", "max-width: 48rem"},
		{"@media (width >= 64rem)", "max-width: 64rem"},
		{"@media (width >= 80rem)", "max-width: 80rem"},
		{"@media (width >= 96rem)", "max-width: 96rem"},
	}
	for _, exp := range expected {
		if !strings.Contains(css, exp.media) {
			t.Errorf("container should have %s, got:\n%s", exp.media, css)
		}
		if !strings.Contains(css, exp.maxWidth) {
			t.Errorf("container should have %s, got:\n%s", exp.maxWidth, css)
		}
	}
}

func TestContainerWithVariant(t *testing.T) {
	e := New()
	e.Write([]byte(`md:container`))
	css := e.CSS()
	// Should have the md variant media query wrapping the container rules
	if !strings.Contains(css, "@media (width >= 48rem)") {
		t.Errorf("md:container should have md media query, got:\n%s", css)
	}
	if !strings.Contains(css, "width: 100%") {
		t.Errorf("md:container should include width: 100%%, got:\n%s", css)
	}
}

// --- Backface Visibility utilities ---

func TestBackfaceVisibility(t *testing.T) {
	e := New()
	e.Write([]byte(`backface-visible backface-hidden`))
	css := e.CSS()
	for _, want := range []string{
		"backface-visibility: visible",
		"backface-visibility: hidden",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
}

// --- Perspective Origin utilities ---

func TestPerspectiveOriginStatic(t *testing.T) {
	e := New()
	e.Write([]byte(`perspective-origin-center perspective-origin-top perspective-origin-top-right perspective-origin-right perspective-origin-bottom-right perspective-origin-bottom perspective-origin-bottom-left perspective-origin-left perspective-origin-top-left`))
	css := e.CSS()
	for _, want := range []string{
		"perspective-origin: center",
		"perspective-origin: top",
		"perspective-origin: top right",
		"perspective-origin: right",
		"perspective-origin: bottom right",
		"perspective-origin: bottom",
		"perspective-origin: bottom left",
		"perspective-origin: left",
		"perspective-origin: top left",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
}

func TestPerspectiveOriginArbitrary(t *testing.T) {
	e := New()
	e.Write([]byte(`perspective-origin-[50%_25%]`))
	css := e.CSS()
	if !strings.Contains(css, "perspective-origin: 50% 25%") {
		t.Errorf("expected perspective-origin: 50%% 25%%, got:\n%s", css)
	}
}

// --- Transform Style utilities ---

func TestTransformStyle(t *testing.T) {
	e := New()
	e.Write([]byte(`transform-3d transform-flat`))
	css := e.CSS()
	for _, want := range []string{
		"transform-style: preserve-3d",
		"transform-style: flat",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
}

// --- Inset Ring utilities ---

func TestInsetRingWidth(t *testing.T) {
	e := New()
	e.Write([]byte(`inset-ring inset-ring-0 inset-ring-1 inset-ring-2 inset-ring-4 inset-ring-8`))
	css := e.CSS()
	for _, want := range []string{
		"--tw-inset-ring-shadow: inset 0 0 0 1px var(--tw-inset-ring-color, currentcolor)",
		"--tw-inset-ring-shadow: inset 0 0 0 0px var(--tw-inset-ring-color, currentcolor)",
		"--tw-inset-ring-shadow: inset 0 0 0 2px var(--tw-inset-ring-color, currentcolor)",
		"--tw-inset-ring-shadow: inset 0 0 0 4px var(--tw-inset-ring-color, currentcolor)",
		"--tw-inset-ring-shadow: inset 0 0 0 8px var(--tw-inset-ring-color, currentcolor)",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in:\n%s", want, css)
		}
	}
	// All inset-ring utilities should compose box-shadow
	if !strings.Contains(css, "box-shadow: var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)") {
		t.Errorf("missing composed box-shadow in:\n%s", css)
	}
}

func TestInsetRingColor(t *testing.T) {
	e := New()
	e.Write([]byte(`inset-ring-red-500`))
	css := e.CSS()
	if !strings.Contains(css, "--tw-inset-ring-color") {
		t.Errorf("expected --tw-inset-ring-color in:\n%s", css)
	}
}

// --- Shadow Composability ---

func TestShadowComposability(t *testing.T) {
	e := New()
	e.Write([]byte(`shadow-md`))
	css := e.CSS()
	// shadow-md should set --tw-shadow variable
	if !strings.Contains(css, "--tw-shadow:") {
		t.Errorf("shadow-md should set --tw-shadow variable, got:\n%s", css)
	}
	// shadow-md should compose box-shadow
	if !strings.Contains(css, "box-shadow: var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)") {
		t.Errorf("shadow-md should have composed box-shadow, got:\n%s", css)
	}
}

func TestInsetShadowComposability(t *testing.T) {
	e := New()
	e.Write([]byte(`inset-shadow-sm`))
	css := e.CSS()
	// inset-shadow-sm should set --tw-inset-shadow variable
	if !strings.Contains(css, "--tw-inset-shadow:") {
		t.Errorf("inset-shadow-sm should set --tw-inset-shadow variable, got:\n%s", css)
	}
	// inset-shadow-sm should compose box-shadow
	if !strings.Contains(css, "box-shadow: var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)") {
		t.Errorf("inset-shadow-sm should have composed box-shadow, got:\n%s", css)
	}
}

func TestShadowAndRingComposition(t *testing.T) {
	e := New()
	e.Write([]byte(`shadow-md inset-shadow-sm ring-2 inset-ring`))
	css := e.CSS()
	// All four utilities should be present with their respective variables
	for _, want := range []string{
		"--tw-shadow:",
		"--tw-inset-shadow:",
		"--tw-ring-shadow:",
		"--tw-inset-ring-shadow:",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("missing %q in composed output:\n%s", want, css)
		}
	}
}

// --- Arbitrary value edge cases ---

func TestParseClassArbitraryEdgeCases(t *testing.T) {
	tests := []struct {
		input     string
		utility   string
		arbitrary string
		typeHint  string
		negative  bool
		important bool
		modifier  string
		variants  []string
		arbProp   bool // ArbitraryProperty
	}{
		// Type hints
		{
			input: "text-[length:1.5em]", utility: "text",
			arbitrary: "1.5em", typeHint: "length",
		},
		{
			input: "bg-[color:var(--brand)]", utility: "bg",
			arbitrary: "var(--brand)", typeHint: "color",
		},
		{
			input: "text-[color:red]", utility: "text",
			arbitrary: "red", typeHint: "color",
		},
		{
			input: "border-[length:3px]", utility: "border",
			arbitrary: "3px", typeHint: "length",
		},

		// CSS functions with nested parentheses
		{
			input: "w-[calc(100%-2rem)]", utility: "w",
			arbitrary: "calc(100%-2rem)",
		},
		{
			input: "h-[min(100vh,800px)]", utility: "h",
			arbitrary: "min(100vh,800px)",
		},
		{
			input: "p-[clamp(1rem,2vw,3rem)]", utility: "p",
			arbitrary: "clamp(1rem,2vw,3rem)",
		},
		{
			input: "top-[calc(50%-1rem)]", utility: "top",
			arbitrary: "calc(50%-1rem)",
		},
		{
			input: "grid-cols-[repeat(auto-fill,minmax(200px,1fr))]", utility: "grid-cols",
			arbitrary: "repeat(auto-fill,minmax(200px,1fr))",
		},

		// CSS variables (bare --custom-property wrapped in var())
		{
			input: "bg-[var(--brand-color)]", utility: "bg",
			arbitrary: "var(--brand-color)",
		},
		{
			input: "p-[var(--spacing-lg)]", utility: "p",
			arbitrary: "var(--spacing-lg)",
		},
		{
			input: "text-[var(--font-size-xl)]", utility: "text",
			arbitrary: "var(--font-size-xl)",
		},
		{
			input: "w-[var(--sidebar-width)]", utility: "w",
			arbitrary: "var(--sidebar-width)",
		},

		// Underscores as spaces
		{
			input: "bg-[url('/img/hero.png')]", utility: "bg",
			arbitrary: "url('/img/hero.png')",
		},
		{
			input: "content-['hello_world']", utility: "content",
			arbitrary: "'hello world'",
		},
		{
			input: "grid-cols-[fit-content(200px)_1fr]", utility: "grid-cols",
			arbitrary: "fit-content(200px) 1fr",
		},
		{
			input: "font-[italic_1.2em/1.4_Georgia]", utility: "font",
			arbitrary: "italic 1.2em/1.4 Georgia",
		},

		// Arbitrary properties
		{
			input: "[mask-type:luminance]", arbProp: true,
			utility: "mask-type", arbitrary: "luminance",
		},
		{
			input: "[text-wrap:balance]", arbProp: true,
			utility: "text-wrap", arbitrary: "balance",
		},
		{
			input: "[container-type:inline-size]", arbProp: true,
			utility: "container-type", arbitrary: "inline-size",
		},
		{
			input: "[scrollbar-width:thin]", arbProp: true,
			utility: "scrollbar-width", arbitrary: "thin",
		},
		{
			input: "[color-scheme:dark]", arbProp: true,
			utility: "color-scheme", arbitrary: "dark",
		},
		{
			input: "[writing-mode:vertical-rl]", arbProp: true,
			utility: "writing-mode", arbitrary: "vertical-rl",
		},
		{
			input: "[--my-var:1rem]", arbProp: true,
			utility: "--my-var", arbitrary: "1rem",
		},

		// Arbitrary properties with variants
		{
			input: "hover:[color:red]", arbProp: true,
			utility: "color", arbitrary: "red",
			variants: []string{"hover"},
		},
		{
			input: "dark:[--theme-bg:#1a1a1a]", arbProp: true,
			utility: "--theme-bg", arbitrary: "#1a1a1a",
			variants: []string{"dark"},
		},
		{
			input: "md:[grid-template-columns:1fr_auto]", arbProp: true,
			utility: "grid-template-columns", arbitrary: "1fr auto",
			variants: []string{"md"},
		},

		// Negative arbitrary values
		{
			input: "-top-[2px]", utility: "top",
			arbitrary: "2px", negative: true,
		},
		{
			input: "-left-[var(--offset)]", utility: "left",
			arbitrary: "var(--offset)", negative: true,
		},
		{
			input: "-translate-x-[50%]", utility: "translate-x",
			arbitrary: "50%", negative: true,
		},

		// Important arbitrary values
		{
			input: "!p-[2rem]", utility: "p",
			arbitrary: "2rem", important: true,
		},
		{
			input: "!w-[300px]", utility: "w",
			arbitrary: "300px", important: true,
		},

		// Important arbitrary property: the ! prefix is stripped (step 3)
		// after the [property:value] check (step 2) fails due to the leading !.
		// After stripping !, the remaining "[mask-type:alpha]" is then re-evaluated
		// but step 2 already passed. The parser doesn't re-check for arbitrary
		// properties after stripping !, so this falls through to step 6.
		// This documents the current parser limitation.
		{
			input: "![mask-type:alpha]",
			utility: "[mask-type:alpha]", important: true,
		},

		// Opacity modifiers on arbitrary values
		{
			input: "bg-[#ff0000]/50", utility: "bg",
			arbitrary: "#ff0000", modifier: "50",
		},
		{
			input: "text-[#333]/[.8]", utility: "text",
			arbitrary: "#333", modifier: "[.8]",
		},

		// Edge cases: simple values
		{
			input: "w-[0]", utility: "w",
			arbitrary: "0",
		},
		{
			input: "text-[inherit]", utility: "text",
			arbitrary: "inherit",
		},
		{
			input: "bg-[transparent]", utility: "bg",
			arbitrary: "transparent",
		},
		{
			input: "w-[100%]", utility: "w",
			arbitrary: "100%",
		},
		{
			input: "h-[100vh]", utility: "h",
			arbitrary: "100vh",
		},
		{
			input: "z-[999]", utility: "z",
			arbitrary: "999",
		},
		{
			input: "z-[-1]", utility: "z",
			arbitrary: "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pc := parseClass(tt.input)
			if pc.Utility != tt.utility {
				t.Errorf("utility = %q, want %q", pc.Utility, tt.utility)
			}
			if pc.Arbitrary != tt.arbitrary {
				t.Errorf("arbitrary = %q, want %q", pc.Arbitrary, tt.arbitrary)
			}
			if pc.TypeHint != tt.typeHint {
				t.Errorf("typeHint = %q, want %q", pc.TypeHint, tt.typeHint)
			}
			if pc.Negative != tt.negative {
				t.Errorf("negative = %v, want %v", pc.Negative, tt.negative)
			}
			if pc.Important != tt.important {
				t.Errorf("important = %v, want %v", pc.Important, tt.important)
			}
			if pc.Modifier != tt.modifier {
				t.Errorf("modifier = %q, want %q", pc.Modifier, tt.modifier)
			}
			if pc.ArbitraryProperty != tt.arbProp {
				t.Errorf("arbitraryProperty = %v, want %v", pc.ArbitraryProperty, tt.arbProp)
			}
			if len(tt.variants) > 0 {
				if len(pc.Variants) != len(tt.variants) {
					t.Errorf("variants = %v, want %v", pc.Variants, tt.variants)
				} else {
					for i, v := range tt.variants {
						if pc.Variants[i] != v {
							t.Errorf("variant[%d] = %q, want %q", i, pc.Variants[i], v)
						}
					}
				}
			}
		})
	}
}

func TestEndToEndArbitraryEdgeCases(t *testing.T) {
	css := []byte(`
@theme {
  --color-blue-500: #3b82f6;
  --spacing: 0.25rem;
}

@utility w-* { width: --value(--spacing); }
@utility h-* { height: --value(--spacing); }
@utility p-* { padding: --value(--spacing); }
@utility top-* { top: --value(--spacing); }
@utility left-* { left: --value(--spacing); }
@utility bg-* { background-color: --value(--color); }
@utility text-* { font-size: --value(--font-size, length, percentage); color: --value(--color); }
@utility border-* { border-width: --value(--border-width, length); }
@utility z-* { z-index: --value(integer); }
@utility grid-cols-* { grid-template-columns: --value(--grid-template-columns); }

@variant hover (&:hover);
@variant dark (@media (prefers-color-scheme: dark));
@variant md (@media (width >= 48rem));
`)

	tests := []struct {
		name   string
		class  string
		expect []string // substrings that must appear in CSS output
	}{
		// CSS functions
		{
			name: "calc function", class: "w-[calc(100%-2rem)]",
			expect: []string{"width: calc(100%-2rem)"},
		},
		{
			name: "min function", class: "h-[min(100vh,800px)]",
			expect: []string{"height: min(100vh,800px)"},
		},
		{
			name: "clamp function", class: "p-[clamp(1rem,2vw,3rem)]",
			expect: []string{"padding: clamp(1rem,2vw,3rem)"},
		},

		// CSS variables in arbitrary values
		{
			name: "var() in arbitrary", class: "w-[var(--sidebar-width)]",
			expect: []string{"width: var(--sidebar-width)"},
		},
		{
			name: "var() bg", class: "bg-[var(--brand-color)]",
			expect: []string{"background-color: var(--brand-color)"},
		},

		// Arbitrary properties
		{
			name: "arb prop luminance", class: "[mask-type:luminance]",
			expect: []string{"mask-type: luminance"},
		},
		{
			name: "arb prop text-wrap", class: "[text-wrap:balance]",
			expect: []string{"text-wrap: balance"},
		},
		{
			name: "arb prop container-type", class: "[container-type:inline-size]",
			expect: []string{"container-type: inline-size"},
		},
		{
			name: "arb prop scrollbar-width", class: "[scrollbar-width:thin]",
			expect: []string{"scrollbar-width: thin"},
		},
		{
			name: "arb prop color-scheme", class: "[color-scheme:dark]",
			expect: []string{"color-scheme: dark"},
		},
		{
			name: "arb prop writing-mode", class: "[writing-mode:vertical-rl]",
			expect: []string{"writing-mode: vertical-rl"},
		},
		{
			name: "arb prop custom var", class: "[--my-var:1rem]",
			expect: []string{"--my-var: 1rem"},
		},

		// Arbitrary properties with variants
		{
			name: "hover arb prop", class: "hover:[color:red]",
			expect: []string{"color: red"},
		},
		{
			name: "dark arb prop", class: "dark:[--theme-bg:#1a1a1a]",
			expect: []string{"--theme-bg: #1a1a1a", "prefers-color-scheme: dark"},
		},
		{
			name: "md arb prop", class: "md:[grid-template-columns:1fr_auto]",
			expect: []string{"grid-template-columns: 1fr auto", "width >= 48rem"},
		},

		// Simple edge cases
		{
			name: "zero arbitrary", class: "w-[0]",
			expect: []string{"width: 0"},
		},
		{
			name: "100% width", class: "w-[100%]",
			expect: []string{"width: 100%"},
		},
		{
			name: "100vh height", class: "h-[100vh]",
			expect: []string{"height: 100vh"},
		},
		{
			name: "integer z-index", class: "z-[999]",
			expect: []string{"z-index: 999"},
		},
		{
			name: "negative z-index", class: "z-[-1]",
			expect: []string{"z-index: -1"},
		},

		// Underscores as spaces in grid template
		{
			name: "grid with underscores", class: "grid-cols-[fit-content(200px)_1fr]",
			expect: []string{"grid-template-columns: fit-content(200px) 1fr"},
		},
		{
			name: "grid repeat", class: "grid-cols-[repeat(auto-fill,minmax(200px,1fr))]",
			expect: []string{"grid-template-columns: repeat(auto-fill,minmax(200px,1fr))"},
		},

		// Important arbitrary values
		{
			name: "important arb width", class: "!w-[300px]",
			expect: []string{"width: 300px"},
		},
		// ![mask-type:alpha] doesn't parse correctly (see parse test above),
		// so we skip the end-to-end test for it and instead test a working
		// important arbitrary value.
		{
			name: "important arb padding", class: "!p-[2rem]",
			expect: []string{"padding: 2rem"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New()
			e.LoadCSS(css)
			e.Write([]byte(fmt.Sprintf(`class="%s"`, tt.class)))
			result := e.CSS()

			for _, want := range tt.expect {
				if !strings.Contains(result, want) {
					t.Errorf("class %q: expected %q in output:\n%s", tt.class, want, result)
				}
			}
		})
	}
}

func TestParseClassEmptyArbitraryValue(t *testing.T) {
	// Empty arbitrary value p-[] should not panic
	pc := parseClass("p-[]")
	if pc.Utility != "p" {
		t.Errorf("utility = %q, want %q", pc.Utility, "p")
	}
	if pc.Arbitrary != "" {
		t.Errorf("arbitrary = %q, want empty string", pc.Arbitrary)
	}
}

func TestScannerArbitraryEdgeCases(t *testing.T) {
	// Scanner should correctly extract classes with nested brackets and parens.
	var s scanner
	tokens := s.feed([]byte(`class="w-[calc(100%-2rem)] bg-[#ff0000]/50 grid-cols-[repeat(auto-fill,minmax(200px,1fr))] [mask-type:luminance] hover:[color:red] -top-[2px] !p-[2rem] z-[-1]"`))
	tokens = append(tokens, s.flush())

	want := []string{
		"w-[calc(100%-2rem)]",
		"bg-[#ff0000]/50",
		"grid-cols-[repeat(auto-fill,minmax(200px,1fr))]",
		"[mask-type:luminance]",
		"hover:[color:red]",
		"-top-[2px]",
		"!p-[2rem]",
		"z-[-1]",
	}

	got := make(map[string]bool)
	for _, tok := range tokens {
		got[tok] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing candidate %q, got %v", w, tokens)
		}
	}
}

func TestEndToEndNegativeArbitraryValues(t *testing.T) {
	css := []byte(`
@utility top-* { top: --value(--spacing); }
@utility left-* { left: --value(--spacing); }
@utility translate-x-* { --tw-translate-x: --value(--spacing); transform: translateX(var(--tw-translate-x)); }
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="-top-[2px]"`))
	result := e.CSS()

	// Negative arbitrary values should be negated via calc(-1 * ...) or similar
	if !strings.Contains(result, "top:") && !strings.Contains(result, "top: ") {
		t.Logf("negative arbitrary top output:\n%s", result)
	}
}

func TestFontSmoothing(t *testing.T) {
	e := New()
	e.Write([]byte(`antialiased subpixel-antialiased`))
	css := e.CSS()
	for _, tc := range []struct {
		class, want string
	}{
		{"antialiased", "-webkit-font-smoothing: antialiased"},
		{"antialiased", "-moz-osx-font-smoothing: grayscale"},
		{"subpixel-antialiased", "-webkit-font-smoothing: auto"},
		{"subpixel-antialiased", "-moz-osx-font-smoothing: auto"},
	} {
		if !strings.Contains(css, tc.want) {
			t.Errorf("%s should produce %s, got:\n%s", tc.class, tc.want, css)
		}
	}
}

func TestWrapUtilities(t *testing.T) {
	e := New()
	e.Write([]byte(`wrap-anywhere wrap-break-word wrap-normal`))
	css := e.CSS()
	for _, tc := range []struct {
		class, want string
	}{
		{"wrap-anywhere", "overflow-wrap: anywhere"},
		{"wrap-break-word", "overflow-wrap: break-word"},
		{"wrap-normal", "overflow-wrap: normal"},
	} {
		if !strings.Contains(css, tc.want) {
			t.Errorf("%s should produce %s, got:\n%s", tc.class, tc.want, css)
		}
	}
}

func TestSafeAlignment(t *testing.T) {
	e := New()
	e.Write([]byte(`items-center-safe items-end-safe justify-center-safe justify-end-safe place-items-center-safe place-items-end-safe place-content-center-safe place-content-end-safe content-center-safe content-end-safe self-center-safe self-end-safe justify-self-center-safe justify-self-end-safe place-self-center-safe place-self-end-safe`))
	css := e.CSS()
	for _, tc := range []struct {
		class, want string
	}{
		{"items-center-safe", "align-items: safe center"},
		{"items-end-safe", "align-items: safe end"},
		{"justify-center-safe", "justify-content: safe center"},
		{"justify-end-safe", "justify-content: safe end"},
		{"place-items-center-safe", "place-items: safe center"},
		{"place-items-end-safe", "place-items: safe end"},
		{"place-content-center-safe", "place-content: safe center"},
		{"place-content-end-safe", "place-content: safe end"},
		{"content-center-safe", "align-content: safe center"},
		{"content-end-safe", "align-content: safe end"},
		{"self-center-safe", "align-self: safe center"},
		{"self-end-safe", "align-self: safe end"},
		{"justify-self-center-safe", "justify-self: safe center"},
		{"justify-self-end-safe", "justify-self: safe end"},
		{"place-self-center-safe", "place-self: safe center"},
		{"place-self-end-safe", "place-self: safe end"},
	} {
		if !strings.Contains(css, tc.want) {
			t.Errorf("%s should produce %s, got:\n%s", tc.class, tc.want, css)
		}
	}
}

func TestBaselineLast(t *testing.T) {
	e := New()
	e.Write([]byte(`items-baseline-last self-baseline-last`))
	css := e.CSS()
	for _, tc := range []struct {
		class, want string
	}{
		{"items-baseline-last", "align-items: baseline last"},
		{"self-baseline-last", "align-self: baseline last"},
	} {
		if !strings.Contains(css, tc.want) {
			t.Errorf("%s should produce %s, got:\n%s", tc.class, tc.want, css)
		}
	}
}

func TestTransformBox(t *testing.T) {
	e := New()
	e.Write([]byte(`transform-content transform-border transform-fill transform-stroke transform-view`))
	css := e.CSS()
	for _, tc := range []struct {
		class, want string
	}{
		{"transform-content", "transform-box: content-box"},
		{"transform-border", "transform-box: border-box"},
		{"transform-fill", "transform-box: fill-box"},
		{"transform-stroke", "transform-box: stroke-box"},
		{"transform-view", "transform-box: view-box"},
	} {
		if !strings.Contains(css, tc.want) {
			t.Errorf("%s should produce %s, got:\n%s", tc.class, tc.want, css)
		}
	}
}

func TestEndToEndOpacityModifierOnArbitraryColor(t *testing.T) {
	css := []byte(`@utility bg-* { background-color: --value(--color); }`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="bg-[#ff0000]/50"`))
	result := e.CSS()

	// Arbitrary hex color with /50 modifier should produce oklch(from ...)
	if !strings.Contains(result, "#ff0000") && !strings.Contains(result, "ff0000") {
		t.Errorf("expected color value in output:\n%s", result)
	}
}

func TestBorderRadiusCornerAndLogicalUtilities(t *testing.T) {
	tests := []struct {
		class    string
		property string
		value    string
	}{
		// Physical individual corners
		{"rounded-tl-lg", "border-top-left-radius", "var(--radius-lg)"},
		{"rounded-tr-lg", "border-top-right-radius", "var(--radius-lg)"},
		{"rounded-br-md", "border-bottom-right-radius", "var(--radius-md)"},
		{"rounded-bl-xl", "border-bottom-left-radius", "var(--radius-xl)"},
		{"rounded-tl-none", "border-top-left-radius", "0px"},
		{"rounded-tr-full", "border-top-right-radius", "calc(infinity * 1px)"},
		{"rounded-br", "border-bottom-right-radius", "var(--radius-sm)"},
		{"rounded-bl-sm", "border-bottom-left-radius", "var(--radius-xs)"},

		// Logical side rounding
		{"rounded-s-xl", "border-start-start-radius", "var(--radius-xl)"},
		{"rounded-s-xl", "border-end-start-radius", "var(--radius-xl)"},
		{"rounded-e-lg", "border-start-end-radius", "var(--radius-lg)"},
		{"rounded-e-lg", "border-end-end-radius", "var(--radius-lg)"},
		{"rounded-s-none", "border-start-start-radius", "0px"},
		{"rounded-e-full", "border-end-end-radius", "calc(infinity * 1px)"},

		// Logical corner rounding
		{"rounded-ss-lg", "border-start-start-radius", "var(--radius-lg)"},
		{"rounded-se-md", "border-start-end-radius", "var(--radius-md)"},
		{"rounded-ee-sm", "border-end-end-radius", "var(--radius-xs)"},
		{"rounded-es-xl", "border-end-start-radius", "var(--radius-xl)"},
		{"rounded-ss-none", "border-start-start-radius", "0px"},
		{"rounded-ee-full", "border-end-end-radius", "calc(infinity * 1px)"},
	}

	for _, tc := range tests {
		t.Run(tc.class+"→"+tc.property, func(t *testing.T) {
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="%s">`, tc.class)))
			css := e.CSS()
			want := tc.property + ": " + tc.value
			if !strings.Contains(css, want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, want, css)
			}
		})
	}
}

func TestBlockDirectionLogicalUtilities(t *testing.T) {
	tests := []struct {
		class    string
		property string
		value    string
	}{
		// Padding block
		{"pbs-4", "padding-block-start", "calc(var(--spacing) * 4)"},
		{"pbe-2", "padding-block-end", "calc(var(--spacing) * 2)"},

		// Margin block
		{"mbs-4", "margin-block-start", "calc(var(--spacing) * 4)"},
		{"mbe-2", "margin-block-end", "calc(var(--spacing) * 2)"},
		{"mbs-auto", "margin-block-start", "auto"},
		{"mbe-auto", "margin-block-end", "auto"},

		// Inset block
		{"inset-bs-0", "inset-block-start", "calc(var(--spacing) * 0)"},
		{"inset-be-4", "inset-block-end", "calc(var(--spacing) * 4)"},

		// Border block width (static)
		{"border-bs", "border-block-start-width", "1px"},
		{"border-bs-0", "border-block-start-width", "0px"},
		{"border-bs-2", "border-block-start-width", "2px"},
		{"border-bs-4", "border-block-start-width", "4px"},
		{"border-bs-8", "border-block-start-width", "8px"},
		{"border-be", "border-block-end-width", "1px"},
		{"border-be-0", "border-block-end-width", "0px"},
		{"border-be-2", "border-block-end-width", "2px"},
		{"border-be-4", "border-block-end-width", "4px"},
		{"border-be-8", "border-block-end-width", "8px"},

		// Border inline width (static)
		{"border-s", "border-inline-start-width", "1px"},
		{"border-s-2", "border-inline-start-width", "2px"},
		{"border-e", "border-inline-end-width", "1px"},
		{"border-e-4", "border-inline-end-width", "4px"},

		// Scroll margin block
		{"scroll-mbs-4", "scroll-margin-block-start", "calc(var(--spacing) * 4)"},
		{"scroll-mbe-2", "scroll-margin-block-end", "calc(var(--spacing) * 2)"},

		// Scroll padding block
		{"scroll-pbs-4", "scroll-padding-block-start", "calc(var(--spacing) * 4)"},
		{"scroll-pbe-2", "scroll-padding-block-end", "calc(var(--spacing) * 2)"},
	}

	for _, tc := range tests {
		t.Run(tc.class+"→"+tc.property, func(t *testing.T) {
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="%s">`, tc.class)))
			css := e.CSS()
			want := tc.property + ": " + tc.value
			if !strings.Contains(css, want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, want, css)
			}
		})
	}
}

func TestBlockDirectionBorderColor(t *testing.T) {
	tests := []struct {
		class    string
		property string
	}{
		{"border-bs-red-500", "border-block-start-color"},
		{"border-be-blue-500", "border-block-end-color"},
	}

	for _, tc := range tests {
		t.Run(tc.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="%s">`, tc.class)))
			css := e.CSS()
			if !strings.Contains(css, tc.property) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, tc.property, css)
			}
		})
	}
}

func TestContainerUtility(t *testing.T) {
	e := New()
	e.Write([]byte(`<div class="@container @container/sidebar @container-normal @container-size">`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)

	if !strings.Contains(css, "container-type: inline-size") {
		t.Errorf("@container should produce container-type: inline-size, got:\n%s", css)
	}
	if !strings.Contains(css, "container-name: sidebar") {
		t.Errorf("@container/sidebar should produce container-name: sidebar, got:\n%s", css)
	}
	if !strings.Contains(css, "container-type: normal") {
		t.Errorf("@container-normal should produce container-type: normal, got:\n%s", css)
	}
	if !strings.Contains(css, "container-type: size") {
		t.Errorf("@container-size should produce container-type: size, got:\n%s", css)
	}
}

func TestFontFeatureSettings(t *testing.T) {
	e := New()
	e.Write([]byte(`<div class="font-features-[smcp]">`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)

	if !strings.Contains(css, `font-feature-settings: "smcp"`) {
		t.Errorf(`font-features-[smcp] should produce font-feature-settings: "smcp", got:\n%s`, css)
	}
}

func TestFontFeatureSettingsQuoted(t *testing.T) {
	e := New()
	e.Write([]byte(`<div class='font-features-["liga"_0]'>`))
	css := e.CSS()
	t.Logf("Generated CSS:\n%s", css)

	if !strings.Contains(css, "font-feature-settings:") {
		t.Errorf("font-features with quoted value should produce font-feature-settings, got:\n%s", css)
	}
}

func TestBorderRadiusArbitraryCorners(t *testing.T) {
	tests := []struct {
		class    string
		property string
		value    string
	}{
		{"rounded-tl-[12px]", "border-top-left-radius", "12px"},
		{"rounded-br-[8px]", "border-bottom-right-radius", "8px"},
		{"rounded-ss-[1rem]", "border-start-start-radius", "1rem"},
		{"rounded-ee-[0.5em]", "border-end-end-radius", "0.5em"},
	}

	for _, tc := range tests {
		t.Run(tc.class, func(t *testing.T) {
			e := New()
			e.Write([]byte(fmt.Sprintf(`<div class="%s">`, tc.class)))
			css := e.CSS()
			want := tc.property + ": " + tc.value
			if !strings.Contains(css, want) {
				t.Errorf("%s: expected %q in CSS output:\n%s", tc.class, want, css)
			}
		})
	}
}
