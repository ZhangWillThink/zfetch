package modules

import (
	"strings"
	"testing"
)

func TestDiskModuleName(t *testing.T) {
	m := &DiskModule{}
	if m.Name() != "disk" {
		t.Errorf("expected 'disk', got %q", m.Name())
	}
}

func TestDiskModuleRun(t *testing.T) {
	m := &DiskModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "Disk" && !strings.HasPrefix(results[0].Key, "Disk") {
		t.Errorf("expected Disk key, got %q", results[0].Key)
	}
	for _, r := range results {
		if r.UsagePercent < 0 || r.UsagePercent > 100 {
			t.Errorf("UsagePercent out of range [0,100]: %v", r.UsagePercent)
		}
	}
}

func TestAllModuleNames(t *testing.T) {
	expected := []string{
		"title", "separator", "os", "kernel", "uptime", "packages",
		"shell", "resolution", "de", "wm", "terminal", "cpu", "gpu",
		"memory", "swap", "disk", "host", "battery", "localip", "locale",
	}
	for _, name := range expected {
		m := Get(name)
		if m == nil {
			t.Errorf("module %q not registered", name)
		}
	}
}
