package modules

import (
	"testing"
)

func TestGPUModuleRun(t *testing.T) {
	m := &GPUModule{}
	results := m.Run()
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	for i, r := range results {
		if r.Value == "" {
			t.Errorf("GPU[%d] has empty value", i)
		}
	}
}
