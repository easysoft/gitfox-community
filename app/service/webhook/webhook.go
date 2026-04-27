package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	webhookStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	pullreqStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pullrequest"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/stream"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

const (
	eventsReaderGroupName = "gitfox:webhook"
)

type Config struct {
	// UserAgentIdentity specifies the identity used for the user agent header
	// IMPORTANT: do not include version.
	UserAgentIdentity string
	// HeaderIdentity specifies the identity used for headers in webhook calls (e.g. X-Gitness-Trigger, ...).
	// NOTE: If no value is provided, the UserAgentIdentity will be used.
	HeaderIdentity      string
	EventReaderName     string
	Concurrency         int
	MaxRetries          int
	AllowPrivateNetwork bool
	AllowLoopback       bool

	// InternalWebhooksURL specifies the internal webhook URL which will be used if webhook is marked internal
	InternalWebhooksURL string
}

func (c *Config) Prepare() error {
	if c == nil {
		return errors.New("config is required")
	}
	if c.Concurrency < 1 {
		return errors.New("config.Concurrency has to be a positive number")
	}
	if c.MaxRetries < 0 {
		return errors.New("config.MaxRetries can't be negative")
	}

	return nil
}

var _ Interface = (*Service)(nil)

type Service struct {
	webhookStore webhookStore.Interface
	repoStore    repoStore.Interface
	pullreqStore pullreqStore.Interface
	urlProvider  url.Provider
	git          git.Interface
	config       Config

	secureHTTPClient           *http.Client
	insecureHTTPClient         *http.Client
	secureHTTPClientInternal   *http.Client
	insecureHTTPClientInternal *http.Client
}

func NewWebhookSvc(
	ctx context.Context,
	config Config,
	git git.Interface,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	prReaderFactory *events.ReaderFactory[*pullreqevents.Reader],
	webhookStore webhookStore.Interface,
	repoStore repoStore.Interface,
	pullreqStore pullreqStore.Interface,
	urlProvider url.Provider,
) (*Service, error) {
	if err := config.Prepare(); err != nil {
		return nil, fmt.Errorf("failed to prepare webhook config: %w", err)
	}

	service := &Service{
		config:                     config,
		git:                        git,
		webhookStore:               webhookStore,
		repoStore:                  repoStore,
		pullreqStore:               pullreqStore,
		urlProvider:                urlProvider,
		secureHTTPClient:           newHTTPClient(config.AllowLoopback, config.AllowPrivateNetwork, false),
		insecureHTTPClient:         newHTTPClient(config.AllowLoopback, config.AllowPrivateNetwork, true),
		secureHTTPClientInternal:   newHTTPClient(true, true, false),
		insecureHTTPClientInternal: newHTTPClient(true, true, true),
	}

	_, err := gitReaderFactory.Launch(ctx, eventsReaderGroupName, config.EventReaderName,
		func(r *gitevents.Reader) error {
			const idleTimeout = 1 * time.Minute
			r.Configure(
				stream.WithConcurrency(config.Concurrency),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(config.MaxRetries),
				))

			// register events
			_ = r.RegisterBranchCreated(service.handleEventBranchCreated)
			_ = r.RegisterBranchUpdated(service.handleEventBranchUpdated)
			_ = r.RegisterBranchDeleted(service.handleEventBranchDeleted)

			_ = r.RegisterTagCreated(service.handleEventTagCreated)
			_ = r.RegisterTagUpdated(service.handleEventTagUpdated)
			_ = r.RegisterTagDeleted(service.handleEventTagDeleted)

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to launch git event reader for webhooks: %w", err)
	}

	_, err = prReaderFactory.Launch(ctx, eventsReaderGroupName, config.EventReaderName,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 1 * time.Minute
			r.Configure(
				stream.WithConcurrency(config.Concurrency),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(config.MaxRetries),
				))

			// register events
			_ = r.RegisterCreated(service.handleEventPullReqCreated)
			_ = r.RegisterReopened(service.handleEventPullReqReopened)
			_ = r.RegisterBranchUpdated(service.handleEventPullReqBranchUpdated)
			_ = r.RegisterClosed(service.handleEventPullReqClosed)
			_ = r.RegisterMerged(service.handleEventPullReqMerged)

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to launch pr event reader for webhooks: %w", err)
	}
	return service, nil
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
		EditedBy:    info.EditedBy,
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

	updates := map[string]any{
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

func (s *Service) ListExecutions(ctx context.Context, webhookID uint, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Execution], error) {
	return s.webhookStore.ListExecutions(ctx, webhookID, filter)
}

func (s *Service) GetExecution(ctx context.Context, executionID uint) (*webhookModel.Execution, error) {
	result, err := s.webhookStore.GetExecutionByID(ctx, executionID)
	if err != nil {
		if errors.Is(err, errors.ErrResourceNotFound) {
			return nil, errors.WithI18nKey(err, i18n.WebhookExecutionNotFound)
		}
		return nil, err
	}
	return result, nil
}
