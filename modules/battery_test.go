package modules

import (
	"testing"
)

func TestBatteryModuleName(t *testing.T) {
	m := &BatteryModule{}
	if m.Name() != "battery" {
		t.Errorf("expected 'battery', got %q", m.Name())
	}
}
