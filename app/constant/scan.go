package constant

import (
	"regexp"
	"strconv"
	"strings"
)

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

// ScanRuleBind 扫描规则绑定
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

// ScanRuleStatus 扫描规则状态
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

// IsValid 判断状态是否有效
func (s ScanRuleStatus) IsValid() bool {
	statuses, _ := GetAllScanRuleStatuses()
	for _, status := range statuses {
		if s == status {
			return true
		}
	}
	return false
}

// IsEnabled 判断状态是否为启用
func (s ScanRuleStatus) IsEnabled() bool {
	return s == ScanRuleStatusEnabled
}

// IsDisabled 判断状态是否为禁用
func (s ScanRuleStatus) IsDisabled() bool {
	return s == ScanRuleStatusDisabled
}

// UpdateStatus 切换状态（enabled ↔ disabled）
func (s ScanRuleStatus) UpdateStatus() ScanRuleStatus {
	if s == ScanRuleStatusEnabled {
		return ScanRuleStatusDisabled
	}
	return ScanRuleStatusEnabled
}

// ScanRuleType 扫描规则类型
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

// ScanType 扫描类型
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

// ScanRuleUnit 扫描规则单元
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

// ScanPlanTriggerType 扫描计划触发器类型
type ScanPlanTriggerType string

const (
	ScanPlanTriggerTypeManual ScanPlanTriggerType = "manual"
	ScanPlanTriggerTypeCron   ScanPlanTriggerType = "cron"
	ScanPlanTriggerTypeAction ScanPlanTriggerType = "action"
)

type ScanTaskSort string

func (ScanTaskSort) Enum() []interface{}                  { return toInterfaceSlice(scanTaskSorts) }
func (s ScanTaskSort) Sanitize() (ScanTaskSort, bool)     { return Sanitize(s, GetAllScanTaskSorts) }
func GetAllScanTaskSorts() ([]ScanTaskSort, ScanTaskSort) { return scanTaskSorts, ScanTaskSortId }

const (
	ScanTaskSortId       ScanTaskSort = "id"
	ScanTaskSortStarted  ScanTaskSort = "started"
	ScanTaskSortFinished ScanTaskSort = "finished"
	ScanTaskSortRepoId   ScanTaskSort = "repo_id"
	ScanTaskSortStatus   ScanTaskSort = "status"
	ScanTaskSortResult   ScanTaskSort = "result"
)

var scanTaskSorts = sortEnum([]ScanTaskSort{
	ScanTaskSortId,
	ScanTaskSortStarted,
	ScanTaskSortFinished,
	ScanTaskSortRepoId,
	ScanTaskSortStatus,
	ScanTaskSortResult,
})

// ScanTaskStatus 扫描任务状态
type ScanTaskStatus string

func (ScanTaskStatus) Enum() []interface{}                { return toInterfaceSlice(scanTaskStatuses) }
func (s ScanTaskStatus) Sanitize() (ScanTaskStatus, bool) { return Sanitize(s, GetAllScanTaskStatuses) }
func GetAllScanTaskStatuses() ([]ScanTaskStatus, ScanTaskStatus) {
	return scanTaskStatuses, ""
}

const (
	ScanTaskStatusInProgress ScanTaskStatus = "in_progress"
	ScanTaskStatusSuccess    ScanTaskStatus = "success"
	ScanTaskStatusFailed     ScanTaskStatus = "failed"
)

var scanTaskStatuses = sortEnum([]ScanTaskStatus{
	ScanTaskStatusInProgress,
	ScanTaskStatusSuccess,
	ScanTaskStatusFailed,
})

// ScanTaskResult 扫描任务结果
type ScanTaskResult string

func (ScanTaskResult) Enum() []interface{}                { return toInterfaceSlice(scanTaskResults) }
func (s ScanTaskResult) Sanitize() (ScanTaskResult, bool) { return Sanitize(s, GetAllScanTaskResults) }
func GetAllScanTaskResults() ([]ScanTaskResult, ScanTaskResult) {
	return scanTaskResults, ""
}

