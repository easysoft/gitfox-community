package webhook

import (
	"context"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

type Store struct {
	db *gorm.DB
}

func NewWebhookStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ctx context.Context, webhook *Webhook) (*webhookModel.Webhook, error) {
	if err := s.db.WithContext(ctx).Create(webhook).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(*webhook), nil
}

func (s *Store) GetByRepoIDAndDisplayName(ctx context.Context, repoID int64, displayName string) (*webhookModel.Webhook, error) {
	var webhook Webhook
	err := s.db.WithContext(ctx).
		Where("repoID = ? AND displayName = ? AND deleted = 0", repoID, displayName).
		First(&webhook).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(webhook), nil
}

func (s *Store) GetByID(ctx context.Context, id uint) (*webhookModel.Webhook, error) {
	var webhook Webhook
	err := s.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&webhook).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapToModel(webhook), nil
}

func (s *Store) Update(ctx context.Context, id uint, updates map[string]interface{}) (*webhookModel.Webhook, error) {
	if err := s.db.WithContext(ctx).Model(&Webhook{}).Where("id = ? AND deleted = 0", id).Updates(updates).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return s.GetByID(ctx, id)
}

func (s *Store) Delete(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&Webhook{}).Where("id = ? AND deleted = 0", id).Update("deleted", 1).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// allowedSortFields 限制排序字段白名单，防止 SQL 注入
var allowedSortFields = map[string]string{
	"displayName":           "displayName",
	"enabled":               "enabled",
	"latestExecutionResult": "latestExecResult",
}

func (s *Store) List(ctx context.Context, repoID int64, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Webhook], error) {
	tx := s.db.WithContext(ctx).Model(&Webhook{}).Where("repoID = ? AND deleted = 0", repoID)

	sort := "createdDate"
	if col, ok := allowedSortFields[filter.Sort]; ok {
		sort = col
	}

	var rows []*Webhook
	pageResult, err := database.Paginate(ctx, tx, database.PageParam{
		Page:      filter.Page,
		PageSize:  filter.PageSize,
		Sort:      sort,
		Order:     filter.Order,
		NeedTotal: true,
	}, &rows)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	data := make([]*webhookModel.Webhook, len(rows))
	for i, row := range rows {
		data[i] = mapToModel(*row)
	}

	return &database.PageResult[*webhookModel.Webhook]{
		Data:      data,
		Total:     pageResult.Total,
		Page:      pageResult.Page,
		PageSize:  pageResult.PageSize,
		TotalPage: pageResult.TotalPage,
	}, nil
}

func mapToModel(data Webhook) *webhookModel.Webhook {
	var triggers []constant.WebhookTrigger
	if data.Triggers != "" {
		for _, t := range strings.Split(data.Triggers, ",") {
			triggers = append(triggers, constant.WebhookTrigger(strings.TrimSpace(t)))
		}
	}

	var latestExecResult *constant.WebhookExecutionResult
	if data.LatestExecResult != nil {
		r := constant.WebhookExecutionResult(*data.LatestExecResult)
		latestExecResult = &r
	}

	parentID := int64(0)
	parentType := ""
	if data.RepoID != nil {
		parentID = *data.RepoID
		parentType = string(constant.WebhookParentRepo)
	} else if data.SpaceID != nil {
		parentID = *data.SpaceID
		parentType = string(constant.WebhookParentSpace)
	}

	return &webhookModel.Webhook{
		ID:                    int64(data.ID),
		ParentID:              parentID,
		ParentType:            parentType,
		CreatedBy:             data.CreatedBy,
		Created:               data.CreatedDate,
		UpdatedBy:             data.EditedBy,
		Updated:               data.EditedDate,
		Internal:              data.Internal,
		DisplayName:           data.DisplayName,
		Description:           data.Desc,
		URL:                   data.URL,
		Secret:                data.Secret,
		Enabled:               data.Enabled,
		Insecure:              !data.IsSSL,
		Triggers:              triggers,
		LatestExecutionResult: latestExecResult,
		AuthMethod:            constant.WebhookAuthMethod(data.AuthMethod),
		AuthHeader:            data.AuthHeader,
	}
}
