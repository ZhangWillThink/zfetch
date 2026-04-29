package modules

import (
	"fmt"
	"math"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&SwapModule{})
}

type SwapModule struct{}

func (m *SwapModule) Name() string { return "swap" }

func (m *SwapModule) Run() []ModuleInfo {
	info := sysinfo.GetSwap()
	if info.Total == 0 {
		return nil
	}

	usedGiB := float64(info.Used) / (1024 * 1024 * 1024)
	totalGiB := float64(info.Total) / (1024 * 1024 * 1024)
	pct := math.Round(float64(info.Used) / float64(info.Total) * 100)

	return []ModuleInfo{{
		Key:          "Swap",
		Value:        fmt.Sprintf("%.2f GiB / %.2f GiB (%.0f%%)", usedGiB, totalGiB, pct),
		UsagePercent: pct,
	}}
}
