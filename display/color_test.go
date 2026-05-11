package display

import (
	"strings"
	"testing"
)

func TestGetColor(t *testing.T) {
	tests := []struct {
		name     string
		contains string
	}{
		{"black", "\033[30m"},
		{"red", "\033[31m"},
		{"green", "\033[32m"},
		{"bright_red", "\033[91m"},
		{"default", "\033[0m"},
		{"nonexistent", "\033[0m"},
	}
	for _, tc := range tests {
		got := GetColor(tc.name)
		if !strings.Contains(got, tc.contains) {
			t.Errorf("GetColor(%q) = %q, want contains %q", tc.name, got, tc.contains)
		}
	}
}

func TestPaint(t *testing.T) {
	SetColorDisabled(false)
	result := Paint("hello", "red")
	if !strings.Contains(result, "\033[31m") {
		t.Error("Paint should contain red color code")
	}
	if !strings.Contains(result, "\033[0m") {
		t.Error("Paint should contain reset code")
	}
	if !strings.Contains(result, "hello") {
		t.Error("Paint should contain text")
	}
	if strings.HasSuffix(result, "\033[0m") == false {
		t.Error("Paint should end with reset code")
	}
}

func TestPaintTitle(t *testing.T) {
	SetColorDisabled(false)
	result := PaintTitle("hello", "green")
	if !strings.Contains(result, "\033[1m") {
		t.Error("PaintTitle should contain bold code")
	}
	if !strings.Contains(result, "\033[32m") {
		t.Error("PaintTitle should contain green color code")
	}
	if !strings.Contains(result, "hello") {
		t.Error("PaintTitle should contain text")
	}
}

func TestColorExists(t *testing.T) {
	if !ColorExists("red") {
		t.Error("'red' should exist")
	}
	if ColorExists("nonexistent_color_xyz") {
		t.Error("unknown color should not exist")
	}
}

func TestUsageColor(t *testing.T) {
	tests := []struct {
		pct      float64
		expected string
	}{
		{-1, "default"},
		{0, "default"},
		{30, "bright_green"},
		{65, "bright_yellow"},
		{84, "bright_yellow"},
		{85, "bright_red"},
		{90, "bright_red"},
		{100, "bright_red"},
	}
	for _, tc := range tests {
		got := usageColor(tc.pct)
		if got != tc.expected {
			t.Errorf("usageColor(%v) = %q, want %q", tc.pct, got, tc.expected)
		}
	}
}
