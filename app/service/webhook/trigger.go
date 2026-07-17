package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

const (
	webhookTimeLimit          = 10 * time.Second
	responseHeadersBytesLimit = 1024
	responseBodyBytesLimit    = 1024
)

var ErrWebhookNotRetriggerable = errors.WithI18nKey(nil, i18n.WebhookExecutionRetriggerError)

func (s *Service) triggerWebhooksFor(
	ctx context.Context,
	parents []webhookModel.ParentInfo,
	triggerID string,
	triggerType constant.WebhookTrigger,
	filter *webhookModel.ReqForList,
	body any,
) ([]webhookModel.TriggerResult, error) {
	webhooks, err := s.webhookStore.ListByParentInfo(ctx, parents)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks for: %w", err)
	}

	return s.triggerWebhooks(ctx, webhooks, triggerID, triggerType, body)
}

func (s *Service) triggerWebhooks(
	ctx context.Context,
	webhooks []*webhookModel.Webhook,
	triggerID string,
	triggerType constant.WebhookTrigger,
	body any,
) ([]webhookModel.TriggerResult, error) {

	// return immediately if webhooks are empty
	if len(webhooks) == 0 {
		return []webhookModel.TriggerResult{}, nil
	}

	// get all previous execution for the same trigger
	executions, err := s.webhookStore.ListExecutionsByTriggerID(ctx, triggerID)
	if err != nil && !errors.Is(err, store.ErrResourceNotFound) {
		return nil, fmt.Errorf("failed to get executions for trigger '%s'", triggerID)
	}

	// precalculate whether a webhook should be executed
	skipExecution := make(map[int64]bool)
	for _, execution := range executions {
		// skip execution in case of success or unrecoverable error
		if execution.Result == constant.WebhookExecutionResultSuccess ||
			execution.Result == constant.WebhookExecutionResultFatalError {
			skipExecution[execution.WebhookID] = true
		}
	}

	results := make([]webhookModel.TriggerResult, 0, len(webhooks))
	for _, webhook := range webhooks {
		result := webhookModel.TriggerResult{
			TriggerID:   triggerID,
			TriggerType: string(triggerType),
			Webhook:     webhook,
		}

		// check if webhook already got executed (success or fatal error)
		if skipExecution[webhook.ID] {
			continue
		}

		if !webhookHasTrigger(webhook, triggerType) {
			results = append(results, result)
			continue
		}

		execution, err := s.executeWebhook(ctx, webhook, triggerID, string(triggerType), body, nil)
		result.Execution = execution
		result.Err = err
		results = append(results, result)
	}
	return results, nil
}

func webhookHasTrigger(webhook *webhookModel.Webhook, triggerType constant.WebhookTrigger) bool {
	if len(webhook.Triggers) == 0 {
		return true
	}
	for _, t := range webhook.Triggers {
		if t == triggerType {
			return true
		}
	}
	return false
}

func (s *Service) RetriggerExecution(ctx context.Context, webhookExecutionID uint) (*webhookModel.TriggerResult, error) {
	webhookExecution, err := s.webhookStore.GetExecutionByID(ctx, webhookExecutionID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.WebhookExecutionNotFound)
		return nil, err
	}

	if !webhookExecution.ReTriggerable {
		return nil, ErrWebhookNotRetriggerable
	}

	webhook, err := s.webhookStore.GetByID(ctx, uint(webhookExecution.WebhookID))
	if err != nil {
		err = errors.WithI18nKey(err, i18n.WebhookNotFound)
		return nil, err
	}

	body := bytes.NewBufferString(webhookExecution.ReqBody)

	newExecution, err := s.executeWebhook(ctx, webhook, webhookExecution.EventID, webhookExecution.TriggerType, body, &webhookExecution.ID)
	return &webhookModel.TriggerResult{
		TriggerID:   webhookExecution.EventID,
		TriggerType: webhookExecution.TriggerType,
		Webhook:     webhook,
		Execution:   newExecution,
		Err:         err,
	}, nil
}

