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
func (s *Store) ListByParentInfo(ctx context.Context, parents []webhookModel.ParentInfo) ([]*webhookModel.Webhook, error) {
	if len(parents) == 0 {
		return nil, nil
	}

	tx := s.db.WithContext(ctx).Model(&Webhook{}).Where("deleted = 0 AND enabled = 1")

	type clause struct {
		col string
		id  int64
	}
	var repoClauses []clause
	var spaceClauses []clause
	for _, p := range parents {
		switch p.Type {
		case string(constant.WebhookParentRepo):
			repoClauses = append(repoClauses, clause{col: "repoID", id: int64(p.ID)})
		case string(constant.WebhookParentSpace):
			spaceClauses = append(spaceClauses, clause{col: "spaceID", id: int64(p.ID)})
		}
	}

	var repoIDs []int64
	for _, c := range repoClauses {
		repoIDs = append(repoIDs, c.id)
	}
	var spaceIDs []int64
	for _, c := range spaceClauses {
		spaceIDs = append(spaceIDs, c.id)
	}

	switch {
	case len(repoIDs) > 0 && len(spaceIDs) > 0:
		tx = tx.Where("repoID IN ? OR spaceID IN ?", repoIDs, spaceIDs)
	case len(repoIDs) > 0:
		tx = tx.Where("repoID IN ?", repoIDs)
	case len(spaceIDs) > 0:
		tx = tx.Where("spaceID IN ?", spaceIDs)
	}

	var rows []*Webhook
	if err := tx.Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]*webhookModel.Webhook, len(rows))
	for i, row := range rows {
		result[i] = mapToModel(*row)
	}
	return result, nil
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

var allowedExecutionSortFields = map[string]string{
	"result":      "result",
	"createdDate": "createdDate",
}

func (s *Store) ListExecutionsByTriggerID(ctx context.Context, triggerID string) ([]*webhookModel.Execution, error) {
	var rows []*Execution
	if err := s.db.WithContext(ctx).Where("eventID = ?", triggerID).Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	result := make([]*webhookModel.Execution, len(rows))
	for i, row := range rows {
		result[i] = mapExecutionToModel(*row)
	}
	return result, nil
}

func (s *Store) GetExecutionByID(ctx context.Context, executionID uint) (*webhookModel.Execution, error) {
	var execution Execution
	err := s.db.WithContext(ctx).Where("id = ?", executionID).First(&execution).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapExecutionToModel(execution), nil
}

func (s *Store) CreateExecution(ctx context.Context, execution *webhookModel.Execution) error {
	entity := &Execution{
		WebhookID:      uint(execution.WebhookID),
		TriggerType:    execution.TriggerType,
		EventID:        execution.EventID,
		Result:         string(execution.Result),
		Duration:       execution.Duration,
		Error:          execution.Error,
		ReqURL:         execution.ReqURL,
		ReqHeaders:     execution.ReqHeaders,
		ReqBody:        execution.ReqBody,
		RespStatusCode: execution.RespStatusCode,
		RespStatus:     execution.RespStatus,
		RespHeaders:    execution.RespHeaders,
		RespBody:       execution.RespBody,
		RetriggerID:    uint(execution.RetriggerID),
		Retriggerable:  execution.ReTriggerable,
		CreatedDate:    execution.CreatedDate,
		EditedDate:     execution.EditedDate,
	}
	if err := s.db.WithContext(ctx).Create(entity).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	execution.ID = int64(entity.ID)
	return nil
}

func (s *Store) ListExecutions(ctx context.Context, webhookID uint, filter webhookModel.ReqForList) (*database.PageResult[*webhookModel.Execution], error) {
	tx := s.db.WithContext(ctx).Model(&Execution{}).Where("webhookID = ?", webhookID)

	sort := "createdDate"
	if col, ok := allowedExecutionSortFields[filter.Sort]; ok {
		sort = col
	}

	var rows []*Execution
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

	data := make([]*webhookModel.Execution, len(rows))
	for i, row := range rows {
		data[i] = mapExecutionToModel(*row)
	}

	return &database.PageResult[*webhookModel.Execution]{
		Data:      data,
		Total:     pageResult.Total,
		Page:      pageResult.Page,
		PageSize:  pageResult.PageSize,
		TotalPage: pageResult.TotalPage,
	}, nil
}

func mapExecutionToModel(data Execution) *webhookModel.Execution {
	return &webhookModel.Execution{
		ID:             int64(data.ID),
		WebhookID:      int64(data.WebhookID),
		TriggerType:    data.TriggerType,
		EventID:        data.EventID,
		Result:         constant.WebhookExecutionResult(data.Result),
		Duration:       data.Duration,
		Error:          data.Error,
		ReqURL:         data.ReqURL,
		ReqHeaders:     data.ReqHeaders,
		ReqBody:        data.ReqBody,
		RespStatusCode: data.RespStatusCode,
		RespStatus:     data.RespStatus,
		RespHeaders:    data.RespHeaders,
		RespBody:       data.RespBody,
		RetriggerID:    int64(data.RetriggerID),
		ReTriggerable:  data.Retriggerable,
		CreatedDate:    data.CreatedDate,
		EditedDate:     data.EditedDate,
	}
}

func mapToModel(data Webhook) *webhookModel.Webhook {
	var triggers []constant.WebhookTrigger
	if data.Triggers != "" {
		for _, t := range strings.Split(data.Triggers, ",") {
			triggers = append(triggers, constant.WebhookTrigger(strings.TrimSpace(t)))
		}
	}

	var latestExecResult *constant.WebhookExecutionResult
	if data.LatestExecResult != "" {
		r := constant.WebhookExecutionResult(data.LatestExecResult)
		latestExecResult = &r
	}

	parentID := int64(0)
	parentType := ""
	if data.RepoID != 0 {
		parentID = data.RepoID
		parentType = string(constant.WebhookParentRepo)
	} else if data.SpaceID != 0 {
		parentID = data.SpaceID
		parentType = string(constant.WebhookParentSpace)
	}

	return &webhookModel.Webhook{
		ID:                    int64(data.ID),
		ParentID:              parentID,
		ParentType:            parentType,
		CreatedBy:             data.CreatedBy,
		CreatedDate:           data.CreatedDate,
		EditedBy:              data.EditedBy,
		EditedDate:            data.EditedDate,
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
