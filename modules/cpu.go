package modules

import (
	"fmt"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&CPUModule{})
}

type CPUModule struct{}

func (m *CPUModule) Name() string { return "cpu" }

func (m *CPUModule) Run() []ModuleInfo {
	info := sysinfo.GetCPU()
	val := info.Name
	if info.Cores > 0 {
		val += fmt.Sprintf(" (%d)", info.Cores)
	}
	return []ModuleInfo{{Key: "CPU", Value: val}}
}
