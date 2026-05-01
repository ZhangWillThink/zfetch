package modules

import (
	"fmt"
	"math"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&MemoryModule{})
}

type MemoryModule struct{}

func (m *MemoryModule) Name() string { return "memory" }

func (m *MemoryModule) Run() []ModuleInfo {
	info := sysinfo.GetMemory()
	if info.Total == 0 {
		return []ModuleInfo{{Key: "Memory", Value: "Unknown"}}
	}

	usedMiB := float64(info.Used) / bytesPerMiB
	totalMiB := float64(info.Total) / bytesPerMiB
	pct := math.Round(clampPercent(float64(info.Used) / float64(info.Total) * 100))
	return []ModuleInfo{{
		Key:          "Memory",
		Value:        fmt.Sprintf("%.0f MiB / %.0f MiB (%.0f%%)", usedMiB, totalMiB, pct),
		UsagePercent: pct,
	}}
}
