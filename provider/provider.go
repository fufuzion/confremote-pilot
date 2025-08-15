package provider

import (
	"errors"
	"github.com/fufuzion/confremote-pilot/codec"
)

type Provider interface {
	Name() string
	Load() (map[string]interface{}, error)
}

func NewProvider(tp CfgProviderType, opts ...Option) (Provider, error) {
	o := &option{
		configType: codec.CfgFileTypeYaml,
	}
	for _, opt := range opts {
		opt(o)
	}
	switch tp {
	case CfgProviderNacos:
		return newNacosProvider(o)
	default:
		return nil, errors.New("unknown provider type")
	}
}
