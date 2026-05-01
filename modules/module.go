package modules

const (
	bytesPerMiB = 1024 * 1024
	bytesPerGiB = 1024 * 1024 * 1024
)

type ModuleInfo struct {
	Key          string
	Value        string
	UsagePercent float64
}

type Module interface {
	Name() string
	Run() []ModuleInfo
}

var registry = make(map[string]Module)

func Register(m Module) {
	registry[m.Name()] = m
}

func Get(name string) Module {
	return registry[name]
}

func AllModules() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

func clampPercent(pct float64) float64 {
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}