const (
	ScanTaskResultPass   ScanTaskResult = "pass"
	ScanTaskResultNoPass ScanTaskResult = "no-pass"
	ScanTaskResultNull   ScanTaskResult = "-"
)

var scanTaskResults = sortEnum([]ScanTaskResult{
	ScanTaskResultPass,
	ScanTaskResultNoPass,
	ScanTaskResultNull,
})

// ScanMethod 扫描方法
type ScanMethod string

const (
	ScanMethodCheck ScanMethod = "check"
	ScanMethodSmell ScanMethod = "smell"
)

var scanMethods = sortEnum([]ScanMethod{
	ScanMethodCheck,
	ScanMethodSmell,
})

func (ScanMethod) Enum() []interface{}              { return toInterfaceSlice(scanMethods) }
func (s ScanMethod) Sanitize() (ScanMethod, bool)   { return Sanitize(s, GetAllScanMethods) }
func GetAllScanMethods() ([]ScanMethod, ScanMethod) { return scanMethods, ScanMethodCheck }
func (s ScanMethod) String() string                 { return string(s) }

// ScanIssueStatus 扫描问题状态
type ScanIssueStatus string

const (
	ScanIssueStatusWait    ScanIssueStatus = "wait"
	ScanIssueStatusTodo    ScanIssueStatus = "todo"
	ScanIssueStatusSolving ScanIssueStatus = "solving"
	ScanIssueStatusSolved  ScanIssueStatus = "solved"
	ScanIssueStatusClosed  ScanIssueStatus = "closed"
	ScanIssueStatusIgnore  ScanIssueStatus = "ignore"
)

var scanIssueStatuses = sortEnum([]ScanIssueStatus{
	ScanIssueStatusWait,
	ScanIssueStatusTodo,
	ScanIssueStatusSolving,
	ScanIssueStatusSolved,
	ScanIssueStatusClosed,
	ScanIssueStatusIgnore,
})

func (ScanIssueStatus) Enum() []interface{} { return toInterfaceSlice(scanIssueStatuses) }
func (s ScanIssueStatus) Sanitize() (ScanIssueStatus, bool) {
	return Sanitize(s, GetAllScanIssueStatuses)
}
func GetAllScanIssueStatuses() ([]ScanIssueStatus, ScanIssueStatus) {
	return scanIssueStatuses, ScanIssueStatusWait
}
func (s ScanIssueStatus) String() string { return string(s) }

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
	ScanRuleSortCreatedDate ScanRuleSort = "createdDate"
	ScanRuleSortLang        ScanRuleSort = "lang"
	ScanRuleSortPlugin      ScanRuleSort = "plugin"
	ScanRuleSortType        ScanRuleSort = "type"
	ScanRuleSortStatus      ScanRuleSort = "status"
	ScanRuleSortDesc        ScanRuleSort = "desc"
	scanRuleSortNone        ScanRuleSort = ""
)

