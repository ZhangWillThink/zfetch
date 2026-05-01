package modules

import (
	"testing"
)

func TestCPUModuleName(t *testing.T) {
	m := &CPUModule{}
	if m.Name() != "cpu" {
		t.Errorf("expected 'cpu', got %q", m.Name())
	}
}

func TestCPUModuleRun(t *testing.T) {
	m := &CPUModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "CPU" {
		t.Errorf("expected CPU key, got %q", results[0].Key)
	}
	if results[0].Value == "" {
		t.Error("expected non-empty CPU value")
	}
}
