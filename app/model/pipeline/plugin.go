package pipeline

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"github.com/ghodss/yaml"
)

// PluginType 插件类型枚举
type PluginType string

const (
	PluginTypeStep  PluginType = "step"
	PluginTypeStage PluginType = "stage"
)

// Plugin 插件 Model 层结构体
type Plugin struct {
	ID          int64           `json:"id"`
	GroupID     int64           `json:"groupId"` // 插件分组 ID
	Name        string          `json:"name"`    // 插件名称（唯一标识符）
	Desc        string          `json:"desc"`    // 插件描述
	Logo        string          `json:"logo"`    // Logo (SVG)
	Yaml        string          `json:"yaml"`    // YAML 格式配置
	Json        string          `json:"json"`    // JSON 格式配置
	Type        PluginType      `json:"type"`    // 插件类型
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

const (
	DefaultPluginGroup = "default"
)

// PluginFilter 插件查询过滤器
type PluginFilter struct {
	Type     PluginType `form:"type" binding:"omitempty"`                   // 插件类型过滤
	Page     int        `form:"page" binding:"omitempty,min=1"`             // 页码，最小1
	PageSize int        `form:"pageSize" binding:"omitempty,min=1,max=100"` // 每页大小，1-100
}

type PluginGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// PluginWithGroup 插件及其分组信息
type PluginWithGroup struct {
	Plugin
	GroupName string `json:"groupName"`
	GroupDesc string `json:"groupDesc"`
}

// Matches 判断插件是否匹配
func (p *Plugin) Matches(v *Plugin) bool {
	if p.GroupID != v.GroupID {
		return false
	}
	if p.Name != v.Name {
		return false
	}
	if p.Desc != v.Desc {
		return false
	}
	if p.Yaml != v.Yaml {
		return false
	}
	if p.Json != v.Json {
		return false
	}
	if p.Logo != v.Logo {
		return false
	}
	return true
}

// yaml

// Config defines the a resource configuration.
type Yaml struct {
	Kind string      `json:"kind,omitempty"`
	Type PluginType  `json:"type,omitempty"`
	Name string      `json:"name,omitempty"`
	Spec interface{} `json:"spec,omitempty"`
}
type YamlSimp struct {
	Kind string     `json:"kind,omitempty"`
	Type PluginType `json:"type,omitempty"`
	Name string     `json:"name,omitempty"`
	Spec SpecSimple `json:"spec,omitempty"`
}

type SpecSimple struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*Yaml, error) {
	return Parse(
		bytes.NewBuffer(b),
	)
}

// Parse parses the configuration from io.Reader r.
func Parse(r io.Reader) (*Yaml, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = yaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}
	out := new(Yaml)
	err = json.Unmarshal(b, out)
	if err != nil {
		return out, err
	}

	// Parse Spec based on Type
	if out.Type == PluginTypeStep {
		var step PluginStep
		specBytes, err := json.Marshal(out.Spec)
		if err != nil {
			return out, err
		}
		if err := json.Unmarshal(specBytes, &step); err != nil {
			return out, err
		}
		out.Spec = &step
	} else if out.Type == PluginTypeStage {
		var stage PluginStage
		specBytes, err := json.Marshal(out.Spec)
		if err != nil {
			return out, err
		}
		if err := json.Unmarshal(specBytes, &stage); err != nil {
			return out, err
		}
		out.Spec = &stage
	}

	return out, nil
}

func ParseBytesSimple(b []byte) (*YamlSimp, error) {
	return ParseSimple(
		bytes.NewBuffer(b),
	)
}

// ParseSimple parses the configuration from io.Reader r.
func ParseSimple(r io.Reader) (*YamlSimp, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = yaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}
	out := new(YamlSimp)
	err = json.Unmarshal(b, out)
	if err != nil {
		return out, err
	}

	var spec SpecSimple
	specBytes, err := json.Marshal(out.Spec)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		return out, err
	}
	out.Spec = spec
	return out, nil
}