func (s *Service) executeWebhook(ctx context.Context, webhook *webhookModel.Webhook, eventID, triggerType string, body any, rerunOfID *int64) (*webhookModel.Execution, error) {
	log := logger.GetLogger(ctx)
	execution := &webhookModel.Execution{
		WebhookID:   webhook.ID,
		EventID:     eventID,
		TriggerType: triggerType,
		Result:      constant.WebhookExecutionResultFatalError,
		Error:       "An unknown error occurred",
	}
	if rerunOfID != nil {
		execution.RetriggerID = *rerunOfID
	}
	start := time.Now()
	defer func(oCtx context.Context) {
		execution.Duration = time.Since(start).Milliseconds()
		execution.CreatedDate = datetime.NowLocalTime()

		if err := s.webhookStore.CreateExecution(oCtx, execution); err != nil {
			log.Errorf(
				"failed to store webhook execution that ended with Result: %s, RespStatus: '%s', Error: '%s', error: %v",
				execution.Result, execution.RespStatus, execution.Error, err)
		}

		if webhook.LatestExecutionResult == nil || *webhook.LatestExecutionResult != execution.Result {
			result := execution.Result
			_, err := s.webhookStore.Update(oCtx, uint(webhook.ID), map[string]interface{}{
				"latestExecResult": string(result),
			})
			if err != nil {
				log.Errorf(
					"failed to update latest execution result to %s for webhook %d",
					execution.Result, webhook.ID)
			}
		}
	}(ctx)

	ctx, cancel := context.WithTimeout(ctx, webhookTimeLimit)
	defer cancel()

	httpReq, err := s.prepareHTTPRequest(ctx, execution, triggerType, webhook, body)
	if err != nil {
		return execution, err
	}

	var resp *http.Response
	switch {
	case webhook.Internal && webhook.Insecure:
		resp, err = s.insecureHTTPClientInternal.Do(httpReq)
	case webhook.Internal:
		resp, err = s.secureHTTPClientInternal.Do(httpReq)
	case webhook.Insecure:
		resp, err = s.insecureHTTPClient.Do(httpReq)
	default:
		resp, err = s.secureHTTPClient.Do(httpReq)
	}

	if resp != nil && resp.Body != nil {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.Errorf("failed to close body after webhook execution %d: %v", execution.ID, closeErr)
			}
		}()
	}

	var dnsError *net.DNSError
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		tErr := fmt.Errorf("request exceeded time limit of %s", webhookTimeLimit)
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultFatalError
		return execution, tErr

	case stderrors.As(err, &dnsError) && dnsError.IsNotFound:
		execution.Error = fmt.Sprintf("host '%s' was not found", dnsError.Name)
		execution.Result = constant.WebhookExecutionResultFatalError
		return execution, fmt.Errorf("failed to resolve host name '%s': %w", dnsError.Name, err)

	case err != nil:
		tErr := fmt.Errorf("an error occurred while sending the request: %w", err)
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultFatalError
		return execution, tErr
	}

	err = handleWebhookResponse(execution, resp)
	return execution, err
}

func (s *Service) prepareHTTPRequest(ctx context.Context, execution *webhookModel.Execution,
	triggerType string, webhook *webhookModel.Webhook, body any) (*http.Request, error) {
	execution.ReqURL = webhook.URL

	bBuff := &bytes.Buffer{}
	switch v := body.(type) {
	case io.Reader:
		bBytes, err := io.ReadAll(v)
		if err != nil {
			tErr := fmt.Errorf("failed to generate request body: %w", err)
			execution.Error = tErr.Error()
			execution.Result = constant.WebhookExecutionResultFatalError
			return nil, tErr
		}
		bBuff.Write(bBytes)
	default:
		if err := json.NewEncoder(bBuff).Encode(v); err != nil {
			execution.Error = "an error occurred preparing the request body"
			execution.Result = constant.WebhookExecutionResultFatalError
			return nil, fmt.Errorf("failed to serialize body to json: %w", err)
		}
	}

	execution.ReqBody = bBuff.String()
	execution.ReTriggerable = true

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bBuff)
	if err != nil {
		tErr := fmt.Errorf("failed to create request: %w", err)
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultFatalError
		return nil, tErr
	}

	httpReq.Header.Add("Content-Type", "application/json")
	httpReq.Header.Add("X-Trigger", triggerType)
	httpReq.Header.Add("X-Webhook-Parent-Type", webhook.ParentType)
	httpReq.Header.Add("X-Webhook-Parent-Id", fmt.Sprint(webhook.ParentID))

	switch webhook.AuthMethod {
	case constant.WebhookAuthMethodSignature:
		if webhook.Secret != "" {
			sig, err := generateHMACSHA256(bBuff.Bytes(), []byte(webhook.Secret))
			if err != nil {
				return nil, fmt.Errorf("failed to generate SHA256 based HMAC: %w", err)
			}
			httpReq.Header.Add("X-Signature", sig)
		}
	case constant.WebhookAuthMethodToken:
		if webhook.Secret != "" && webhook.AuthHeader != "" {
			httpReq.Header.Add(webhook.AuthHeader, webhook.Secret)
		}
	}

	hBuff := &bytes.Buffer{}
	if err = httpReq.Header.Write(hBuff); err != nil {
		tErr := fmt.Errorf("failed to write request headers: %w", err)
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultRetriableError
		return nil, tErr
	}
	rawHeaders := hBuff.String()
	if webhook.Secret != "" && len(webhook.Secret) > 4 {
		rawHeaders = strings.ReplaceAll(rawHeaders, webhook.Secret,
			fmt.Sprintf("%s*****%s", webhook.Secret[:2], webhook.Secret[len(webhook.Secret)-2:]))
	}
	execution.ReqHeaders = rawHeaders

	return httpReq, nil
}

