package modules

import (
	"strings"
	"testing"
)

func TestUptimeModuleRun(t *testing.T) {
	m := &UptimeModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "Uptime" {
		t.Errorf("expected Uptime key, got %q", results[0].Key)
	}
	val := results[0].Value
	if val == "" {
		t.Error("expected non-empty uptime value")
	}
	if !strings.Contains(val, "h") && !strings.Contains(val, "m") && val != "Unknown" {
		t.Errorf("unexpected uptime format: %q", val)
	}
}
