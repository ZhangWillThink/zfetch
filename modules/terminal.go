package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&TerminalModule{})
}

type TerminalModule struct{}

func (m *TerminalModule) Name() string { return "terminal" }

func (m *TerminalModule) Run() []ModuleInfo {
	info := sysinfo.GetTerminal()
	val := info.Name
	if info.Version != "" {
		val += " " + info.Version
	}
	return []ModuleInfo{{Key: "Terminal", Value: val}}
}
