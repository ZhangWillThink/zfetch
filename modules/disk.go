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
	disks := sysinfo.GetDisk()
	if len(disks) == 0 {
		return []ModuleInfo{{Key: "Disk", Value: "Unknown"}}
	}

	var result []ModuleInfo
	for _, disk := range disks {
		usedGiB := float64(disk.Used) / bytesPerGiB
		totalGiB := float64(disk.Total) / bytesPerGiB
		pct := math.Round(clampPercent(float64(disk.Used) / float64(disk.Total) * 100))

		key := "Disk"
		if len(disks) > 1 {
			key = fmt.Sprintf("Disk (%s)", disk.Path)
		}

		fs := ""
		if disk.Filesystem != "" {
			fs = fmt.Sprintf(" - %s", disk.Filesystem)
		}

		result = append(result, ModuleInfo{
			Key:          key,
			Value:        fmt.Sprintf("%.2f GiB / %.2f GiB (%.0f%%)%s", usedGiB, totalGiB, pct, fs),
			UsagePercent: pct,
		})
	}

	return result
}
