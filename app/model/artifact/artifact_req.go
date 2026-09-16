package artifact

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

type EntityID string

func (e EntityID) GetEntityID() SplitEntityID {
	return GetEntityID(e)
}

func isEntityIDEmpty(EntityID EntityID) bool {
	return EntityID == ""
}

func (r *EntityID) Validate() error {
	if isEntityIDEmpty(*r) {
		return errors.WithI18nKey(errors.New("entity id is empty"), i18n.ArtifactEntityIdEmpty)
	}
	if err := validateEntityID(*r); err != nil {
		return err
	}
	return nil
}

// ArtifactEntitiesReq 恢复请求体，复用 EntityID 串（与软删保持一致）。
type ArtifactEntitiesReq struct {
	EntityIDs []EntityID `json:"entityIDs,omitempty"`
}

func (r *ArtifactEntitiesReq) Validate() error {
	if len(r.EntityIDs) == 0 {
		return errors.WithI18nKey(errors.New("entity ids is empty"), i18n.ArtifactEntityIdEmpty)
	}
	for _, id := range r.EntityIDs {
		if err := validateEntityID(id); err != nil {
			return err
		}
	}
	return nil
}

type ArtifactEntitieReq struct {
	EntityID EntityID `json:"entityID,omitempty"`
}

func (r *ArtifactEntitieReq) Validate() error {
	if r.EntityID == "" {
		return errors.WithI18nKey(errors.New("entity id is empty"), i18n.ArtifactEntityIdEmpty)
	}
	if err := validateEntityID(r.EntityID); err != nil {
		return err
	}
	return nil
}

func (r *ArtifactEntitieReq) GetEntityID() SplitEntityID {
	return GetEntityID(r.EntityID)
}

func validateEntityID(EntityID EntityID) error {
	// 校验 entityID 格式
	if strings.Count(string(EntityID), ".") != 1 {
		return errors.WithI18nKey(errors.New("invalid entity id format"), i18n.ArtifactEntityIdInvalid)
	}
	parts := strings.SplitN(string(EntityID), ".", 2)
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return errors.WithI18nKey(errors.New("invalid entity id number"), i18n.ArtifactEntityIdInvalid)
	}
	// 校验 kind 是否为有效 ArtifactEntityKind
	kind := constant.ArtifactEntityKind(parts[0])
	_, ok := kind.Sanitize()
	if !ok {
		return errors.WithI18nKey(fmt.Errorf("unsupported kind: %s", kind), i18n.ArtifactEntityIdKindInvalid)
	}
	return nil
}

type SplitEntityID struct {
	Kind constant.ArtifactEntityKind
	ID   int64
}

func GetEntityID(EntityID EntityID) SplitEntityID {
	parts := strings.SplitN(string(EntityID), ".", 2)
	if len(parts) != 2 {
		return SplitEntityID{"", 0}
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return SplitEntityID{"", 0}
	}

	return SplitEntityID{Kind: constant.ArtifactEntityKind(parts[0]), ID: id}
}
