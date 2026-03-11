package tailwind

import (
	"strings"
	"testing"
)

func TestParseKeyframes(t *testing.T) {
	css := []byte(`
@keyframes spin {
  to { transform: rotate(360deg); }
}
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Keyframes) != 1 {
		t.Fatalf("got %d keyframes, want 1", len(ss.Keyframes))
	}
	if ss.Keyframes[0].Name != "spin" {
		t.Errorf("name = %q, want %q", ss.Keyframes[0].Name, "spin")
	}
	if !strings.Contains(ss.Keyframes[0].Body, "@keyframes spin") {
		t.Errorf("body missing @keyframes spin: %s", ss.Keyframes[0].Body)
	}
}

func TestParseMultipleKeyframes(t *testing.T) {
	css := []byte(`
@keyframes spin {
  to { transform: rotate(360deg); }
}
@keyframes ping {
  75%, 100% { transform: scale(2); opacity: 0; }
}
`)
	ss, err := parseStylesheet(css)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss.Keyframes) != 2 {
		t.Fatalf("got %d keyframes, want 2", len(ss.Keyframes))
	}
	if ss.Keyframes[0].Name != "spin" {
		t.Errorf("first keyframe name = %q, want %q", ss.Keyframes[0].Name, "spin")
	}
	if ss.Keyframes[1].Name != "ping" {
		t.Errorf("second keyframe name = %q, want %q", ss.Keyframes[1].Name, "ping")
	}
}

func TestEndToEndAnimationKeyframes(t *testing.T) {
	css := []byte(`
@keyframes spin {
  to { transform: rotate(360deg); }
}
@utility animate-spin {
  animation: spin 1s linear infinite;
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="animate-spin"`))
	result := e.CSS()
	t.Logf("Generated CSS:\n%s", result)

	if !strings.Contains(result, "@keyframes spin") {
		t.Errorf("missing @keyframes spin: %s", result)
	}
	if !strings.Contains(result, "animation: spin") {
		t.Errorf("missing animation declaration: %s", result)
	}
}

func TestKeyframesBeforeUtilityRules(t *testing.T) {
	css := []byte(`
@keyframes spin {
  to { transform: rotate(360deg); }
}
@utility animate-spin {
  animation: spin 1s linear infinite;
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="animate-spin"`))
	result := e.CSS()

	kfIdx := strings.Index(result, "@keyframes spin")
	animIdx := strings.Index(result, ".animate-spin")
	if kfIdx < 0 || animIdx < 0 {
		t.Fatalf("missing content: kf=%d anim=%d\n%s", kfIdx, animIdx, result)
	}
	if kfIdx >= animIdx {
		t.Errorf("@keyframes should appear before utility rules:\n%s", result)
	}
}

func TestUnreferencedKeyframesNotEmitted(t *testing.T) {
	css := []byte(`
@keyframes spin {
  to { transform: rotate(360deg); }
}
@utility flex {
  display: flex;
}
`)
	e := New()
	e.LoadCSS(css)
	e.Write([]byte(`class="flex"`))
	result := e.CSS()

	if strings.Contains(result, "@keyframes") {
		t.Errorf("unreferenced @keyframes should not be emitted:\n%s", result)
	}
}

func TestMalformedKeyframesNoPanic(t *testing.T) {
	// Empty body
	css1 := []byte(`@keyframes {}`)
	ss1, err := parseStylesheet(css1)
	if err != nil {
		t.Fatal(err)
	}
	// Empty name should be skipped
	if len(ss1.Keyframes) != 0 {
		t.Errorf("expected 0 keyframes for empty name, got %d", len(ss1.Keyframes))
	}

	// Missing body
	css2 := []byte(`@keyframes spin;`)
	ss2, err := parseStylesheet(css2)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss2.Keyframes) != 0 {
		t.Errorf("expected 0 keyframes for missing body, got %d", len(ss2.Keyframes))
	}

	// Just @keyframes at EOF
	css3 := []byte(`@keyframes`)
	ss3, err := parseStylesheet(css3)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss3.Keyframes) != 0 {
		t.Errorf("expected 0 keyframes at EOF, got %d", len(ss3.Keyframes))
	}
}
