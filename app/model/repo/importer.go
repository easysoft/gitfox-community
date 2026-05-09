package repo

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
)

type ImportInput struct {
	SpaceID      uint     `json:"spaceID"`
	Identifier   string   `json:"identifier"`
	Description  string   `json:"description"`
	Provider     Provider `json:"provider"`
	ProviderRepo string   `json:"providerRepo"`
}

type RepositoryOutput struct {
	Repo
	IsPublic  bool `json:"isPublic" yaml:"isPublic"`
	Importing bool `json:"importing" yaml:"-"`
}

type Provider struct {
	Type     constant.RepoProviderType `json:"type"`
	Host     string                    `json:"host"`
	Username string                    `json:"username"`
	Password string                    `json:"password"`
}
