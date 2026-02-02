package yaml

type StepAssetDownload struct {
	AssetDir string                   `json:"assetDir,omitempty"`
	Assets   []*StepAssetDownloadItem `json:"assets,omitempty"`
}
