package provider

import (
	"context"
	"errors"
	"github.com/fufuzion/confremote-pilot/codec"
)

type Provider interface {
	Name() string
	Load() (map[string]interface{}, error)
}

func NewProvider(ctx context.Context, tp CfgProviderType, opts ...Option) (Provider, error) {
	o := &option{
		configType: codec.CfgFileTypeYaml,
	}
	for _, opt := range opts {
		opt(o)
	}
	switch tp {
	case CfgProviderNacos:
		return newNacosProvider(ctx, o)
	case CfgProviderConsul, CfgProviderEtcd, CfgProviderFirestore:
		return newViperBaseProvider(ctx, tp, o)
	default:
		return nil, errors.New("unknown provider type")
	}
}
