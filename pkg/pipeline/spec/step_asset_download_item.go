package yaml

type StepAssetDownloadItem struct {
	SpaceRef          string `json:"spaceRef,omitempty"`
	Format            string `json:"format,omitempty"`
	Type              string `json:"type,omitempty"`
	FilePath          string `json:"filePath,omitempty"`
	DisableUnCompress bool   `json:"disableUnCompress,omitempty"`
	GroupId           int64  `json:"groupId,omitempty"`
	DestPath          string `json:"destPath,omitempty"`
	Location          string `json:"location,omitempty"`
}
