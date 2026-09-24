// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
)

type Container struct {
	Image       string         `json:"image,omitempty"`
	Connector   string         `json:"connector,omitempty"`
	Credentials *Credentials   `json:"credentials,omitempty"`
	Pull        string         `json:"pull,omitempty"`
	Entrypoint  string         `json:"entrypoint,omitempty"`
	Args        []string       `json:"args,omitempty"`
	Dns         Stringorslice  `json:"dns,omitempty"`
	DsnSearch   Stringorslice  `json:"dsn_search,omitempty"`
	ExtraHosts  Stringorslice  `json:"extra_hosts,omitempty"`
	Network     string         `json:"network,omitempty"`
	NetworkMode string         `json:"network_mode,omitempty"`
	Privileged  bool           `json:"privileged,omitempty"`
	Workdir     string         `json:"workdir,omitempty"`
	Ports       []string       `json:"ports,omitempty"`
	Volumes     []string       `json:"volumes,omitempty"`
	User        string         `json:"user,omitempty"`
	Group       string         `json:"group,omitempty"`
	Cpu         StringorInt    `json:"cpu,omitempty"`
	Memory      MemStringorInt `json:"memory,omitempty"`
	ShmSize     MemStringorInt `json:"shm_size,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Container) UnmarshalJSON(data []byte) error {
	type S Container
	type T struct {
		*S
	}

	var out string
	if err := json.Unmarshal(data, &out); err == nil {
		v.Image = out
		return nil
	}

	obj := &T{S: (*S)(v)}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}

	return nil
}
