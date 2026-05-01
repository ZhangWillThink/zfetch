package modules

import (
	"testing"
)

func TestPackagesModuleRun(t *testing.T) {
	m := &PackagesModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Key != "Packages" {
		t.Errorf("expected Packages key, got %q", results[0].Key)
	}
}
