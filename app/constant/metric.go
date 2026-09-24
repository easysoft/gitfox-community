// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package constant

type MetricQueryStep string

func (MetricQueryStep) Enum() []interface{} {
	return toInterfaceSlice(metricQuerySteps)
}
func (s MetricQueryStep) Sanitize() (MetricQueryStep, bool) {
	return Sanitize(s, GetAllMetricQuerySteps)
}
func GetAllMetricQuerySteps() ([]MetricQueryStep, MetricQueryStep) {
	return metricQuerySteps, MetricQueryStepDay
}

const (
	MetricQueryStepDay   MetricQueryStep = "day"
	MetricQueryStepWeek  MetricQueryStep = "week"
	MetricQueryStepMonth MetricQueryStep = "month"
)

var metricQuerySteps = sortEnum([]MetricQueryStep{
	MetricQueryStepDay, MetricQueryStepWeek, MetricQueryStepMonth,
})

type MetricName string

func (s MetricName) String() string { return string(s) }
func (s MetricName) Sanitize() (MetricName, bool) {
	return Sanitize(s, GetAllMetricNames)
}

func GetAllMetricNames() ([]MetricName, MetricName) {
	return allMetricNames, ""
}

var allMetricNames = []MetricName{
	MetricName(MetricScanQueryAdded), MetricName(MetricScanQueryFixed),
}

type MetricScanQuery MetricName

const (
	MetricScanQueryAdded MetricScanQuery = "issue_added"
	MetricScanQueryFixed MetricScanQuery = "issue_fixed"
)

func (MetricScanQuery) Enum() []interface{} {
	return toInterfaceSlice(metricScanQueryNames)
}
func (s MetricScanQuery) Sanitize() (MetricScanQuery, bool) {
	return Sanitize(s, GetAllMetricScanQueryNames)
}

func GetAllMetricScanQueryNames() ([]MetricScanQuery, MetricScanQuery) {
	return metricScanQueryNames, ""
}

var metricScanQueryNames = sortEnum([]MetricScanQuery{
	MetricScanQueryAdded, MetricScanQueryFixed,
})

type MetricSourceKind string

const (
	MetricSourceKindScanIssue MetricSourceKind = "scan_issue"
	MetricSourceKindScanTask  MetricSourceKind = "scan_task"
)
