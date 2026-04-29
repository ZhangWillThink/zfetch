package modules

import (
	"fmt"
	"math"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&DiskModule{})
}

type DiskModule struct{}

func (m *DiskModule) Name() string { return "disk" }

func (m *DiskModule) Run() []ModuleInfo {
	info := sysinfo.GetDisk()
	if info.Total == 0 {
		return []ModuleInfo{{Key: "Disk", Value: "Unknown"}}
	}

	usedGiB := float64(info.Used) / (1024 * 1024 * 1024)
	totalGiB := float64(info.Total) / (1024 * 1024 * 1024)
	pct := math.Round(float64(info.Used) / float64(info.Total) * 100)
	return []ModuleInfo{{
		Key:          "Disk",
		Value:        fmt.Sprintf("%.1f GiB / %.1f GiB (%.0f%%)", usedGiB, totalGiB, pct),
		UsagePercent: pct,
	}}
}
