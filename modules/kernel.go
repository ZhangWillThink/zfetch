package modules

import (
	"fmt"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&KernelModule{})
}

type KernelModule struct{}

func (m *KernelModule) Name() string { return "kernel" }

func (m *KernelModule) Run() []ModuleInfo {
	info := sysinfo.GetKernel()
	return []ModuleInfo{{Key: "Kernel", Value: fmt.Sprintf("%s %s", info.Release, info.Machine)}}
}
