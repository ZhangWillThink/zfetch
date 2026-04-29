package modules

type ModuleInfo struct {
	Key           string
	Value         string
	UsagePercent  float64
}

type Module interface {
	Name() string
	Run() []ModuleInfo
}

type CategorizedModule struct {
	Category string
	Modules  []ModuleInfo
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
