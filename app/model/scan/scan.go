package scan

// ScanRulePriority 规则优先级
type ScanRulePriority string

const (
	ScanRulePriorityLow    ScanRulePriority = "low"
	ScanRulePriorityMedium ScanRulePriority = "medium"
	ScanRulePriorityHigh   ScanRulePriority = "high"
	scanRulePriorityNone   ScanRulePriority = ""
)

var scanRulePriorities = sortEnum([]ScanRulePriority{
	ScanRulePriorityLow,
	ScanRulePriorityMedium,
	ScanRulePriorityHigh,
})

func (s ScanRulePriority) String() string    { return string(s) }
func (ScanRulePriority) Enum() []interface{} { return toInterfaceSlice(scanRulePriorities) }
func (s ScanRulePriority) Sanitize() (ScanRulePriority, bool) {
	return Sanitize(s, GetAllScanRulePriorities)
}
func (s ScanRulePriority) IsValid() bool {
	priorities, _ := GetAllScanRulePriorities()
	for _, p := range priorities {
		if s == p {
			return true
		}
	}
	return false
}
func GetAllScanRulePriorities() ([]ScanRulePriority, ScanRulePriority) {
	return scanRulePriorities, scanRulePriorityNone
}

// ScanRuleSeverity 规则严重性
type ScanRuleSeverity string

const (
	ScanRuleSeverityLow      ScanRuleSeverity = "low"
	ScanRuleSeverityMedium   ScanRuleSeverity = "medium"
	ScanRuleSeverityHigh     ScanRuleSeverity = "high"
	ScanRuleSeverityCritical ScanRuleSeverity = "critical"
	scanRuleSeverityNone     ScanRuleSeverity = ""
)

var scanRuleSeverities = sortEnum([]ScanRuleSeverity{
	ScanRuleSeverityLow,
	ScanRuleSeverityMedium,
	ScanRuleSeverityHigh,
	ScanRuleSeverityCritical,
})

func (ScanRuleSeverity) Enum() []interface{} { return toInterfaceSlice(scanRuleSeverities) }
func (s ScanRuleSeverity) Sanitize() (ScanRuleSeverity, bool) {
	return Sanitize(s, GetAllScanRuleSeverities)
}
func (s ScanRuleSeverity) IsValid() bool {
	severities, _ := GetAllScanRuleSeverities()
	for _, sev := range severities {
		if s == sev {
			return true
		}
	}
	return false
}
func GetAllScanRuleSeverities() ([]ScanRuleSeverity, ScanRuleSeverity) {
	return scanRuleSeverities, scanRuleSeverityNone
}

type ScanRuleBind string

const (
	ScanRuleBindBind   ScanRuleBind = "bind"
	ScanRuleBindUnbind ScanRuleBind = "unbind"
	scanRuleBindNone   ScanRuleBind = ""
)

var scanRuleBinds = sortEnum([]ScanRuleBind{
	ScanRuleBindBind,
	ScanRuleBindUnbind,
})

func (ScanRuleBind) Enum() []interface{} { return toInterfaceSlice(scanRuleBinds) }
func (s ScanRuleBind) Sanitize() (ScanRuleBind, bool) {
	return Sanitize(s, GetAllScanRuleBinds)
}
func GetAllScanRuleBinds() ([]ScanRuleBind, ScanRuleBind) {
	return scanRuleBinds, scanRuleBindNone
}
func (s ScanRuleBind) IsValid() bool {
	binds, _ := GetAllScanRuleBinds()
	for _, b := range binds {
		if s == b {
			return true
		}
	}
	return false
}

type ScanRuleStatus string

const (
	ScanRuleStatusEnabled  ScanRuleStatus = "enabled"
	ScanRuleStatusDisabled ScanRuleStatus = "disabled"
	scanRuleStatusNone     ScanRuleStatus = ""
)

var scanRuleStatuses = sortEnum([]ScanRuleStatus{
	ScanRuleStatusEnabled,
	ScanRuleStatusDisabled,
})

func (ScanRuleStatus) Enum() []interface{}                { return toInterfaceSlice(scanRuleStatuses) }
func (s ScanRuleStatus) Sanitize() (ScanRuleStatus, bool) { return Sanitize(s, GetAllScanRuleStatuses) }
func GetAllScanRuleStatuses() ([]ScanRuleStatus, ScanRuleStatus) {
	return scanRuleStatuses, scanRuleStatusNone
}
func (s ScanRuleStatus) String() string {
	return string(s)
}

func (s ScanRuleStatus) IsValid() bool {
	statuses, _ := GetAllScanRuleStatuses()
	for _, st := range statuses {
		if s == st {
			return true
		}
	}
	return false
}

