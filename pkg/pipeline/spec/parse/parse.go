// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package parse

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	schema "github.com/easysoft/gitfox-community/pkg/pipeline/spec"

	"github.com/ghodss/yaml"
)

// Parse parses the configuration from io.Reader r.
func Parse(r io.Reader) (*schema.Config, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	b, err = yaml.YAMLToJSON(b)
	if err != nil {
		return nil, err
	}
	out := new(schema.Config)
	err = json.Unmarshal(b, out)
	return out, err
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*schema.Config, error) {
	return Parse(
		bytes.NewBuffer(b),
	)
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*schema.Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}

// ParseFile parses the configuration from path p.
func ParseFile(p string) (*schema.Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

// ParseMulti parses a multi-document configuration
// from io.Reader r.
func ParseMulti(r io.Reader) ([]*schema.Config, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseMultiBytes(b)
}

// ParseMultiBytes parses a multi-document configuration
// from bytes b.
func ParseMultiBytes(b []byte) ([]*schema.Config, error) {
	var out []*schema.Config
	parts := bytes.Split(b, []byte("\n---\n"))
	for _, part := range parts {
		resource, err := ParseBytes(part)
		if err != nil {
			return nil, err
		}
		out = append(out, resource)
	}
	return out, nil
}

// ParseMultiFile parses a multi-document configuration
// from path p.
func ParseMultiFile(p string) ([]*schema.Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseMulti(f)
}

// ToConfig converts any config-like value into a *schema.Config by
// round-tripping through JSON.
func ToConfig(src any) (*schema.Config, error) {
	if src == nil {
		return nil, errors.New("source config is nil")
	}
	// Marshal src to JSON
	data, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	// Unmarshal JSON to schema.Config
	dst := new(schema.Config)
	if err := json.Unmarshal(data, dst); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to schema.Config: %w", err)
	}
	return dst, nil
}