func handleWebhookResponse(execution *webhookModel.Execution, resp *http.Response) error {
	execution.RespStatusCode = resp.StatusCode
	execution.RespStatus = resp.Status

	hBuff := &bytes.Buffer{}
	if err := resp.Header.Write(hBuff); err != nil {
		tErr := fmt.Errorf("failed to read response headers: %w", err)
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultRetriableError
		return tErr
	}
	headerLen := hBuff.Len()
	if headerLen > responseHeadersBytesLimit {
		headerLen = responseHeadersBytesLimit
	}
	execution.RespHeaders = string(hBuff.Bytes()[:headerLen])

	if resp.Body != nil {
		bodyRaw, err := io.ReadAll(io.LimitReader(resp.Body, responseBodyBytesLimit))
		if err != nil {
			tErr := fmt.Errorf("an error occurred while reading the response body: %w", err)
			execution.Error = tErr.Error()
			execution.Result = constant.WebhookExecutionResultRetriableError
			return tErr
		}
		execution.RespBody = string(bodyRaw)
	}

	switch code := resp.StatusCode; {
	case code < 200:
		execution.Error = "1xx response codes are not supported"
		execution.Result = constant.WebhookExecutionResultFatalError
		return fmt.Errorf("received response with unsupported status code %d", code)
	case code < 300:
		execution.Error = ""
		execution.Result = constant.WebhookExecutionResultSuccess
		return nil
	case code < 400:
		execution.Error = "3xx response codes are not supported"
		execution.Result = constant.WebhookExecutionResultFatalError
		return fmt.Errorf("received response with unsupported status code %d", code)
	case code == 408:
		tErr := errors.New("request timed out")
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultRetriableError
		return tErr
	case code == 429:
		tErr := errors.New("request got throttled")
		execution.Error = tErr.Error()
		execution.Result = constant.WebhookExecutionResultRetriableError
		return tErr
	case code < 500:
		execution.Error = "4xx response codes are not supported (apart from 408 and 429)"
		execution.Result = constant.WebhookExecutionResultFatalError
		return fmt.Errorf("received response with unrecoverable status code %d", code)
	case code == 501:
		execution.Error = "remote server does not implement requested action"
		execution.Result = constant.WebhookExecutionResultFatalError
		return fmt.Errorf("received response with unrecoverable status code %d", code)
	case code < 600:
		execution.Error = "remote server encountered an error"
		execution.Result = constant.WebhookExecutionResultRetriableError
		return fmt.Errorf("remote server encountered an error: %d", code)
	default:
		execution.Error = "response code not supported"
		execution.Result = constant.WebhookExecutionResultFatalError
		return fmt.Errorf("received response with unsupported status code %d", code)
	}
}

func generateHMACSHA256(data []byte, key []byte) (string, error) {
	h := hmac.New(sha256.New, key)
	if _, err := h.Write(data); err != nil {
		return "", fmt.Errorf("failed to write data into hash: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
