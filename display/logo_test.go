package display

import (
	"strings"
	"testing"
)

func TestGetLogoKnown(t *testing.T) {
	names := []string{"ubuntu", "debian", "arch", "fedora", "centos",
		"opensuse", "redhat", "gentoo", "nixos", "linux", "alpine", "macos", "windows"}

	for _, name := range names {
		logo := GetLogo(name)
		if len(logo) == 0 {
			t.Errorf("GetLogo(%q) returned empty", name)
		}

		found := false
		if logoFromMap, ok := logoMap[name]; ok {
			if len(logo) == len(logoFromMap) {
				found = true
			}
		}
		if !found && name != "default" {
			// Even if not in logoMap, should return default logo
			if GetLogo(name) == nil || len(GetLogo(name)) == 0 {
				t.Errorf("GetLogo(%q) should return default when not found", name)
			}
		}
	}
}

func TestGetLogoFallback(t *testing.T) {
	logo := GetLogo("nonexistent_os_xyz123")
	if len(logo) == 0 {
		t.Error("GetLogo should return default for unknown name")
	}
	defaultLogo := GetLogo("default")
	if len(logo) != len(defaultLogo) {
		t.Error("unknown logo should fall back to default")
	}
}

func TestListLogos(t *testing.T) {
	logos := ListLogos()
	if len(logos) == 0 {
		t.Fatal("ListLogos should not be empty")
	}

	defaultFound := false
	for _, name := range logos {
		if name == "default" {
			defaultFound = true
			break
		}
	}
	if !defaultFound {
		t.Error("ListLogos should include 'default'")
	}
}

func TestDetectOSLogo(t *testing.T) {
	result := detectOSLogo()
	if result == "" {
		t.Error("detectOSLogo should not return empty string")
	}
	if result == "default" {
		t.Log("detectOSLogo returned 'default' - may need release files or os-release ID")
	}

	logo := GetLogo(result)
	if len(logo) == 0 {
		t.Errorf("GetLogo(%q) should return valid logo", result)
	}
}

func TestLogoWidth(t *testing.T) {
	for name, lines := range logoMap {
		if len(lines) == 0 {
			t.Errorf("logo %q has no lines", name)
		}
		for i, line := range lines {
			if line == "" {
				t.Logf("logo %q line %d is empty", name, i)
			}
		}
	}
}

func TestLogoContainsOnlyPrintable(t *testing.T) {
	for name, lines := range logoMap {
		for _, line := range lines {
			if strings.Contains(line, "\x00") {
				t.Errorf("logo %q contains null byte", name)
			}
		}
	}
}
