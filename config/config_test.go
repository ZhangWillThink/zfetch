package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig should not return nil")
	}
	if cfg.Separator != ": " {
		t.Errorf("expected separator ': ', got %q", cfg.Separator)
	}
	if cfg.ColorKeys != "default" {
		t.Errorf("expected ColorKeys 'default', got %q", cfg.ColorKeys)
	}
	if !strings.Contains(cfg.Structure, "title") {
		t.Error("default structure should contain 'title'")
	}
	if !strings.Contains(cfg.Structure, "separator") {
		t.Error("default structure should contain 'separator'")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.jsonc")

	content := `{
		"structure": "title:os",
		"separator": "=",
		"colorKeys": "red",
		"colorTitle": "bright_red"
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Structure != "title:os" {
		t.Errorf("expected structure 'title:os', got %q", cfg.Structure)
	}
	if cfg.Separator != "=" {
		t.Errorf("expected separator '=', got %q", cfg.Separator)
	}
	if cfg.ColorKeys != "red" {
		t.Errorf("expected ColorKeys 'red', got %q", cfg.ColorKeys)
	}
}

func TestLoadFromFileWithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.jsonc")

	content := `{
		// This is a comment
		"structure": "title:cpu:gpu",
		/* block comment */
		"separator": " ~ "
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Structure != "title:cpu:gpu" {
		t.Errorf("expected structure, got %q", cfg.Structure)
	}
	if cfg.Separator != " ~ " {
		t.Errorf("expected separator, got %q", cfg.Separator)
	}
}

func TestLoadFromFileMissing(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.jsonc")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFromFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.jsonc")
	os.WriteFile(path, []byte("{invalid}"), 0644)

	_, err := LoadFromFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFindPreset(t *testing.T) {
	dir := t.TempDir()
	presetPath := filepath.Join(dir, "mypreset.jsonc")
	if err := os.WriteFile(presetPath, []byte(`{"structure":"os"}`), 0644); err != nil {
		t.Fatal(err)
	}

	found, err := FindPreset(presetPath)
	if err != nil {
		t.Fatalf("FindPreset(absolute) should work: %v", err)
	}
	if found != presetPath {
		t.Errorf("expected %q, got %q", presetPath, found)
	}
}

func TestFindPresetNotFound(t *testing.T) {
	_, err := FindPreset("nonexistent_preset_xyz123")
	if err == nil {
		t.Error("expected error for nonexistent preset")
	}
}

func TestFindDefaultConfig(t *testing.T) {
	path := FindDefaultConfig()
	if path == "" {
		t.Skip("cannot determine home directory in test environment")
	}
	if !strings.HasSuffix(path, "config.jsonc") {
		t.Errorf("expected path ending with config.jsonc, got %q", path)
	}
}

func TestListConfigPaths(t *testing.T) {
	paths := ListConfigPaths()
	if len(paths) == 0 {
		t.Skip("cannot determine home directory in test environment")
	}
	for _, p := range paths {
		if !strings.Contains(p, "zfetch") {
			t.Errorf("expected 'zfetch' in path %q", p)
		}
	}
}

func TestConfigDefaultsPreserveOnPartialLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.jsonc")

	content := `{ "separator": "= " }`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Separator != "= " {
		t.Errorf("expected separator from file, got %q", cfg.Separator)
	}
	if cfg.ColorKeys != "default" {
		t.Errorf("expected default ColorKeys preserved, got %q", cfg.ColorKeys)
	}
	if !strings.Contains(cfg.Structure, "title") {
		t.Error("default structure should be preserved when not specified in file")
	}
}
