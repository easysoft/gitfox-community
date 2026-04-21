package webhook

import (
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	webhookStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

var _ Interface = (*Service)(nil)

type Service struct {
	webhookStore webhookStore.Interface
}

func NewWebhookSvc(webhookStore webhookStore.Interface) *Service {
	return &Service{webhookStore: webhookStore}
}

func (s *Service) Create(ctx context.Context, info webhookModel.Webhook) (*webhookModel.Webhook, error) {
	_, err := s.webhookStore.GetByRepoIDAndDisplayName(ctx, info.ParentID, info.DisplayName)
	if err == nil {
		return nil, errors.WithI18nKey(nil, i18n.WebhookDuplicateName)
	}
	if !errors.Is(err, errors.ErrResourceNotFound) {
		return nil, errors.WithI18nKey(err, i18n.WebhookCreateError)
	}

	triggerParts := make([]string, len(info.Triggers))
	for i, t := range info.Triggers {
		triggerParts[i] = string(t)
	}

	var repoID *int64
	var spaceID *int64
	if info.ParentType == string(constant.WebhookParentRepo) {
		repoID = &info.ParentID
	} else if info.ParentType == string(constant.WebhookParentSpace) {
		spaceID = &info.ParentID
	}

	webhook := &webhookStore.Webhook{
		SpaceID:     spaceID,
		RepoID:      repoID,
		DisplayName: info.DisplayName,
		Desc:        info.Description,
		URL:         info.URL,
		Secret:      info.Secret,
		IsSSL:       !info.Insecure,
		Triggers:    strings.Join(triggerParts, ","),
		Internal:    info.Internal,
		AuthMethod:  string(info.AuthMethod),
		AuthHeader:  info.AuthHeader,
		Enabled:     info.Enabled,
		CreatedBy:   info.CreatedBy,
		EditedBy:    info.UpdatedBy,
	}

	return s.webhookStore.Create(ctx, webhook)
}

func (s *Service) Get(ctx context.Context, webhookID uint) (*webhookModel.Webhook, error) {
	result, err := s.webhookStore.GetByID(ctx, webhookID)
	if err != nil {
		if errors.Is(err, errors.ErrResourceNotFound) {
			return nil, errors.WithI18nKey(err, i18n.WebhookNotFound)
		}
		return nil, err
	}
	return result, nil
}

func (s *Service) List(ctx context.Context, repoID int64, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Webhook], error) {
	return s.webhookStore.List(ctx, repoID, filter)
}

func (s *Service) Update(ctx context.Context, webhookID uint, req webhookModel.ReqForUpdate, updatedBy string) (*webhookModel.Webhook, error) {
	existing, err := s.webhookStore.GetByID(ctx, webhookID)
	if err != nil {
		if errors.Is(err, errors.ErrResourceNotFound) {
			return nil, errors.WithI18nKey(err, i18n.WebhookNotFound)
		}
		return nil, errors.WithI18nKey(err, i18n.WebhookUpdateError)
	}

	// 如果 DisplayName 有变更，检查同仓库下是否重名
	if req.DisplayName != nil && *req.DisplayName != "" && *req.DisplayName != existing.DisplayName {
		_, dupErr := s.webhookStore.GetByRepoIDAndDisplayName(ctx, existing.ParentID, *req.DisplayName)
		if dupErr == nil {
			return nil, errors.WithI18nKey(nil, i18n.WebhookDuplicateName)
		}
		if !errors.Is(dupErr, errors.ErrResourceNotFound) {
			return nil, errors.WithI18nKey(dupErr, i18n.WebhookUpdateError)
		}
	}

	updates := map[string]interface{}{
		"editedBy": updatedBy,
	}

	if req.DisplayName != nil && *req.DisplayName != "" {
		updates["displayName"] = *req.DisplayName
	}
	if req.URL != nil && *req.URL != "" {
		updates["url"] = *req.URL
	}
	if len(req.Triggers) > 0 {
		parts := make([]string, len(req.Triggers))
		for i, t := range req.Triggers {
			parts[i] = string(t)
		}
		updates["triggers"] = strings.Join(parts, ",")
	}
	if req.Description != nil {
		updates["desc"] = *req.Description
	}
	if req.Secret != nil {
		updates["secret"] = *req.Secret
	}
	if req.Insecure != nil {
		updates["isSSL"] = !*req.Insecure
	}
	if req.AuthMethod != nil {
		updates["authMethod"] = string(*req.AuthMethod)
	}
	if req.AuthHeader != nil {
		updates["authHeader"] = *req.AuthHeader
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	return s.webhookStore.Update(ctx, webhookID, updates)
}

func (s *Service) Delete(ctx context.Context, webhookID uint) error {
	_, err := s.webhookStore.GetByID(ctx, webhookID)
	if err != nil {
		if errors.Is(err, errors.ErrResourceNotFound) {
			return errors.WithI18nKey(err, i18n.WebhookNotFound)
		}
		return errors.WithI18nKey(err, i18n.WebhookDeleteError)
	}
	return s.webhookStore.Delete(ctx, webhookID)
}
