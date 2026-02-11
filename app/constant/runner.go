package constant

// RunnerStatus Runner 在线/离线状态
type RunnerStatus string

const (
	RunnerOnline  RunnerStatus = "online"
	RunnerOffline RunnerStatus = "offline"
)

// RunnerActiveStatus Runner 启用/禁用状态
type RunnerActiveStatus string

const (
	RunnerEnable  RunnerActiveStatus = "enable"
	RunnerDisable RunnerActiveStatus = "disable"
)