var scanRuleSorts = sortEnum([]ScanRuleSort{
	ScanRuleSortID,
	ScanRuleSortName,
	ScanRuleSortPriority,
	ScanRuleSortCreatedDate,
	ScanRuleSortLang,
	ScanRuleSortPlugin,
	ScanRuleSortType,
	ScanRuleSortStatus,
	ScanRuleSortDesc,
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

type ScanIssueSort string

func (ScanIssueSort) Enum() []interface{}                    { return toInterfaceSlice(scanIssueSorts) }
func (s ScanIssueSort) Sanitize() (ScanIssueSort, bool)      { return Sanitize(s, GetAllScanIssueSorts) }
func GetAllScanIssueSorts() ([]ScanIssueSort, ScanIssueSort) { return scanIssueSorts, ScanIssueSortId }

const (
	ScanIssueSortId       ScanIssueSort = "id"
	ScanIssueSortPriority ScanIssueSort = "priority"
	ScanIssueSortType     ScanIssueSort = "type"
	ScanIssueSortStatus   ScanIssueSort = "status"
	ScanIssueSortCreated  ScanIssueSort = "created"
	ScanIssueSortLine     ScanIssueSort = "line"
)

var scanIssueSorts = sortEnum([]ScanIssueSort{
	ScanIssueSortId,
	ScanIssueSortPriority,
	ScanIssueSortType,
	ScanIssueSortStatus,
	ScanIssueSortCreated,
	ScanIssueSortLine,
})

// MaxIssueIDsPerRequest 单次请求允许的最大 issue ID 数量
const MaxIssueIDsPerRequest = 200

// ScanIssueResolution 扫描问题解决方案
type ScanIssueResolution string

const (
	ScanIssueResolutionNull         ScanIssueResolution = ""
	ScanIssueResolutionByDesign     ScanIssueResolution = "bydesign"
	ScanIssueResolutionDuplicate    ScanIssueResolution = "duplicate"
	ScanIssueResolutionExternal     ScanIssueResolution = "external"
	ScanIssueResolutionFixed        ScanIssueResolution = "fixed"
	ScanIssueResolutionNotReproduce ScanIssueResolution = "notrepro"
	ScanIssueResolutionPostponed    ScanIssueResolution = "postponed"
	ScanIssueResolutionWillNotFix   ScanIssueResolution = "willnotfix"
	ScanIssueResolutionToStory      ScanIssueResolution = "tostory"
)

var scanIssueResolutions = sortEnum([]ScanIssueResolution{
	ScanIssueResolutionByDesign, ScanIssueResolutionDuplicate, ScanIssueResolutionExternal,
	ScanIssueResolutionFixed, ScanIssueResolutionNotReproduce, ScanIssueResolutionPostponed,
	ScanIssueResolutionWillNotFix, ScanIssueResolutionToStory,
})

func (s ScanIssueResolution) String() string    { return string(s) }
func (ScanIssueResolution) Enum() []interface{} { return toInterfaceSlice(scanIssueResolutions) }
func (s ScanIssueResolution) Sanitize() (ScanIssueResolution, bool) {
	return Sanitize(s, GetAllScanIssueResolutions)
}
func GetAllScanIssueResolutions() ([]ScanIssueResolution, ScanIssueResolution) {
	return scanIssueResolutions, ScanIssueResolutionNull
}

type ScanLangID int

const (
	ScanLangIDCustom ScanLangID = iota
	ScanLangGo
	ScanLangJava
	ScanLangJavaScript
	ScanLangKotlin
	ScanLangPHP
	ScanLangPython
	ScanLangRuby
	ScanLangRust
	ScanLangTypeScript
	ScanLangC
	ScanLangCsharp
)

// String 实现 Stringer 接口，返回语言名称字符串
func (s ScanLangID) String() string {
	return string(s.ToName())
}

var scanLangIDs = sortEnum([]ScanLangID{
	ScanLangIDCustom,
	ScanLangGo,
	ScanLangJava,
	ScanLangJavaScript,
	ScanLangKotlin,
	ScanLangPHP,
	ScanLangPython,
	ScanLangRuby,
	ScanLangRust,
	ScanLangTypeScript,
	ScanLangC,
	ScanLangCsharp,
})

// ToName 将 ScanLangID 转换为对应的语言名称
func (s ScanLangID) ToName() ScanLangName {
	switch s {
	case ScanLangIDCustom:
		return ScanLangCustomName
	case ScanLangGo:
		return ScanLangGoName
	case ScanLangJava:
		return ScanLangJavaName
	case ScanLangJavaScript:
		return ScanLangJavaScriptName
	case ScanLangKotlin:
		return ScanLangKotlinName
	case ScanLangPHP:
		return ScanLangPHPName
	case ScanLangPython:
		return ScanLangPythonName
	case ScanLangRuby:
		return ScanLangRubyName
	case ScanLangRust:
		return ScanLangRustName
	case ScanLangTypeScript:
		return ScanLangTypeScriptName
	case ScanLangC:
		return ScanLangCName
	case ScanLangCsharp:
		return ScanLangCsharpName
	default:
		return ScanLangCustomName
	}
}
func (ScanLangID) Enum() []interface{}              { return toInterfaceSlice(scanLangIDs) }
func (s ScanLangID) Sanitize() (ScanLangID, bool)   { return Sanitize(s, GetAllScanLangIDs) }
func GetAllScanLangIDs() ([]ScanLangID, ScanLangID) { return scanLangIDs, ScanLangIDCustom }

type ScanLangName string

const (
	ScanLangCustomName     ScanLangName = "Custom"
	ScanLangGoName         ScanLangName = "Go"
	ScanLangJavaName       ScanLangName = "Java"
	ScanLangJavaScriptName ScanLangName = "JavaScript"
	ScanLangKotlinName     ScanLangName = "Kotlin"
	ScanLangPHPName        ScanLangName = "PHP"
	ScanLangPythonName     ScanLangName = "Python"
	ScanLangRubyName       ScanLangName = "Ruby"
	ScanLangRustName       ScanLangName = "Rust"
	ScanLangTypeScriptName ScanLangName = "TypeScript"
	ScanLangCName          ScanLangName = "C"
	ScanLangCsharpName     ScanLangName = "C#"
)

func (s ScanLangName) String() string {
	return string(s)
}

func (s ScanLangName) LowerString() string {
	return strings.ToLower(string(s))
}

func (s ScanLangName) ID() ScanLangID {
	switch s {
	case ScanLangCustomName:
		return ScanLangIDCustom
	case ScanLangGoName:
		return ScanLangGo
	case ScanLangJavaName:
		return ScanLangJava
	case ScanLangJavaScriptName:
		return ScanLangJavaScript
	case ScanLangKotlinName:
		return ScanLangKotlin
	case ScanLangPHPName:
		return ScanLangPHP
	case ScanLangPythonName:
		return ScanLangPython
	case ScanLangRubyName:
		return ScanLangRuby
	case ScanLangRustName:
		return ScanLangRust
	case ScanLangTypeScriptName:
		return ScanLangTypeScript
	case ScanLangCName:
		return ScanLangC
	case ScanLangCsharpName:
		return ScanLangCsharp
	default:
		return ScanLangIDCustom
	}
}

var scanLangs = sortEnum([]ScanLangName{
	ScanLangGoName,
	ScanLangJavaName,
	ScanLangJavaScriptName,
	ScanLangKotlinName,
	ScanLangPHPName,
	ScanLangPythonName,
	ScanLangRubyName,
	ScanLangRustName,
	ScanLangTypeScriptName,
	ScanLangCustomName,
	ScanLangCName,
	ScanLangCsharpName,
})

func (ScanLangName) Enum() []interface{}              { return toInterfaceSlice(scanLangs) }
func (s ScanLangName) Sanitize() (ScanLangName, bool) { return Sanitize(s, GetAllScanLangs) }
func GetAllScanLangs() ([]ScanLangName, ScanLangName) { return scanLangs, ScanLangCustomName }

// ParseLang 解析语言类型
func ParseLang(v string) ScanLangName {
	// 尝试将输入解析为数字
	if matched, _ := regexp.MatchString(`^[0-9]+$`, v); matched {
		if id, err := strconv.Atoi(v); err == nil {
			if id >= int(ScanLangIDCustom) && id <= int(ScanLangCsharp) {
				return ScanLangID(id).ToName()
			}
		}
	}
	// 按语言名称匹配
	switch strings.ToLower(v) {
	case ScanLangGoName.LowerString(), "golang":
		return ScanLangGoName
	case ScanLangJavaName.LowerString():
		return ScanLangJavaName
	case ScanLangJavaScriptName.LowerString():
		return ScanLangJavaScriptName
	case ScanLangKotlinName.LowerString():
		return ScanLangKotlinName
	case ScanLangPHPName.LowerString():
		return ScanLangPHPName
	case ScanLangPythonName.LowerString():
		return ScanLangPythonName
	case ScanLangRubyName.LowerString():
		return ScanLangRubyName
	case ScanLangRustName.LowerString():
		return ScanLangRustName
	case ScanLangTypeScriptName.LowerString():
		return ScanLangTypeScriptName
	case ScanLangCName.LowerString():
		return ScanLangCName
	case ScanLangCsharpName.LowerString():
		return ScanLangCsharpName
	default:
		return ScanLangCustomName
	}
}
