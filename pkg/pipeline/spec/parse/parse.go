package parse

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	schema "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pipeline/spec"

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
