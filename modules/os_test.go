package modules

import (
	"testing"
)

func TestOSModuleRun(t *testing.T) {
	m := &OSModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "OS" {
		t.Errorf("expected OS key, got %q", results[0].Key)
	}
	if results[0].Value == "" {
		t.Error("expected non-empty OS value")
	}
}
