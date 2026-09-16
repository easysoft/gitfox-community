package pipeline

import "github.com/easysoft/gitfox-community/pkg/utils/datetime"

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
	Kind   string `json:"kind"`
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

// PipelineFile represents the raw file contents in the pipeline content.
type PipelineFile struct {
	Data []byte
}

type Pipeline struct {
	ID               int64           `json:"id"`
	Name             string          `json:"name"`
	Engine           string          `json:"engine"`
	ProviderID       int64           `json:"providerID"`
	Scope            string          `json:"scope"`
	SpaceID          int64           `json:"spaceID"`
	RepoID           int64           `json:"repoID"`
	Desc             string          `json:"desc"`
	Status           string          `json:"status"`
	LatestVersion    int             `json:"latestVersion"`
	DefaultBranch    string          `json:"defaultBranch"`
	YamlPath         string          `json:"yamlPath"`
	CustomParam      string          `json:"customParam"`
	ExternalPipeline string          `json:"externalPipeline"`
	LastExec         *datetime.Local `json:"lastExec"`
	LastResult       string          `json:"lastResult"`
	CreatedBy        string          `json:"createdBy"`
	CreatedDate      *datetime.Local `json:"createdDate"`
	EditedBy         string          `json:"editedBy"`
	EditedDate       *datetime.Local `json:"editedDate"`
	Deleted          int8            `json:"deleted"`
}

type PipelineContent struct {
	ID          int64           `json:"id"`
	PipelineID  int64           `json:"pipelineID"`
	Version     int             `json:"version"`
	Data        []byte          `json:"data"`
	DataStr     string          `json:"dataStr"`
	Variables   string          `json:"variables"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"`
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"`
}

type PipelineVariable struct {
	Name         string `json:"name"`
	Key          string `json:"key"`
	Runtime      bool   `json:"runtime"`
	Value        string `json:"value"`
	DefaultValue string `json:"defaultValue"`
}
