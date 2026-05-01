package modules

import (
	"testing"
)

func TestShellModuleRun(t *testing.T) {
	m := &ShellModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "Shell" {
		t.Errorf("expected Shell key, got %q", results[0].Key)
	}
	if results[0].Value == "" {
		t.Error("expected non-empty Shell value")
	}
}
