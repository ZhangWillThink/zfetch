package display

import (
	"strings"
	"testing"

	"github.com/WillZhang/zfetch/config"
	"github.com/WillZhang/zfetch/modules"
)

func TestNewDisplay(t *testing.T) {
	d := New(config.DefaultConfig(), false)
	if d == nil {
		t.Fatal("New should not return nil")
	}
	if d.cfg.ColorKeys != "bright_green" {
		t.Errorf("expected ColorKeys 'bright_green' after New, got %q", d.cfg.ColorKeys)
	}
	if d.cfg.ColorTitle != "bright_white" {
		t.Errorf("expected ColorTitle 'bright_white' after New, got %q", d.cfg.ColorTitle)
	}
}

func TestNewDisplayNilConfig(t *testing.T) {
	d := New(nil, false)
	if d == nil {
		t.Fatal("New(nil) should not return nil")
	}
	if d.cfg == nil {
		t.Fatal("New(nil) should produce a valid config")
	}
}

func TestNewDisplayPipe(t *testing.T) {
	d := New(config.DefaultConfig(), true)
	if !d.pipe {
		t.Error("expected pipe=true")
	}
}

func TestSplitColored(t *testing.T) {
	tests := []struct {
		input    string
		isTitle  bool
		wantKey  string
		wantRest string
	}{
		{"hello world", false, "hello", " world"},
		{"hello", false, "hello", ""},
		{"key = value", false, "key", " = value"},
	}

	for _, tc := range tests {
		coloredKey, rest := splitColored(tc.input, 0, "bright_green", tc.isTitle)

		if tc.wantKey == "hello" {
			if !strings.Contains(coloredKey, "\033[") {
				t.Error("colored key should contain ANSI codes")
			}
		}

		if rest != tc.wantRest {
			t.Errorf("splitColored(%q) rest = %q, want %q", tc.input, rest, tc.wantRest)
		}
	}

	coloredKey, rest := splitColored("hello", 0, "bright_white", true)
	if !strings.Contains(coloredKey, "\033[1m") {
		t.Error("title should contain bold code")
	}
	if rest != "" {
		t.Errorf("title with no separator should have empty rest, got %q", rest)
	}
}

func TestRenderPipe(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Separator = ": "
	d := &Display{cfg: cfg, pipe: true}

	infos := []modules.ModuleInfo{
		{Key: "OS", Value: "Ubuntu 24.04"},
		{Key: "Kernel", Value: "6.8.0"},
		{Key: "", Value: ""},
		{Key: "separator", Value: ""},
	}

	d.renderPipe(infos)
}

func TestDisplayCfgNotMutated(t *testing.T) {
	cfg := config.DefaultConfig()
	origKeys := cfg.ColorKeys
	origTitle := cfg.ColorTitle

	_ = New(cfg, false)

	if cfg.ColorKeys != origKeys {
		t.Error("New should not mutate original config ColorKeys")
	}
	if cfg.ColorTitle != origTitle {
		t.Error("New should not mutate original config ColorTitle")
	}
}

func TestGetTerminalWidth(t *testing.T) {
	w := getTerminalWidth()
	if w < 0 {
		t.Errorf("getTerminalWidth returned negative: %d", w)
	}
}

func TestRenderInline(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Separator = ": "
	d := &Display{cfg: cfg, pipe: false}

	infos := []modules.ModuleInfo{
		{Key: "User@host", Value: ""},
		{Key: "separator", Value: ""},
		{Key: "OS", Value: "Ubuntu 24.04"},
		{Key: "CPU", Value: "Intel i7 (8)"},
		{Key: "Memory", Value: "2048 MiB / 16384 MiB (13%)", UsagePercent: 13},
	}

	d.renderInline(infos)
}
