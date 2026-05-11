package display

import (
	"strings"
	"testing"
)

func TestConfigureColorPolicyWithEnv_NoColorDisablesANSI(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("FORCE_COLOR", "")
	t.Cleanup(func() { SetColorDisabled(false) })
	ConfigureColorPolicyWithEnv(false, "", "")
	got := Paint("x", "red")
	if strings.Contains(got, "\033[") {
		t.Fatalf("expected plain text without ANSI, got %q", got)
	}
}

func TestWrapParagraphSplitsWords(t *testing.T) {
	lines := wrapParagraph("aa bb cc dd ee", 5)
	if len(lines) < 2 {
		t.Fatalf("expected multiple lines, got %v", lines)
	}
}