type ScanRuleType string

const (
	ScanRuleTypeDefect   ScanRuleType = "defect"     // 缺陷
	ScanRuleTypeVuln     ScanRuleType = "security"   // 安全
	ScanRuleTypeComplian ScanRuleType = "compliance" // 合规
	ScanRuleTypeOptimize ScanRuleType = "optimize"   // 优化 // 代码异味
	scanRuleTypeNone     ScanRuleType = ""
)

var scanRuleTypes = sortEnum([]ScanRuleType{
	ScanRuleTypeDefect,
	ScanRuleTypeVuln,
	ScanRuleTypeComplian,
	ScanRuleTypeOptimize,
})

func (s ScanRuleType) String() string    { return string(s) }
func (ScanRuleType) Enum() []interface{} { return toInterfaceSlice(scanRuleTypes) }
func (s ScanRuleType) Sanitize() (ScanRuleType, bool) {
	return Sanitize(s, GetAllScanRuleTypes)
}
func (s ScanRuleType) IsValid() bool {
	types, _ := GetAllScanRuleTypes()
	for _, t := range types {
		if s == t {
			return true
		}
	}
	return false
}
func GetAllScanRuleTypes() ([]ScanRuleType, ScanRuleType) {
	return scanRuleTypes, scanRuleTypeNone
}

type ScanType string

const (
	ScanTypeFull        ScanType = "full"
	ScanTypeIncremental ScanType = "incremental"
)

var scanTypes = sortEnum([]ScanType{
	ScanTypeFull,
	ScanTypeIncremental,
})

func (ScanType) Enum() []interface{} { return toInterfaceSlice(scanTypes) }
func (s ScanType) Sanitize() (ScanType, bool) {
	return Sanitize(s, GetAllScanTypes)
}
func GetAllScanTypes() ([]ScanType, ScanType) {
	return scanTypes, ScanTypeFull
}

type ScanRuleUnit string

const (
	ScanRuleUnitCount   ScanRuleUnit = "count"
	ScanRuleUnitPercent ScanRuleUnit = "percent"
)

var scanRuleUnits = sortEnum([]ScanRuleUnit{
	ScanRuleUnitCount,
	ScanRuleUnitPercent,
})

func (ScanRuleUnit) Enum() []interface{} { return toInterfaceSlice(scanRuleUnits) }
func (s ScanRuleUnit) Sanitize() (ScanRuleUnit, bool) {
	return Sanitize(s, GetAllScanRuleUnits)
}
func GetAllScanRuleUnits() ([]ScanRuleUnit, ScanRuleUnit) {
	return scanRuleUnits, ScanRuleUnitCount
}

// Order 排序方向
type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
	orderNone Order = ""
)

var orders = sortEnum([]Order{
	OrderAsc,
	OrderDesc,
})

func (Order) Enum() []interface{}       { return toInterfaceSlice(orders) }
func (o Order) Sanitize() (Order, bool) { return Sanitize(o, GetAllOrders) }
func GetAllOrders() ([]Order, Order) {
	return orders, OrderDesc
}
func (o Order) String() string {
	return string(o)
}

// IsValid 判断排序方向是否有效
func (o Order) IsValid() bool {
	return o == OrderAsc || o == OrderDesc
}

// ScanRuleSort 扫描规则排序字段
type ScanRuleSort string

const (
	ScanRuleSortID          ScanRuleSort = "id"
	ScanRuleSortName        ScanRuleSort = "name"
	ScanRuleSortPriority    ScanRuleSort = "priority"
	ScanRuleSortCreatedDate ScanRuleSort = "created_date"
	scanRuleSortNone        ScanRuleSort = ""
)

var scanRuleSorts = sortEnum([]ScanRuleSort{
	ScanRuleSortID,
	ScanRuleSortName,
	ScanRuleSortPriority,
	ScanRuleSortCreatedDate,
})

func (ScanRuleSort) Enum() []interface{} { return toInterfaceSlice(scanRuleSorts) }
func (s ScanRuleSort) Sanitize() (ScanRuleSort, bool) {
	return Sanitize(s, GetAllScanRuleSorts)
}
func GetAllScanRuleSorts() ([]ScanRuleSort, ScanRuleSort) {
	return scanRuleSorts, ScanRuleSortID
}
func (s ScanRuleSort) String() string {
	return string(s)
}

// IsValid 判断排序字段是否有效
func (s ScanRuleSort) IsValid() bool {
	sorts, _ := GetAllScanRuleSorts()
	for _, sort := range sorts {
		if s == sort {
			return true
		}
	}
	return false
}
