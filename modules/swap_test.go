package modules

import (
	"testing"
)

func TestSwapModuleRun(t *testing.T) {
	m := &SwapModule{}
	results := m.Run()
	for _, r := range results {
		if r.UsagePercent < 0 || r.UsagePercent > 100 {
			t.Errorf("UsagePercent out of range: %v", r.UsagePercent)
		}
	}
}
