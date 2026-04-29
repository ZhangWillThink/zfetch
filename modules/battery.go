package modules

import (
	"fmt"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&BatteryModule{})
}

type BatteryModule struct{}

func (m *BatteryModule) Name() string { return "battery" }

func (m *BatteryModule) Run() []ModuleInfo {
	info := sysinfo.GetBattery()
	if info.Status == "" && info.Percentage == 0 {
		return nil
	}

	value := fmt.Sprintf("%d%%", info.Percentage)
	if info.Status != "" {
		value += fmt.Sprintf(" [%s]", info.Status)
	}

	return []ModuleInfo{{Key: "Battery", Value: value}}
}
