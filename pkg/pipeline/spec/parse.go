// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package yaml

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/ghodss/yaml"
)

// Parse parses the configuration from io.Reader r.
func Parse(r io.Reader) (*Config, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = yaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}
	out := new(Config)
	err = json.Unmarshal(b, out)
	return out, err
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*Config, error) {
	return Parse(
		bytes.NewBuffer(b),
	)
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}

// ParseFile parses the configuration from path p.
func ParseFile(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}
