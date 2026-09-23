// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"encoding/json"
	"errors"
)

// Registry provides container registry details.
type Registry struct {
	Connector []*RegistryConnector `json:"connector,omitempty"`
	Mirror    Stringorslice        `json:"mirror,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *Registry) UnmarshalJSON(data []byte) error {
	var out1 = struct {
		Connector json.RawMessage `json:"connector,omitempty"`
	}{}
	// unmarshal into a temporary struct
	if err := json.Unmarshal(data, &out1); err != nil {
		return err
	}

	var out2 string
	var out3 *RegistryConnector
	var out4 []*RegistryConnector

	if err := json.Unmarshal(out1.Connector, &out2); err == nil {
		v.Connector = append(v.Connector, &RegistryConnector{Name: out2})
		return nil
	}

	if err := json.Unmarshal(out1.Connector, &out3); err == nil {
		v.Connector = append(v.Connector, out3)
		return nil
	}

	if err := json.Unmarshal(out1.Connector, &out4); err == nil {
		v.Connector = append(v.Connector, out4...)
		return nil
	}

	return errors.New("yaml: cannot unmarshal registry connector")
}

// RegistryConnector provides registry credentials.
type RegistryConnector struct {
	Name  string `json:"name,omitempty"`
	Match string `json:"match,omitempty"`
}

// UnmarshalJSON implement the json.Unmarshaler interface.
func (v *RegistryConnector) UnmarshalJSON(data []byte) error {
	var out1 string
	var out2 = struct {
		Name  string `json:"name,omitempty"`
		Match string `json:"match,omitempty"`
	}{}

	if err := json.Unmarshal(data, &out1); err == nil {
		v.Name = out1
		return nil
	}

	if err := json.Unmarshal(data, &out2); err == nil {
		v.Name = out2.Name
		v.Match = out2.Match
		return nil
	}

	return errors.New("yaml: cannot unmarshal registry connector")
}
