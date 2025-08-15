package codec

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type Codec interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
}

type CfgFileType string

func (t CfgFileType) ToString() string {
	return string(t)
}

const (
	CfgFileTypeYaml CfgFileType = "yaml"
	CfgFileTypeJson CfgFileType = "json"
)

func NewCodec(tp CfgFileType) Codec {
	switch tp {
	case CfgFileTypeJson:
		return &JsonCodec{}
	default:
		return &YamlCodec{}
	}
}

type YamlCodec struct {
}

func (c *YamlCodec) Encode(v any) ([]byte, error) {
	return yaml.Marshal(v)
}
func (c *YamlCodec) Decode(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

type JsonCodec struct {
}

func (c *JsonCodec) Encode(v any) ([]byte, error) {
	return json.Marshal(v)
}
func (c *JsonCodec) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
