// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package runner

import (
	"context"
	"sync"
	"time"

	"github.com/easysoft/gitfox-community/pkg/errors"
)

// TokenCache stores temporary tokens for runner registration
type TokenCache struct {
	mx     sync.RWMutex
	tokens map[string]tokenEntry
	maxAge time.Duration
	stopCh chan struct{}
}

type tokenEntry struct {
	account string
	addedAt time.Time
}

// NewTokenCache creates a new token cache with TTL
func NewTokenCache(maxAge time.Duration) *TokenCache {
	c := &TokenCache{
		tokens: make(map[string]tokenEntry),
		maxAge: 60 * 7 * 24 * time.Minute,
		stopCh: make(chan struct{}),
	}
	go c.purger()
	return c
}

// purger periodically evicts stale tokens from the cache
func (c *TokenCache) purger() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			return
		case now := <-ticker.C:
			c.mx.Lock()
			for token, entry := range c.tokens {
				if now.Sub(entry.addedAt) >= c.maxAge {
					delete(c.tokens, token)
				}
			}
			c.mx.Unlock()
		}
	}
}

// Stop stops the internal purger
func (c *TokenCache) Stop() {
	close(c.stopCh)
}

// Set stores a token with associated account
func (c *TokenCache) Set(token, account string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.tokens[token] = tokenEntry{
		account: account,
		addedAt: time.Now(),
	}
}

// Get retrieves the account associated with a token
func (c *TokenCache) Get(token string) (string, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	entry, ok := c.tokens[token]
	if !ok {
		return "", false
	}

	if time.Since(entry.addedAt) >= c.maxAge {
		return "", false
	}

	return entry.account, true
}

// Delete removes a token from the cache
func (c *TokenCache) Delete(token string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	delete(c.tokens, token)
}

// TempStoreToken stores a temporary token for runner registration
func (s *Store) TempStoreToken(ctx context.Context, token, account string) error {
	if s.tokenCache == nil {
		return errors.SystemErr.New("token cache not initialized")
	}
	s.tokenCache.Set(token, account)
	return nil
}

// FindToken retrieves the account associated with a token
func (s *Store) FindToken(ctx context.Context, token string) (string, error) {
	if s.tokenCache == nil {
		return "", errors.SystemErr.New("token cache not initialized")
	}
	account, ok := s.tokenCache.Get(token)
	if !ok {
		return "", errors.NotFound.New("token not found or expired")
	}
	return account, nil
}
