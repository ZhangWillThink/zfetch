package modules

import (
	"fmt"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&HostModule{})
}

type HostModule struct{}

func (m *HostModule) Name() string { return "host" }

func (m *HostModule) Run() []ModuleInfo {
	info := sysinfo.GetHost()
	if info.Product == "" {
		return nil
	}

	value := info.Product
	if info.Version != "" {
		value += fmt.Sprintf(" (%s)", info.Version)
	}

	return []ModuleInfo{{Key: "Host", Value: value}}
}
