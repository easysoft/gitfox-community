package provider

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/provider_store.go -package=mock

import (
	"context"

	providerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/provider"
)

// Interface defines the provider store operations.
type Interface interface {
	// GetByID returns a provider by its ID.
	GetByID(ctx context.Context, id uint) (*providerModel.Provider, error)
	GetByIDs(ctx context.Context, ids []uint) (map[uint]*providerModel.Provider, error)
}
