// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"errors"
	"fmt"
	"time"
)

const (
	// streamPayloadKey is the key used for storing the event in a stream message.
	streamPayloadKey = "event"
)

type Event[T interface{}] struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Payload   T         `json:"payload"`
}

// EventType describes the type of event.
type EventType string

// getStreamID generates the streamID for a given category and type of event.
func getStreamID(category string, event EventType) string {
	return fmt.Sprintf("events:%s:%s", category, event)
}

// Mode defines the different modes of the event framework.
type Mode string

const (
	ModeRedis    Mode = "redis"
	ModeInMemory Mode = "inmemory"
)

// Config defines the config of the events system.
type Config struct {
	Mode                  Mode
	Namespace             string
	MaxStreamLength       int64
	ApproxMaxStreamLength bool
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("config is required")
	}
	if c.Mode != ModeRedis && c.Mode != ModeInMemory {
		return fmt.Errorf("config.Mode '%s' is not supported", c.Mode)
	}
	if c.MaxStreamLength < 1 {
		return errors.New("config.MaxStreamLength has to be a positive number")
	}

	return nil
}
