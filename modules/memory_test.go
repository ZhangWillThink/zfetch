package modules

import (
	"strings"
	"testing"
)

func TestMemoryModuleRun(t *testing.T) {
	m := &MemoryModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "Memory" {
		t.Errorf("expected Memory key, got %q", results[0].Key)
	}
	if !strings.Contains(results[0].Value, "MiB") {
		t.Errorf("expected MiB in value, got %q", results[0].Value)
	}
	if results[0].UsagePercent < 0 || results[0].UsagePercent > 100 {
		t.Errorf("UsagePercent out of range: %v", results[0].UsagePercent)
	}
}
