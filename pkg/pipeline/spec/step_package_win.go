package yaml

type StepWinPackage struct {
	Tool          string `json:"tool,omitempty"`
	AppName       string `json:"appName,omitempty"`
	AppVersion    string `json:"appVersion,omitempty"`
	AppPublisher  string `json:"appPublisher,omitempty"`
	AppBaseDir    string `json:"appBaseDir,omitempty"`
	AppInstallDir string `json:"appInstallDir,omitempty"`
	AppMainFile   string `json:"appMainFile,omitempty"`
	AssetDir      string `json:"assetDir,omitempty"`
	FileName      string `json:"fileName,omitempty"`
}
