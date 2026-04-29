package modules

import (
	"fmt"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

func init() {
	Register(&GPUModule{})
}

type GPUModule struct{}

func (m *GPUModule) Name() string { return "gpu" }

func (m *GPUModule) Run() []ModuleInfo {
	gpus := sysinfo.GetGPU()
	if len(gpus) == 0 {
		return []ModuleInfo{{Key: "GPU", Value: "Unknown"}}
	}

	var result []ModuleInfo
	for i, gpu := range gpus {
		key := "GPU"
		if i > 0 {
			key = fmt.Sprintf("GPU %d", i+1)
		}

		value := gpu.Name
		if gpu.MemoryMiB > 0 {
			value += fmt.Sprintf(" (%.2f GiB)", float64(gpu.MemoryMiB)/1024)
		}
		if gpu.Type != "" {
			value += fmt.Sprintf(" [%s]", gpu.Type)
		}

		result = append(result, ModuleInfo{Key: key, Value: value})
	}

	return result
}
