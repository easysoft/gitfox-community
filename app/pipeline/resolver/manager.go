package resolver

import (
	"context"

	"github.com/easysoft/gitfox-community/app/service/pipeline"

	v1yaml "github.com/easysoft/gitfox-community/pkg/pipeline/spec"
	"github.com/easysoft/gitfox-community/pkg/pipeline/spec/parse"
)

type Manager struct {
	pipelineSvc	*pipeline.Service
}

func NewManager(
	pipelineSvc *pipeline.Service,
) *Manager {
	return &Manager{
		pipelineSvc: pipelineSvc,
	}
}

// Lookup returns a resource by name, kind and type. It also sends in the
// execution ID.
type LookupFunc func(name, kind, typ, version string, id int64) (*v1yaml.Config, error)

// GetLookupFn returns a lookup function for plugins and templates which can be used in the resolver
// passed to the drone runner.
func (m *Manager) GetLookupFn() LookupFunc {
    return func(name, kind, typ, version string, executionID int64) (*v1yaml.Config, error) {
        config, err := m.pipelineSvc.Resolve(context.Background(), kind, typ, name)
        if err != nil {
            return nil, err
        }
        return parse.ToConfig(config)
    }
}