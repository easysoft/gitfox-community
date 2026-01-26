package pipeline

type PipelineConfigSource string

const (
	PipelineConfigSourceInner PipelineConfigSource = "inner"
	PipelineConfigSourceGit   PipelineConfigSource = "git"

	PipelineConfigPathInnerPrefix = "_inner/"
)

type PipelineConfigPath struct {
	Source PipelineConfigSource
	Path   string
	Name   string
}

type PipelineContentResponse struct {
	Content string `json:"content"`
}

type PipelineContent struct {
	PipelineId int64  `json:"pipeline_id"`
	Version    int64  `json:"version"`
	Success    bool   `json:"success"`
	Created    int64  `json:"created"`
	Updated    int64  `json:"updated"`
	Data       []byte `json:"data"`
}

type PipelineContentValidateResponse []SchemaInvalidError

type SchemaInvalidError struct {
	Type   string                 `json:"type"`
	Field  string                 `json:"field"`
	Value  interface{}            `json:"value"`
	Detail map[string]interface{} `json:"detail"`
}

type SimpleStep struct {
	Name   string `json:"name"`
	NameCn string `json:"alias"`
	Svg    string `json:"svg"`
}

// GroupedPlugins 按分组归类的插件
type GroupedPlugins struct {
	GroupID   int64         `json:"groupId"`
	GroupName string        `json:"groupName"`
	Desc      string        `json:"desc"`
	Plugins   []*SimpleStep `json:"plugins"`
}

type StepFilter struct {
	Name     string
	Category string
}
