// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package runner

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	runnerModel "github.com/easysoft/gitfox-community/app/model/runner"
	"github.com/easysoft/gitfox-community/app/store/runner"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/errors"
)

type Service struct {
	runnerStore runner.Interface
}

var _ Interface = (*Service)(nil)

// NewService creates a new runner service instance
func NewService(runnerStore runner.Interface) *Service {
	return &Service{
		runnerStore: runnerStore,
	}
}

// Create creates a new runner
func (s *Service) Create(ctx context.Context, runner *runnerModel.Runner, runnerToken, runnerIP string) error {

	// Set default data
	now := int(time.Now().Unix())
	runner.Token = runnerToken
	runner.Name = fmt.Sprintf("%s-%s", runner.Name, runnerIP)
	runner.IP = runnerIP
	runner.Online = "online"
	runner.Status = "enable"
	runner.HeartBeat = now

	return s.runnerStore.Create(ctx, runner)
}

// Update 更新 runner
func (s *Service) Update(ctx context.Context, runner *runnerModel.Runner) error {
	// Check if runner exists
	existing, err := s.runnerStore.GetByID(ctx, runner.ID)
	if err != nil {
		return errors.NotFound.Newf("runner not found: %d", runner.ID)
	}

	// 内置 runner 不允许通过 API 修改
	if existing.IsDefault {
		return errors.BadRequest.Newf("cannot modify default runner")
	}

	return s.runnerStore.Update(ctx, runner)
}

// UpdateRunnerHeartbeat 更新 runner 心跳
func (s *Service) UpdateRunnerHeartbeat(ctx context.Context, runnerID uint) error {
	// Check if runner exists
	existing, err := s.runnerStore.GetByID(ctx, runnerID)
	if err != nil {
		return errors.NotFound.Newf("runner not found: %d", runnerID)
	}

	// 内置 runner 不需要外部心跳更新
	if existing.IsDefault {
		return errors.BadRequest.Newf("cannot update heartbeat for default runner")
	}

	now := int(time.Now().Unix())
	runnerInfo := &runnerModel.Runner{
		ID:        runnerID,
		HeartBeat: now,
		Online:    "online", // 心跳更新时设置为在线
	}
	return s.runnerStore.Update(ctx, runnerInfo)
}

// GetByID gets runner by ID
func (s *Service) GetByID(ctx context.Context, id uint) (*runnerModel.Runner, error) {
	return s.runnerStore.GetByID(ctx, id)
}

// GetByToken gets runner by token
func (s *Service) GetByToken(ctx context.Context, token string) (*runnerModel.Runner, error) {
	if token == "" {
		return nil, errors.BadRequest.New("A valid runner token must be provided.")
	}

	runner, err := s.runnerStore.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find runner by token")
	}

	return runner, nil
}

// List returns a list of runners with optional filter
func (s *Service) List(ctx context.Context, filter *runnerModel.ReqFilter) (*database.PageResult[*runnerModel.Runner], error) {
	return s.runnerStore.List(ctx, filter)
}

// Delete soft deletes a runner by ID
func (s *Service) Delete(ctx context.Context, id uint) error {
	// Check if runner exists
	existing, err := s.runnerStore.GetByID(ctx, id)
	if err != nil {
		return errors.NotFound.Newf("runner not found: %d", id)
	}

	// 内置 runner 不允许删除
	if existing.IsDefault {
		return errors.BadRequest.Newf("cannot delete default runner")
	}

	return s.runnerStore.Delete(ctx, id)
}

// GenerateRunnerToken 为 runner 生成临时访问 token
func (s *Service) GenerateRunnerToken(ctx context.Context, account string) (string, error) {
	if account == "" {
		return "", errors.BadRequest.New("account is required")
	}

	// 生成安全 token
	token, err := s.generateSecureToken(account)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate token")
	}

	// 存储到缓存
	if err := s.runnerStore.TempStoreToken(ctx, token, account); err != nil {
		return "", errors.Wrapf(err, "failed to store token")
	}

	return token, nil
}

// ValidateRunnerToken 验证数据库中的 runner token 并返回对应的 runner
func (s *Service) ValidateRunnerToken(ctx context.Context, token string) (*runnerModel.Runner, error) {
	if token == "" {
		return nil, errors.BadRequest.New("token is required")
	}

	runner, err := s.runnerStore.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.NotFound.New("invalid runner token")
	}

	return runner, nil
}

// ValidateTempRunnerToken 验证缓存中的 runner token 并返回对应的 account 信息
func (s *Service) ValidateTempRunnerToken(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", errors.BadRequest.New("token is required")
	}

	account, err := s.runnerStore.FindToken(ctx, token)
	if err != nil {
		return "", errors.NotFound.New("token not found or expired")
	}

	return account, nil
}

// generateSecureToken 生成一个安全的随机 token，格式：uid_timestamp_random，使用 MD5 编码
// 同时生成可解析的格式来提取 UID
func (s *Service) generateSecureToken(account string) (string, error) {
	// 生成 16 字节的随机数据
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixMilli()

	// 构造 token 原始字符串：uid_timestamp_random
	randomStr := base64.RawURLEncoding.EncodeToString(randomBytes)
	tokenRaw := fmt.Sprintf("%s_%d_%s", account, timestamp, randomStr)

	// 创建复合 token：base64(uid) + "." + md5(tokenRaw)
	// 这样既保持 MD5 的安全性，又能解析出 UID
	uidBase64 := base64.RawURLEncoding.EncodeToString([]byte(account))
	hash := md5.Sum([]byte(tokenRaw))
	hashStr := hex.EncodeToString(hash[:])

	token := fmt.Sprintf("%s.%s", uidBase64, hashStr)

	return token, nil
}

// parseTokenToUID 从 token 中解析出 UID
func (s *Service) parseTokenToUID(token string) (string, error) {
	// 分割 token：base64(uid).md5(hash)
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token format")
	}

	// 解码 UID 部分
	// 使用 RawURLEncoding 与生成函数保持一致
	uidBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid UID encoding: %w", err)
	}

	uid := string(uidBytes)
	if uid == "" {
		return "", fmt.Errorf("empty UID in token")
	}

	return uid, nil
}
