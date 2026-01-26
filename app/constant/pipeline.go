package constant

type PipelineOs string

const (
	PipelineOsWin   PipelineOs = "windows"
	PipelineOsMac   PipelineOs = "darwin"
	PipelineOsLinux PipelineOs = "linux"
)

var pipelineOsList = sortEnum([]PipelineOs{
	PipelineOsWin, PipelineOsMac, PipelineOsLinux,
})

func (PipelineOs) Enum() []interface{}            { return toInterfaceSlice(pipelineOsList) }
func (s PipelineOs) Sanitize() (PipelineOs, bool) { return Sanitize(s, GetAllPipelineOs) }
func GetAllPipelineOs() ([]PipelineOs, PipelineOs) {
	return pipelineOsList, PipelineOsLinux
}
func (s PipelineOs) String() string {
	return string(s)
}

type PipelineRunner string

const (
	PipelineRunnerDocker     PipelineRunner = "docker"
	PipelineRunnerKubernetes PipelineRunner = "kubernetes"
	PipelineRunnerAgent      PipelineRunner = "agent"
)

var pipelineRunners = sortEnum([]PipelineRunner{
	PipelineRunnerDocker, PipelineRunnerKubernetes, PipelineRunnerAgent,
})

func (PipelineRunner) Enum() []interface{}                { return toInterfaceSlice(pipelineRunners) }
func (s PipelineRunner) Sanitize() (PipelineRunner, bool) { return Sanitize(s, GetAllPipelineRunners) }
func GetAllPipelineRunners() ([]PipelineRunner, PipelineRunner) {
	return pipelineRunners, PipelineRunnerDocker
}
func (s PipelineRunner) String() string {
	return string(s)
}

type StepCategory string

const (
	StepCategoryBuild    StepCategory = "build"
	StepCategoryTest     StepCategory = "test"
	StepCategoryDeploy   StepCategory = "deploy"
	StepCategoryArtifact StepCategory = "artifact"
	//StepCategoryPublish  StepCategory = "publish"
)

var StepCategories = []StepCategory{
	StepCategoryBuild, StepCategoryArtifact, StepCategoryDeploy, StepCategoryTest,
}

func (StepCategory) Enum() []interface{}              { return toInterfaceSlice(StepCategories) }
func (s StepCategory) Sanitize() (StepCategory, bool) { return Sanitize(s, GetAllStepCategories) }
func GetAllStepCategories() ([]StepCategory, StepCategory) {
	return StepCategories, StepCategoryBuild
}
func (s StepCategory) String() string {
	return string(s)
}
func (s StepCategory) AliasName() string {
	switch s {
	case StepCategoryBuild:
		return "编译"
	case StepCategoryArtifact:
		return "制品"
	case StepCategoryDeploy:
		return "部署"
	case StepCategoryTest:
		return "测试"
	default:
		return s.String()
	}
}

type BuildEnvType string

const (
	BuildEnvTypeMaven  BuildEnvType = "maven"
	BuildEnvTypeGradle BuildEnvType = "gradle"
	BuildEnvTypeNpm    BuildEnvType = "npm"
	BuildEnvTypeYarn   BuildEnvType = "yarn"
	BuildEnvTypeGo     BuildEnvType = "go"

	BuildEnvCustom BuildEnvType = "custom"
)

var buildEnvTypes = sortEnum([]BuildEnvType{
	BuildEnvTypeMaven,
	BuildEnvTypeGradle,
	BuildEnvTypeNpm,
	BuildEnvTypeYarn,
	BuildEnvTypeGo,
})

func (BuildEnvType) Enum() []interface{}                  { return toInterfaceSlice(buildEnvTypes) }
func (s BuildEnvType) Sanitize() (BuildEnvType, bool)     { return Sanitize(s, GetAllBuildEnvTypes) }
func GetAllBuildEnvTypes() ([]BuildEnvType, BuildEnvType) { return buildEnvTypes, "" }

type BuildEnvCategory string

const (
	BuildEnvCategoryInternal BuildEnvCategory = "internal"
	BuildEnvCategoryExternal BuildEnvCategory = "external"
)

var buildEnvCategories = sortEnum([]BuildEnvCategory{
	BuildEnvCategoryInternal,
	BuildEnvCategoryExternal,
})

func (BuildEnvCategory) Enum() []interface{} { return toInterfaceSlice(buildEnvCategories) }
func (c BuildEnvCategory) Sanitize() (BuildEnvCategory, bool) {
	return Sanitize(c, GetAllBuildEnvCategories)
}
func GetAllBuildEnvCategories() ([]BuildEnvCategory, BuildEnvCategory) { return buildEnvCategories, "" }

func (c BuildEnvCategory) ID() BuildEnvCategoryID {
	switch c {
	case BuildEnvCategoryExternal:
		return BuildEnvCategoryExternalID
	default:
		return BuildEnvCategoryInternalID
	}
}

type BuildEnvCategoryID int

const (
	BuildEnvCategoryInternalID BuildEnvCategoryID = iota + 1
	BuildEnvCategoryExternalID
)

func (cid BuildEnvCategoryID) String() BuildEnvCategory {
	switch cid {
	case BuildEnvCategoryExternalID:
		return BuildEnvCategoryExternal
	default:
		return BuildEnvCategoryInternal
	}
}

type ImagePullPolicy string

func (s ImagePullPolicy) String() string { return string(s) }

const (
	ImagePullPolicyAlways       ImagePullPolicy = "always"
	ImagePullPolicyIfNotPresent ImagePullPolicy = "if-not-exists"
)

type BuildEnvScope string

const (
	BuildEnvScopeGlobal BuildEnvScope = "global"
)

func (c BuildEnvScope) ID() BuildEnvScopeID {
	switch c {
	default:
		return BuildEnvScopeGlobalID
	}
}

type BuildEnvScopeID int

const (
	BuildEnvScopeGlobalID BuildEnvScopeID = iota + 1
)

func (c BuildEnvScopeID) String() BuildEnvScope {
	switch c {
	default:
		return BuildEnvScopeGlobal
	}
}

type BuildEnvSort string

func (BuildEnvSort) Enum() []interface{}                  { return toInterfaceSlice(buildEnvSorts) }
func (s BuildEnvSort) Sanitize() (BuildEnvSort, bool)     { return Sanitize(s, GetAllBuildEnvSorts) }
func GetAllBuildEnvSorts() ([]BuildEnvSort, BuildEnvSort) { return buildEnvSorts, BuildEnvSortId }

const (
	BuildEnvSortId         BuildEnvSort = "id"
	BuildEnvSortType       BuildEnvSort = "type"
	BuildEnvSortIdentifier BuildEnvSort = "identifier"
	BuildEnvSortCreated    BuildEnvSort = "created"
	BuildEnvSortUpdated    BuildEnvSort = "updated"
)

var buildEnvSorts = sortEnum([]BuildEnvSort{
	BuildEnvSortId,
	BuildEnvSortType,
	BuildEnvSortIdentifier,
	BuildEnvSortCreated,
	BuildEnvSortUpdated,
})
