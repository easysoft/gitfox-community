package constant

import (
	"regexp"
	"strconv"
	"strings"
)

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
