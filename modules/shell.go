package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&ShellModule{})
}

type ShellModule struct{}

func (m *ShellModule) Name() string { return "shell" }

func (m *ShellModule) Run() []ModuleInfo {
	info := sysinfo.GetShell()
	val := info.Name
	if info.Version != "" {
		val += " " + info.Version
	}
	return []ModuleInfo{{Key: "Shell", Value: val}}
}
