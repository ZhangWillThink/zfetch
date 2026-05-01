package modules

import (
	"testing"
)

type testModule struct {
	name string
	info []ModuleInfo
}

func (m *testModule) Name() string     { return m.name }
func (m *testModule) Run() []ModuleInfo { return m.info }

func TestRegisterAndGet(t *testing.T) {
	m := &testModule{name: "test_module", info: []ModuleInfo{{Key: "K", Value: "V"}}}
	Register(m)

	got := Get("test_module")
	if got == nil {
		t.Fatal("expected module, got nil")
	}
	if got.Name() != "test_module" {
		t.Errorf("expected name 'test_module', got %q", got.Name())
	}
}

func TestGetNotFound(t *testing.T) {
	got := Get("nonexistent")
	if got != nil {
		t.Errorf("expected nil for unknown module, got %v", got)
	}
}

func TestAllModules(t *testing.T) {
	names := AllModules()
	if len(names) == 0 {
		t.Error("AllModules returned empty list, expected at least registered modules")
	}

	found := false
	for _, n := range names {
		if n == "test_module" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'test_module' in AllModules result")
	}
}

func TestClampPercent(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0, 0},
		{50, 50},
		{100, 100},
		{-5, 0},
		{150, 100},
		{-0.1, 0},
		{100.5, 100},
	}
	for _, tc := range tests {
		got := clampPercent(tc.input)
		if got != tc.expected {
			t.Errorf("clampPercent(%v) = %v, want %v", tc.input, got, tc.expected)
		}
	}
}
