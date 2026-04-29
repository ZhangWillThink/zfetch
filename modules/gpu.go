package modules

import (
	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&GPUModule{})
}

type GPUModule struct{}

func (m *GPUModule) Name() string { return "gpu" }

func (m *GPUModule) Run() []ModuleInfo {
	info := sysinfo.GetGPU()
	return []ModuleInfo{{Key: "GPU", Value: info.Name}}
}
