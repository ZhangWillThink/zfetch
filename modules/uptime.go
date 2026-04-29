package modules

import (
	"fmt"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&UptimeModule{})
}

type UptimeModule struct{}

func (m *UptimeModule) Name() string { return "uptime" }

func (m *UptimeModule) Run() []ModuleInfo {
	info := sysinfo.GetUptime()
	if info.Uptime == 0 {
		return []ModuleInfo{{Key: "Uptime", Value: "Unknown"}}
	}

	days := info.Uptime / 86400
	hours := (info.Uptime % 86400) / 3600
	mins := (info.Uptime % 3600) / 60

	var val string
	if days > 0 {
		val = fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	} else {
		val = fmt.Sprintf("%dh %dm", hours, mins)
	}

	return []ModuleInfo{{Key: "Uptime", Value: val}}
}
