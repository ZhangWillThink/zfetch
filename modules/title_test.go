package modules

import (
	"testing"
)

func TestTitleModuleRun(t *testing.T) {
	m := &TitleModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Value != "" {
		t.Errorf("title value should be empty, got %q", results[0].Value)
	}
	if results[0].Key == "" {
		t.Error("title key should not be empty")
	}
}
