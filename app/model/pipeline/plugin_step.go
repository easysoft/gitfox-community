// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

type PluginStep struct {
	Name        string            `json:"name,omitempty"`
	Description string            `json:"description,omitempty"`
	Step        *Step             `json:"step,omitempty"`
	Inputs      map[string]*Input `json:"inputs,omitempty"`
}

type Step struct {
	Id       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Desc     string                 `json:"desc,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Timeout  string                 `json:"timeout,omitempty"`
	Strategy *Strategy              `json:"strategy,omitempty"`
	When     *When                  `json:"when,omitempty"`
	Failure  *FailureList           `json:"failure,omitempty"`
	Inputs   map[string]interface{} `json:"inputs,omitempty"`
	Spec     interface{}            `json:"spec,omitempty"`
}

// Input defines an input parameter
// 输入参数定义
type Input struct {
	Type        string      `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Items       *InputItems `json:"items,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Mask        bool        `json:"mask,omitempty"`
}

// InputItems defines an array input type
// 数组输入参数定义
type InputItems struct {
	Type string `json:"type,omitempty"`
}
